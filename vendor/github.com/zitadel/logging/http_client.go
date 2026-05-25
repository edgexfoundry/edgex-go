package logging

import (
	"context"
	"net/http"
	"time"

	"log/slog"
)

type ClientLoggerOption func(*logRountTripper)

// WithFallbackLogger uses the passed logger if none was
// found in the context.
func WithFallbackLogger(logger *slog.Logger) ClientLoggerOption {
	return func(lrt *logRountTripper) {
		lrt.fallback = logger
	}
}

// WithClientDurationFunc allows overiding the request duration
// for testing.
func WithClientDurationFunc(df func(time.Time) time.Duration) ClientLoggerOption {
	return func(lrt *logRountTripper) {
		lrt.duration = df
	}
}

// WithClientGroup groups the log attributes
// produced by the client.
func WithClientGroup(name string) ClientLoggerOption {
	return func(lrt *logRountTripper) {
		lrt.group = name
	}
}

// WithClientRequestAttr allows customizing the information used
// from a request as request attributes.
func WithClientRequestAttr(requestToAttr func(*http.Request) slog.Attr) ClientLoggerOption {
	return func(lrt *logRountTripper) {
		lrt.reqToAttr = requestToAttr
	}
}

// WithClientResponseAttr allows customizing the information used
// from a response as response attributes.
func WithClientResponseAttr(responseToAttr func(*http.Response) slog.Attr) ClientLoggerOption {
	return func(lrt *logRountTripper) {
		lrt.resToAttr = responseToAttr
	}
}

// EnableHTTPClient adds slog functionality to the HTTP client.
// It attempts to obtain a logger with [FromContext].
// If no logger is in the context, it tries to use a fallback logger,
// which might be set by [WithFallbackLogger].
// If no logger was found finally, the Transport is
// executed without logging.
func EnableHTTPClient(c *http.Client, opts ...ClientLoggerOption) {
	lrt := &logRountTripper{
		next:      c.Transport,
		duration:  time.Since,
		reqToAttr: requestToAttr,
		resToAttr: responseToAttr,
	}
	if lrt.next == nil {
		lrt.next = http.DefaultTransport
	}
	for _, opt := range opts {
		opt(lrt)
	}
	c.Transport = lrt
}

type logRountTripper struct {
	next     http.RoundTripper
	duration func(time.Time) time.Duration
	fallback *slog.Logger

	group     string
	reqToAttr func(*http.Request) slog.Attr
	resToAttr func(*http.Response) slog.Attr
}

// RoundTrip implements [http.RoundTripper].
func (l *logRountTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	logger, ok := l.fromContextOrFallback(req.Context())
	if !ok {
		return l.next.RoundTrip(req)
	}
	start := time.Now()

	resp, err := l.next.RoundTrip(req)
	logger = logger.WithGroup(l.group).With(
		l.reqToAttr(req),
		slog.Duration("duration", l.duration(start)),
	)
	if err != nil {
		logger.Error("request roundtrip", "error", err)
		return resp, err
	}
	logger.Info("request roundtrip", l.resToAttr(resp))
	return resp, nil
}

func (l *logRountTripper) fromContextOrFallback(ctx context.Context) (*slog.Logger, bool) {
	if logger, ok := FromContext(ctx); ok {
		return logger, ok
	}
	return l.fallback, l.fallback != nil
}
