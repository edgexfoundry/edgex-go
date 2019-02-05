package correlation

import (
	"context"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
)

func FromContext(ctx context.Context) string {
	hdr, ok := ctx.Value(clients.CorrelationHeader).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}
