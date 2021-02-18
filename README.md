# CREAM Tracking Hacker Transactions

## Getting started

1. Get an API key from https://streamingfast.io
1. Download a release from [the releases](https://github.com/streamingfast/cream-track-hacker/releases)
1. Start the tracker

## Install

Build from source:

    GO111MODULE=on go get github.com/streamingfast/cream-track-hacker/cmd/tracker

Or download a self-contained binary for [Windows, macOS or Linux](https://github.com/streamingfast/cream-track-hacker/releases).

## Usage

```bash
export STREAMINGFAST_API_KEY="server_......................"
```

Then simply launch the tracker:

```bash
tracker
```

The binart starts a never ending stream that receives all Ethereum blocks, analyze their transactions
and notify when a transaction is doing a pure Ether Transfer (from/to) or a ERC20 Transfer
(from/to) coming from one of the tracked address (defaults to `0x560a8e3b79d23b0a525e15c6f3486c6a293ddad2`
and `0x905315602ed9a854e325f692ff82f58799beab57`).

The notification is sent to the standard output, can be enhanced to send email, send a Slack
or WeChat message, etc.

The file `cursor.txt` saved in the current directory must be kept and persisted, it's the marker
that tells the stream where to start back at. If the file is non existant, the stream starts by default
from block #11 850 00. If it starts from that location, it will take sometime for the stream to catch
up with live block since it will need to inspect all blocks between the default start block and
current head block. Try to save and backup the `cursor.txt` to avoid long delays before being live
and tracking actual real-time transactions.

### Flags

Use

```
go run ./cmd/tracker --help
```

To list available flags to tweak default start block, tracked addresses, status frequency and cursor
file location.

## License

[Apache 2.0](./LICENSE)
