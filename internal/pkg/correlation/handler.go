package correlation

import (
	"context"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

func FromContext(ctx context.Context) string {
	hdr, ok := ctx.Value(clients.CorrelationHeader).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}
