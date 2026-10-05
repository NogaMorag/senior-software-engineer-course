package telemetry

import (
	"context"
	"os"
)

// Version is the build's version, set with -ldflags "-X example.com/hookline/internal/telemetry.Version=...".
var Version = "dev"

// Enabled reports whether a collector is configured. Without OTEL_EXPORTER_OTLP_ENDPOINT, Hookline's
// binaries keep the global no-op providers: tests, the compose runs of milestones M2 to M7 and the labs
// produce no telemetry and no failed exports.
func Enabled() bool { return os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" }

// Start is Setup when Enabled, and otherwise a no-op whose Shutdown does nothing (milestone M8).
func Start(ctx context.Context, service string) (Shutdown, error) {
	if !Enabled() {
		return func(context.Context) error { return nil }, nil
	}
	return Setup(ctx, service, Version)
}
