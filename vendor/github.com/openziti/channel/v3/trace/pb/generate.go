//go:generate protoc -I ./ ./trace.proto --go_out=paths=source_relative:./

package trace_pb

// Here to provide the go:generate line above
