package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	dfuse "github.com/dfuse-io/client-go"
	"github.com/dfuse-io/dgrpc"
	"github.com/dfuse-io/logging"
	pbbstream "github.com/dfuse-io/pbgo/dfuse/bstream/v1"
	"github.com/golang/protobuf/ptypes"
	pbcodec "github.com/streamingfast/cream-track-hacker/pb/dfuse/ethereum/codec/v1"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

var flagAddresses = flag.String("a", "0x560a8e3b79d23b0a525e15c6f3486c6a293ddad2, 0x905315602ed9a854e325f692ff82f58799beab57", "Specify the list of addresses to track, comma separated")
var flagCursorFile = flag.String("c", "cursor.txt", "The file containing the last cursor ever seen, when present, the cursor in it will be used to reconnect to last seen block")
var flagEndpoint = flag.String("e", "api.streamingfast.io:443", "The endpoint to connect the stream of blocks to")
var flagStatusFrequency = flag.Duration("f", 30*time.Second, "How often current state is logged to logger so it's possible to stream how the stream is behaving")
var flagStartBlock = flag.Int64("s", 11878000, "The block num to start from when no previous cursor exists")
var flagSkipSSLVerify = flag.Bool("i", false, "When set to true, skips SSL certificate verification")

var zlog = logging.NewSimpleLogger("tracker", "github.com/streaminfast/cream-track-hacker")

func usage() string {
	return `usage: tracker

Starts a never ending stream that receives all Ethereum blocks, analyze their transactions
and notify when a transaction is doing a pure Ether Transfer (from/to) or a ERC20 Transfer
(from/to).

The notification is sent to the standard output, can be enhanced to send email, send a Slack
or WeChat message, etc.

Flags:
` + flagUsage()
}

func main() {
	setupFlag()

	apiKey := os.Getenv("STREAMINGFAST_API_KEY")
	ensure(apiKey != "", errorUsage("the environment variable STREAMINGFAST_API_KEY must be set to a valid dfuse API key value"))
	addresses, celAddresses, err := parseAddressesFlag(*flagAddresses)

	dfuse, err := dfuse.NewClient(*flagEndpoint, apiKey)
	noError(err, "unable to create client")

	var dialOptions []grpc.DialOption
	if *flagSkipSSLVerify {
		dialOptions = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))}
	}

	conn, err := dgrpc.NewExternalClient(*flagEndpoint, dialOptions...)
	noError(err, "unable to create external gRPC client")

	cursor, err := loadCursor(*flagCursorFile)
	noError(err, "unable to load latest cursor")

	filter := fmt.Sprintf("to in %[1]s || from in %[1]s || erc20_to in %[1]s || erc20_from in %[1]s", celAddresses)
	zlog.Debug("using cel filter", zap.String("filter", filter))

	state := newStreamState()
	startingAt := strconv.FormatInt(*flagStartBlock, 10)
	if cursor != "" {
		startingAt = fmt.Sprintf("cursor (%s)", cursor)
	}

	streamClient := pbbstream.NewBlockStreamV2Client(conn)
	nextStatus := time.Now().Add(*flagStatusFrequency)

	zlog.Info("Starting stream (never ending)", zap.Strings("addresses", addresses), zap.String("starting_at", startingAt))
	for {
		tokenInfo, err := dfuse.GetAPITokenInfo(context.Background())
		noError(err, "unable to retrieve StreamingFast API token")

		credentials := oauth.NewOauthAccess(&oauth2.Token{AccessToken: tokenInfo.Token, TokenType: "Bearer"})
		stream, err := streamClient.Blocks(context.Background(), &pbbstream.BlocksRequestV2{
			StartBlockNum:     *flagStartBlock,
			StartCursor:       cursor,
			ForkSteps:         []pbbstream.ForkStep{pbbstream.ForkStep_STEP_NEW},
			IncludeFilterExpr: filter,
			Details:           pbbstream.BlockDetails_BLOCK_DETAILS_LIGHT,
		}, grpc.PerRPCCredentials(credentials))
		noError(err, "unable to start blocks stream")

		for {
			block, newCursor, shouldBreak := readBlock(stream)
			if shouldBreak {
				break
			}

			// Those will be transaction that matches our filter (Ether from/to, ERC20 Transfer from/to)
			for _, trxTrace := range block.TransactionTraces {
				notifyTransactionSeen(block, trxTrace, addresses)
			}

			cursor = newCursor
			state.recordBlock(block.Number)
			err = writeCursor(*flagCursorFile, cursor)
			noError(err, "unable to write cursor to persistent storage")

			now := time.Now()
			if now.After(nextStatus) {
				zlog.Info(fmt.Sprintf("Stream state each %s", *flagStatusFrequency), state.LogFields()...)
				nextStatus = now.Add(*flagStatusFrequency)
			}
		}

		zlog.Info("Waiting 5s before retrying")
		time.Sleep(5 * time.Second)
		state.reconnectCount++
	}
}

