//go:generate protoc -I ./ ./ziti_metrics.proto --go_out=paths=source_relative:./

package metrics_pb

// Here to provide the go:generate line above
