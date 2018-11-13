package correlation

import (
	"context"
	"github.com/edgexfoundry/edgex-go/internal"
)

func FromContext(ctx context.Context) string {
	hdr, ok := ctx.Value(internal.CorrelationHeader).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}
