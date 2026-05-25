package logging

import (
	"net/http"
	"time"

	"log/slog"
)

type MiddlewareOption func(*middleware)

// WitLogger sets the passed logger with request attributes
// into the Request's context.
func WithLogger(logger *slog.Logger) MiddlewareOption {
	return func(m *middleware) {
		m.logger = logger
	}
}

// WithGroup groups the log attributes
// produced by the middleware.
func WithGroup(name string) MiddlewareOption {
	return func(m *middleware) {
		m.group = name
	}
}

// WithIDFunc enables the creating of request IDs
// in the middleware, which are then attached to
// the logger.
func WithIDFunc(nextID func() slog.Attr) MiddlewareOption {
	return func(m *middleware) {
		m.nextID = nextID
	}
}

// WithDurationFunc allows overriding the request duration for testing.
func WithDurationFunc(df func(time.Time) time.Duration) MiddlewareOption {
	return func(m *middleware) {
		m.duration = df
	}
}

// WithRequestAttr allows customizing the information used
// from a request as request attributes.
func WithRequestAttr(requestToAttr func(*http.Request) slog.Attr) MiddlewareOption {
	return func(m *middleware) {
		m.reqAttr = requestToAttr
	}
}

// WithLoggedWriter allows customizing the writer from
// which post-request attributes are taken.
func WithLoggedWriter(wrap func(w http.ResponseWriter) LoggedWriter) MiddlewareOption {
	return func(m *middleware) {
		m.wrapWriter = wrap
	}
}

// Middleware enables request logging and sets a logger
// to the request context.
// Use [FromContext] to obtain the logger anywhere in the request liftime.
//
// The default logger is [slog.Default], with the request's URL and Method
// as preset attributes.
// When the request terminates, a INFO line with the Status Code and
// amount written to the client is printed.
// This behaviors can be modified with options.
func Middleware(options ...MiddlewareOption) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		mw := &middleware{
			logger:     slog.Default(),
			duration:   time.Since,
			next:       next,
			reqAttr:    requestToAttr,
			wrapWriter: newLoggedWriter,
		}
		for _, opt := range options {
			opt(mw)
		}
		return mw
	}
}

type middleware struct {
	logger     *slog.Logger
	group      string
	nextID     func() slog.Attr
	next       http.Handler
	duration   func(time.Time) time.Duration
	reqAttr    func(*http.Request) slog.Attr
	wrapWriter func(http.ResponseWriter) LoggedWriter
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	logger := m.logger.With(slog.Group(m.group, m.reqAttr(r)))
	if m.nextID != nil {
		logger = logger.With(slog.Group(m.group, m.nextID()))
	}
	r = r.WithContext(ToContext(r.Context(), logger))

	lw := m.wrapWriter(w)
	m.next.ServeHTTP(lw, r)
	logger = logger.With(slog.Group(m.group,
		slog.Duration("duration", m.duration(start)),
		lw.Attr(),
	))
	if err := lw.Err(); err != nil {
		logger.WarnContext(r.Context(), "write response", "error", err)
		return
	}
	logger.InfoContext(r.Context(), "request served")
}

type loggedWriter struct {
	http.ResponseWriter

	statusCode int
	written    int
	err        error
}

func newLoggedWriter(w http.ResponseWriter) LoggedWriter {
	return &loggedWriter{
		ResponseWriter: w,
	}
}

func (w *loggedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggedWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.written += n
	w.err = err
	return n, err
}

func (lw *loggedWriter) Attr() slog.Attr {
	return slog.Group("response",
		"status", lw.statusCode,
		"written", lw.written,
	)
}

func (lw *loggedWriter) Err() error {
	return lw.err
}
