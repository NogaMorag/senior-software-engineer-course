package platform

import (
	"context"
	"io"
	"log/slog"
)

// NewLogger returns a JSON logger. Every record logged through a ...Context
// method (InfoContext, ErrorContext, LogAttrs) gets the request ID found in
// the context, so all the lines of one request can be found together.
func NewLogger(w io.Writer, level slog.Leveler) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	return slog.New(contextHandler{h})
}

// contextHandler decorates another handler with values taken from the context.
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := RequestIDFrom(ctx); id != "" {
		r = r.Clone() // records share state; never modify one you were handed
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}
