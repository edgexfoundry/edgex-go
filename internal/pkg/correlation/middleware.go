package correlation

import (
	"context"
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

var LoggingClient logger.LoggingClient

func ManageHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get(clients.CorrelationHeader)
		if hdr == "" {
			hdr = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), clients.CorrelationHeader, hdr)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func OnResponseComplete(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		next.ServeHTTP(w, r)
		correlationId := FromContext(r.Context())
		if LoggingClient != nil {
			LoggingClient.Info("Response complete", clients.CorrelationHeader, correlationId, internal.LogDurationKey, time.Since(begin).String())
		}
	})
}

func OnRequestBegin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationId := FromContext(r.Context())
		if LoggingClient != nil {
			LoggingClient.Info("Begin request", clients.CorrelationHeader, correlationId, "path", r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}
