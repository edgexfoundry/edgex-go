package correlation

import (
	"context"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

func FromContext(ctx context.Context) string {
	hdr, ok := ctx.Value(common.CorrelationHeader).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}
