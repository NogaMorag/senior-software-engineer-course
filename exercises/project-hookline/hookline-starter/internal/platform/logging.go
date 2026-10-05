package platform

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"example.com/hookline/internal/redact"
)

// NewLogger returns a JSON logger. Every record logged through a ...Context
// method (InfoContext, ErrorContext, LogAttrs) gets the request ID and, in
// chapter 24, the trace and span IDs found in the context, so a log line
// leads to its trace and a trace leads to its log lines.
//
// Records below level are dropped. Records at or above it go to w and to
// every handler in also (chapter 24 passes the OpenTelemetry bridge, which
// ships the same records to Loki through the collector).
func NewLogger(w io.Writer, level slog.Leveler, also ...slog.Handler) *slog.Logger {
	var h slog.Handler = contextHandler{slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})}
	if len(also) > 0 {
		h = slog.NewMultiHandler(append([]slog.Handler{h}, also...)...)
	}
	// Redaction in front of every output (milestone M9, spec N7): API keys, signing secrets and bearer
	// tokens never reach stdout or Loki, whoever logs them by mistake.
	return slog.New(minLevel{redact.Handler(h), level})
}

// contextHandler decorates another handler with values taken from the context.
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	id := RequestIDFrom(ctx)
	sc := trace.SpanContextFromContext(ctx)
	if id != "" || sc.IsValid() {
		r = r.Clone() // records share state; never modify one you were handed
		if id != "" {
			r.AddAttrs(slog.String("request_id", id))
		}
		if sc.IsValid() {
			r.AddAttrs(
				slog.String("trace_id", sc.TraceID().String()),
				slog.String("span_id", sc.SpanID().String()),
			)
		}
	}
	return h.Handler.Handle(ctx, r)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}

// minLevel applies one level to every handler behind it, including handlers
// such as the OpenTelemetry bridge that have no level option of their own.
type minLevel struct {
	slog.Handler
	level slog.Leveler
}

func (h minLevel) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= h.level.Level() && h.Handler.Enabled(ctx, l)
}

func (h minLevel) WithAttrs(attrs []slog.Attr) slog.Handler {
	return minLevel{h.Handler.WithAttrs(attrs), h.level}
}

func (h minLevel) WithGroup(name string) slog.Handler {
	return minLevel{h.Handler.WithGroup(name), h.level}
}
