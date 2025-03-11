# Prerequisites

1. Install the protoc binary from: https://github.com/protocolbuffers/protobuf/releases
2. Install the protoc plugin for Go ```go install google.golang.org/protobuf/cmd/protoc-gen-go@latest```
3. Ensure ```protoc``` is on your path.
4. Ensure your Go bin directory is on your path


# Generate Go Code

Two options, run the command manually or use `go generate`

## Go Generate

1. Navigate to the root project directory `sdk-golang`
2. run `go generate ./pb/edge_client_pb/...` or  `go generate /pb/edge_client_pb/...`

Note: Running a naked `go generate` will trigger all `go:generate` tags in the project, which you most likely do not want

## Manually

1. Navigate to the project root
2. Run: ```protoc -I ./pb/ ./pb/edge_client_pb/edge_client.proto --go_out=./pb/edge_client_pb```
