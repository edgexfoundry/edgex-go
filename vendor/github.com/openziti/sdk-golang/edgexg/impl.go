package edgexg

import "github.com/openziti/sdk-golang/ziti/edge"

const (
	PayloadFlagsHeader uint8 = 0x10
)

// HeadersToFabric tracks the headers to pass through fabric to the other side
var HeadersToFabric = map[int32]uint8{
	edge.FlagsHeader: PayloadFlagsHeader,
}

var HeadersFromFabric = map[uint8]int32{
	PayloadFlagsHeader: edge.FlagsHeader,
}
