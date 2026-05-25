package logging

import (
	"context"

	"log/slog"
)

type ctxKeyType struct{}

var ctxKey ctxKeyType

// FromContext takes a Logger from the context, if it was
// previously set by [ToContext]
func FromContext(ctx context.Context) (logger *slog.Logger, ok bool) {
	logger, ok = ctx.Value(ctxKey).(*slog.Logger)
	return logger, ok
}

// ToContext sets a Logger to the context.
func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey, logger)
}
