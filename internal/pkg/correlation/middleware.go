package correlation

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
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

// UrlDecodeMiddleware decode the path variables
// After invoking the router.UseEncodedPath() func, the path variables needs to decode before passing to the controller
func UrlDecodeMiddleware(lc logger.LoggingClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			for k, v := range vars {
				unescape, err := url.QueryUnescape(v)
				if err != nil {
					lc.Debugf("failed to decode the %s from the value %s", k, v)
					return
				}
				vars[k] = unescape
			}
			next.ServeHTTP(w, r)
		})
	}
}
