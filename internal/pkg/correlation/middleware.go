package correlation

import (
	"context"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/logger"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/google/uuid"
	"net/http"
	"time"
)

var LoggingClient logger.LoggingClient

func ManageHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get(internal.CorrelationHeader)
		if hdr == "" {
			hdr = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), internal.CorrelationHeader, hdr)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func OnResponseComplete(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// handle the SMA's snowflake logging setup
		if LoggingClient == nil {
			LoggingClient = logs.LoggingClient
		}

		begin := time.Now()
		next.ServeHTTP(w, r)
		correlationId := FromContext(r.Context())
		LoggingClient.Info("Response complete", internal.CorrelationHeader, correlationId, internal.LogDurationKey, time.Since(begin).String())
	})
}

func OnRequestBegin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// handle the SMA's snowflake logging setup
		if LoggingClient == nil {
			LoggingClient = logs.LoggingClient
		}

		correlationId := FromContext(r.Context())
		LoggingClient.Info("Begin request", internal.CorrelationHeader, correlationId, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
