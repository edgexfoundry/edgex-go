package correlation

import (
	"context"

	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
)

func FromContext(ctx context.Context) string {
	hdr, ok := ctx.Value(common.CorrelationHeader).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}

// FromContextOrNew returns the correlation ID from the context if it exists, otherwise it generates a new one and put it back to the context.
func FromContextOrNew(ctx context.Context) (context.Context, string) {
	hdr := FromContext(ctx)
	if hdr == "" {
		hdr = uuid.New().String()
		// lint:ignore SA1029 legacy
		// nolint:staticcheck // See golangci-lint #741
		ctx = context.WithValue(ctx, common.CorrelationHeader, hdr)
	}
	return ctx, hdr
}
