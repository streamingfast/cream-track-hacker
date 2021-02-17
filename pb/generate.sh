ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"

# Protobuf definitions (required to be a sibling of this repository, no check perform yet, needs to be done manually)
PROTO_ETHEREUM=${1:-"$ROOT/../proto-ethereum"}

function main() {
  set -e
  pushd "$ROOT/pb" &> /dev/null

  # **Imporant** Requires proto-gen-go >= 1.20 && protoc-gen-go-grpc >= 1.1.0 (So the second majour revision of Go protocol buffer, a.k.a APIv2)
  generate "dfuse/ethereum/codec/v1/codec.proto"

  echo "generate.sh - `date` - `whoami`" > $ROOT/pb/last_generate.txt
  echo "dfuse-io/proto-ethereum revision: `GIT_DIR=$PROTO_ETHEREUM/.git git rev-parse HEAD`" >> $ROOT/pb/last_generate.txt
}

# usage:
# - generate <protoPath>
# - generate <protoBasePath/> [<file.proto> ...]
function generate() {
    base=""
    if [[ "$#" -gt 1 ]]; then
      base="$1"; shift
    fi

    for file in "$@"; do
      protoc -I$PROTO_ETHEREUM \
        --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
         $base$file
    done
}

main "$@"
