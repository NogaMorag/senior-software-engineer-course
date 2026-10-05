package dispatch_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"

	"example.com/hookline/internal/bus"
	"example.com/hookline/internal/cache"
	"example.com/hookline/internal/dispatch"
	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/store/postgres"
	"example.com/hookline/internal/subcache"
	"example.com/hookline/internal/testdb"
	"example.com/hookline/internal/testenv"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

// scene: a tenant whose endpoints match invoice.paid in every way that matters.
type scene struct {
	pool                        *pgxpool.Pool
	st                          *postgres.Store
	tenant                      domain.Tenant
	exact, wildcard, other, off domain.Endpoint
}

func newScene(t *testing.T) scene {
	pool := testdb.Pool(t)
	st := postgres.New(pool)
	tn := testenv.Tenant(t, st)
	ep := func() domain.Endpoint {
		return testenv.Endpoint(t, st, tn.ID, "https://receiver.example/hooks", testenv.SinkSecret)
	}
	s := scene{pool: pool, st: st, tenant: tn, exact: ep(), wildcard: ep(), other: ep(), off: ep()}
	testenv.Subscribe(t, st, tn.ID, s.exact.ID, "invoice.paid")
	testenv.Subscribe(t, st, tn.ID, s.wildcard.ID, "*")
	testenv.Subscribe(t, st, tn.ID, s.wildcard.ID, "invoice.paid") // both: still one delivery
	testenv.Subscribe(t, st, tn.ID, s.other.ID, "shipment.moved")
	testenv.Subscribe(t, st, tn.ID, s.off.ID, "*")
	if _, err := st.DisableEndpoint(context.Background(), tn.ID, s.off.ID); err != nil {
		t.Fatal(err)
	}
	return s
}

func (s scene) endpointsOf(t *testing.T, eventID string) []string {
	t.Helper()
	rows, err := s.pool.Query(context.Background(), `SELECT endpoint_id::text FROM deliveries WHERE event_id = $1 ORDER BY 1`, eventID)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	return ids
}

func (s scene) want() []string {
	w := []string{s.exact.ID, s.wildcard.ID}
	slices.Sort(w)
	return w
}

func (s scene) dispatcher() *dispatch.Dispatcher {
	subs := subcache.New(subcache.Disabled(), s.st, cache.Options{TTL: time.Minute})
	return &dispatch.Dispatcher{DB: s.pool, Matcher: subs, Log: quiet}
}

func record(ev domain.Event) *kgo.Record {
	m := outbox.Message{EventID: ev.ID, TenantID: ev.TenantID, Type: ev.Type}
	return &kgo.Record{Topic: outbox.Topic, Key: []byte(ev.TenantID), Value: m.Value()}
}

func TestHandleFansOutToMatchingEnabledEndpointsOnce(t *testing.T) {
	s := newScene(t)
	ev := testenv.Event(t, s.st, s.tenant.ID, "invoice.paid", map[string]int{"amount": 1})
	d := s.dispatcher()
	for range 3 { // a relay duplicate, then a redelivery after a crash: no new rows
		if err := d.Handle(context.Background(), record(ev)); err != nil {
			t.Fatal(err)
		}
	}
	if got := s.endpointsOf(t, ev.ID); !slices.Equal(got, s.want()) {
		t.Fatalf("deliveries for %v, want exactly %v", got, s.want())
	}
}

func TestFanoutReportsOnlyNewRows(t *testing.T) {
	s := newScene(t)
	ev := testenv.Event(t, s.st, s.tenant.ID, "invoice.paid", map[string]int{"amount": 1})
	m := outbox.Message{EventID: ev.ID, TenantID: ev.TenantID, Type: ev.Type}
	for i, want := range []int{2, 0} {
		n, err := dispatch.Fanout(context.Background(), s.pool, m, s.want())
		if err != nil || n != want {
			t.Fatalf("Fanout #%d = %d, %v; want %d", i+1, n, err, want)
		}
	}
	if n, err := dispatch.Fanout(context.Background(), s.pool, m, nil); err != nil || n != 0 {
		t.Fatalf("Fanout with no endpoints = %d, %v", n, err)
	}
}

func TestPoisonRecordIsSkipped(t *testing.T) {
	s := newScene(t)
	if err := s.dispatcher().Handle(context.Background(), &kgo.Record{Value: []byte("not an event")}); err != nil {
		t.Fatalf("Handle(poison) = %v, want nil: retrying it would block the partition", err)
	}
}

type failingMatcher struct{}

func (failingMatcher) Endpoints(context.Context, string, string) ([]string, error) {
	return nil, errors.New("database down")
}

func TestMatcherFailureIsReturnedForRetry(t *testing.T) {
	s := newScene(t)
	ev := testenv.Event(t, s.st, s.tenant.ID, "invoice.paid", map[string]int{})
	d := &dispatch.Dispatcher{DB: s.pool, Matcher: failingMatcher{}, Log: quiet}
	if err := d.Handle(context.Background(), record(ev)); err == nil {
		t.Fatal("Handle succeeded although matching failed; the offset would be committed and the event lost")
	}
}

// TestRelayToDispatcherThroughRedpanda: ingest, relay with a crash between publish and mark (so every
// event is on the topic twice), dispatch from a consumer group: every event ends with exactly one
// delivery per matching endpoint (chapter 37, section 37.6, rows 2 and 4).
func TestRelayToDispatcherThroughRedpanda(t *testing.T) {
	brokers := testenv.Brokers(t)
	s := newScene(t)
	topic := testenv.Topic(t, brokers, 3)
	var events []domain.Event
	for i := range 10 {
		events = append(events, testenv.Event(t, s.st, s.tenant.ID, "invoice.paid", map[string]int{"n": i}))
	}
	producer, err := bus.NewProducer(brokers)
	if err != nil {
		t.Fatal(err)
	}
	defer producer.Close()
	pub := bus.KafkaPublisher{Client: producer, Topic: topic}
	ctx := context.Background()
	crash := errors.New("relay killed after publish")
	if _, err := (&outbox.Relay{DB: s.pool, Publisher: pub, AfterPublish: func(context.Context, []outbox.Message) error { return crash }}).RunOnce(ctx); !errors.Is(err, crash) {
		t.Fatalf("crashing relay = %v", err)
	}
	for {
		n, err := (&outbox.Relay{DB: s.pool, Publisher: pub}).RunOnce(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			break
		}
	}

	consumer, err := bus.NewGroupConsumer(brokers, testenv.Unique("dispatchers"), topic)
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Close()
	runCtx, stop := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- dispatch.Run(runCtx, consumer, s.dispatcher().Handle, quiet) }()

	deadline := time.Now().Add(30 * time.Second)
	for _, ev := range events {
		for !slices.Equal(s.endpointsOf(t, ev.ID), s.want()) {
			if time.Now().After(deadline) {
				t.Fatalf("event %s has deliveries %v, want %v", ev.ID, s.endpointsOf(t, ev.ID), s.want())
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	stop()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM deliveries WHERE tenant_id = $1`, s.tenant.ID).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 2*len(events) {
		t.Fatalf("deliveries = %d, want %d: the duplicate publish created extra rows", total, 2*len(events))
	}
}
