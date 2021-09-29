package correlation

import (
	"context"
	"net/http"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/google/uuid"
)

func ManageHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(common.CorrelationHeader)
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		// lint:ignore SA1029 legacy
		// nolint:staticcheck // See golangci-lint #741
		ctx := context.WithValue(r.Context(), common.CorrelationHeader, correlationID)

		contentType := r.Header.Get(common.ContentType)
		// lint:ignore SA1029 legacy
		// nolint:staticcheck // See golangci-lint #741
		ctx = context.WithValue(ctx, common.ContentType, contentType)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(lc logger.LoggingClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if lc.LogLevel() == models.TraceLog {
				begin := time.Now()
				correlationId := FromContext(r.Context())
				lc.Trace("Begin request", common.CorrelationHeader, correlationId, "path", r.URL.Path)
				next.ServeHTTP(w, r)
				lc.Trace("Response complete", common.CorrelationHeader, correlationId, "duration", time.Since(begin).String())
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
