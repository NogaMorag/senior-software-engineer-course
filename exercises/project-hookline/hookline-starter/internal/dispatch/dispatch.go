// Package dispatch is the dispatcher (milestone M4): a consumer-group member on hookline.events that
// turns each event into one delivery row per matching, enabled endpoint, and commits the offset only
// after the rows are committed (chapter 13). Reprocessing an event, after a relay duplicate or a crash
// before the commit, inserts nothing new: deliveries has UNIQUE (event_id, endpoint_id).
package dispatch

import (
	"context"

	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"

	"example.com/hookline/internal/metrics"

	"example.com/hookline/internal/outbox"
)

// Matcher answers "which endpoints get this event?"; *subcache.Cache is the real one.
type Matcher interface {
	Endpoints(ctx context.Context, tenantID, eventType string) ([]string, error)
}

// Dispatcher fans events out into deliveries.
type Dispatcher struct {
	DB      *pgxpool.Pool
	Matcher Matcher
	Log     *slog.Logger
	// Tracer, when set, is the kotel tracer of the consuming client: each record gets a process span
	// that continues the trace from its headers (milestone M8).
	Tracer *kotel.Tracer
	// Metrics counts the created rows; nil means none.
	Metrics *metrics.Metrics
}

// Fanout inserts one pending delivery per endpoint and reports how many rows are new. uuidv7() is
// built into PostgreSQL 18. The insert is one statement, so either all rows of the event exist or none.
func Fanout(ctx context.Context, db *pgxpool.Pool, m outbox.Message, endpointIDs []string) (int, error) {
	panic("TODO: see chapter 38")

}

// Handle processes one record. A record that cannot be decoded is logged and skipped: retrying it
// would block its partition forever (a poison message). Any other error is returned, and Run retries.
func (d *Dispatcher) Handle(ctx context.Context, r *kgo.Record) error {
	panic("TODO: see chapter 38")

}

// Run polls cl until ctx is done and commits offsets only for records that handle finished. A failing
// handle (the database is down) is retried in place with a doubling wait up to 5 seconds, so records
// are neither skipped nor reordered. cl must come from bus.NewGroupConsumer (no auto-commit, rebalances
// blocked while a poll's records are being handled).
func Run(ctx context.Context, cl *kgo.Client, handle func(context.Context, *kgo.Record) error, log *slog.Logger) error {
	panic("TODO: see chapter 38")

}

// retry runs fn until it succeeds (true) or ctx is done (false).
func retry(ctx context.Context, log *slog.Logger, fn func() error) bool {
	panic("TODO: see chapter 38")

}
