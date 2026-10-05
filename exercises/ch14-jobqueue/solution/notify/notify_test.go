package notify_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"example.com/ch14-jobqueue/notify"
	"example.com/ch14-jobqueue/queue"
	"example.com/ch14-jobqueue/testenv"
)

// provider is a fake external service that dedupes on the idempotency key, as real
// providers that accept such a key promise to.
type provider struct {
	mu        sync.Mutex
	calls     int
	delivered map[string]string
	failNext  bool
}

func (p *provider) Send(_ context.Context, key, user, msg string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	if p.failNext {
		p.failNext = false
		return errors.New("provider timeout")
	}
	if p.delivered == nil {
		p.delivered = map[string]string{}
	}
	if _, dup := p.delivered[key]; !dup {
		p.delivered[key] = user + ": " + msg
	}
	return nil
}

func job(t *testing.T, id int64, user string) queue.Job {
	t.Helper()
	payload, err := json.Marshal(notify.Payload{User: user, Message: "your export is ready"})
	if err != nil {
		t.Fatal(err)
	}
	return queue.Job{ID: id, Kind: notify.Kind, Payload: payload}
}

func TestRunningTheSameJobTwiceSendsOnce(t *testing.T) {
	db := testenv.Postgres(t)
	p := &provider{}
	h := notify.Handler(db, p)
	j := job(t, 7, "alice")
	for i := range 2 {
		if err := h(context.Background(), j); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}
	}
	if p.calls != 1 {
		t.Errorf("provider called %d times, want 1 (the second run should find our record)", p.calls)
	}
	if len(p.delivered) != 1 {
		t.Errorf("delivered %d messages, want 1", len(p.delivered))
	}
}

func TestRetryAfterProviderFailureSendsOnce(t *testing.T) {
	db := testenv.Postgres(t)
	p := &provider{failNext: true}
	h := notify.Handler(db, p)
	j := job(t, 8, "bob")
	if err := h(context.Background(), j); err == nil {
		t.Fatal("first run succeeded, want the provider error")
	}
	if err := h(context.Background(), j); err != nil {
		t.Fatal(err)
	}
	if len(p.delivered) != 1 {
		t.Errorf("delivered %d messages, want 1", len(p.delivered))
	}
}

// The crash point: the provider got the message, our insert never ran. The re-run sends
// again with the same key, and the provider's dedupe makes it one delivery.
func TestCrashAfterSendBeforeRecordIsSafe(t *testing.T) {
	db := testenv.Postgres(t)
	p := &provider{}
	j := job(t, 9, "carol")
	key := "notify-job-9"
	if err := p.Send(context.Background(), key, "carol", "your export is ready"); err != nil { // the lost run
		t.Fatal(err)
	}
	if err := notify.Handler(db, p)(context.Background(), j); err != nil {
		t.Fatal(err)
	}
	if p.calls != 2 || len(p.delivered) != 1 {
		t.Errorf("calls = %d, delivered = %d; want 2 calls and 1 delivery", p.calls, len(p.delivered))
	}
}

func TestBadPayloadIsPermanent(t *testing.T) {
	db := testenv.Postgres(t)
	err := notify.Handler(db, &provider{})(context.Background(), queue.Job{ID: 1, Payload: []byte(`{"user":""}`)})
	if !queue.IsPermanent(err) {
		t.Errorf("err = %v, want a permanent error", err)
	}
}
