package correlation

import (
	"context"
	"net/http"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/google/uuid"
)

func ManageHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(clients.CorrelationHeader)
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), clients.CorrelationHeader, correlationID)

		contentType := r.Header.Get(clients.ContentType)
		ctx = context.WithValue(ctx, clients.ContentType, contentType)

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
				lc.Trace("Begin request", clients.CorrelationHeader, correlationId, "path", r.URL.Path)
				next.ServeHTTP(w, r)
				lc.Trace("Response complete", clients.CorrelationHeader, correlationId, "duration", time.Since(begin).String())
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
