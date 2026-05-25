package logging

import (
	"fmt"
	"net/http"

	"log/slog"
)

// StringValuer returns a Valuer that
// forces the logger to use the type's String
// method, even in json ouput mode.
// By wrapping the type we defer String
// being called to the point we actually log.
func StringerValuer(s fmt.Stringer) slog.LogValuer {
	return stringerValuer{s}
}

type stringerValuer struct {
	fmt.Stringer
}

func (v stringerValuer) LogValue() slog.Value {
	return slog.StringValue(v.String())
}

func requestToAttr(req *http.Request) slog.Attr {
	return slog.Group("request",
		slog.String("method", req.Method),
		slog.Any("url", StringerValuer(req.URL)),
	)
}

func responseToAttr(resp *http.Response) slog.Attr {
	return slog.Group("response",
		slog.String("status", resp.Status),
		slog.Int64("content_length", resp.ContentLength),
	)
}

// LoggedWriter stores information regarding the response.
// This might be status code, amount of data written or header.
type LoggedWriter interface {
	http.ResponseWriter

	// Attr is called after the next handler
	// in the Middleware returns and
	// the complete reponse should have been written.
	//
	// The returned Attribute should be a [slog.Group]
	// containing response Attributes.
	Attr() slog.Attr

	// Err() is called by the middleware to check
	// if the underlying writer returned an error.
	// If so, the middleware will print an ERROR line.
	Err() error
}