func notifyTransactionSeen(block *pbcodec.Block, trxTrace *pbcodec.TransactionTrace, trackedAddresses []string) {
	trackedSet := addressSet(trackedAddresses)

	fmt.Printf("Matching transaction %[1]s in block #%d (Links https://ethq.app/tx/%[1]s ,https://etherscan.io/tx/%[1]s)\n", hash(trxTrace.Hash).Pretty(), block.Number)
	for i, call := range trxTrace.Calls {
		callFromTracked := address(call.Caller).Pretty()
		if trackedSet.contains(callFromTracked) {
			callFromTracked += " *tracked*"
		}

		callToTracked := address(call.Address).Pretty()
		if trackedSet.contains(callToTracked) {
			callToTracked += " *tracked*"
		}

		etherTransfer := ""
		value := call.Value.Native()
		if value.Sign() > 0 {
			etherTransfer = fmt.Sprintf(", transferred %s ETH", formatTokenAmount(value, 18, 4))
		}

		fmt.Printf(" Internal call #%d %s -> %s matched%s\n", i, callFromTracked, callToTracked, etherTransfer)
	}
}

func loadCursor(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", fmt.Errorf("unable to read file %q: %w", filePath, err)
	}

	return string(content), nil
}

func writeCursor(filePath string, cursor string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("unable to create directory %q (and its parents): %w", dir, err)
		}
	}

	return ioutil.WriteFile(filePath, []byte(cursor), os.ModePerm)
}

func readBlock(stream pbbstream.BlockStreamV2_BlocksClient) (block *pbcodec.Block, cursor string, shouldBreak bool) {
	response, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			zlog.Error("We received a termination signal, this is unexpected as we expected the stream to be never ending")
			return nil, "", true
		}

		zlog.Error("Stream encountered a remote error, going to retry", zap.Error(err))
		return nil, "", true
	}

	zlog.Debug("Decoding received message's block")
	block = &pbcodec.Block{}
	err = ptypes.UnmarshalAny(response.Block, block)
	noError(err, "should have been able to unmarshal received block payload")

	return block, response.Cursor, false
}

type streamState struct {
	startTime      time.Time
	headBlock      uint64
	blockCount     uint64
	reconnectCount uint64
}

func newStreamState() *streamState {
	return &streamState{
		startTime:      time.Now(),
		headBlock:      0,
		blockCount:     0,
		reconnectCount: 0,
	}
}

func (s *streamState) LogFields() []zap.Field {
	return []zap.Field{
		zap.Uint64("head_block", s.headBlock),
		zap.Uint64("block_count", s.blockCount),
		zap.Uint64("reconnect_count", s.reconnectCount),
	}
}

func (s *streamState) recordBlock(blockNum uint64) {
	s.headBlock = blockNum
	s.blockCount++
}
