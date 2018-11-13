package clients

import (
	"context"
)

func fromContext(key string, ctx context.Context) string {
	hdr, ok := ctx.Value(key).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}
