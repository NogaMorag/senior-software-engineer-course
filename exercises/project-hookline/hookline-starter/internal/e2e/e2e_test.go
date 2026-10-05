// Package e2e runs the whole pipeline against the compose stack: the API over HTTP, the relay, a
// dispatcher with the Redis subscription cache, and a worker pool sending signed requests to the sink.
// It is milestone M5's completion test (spec F4, F6, F7, F8).
package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"example.com/hookline/internal/api"
	"example.com/hookline/internal/backoff"
	"example.com/hookline/internal/bus"
	"example.com/hookline/internal/cache"
	"example.com/hookline/internal/delivery"
	"example.com/hookline/internal/dispatch"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/store/postgres"
	"example.com/hookline/internal/store/storetest"
	"example.com/hookline/internal/subcache"
	"example.com/hookline/internal/testdb"
	"example.com/hookline/internal/testenv"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

// client calls one HTTP service and decodes JSON answers.
type client struct {
	t    *testing.T
	base string
	key  string
}

func (c client) call(method, path string, body any, headers map[string]string) (int, map[string]any) {
	c.t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base+path, r)
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.key != "" {
		req.Header.Set("Authorization", "Bearer "+c.key)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func (c client) must(want int, method, path string, body any, headers map[string]string) map[string]any {
	c.t.Helper()
	code, out := c.call(method, path, body, headers)
	if code != want {
		c.t.Fatalf("%s %s = %d %v, want %d", method, path, code, out, want)
	}
	return out
}

// eventually polls cond every 100ms for up to a minute.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Minute)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestPipelineEndToEnd(t *testing.T) {
	brokers, redisAddr, sinkURL := testenv.Brokers(t), testenv.RedisAddr(t), testenv.SinkURL(t)
	pool := testdb.Pool(t)
	st := postgres.New(pool)
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	// The services, as the four binaries wire them, in this process.
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()
	subs := subcache.New(cache.NewRedisStore(rdb), st, cache.Options{TTL: time.Minute})
	apiSrv := httptest.NewServer(api.New(st, api.Options{DevMode: true, Subscriptions: subs}))
	defer apiSrv.Close()

	topic := testenv.Topic(t, brokers, 3)
	producer, err := bus.NewProducer(brokers)
	if err != nil {
		t.Fatal(err)
	}
	defer producer.Close()
	consumer, err := bus.NewGroupConsumer(brokers, testenv.Unique("e2e-dispatchers"), topic)
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Close()
	relay := &outbox.Relay{DB: pool, Publisher: bus.KafkaPublisher{Client: producer, Topic: topic}}
	dispatcher := &dispatch.Dispatcher{DB: pool, Matcher: subs, Log: quiet}
	workers := &delivery.Pool{
		DB: pool, Sender: delivery.Sender{Timeout: 2 * time.Second}, Workers: 4, Lease: 10 * time.Second,
		Poll: 50 * time.Millisecond, MaxAttempts: 30, Log: quiet, // room for the outage of subtest 2
		Backoff: backoff.Policy{Base: 100 * time.Millisecond, Cap: 500 * time.Millisecond},
	}
	done := make(chan struct{}, 3)
	go func() { _ = relay.Run(ctx, 50*time.Millisecond, quiet); done <- struct{}{} }()
	go func() { _ = dispatch.Run(ctx, consumer, dispatcher.Handle, quiet); done <- struct{}{} }()
	go func() { _ = workers.Run(ctx); done <- struct{}{} }()
	defer func() {
		stop()
		for range 3 {
			<-done
		}
	}()

	sink := client{t: t, base: sinkURL}
	sink.must(200, "POST", "/reset", nil, nil)
	defer sink.call("POST", "/reset", nil, nil)

	_, key := storetest.NewTenant(t, st, "Acme")
	hl := client{t: t, base: apiSrv.URL, key: key}
	hooks := map[string]string{} // endpoint id -> sink hook name
	endpoint := func(name, eventType string) string {
		hook := testenv.Unique(name)
		id := hl.must(201, "POST", "/v1/endpoints", map[string]string{"url": sinkURL + "/hooks/" + hook}, nil)["id"].(string)
		// The sink verifies every hook with one known secret; give the endpoint that secret instead of
		// the generated one, which only the create response showed (a test fixture, not an API).
		if _, err := pool.Exec(ctx, `UPDATE endpoints SET secret = $1 WHERE id = $2`, testenv.SinkSecret, id); err != nil {
			t.Fatal(err)
		}
		hl.must(201, "POST", "/v1/subscriptions", map[string]string{"endpoint_id": id, "event_type": eventType}, nil)
		hooks[id] = hook
		return hook
	}
	paidHook, allHook, shipHook := endpoint("paid", "invoice.paid"), endpoint("all", "*"), endpoint("ship", "shipment.moved")

	ingest := func(idemKey, typ string, n int) string {
		return hl.must(202, "POST", "/v1/events", map[string]any{"type": typ, "payload": map[string]int{"n": n}},
			map[string]string{"Idempotency-Key": idemKey})["id"].(string)
	}
	stats := func() map[string]map[string]float64 {
		_, s := sink.call("GET", "/stats", nil, nil)
		out := map[string]map[string]float64{}
		hooksOut, _ := s["hooks"].(map[string]any)
		for name, c := range hooksOut {
			out[name] = map[string]float64{}
			for k, v := range c.(map[string]any) {
				out[name][k] = v.(float64)
			}
		}
		return out
	}

	t.Run("every event reaches every matching endpoint, signed, once per endpoint", func(t *testing.T) {
		var ids []string
		for i := range 20 {
			ids = append(ids, ingest(fmt.Sprintf("paid-%d", i), "invoice.paid", i))
		}
		for i := range 5 {
			ingest(fmt.Sprintf("ship-%d", i), "shipment.moved", i)
		}
		for i := range 3 { // client retries: the same keys again, no new events
			if again := ingest(fmt.Sprintf("paid-%d", i), "invoice.paid", i); again != ids[i] {
				t.Fatalf("retry of paid-%d created event %s, want %s", i, again, ids[i])
			}
		}
		eventually(t, "the sink to receive 20, 25 and 5 unique events", func() bool {
			s := stats()
			return s[paidHook]["unique"] == 20 && s[allHook]["unique"] == 25 && s[shipHook]["unique"] == 5
		})
		s := stats()
		for _, h := range []string{paidHook, allHook, shipHook} {
			if s[h]["rejected"] != 0 {
				t.Errorf("hook %s rejected %v requests: a bad signature", h, s[h]["rejected"])
			}
		}
		ds := hl.must(200, "GET", "/v1/events/"+ids[0]+"/deliveries", nil, nil)["items"].([]any)
		if len(ds) != 2 {
			t.Fatalf("event %s has %d deliveries, want 2 (paid and all)", ids[0], len(ds))
		}
		for _, d := range ds {
			d := d.(map[string]any)
			history := d["history"].([]any)
			if d["status"] != "delivered" || len(history) != 1 || history[0].(map[string]any)["status_code"] != float64(200) {
				t.Errorf("delivery = %v", d)
			}
		}
	})

	t.Run("a failing receiver is retried until it recovers", func(t *testing.T) {
		sink.must(200, "POST", "/control", map[string]any{"failure_rate": 1, "failure_status": 503}, nil)
		id := ingest("flaky-1", "shipment.moved", 0) // matches ship and all: two deliveries
		deliveries := func() []map[string]any {
			var out []map[string]any
			for _, d := range hl.must(200, "GET", "/v1/events/"+id+"/deliveries", nil, nil)["items"].([]any) {
				out = append(out, d.(map[string]any))
			}
			return out
		}
		eventually(t, "two failed attempts on each delivery", func() bool {
			ds := deliveries()
			for _, d := range ds {
				if len(d["history"].([]any)) < 2 {
					return false
				}
			}
			return len(ds) == 2
		})
		sink.must(200, "POST", "/control", map[string]any{"failure_rate": 0}, nil)
		eventually(t, "both deliveries after recovery", func() bool {
			for _, d := range deliveries() {
				if d["status"] != "delivered" || d["attempts"].(float64) < 3 {
					return false
				}
			}
			return true
		})
	})

	t.Run("a 4xx is not retried, the delivery dies, and replay delivers it", func(t *testing.T) {
		sink.must(200, "POST", "/control", map[string]any{"failure_rate": 1, "failure_status": 422}, nil)
		id := ingest("rejected-1", "invoice.paid", 0)
		dead := func() []map[string]any {
			var out []map[string]any
			for _, d := range hl.must(200, "GET", "/v1/deliveries?status=dead", nil, nil)["items"].([]any) {
				if d := d.(map[string]any); d["event_id"] == id {
					out = append(out, d)
				}
			}
			return out
		}
		eventually(t, "two dead deliveries", func() bool { return len(dead()) == 2 })
		sink.must(200, "POST", "/control", map[string]any{"failure_rate": 0}, nil)
		for _, d := range dead() {
			if d["attempts"] != float64(1) || !strings.Contains(d["last_error"].(string), "422") {
				t.Errorf("dead delivery = %v; want one attempt (422 is final)", d)
			}
			hl.must(202, "POST", "/v1/deliveries/"+d["id"].(string)+"/replay", nil, nil)
		}
		eventually(t, "both replays delivered", func() bool {
			ds := hl.must(200, "GET", "/v1/events/"+id+"/deliveries", nil, nil)["items"].([]any)
			for _, d := range ds {
				d := d.(map[string]any)
				if d["status"] != "delivered" || d["attempts"] != float64(2) {
					return false
				}
			}
			return len(ds) == 2
		})
	})
}
