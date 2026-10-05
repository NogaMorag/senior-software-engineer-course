package platform_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"

	"example.com/hookline/internal/platform"
)

// spanContext returns ctx carrying a fixed, valid span context, as otelhttp
// would leave it inside a handler.
func spanContext(t *testing.T) (context.Context, trace.SpanContext) {
	t.Helper()
	tid, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	sid, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: tid, SpanID: sid, TraceFlags: trace.FlagsSampled})
	return trace.ContextWithSpanContext(context.Background(), sc), sc
}

func TestLoggerAddsTraceAndSpanIDs(t *testing.T) {
	ctx, sc := spanContext(t)
	ctx = platform.WithRequestID(ctx, "req-1")
	var buf bytes.Buffer
	log := platform.NewLogger(&buf, slog.LevelInfo)

	log.InfoContext(ctx, "with context")
	log.Info("without context")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), buf.String())
	}
	var with, without map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &with); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &without); err != nil {
		t.Fatal(err)
	}
	if with["trace_id"] != sc.TraceID().String() || with["span_id"] != sc.SpanID().String() || with["request_id"] != "req-1" {
		t.Errorf("line with context = %v, want trace_id, span_id and request_id", with)
	}
	if _, ok := without["trace_id"]; ok {
		t.Errorf("line without context has a trace_id: %v", without)
	}
}

// recordingHandler stands in for the OpenTelemetry bridge.
type recordingHandler struct{ records *[]slog.Record }

func (h recordingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h recordingHandler) Handle(_ context.Context, r slog.Record) error {
	*h.records = append(*h.records, r)
	return nil
}
func (h recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h recordingHandler) WithGroup(string) slog.Handler      { return h }

func TestLoggerFeedsExtraHandlersAtTheSameLevel(t *testing.T) {
	var buf bytes.Buffer
	var got []slog.Record
	log := platform.NewLogger(&buf, slog.LevelInfo, recordingHandler{&got})
	ctx, _ := spanContext(t)

	log.DebugContext(ctx, "below the level")
	log.InfoContext(ctx, "at the level")

	if len(got) != 1 || got[0].Message != "at the level" {
		t.Fatalf("extra handler got %d records, want only the info one", len(got))
	}
	// The bridge reads the trace from ctx itself, so the extra handler must
	// not receive the trace_id attribute a second time.
	got[0].Attrs(func(a slog.Attr) bool {
		if a.Key == "trace_id" {
			t.Errorf("extra handler got a trace_id attribute; it reads the trace from the context")
		}
		return true
	})
	if strings.Count(buf.String(), "\n") != 1 {
		t.Errorf("JSON output = %q, want one line", buf.String())
	}
}
