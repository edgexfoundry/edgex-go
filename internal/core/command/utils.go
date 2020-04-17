package command

import (
	"context"
	"github.com/edgexfoundry/edgex-go/internal/core/command/errors"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

// propagatedHeader is a slice containing request headers that we want to propagate. It is
// used by the function addHeadersToRequest, which is responsible for adding the headers
// by iterating over the slice, adding any header values contained in this slice.
var propagatedHeader = []string{clients.ContentType}

// addHeadersToRequest populates deviceServiceRequest with matching headers from
// originalRequest to retain information that can be used by the device service.
func addHeadersToRequest(
	originalRequest *http.Request,
	deviceServiceProxiedRequest *http.Request,
	context context.Context) error {

	if originalRequest == nil {
		return errors.NewErrParsingOriginalRequest("header")
	}

	for _, headerName := range propagatedHeader {
		originalHeader := originalRequest.Header.Get(headerName)
		if originalHeader != "" {
			deviceServiceProxiedRequest.Header.Set(headerName, originalHeader)
		}
	}

	// Also populate deviceServiceProxiedRequest with clients.CorrelationHeader value
	// from originalRequest (via the context.Value() part of the request) since
	// inclusion of a Correlation ID (in deviceServiceProxiedRequest) is a requirement.
	correlationID := context.Value(clients.CorrelationHeader)
	if correlationID != nil {
		deviceServiceProxiedRequest.Header.Set(clients.CorrelationHeader, correlationID.(string))
	}

	return nil
}
