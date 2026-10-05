package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/ch35-drills/clock"
)

var t0 = time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

// deliver sends one webhook to h and returns the status code and the "status" field.
func deliver(t *testing.T, h http.Handler, id, body string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/webhooks", strings.NewReader(body))
	if id != "" {
		req.Header.Set(EventIDHeader, id)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var r struct {
		Status string `json:"status"`
	}
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
			t.Errorf("200 with a body that is not the JSON reply: %q", rec.Body.String())
		}
	}
	return rec.Code, r.Status
}

// counting returns a processor that counts calls per id.
func counting() (Processor, func(string) int) {
	var mu sync.Mutex
	calls := map[string]int{}
	p := func(_ context.Context, ev Event) error {
		mu.Lock()
		defer mu.Unlock()
		calls[ev.ID]++
		return nil
	}
	return p, func(id string) int { mu.Lock(); defer mu.Unlock(); return calls[id] }
}

func TestRejectsBadRequests(t *testing.T) {
	p, calls := counting()
	h := New(p, time.Hour, clock.NewFake(t0).Now)
	cases := []struct {
		name   string
		method string
		id     string
		body   string
		want   int
	}{
		{"GET", http.MethodGet, "evt_1", `{}`, http.StatusMethodNotAllowed},
		{"no event id", http.MethodPost, "", `{}`, http.StatusBadRequest},
		{"not JSON", http.MethodPost, "evt_1", `{oops`, http.StatusBadRequest},
		{"too large", http.MethodPost, "evt_1", `"` + strings.Repeat("x", MaxBody) + `"`, http.StatusRequestEntityTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/webhooks", strings.NewReader(tc.body))
			if tc.id != "" {
				req.Header.Set(EventIDHeader, tc.id)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status %d, want %d", rec.Code, tc.want)
			}
		})
	}
	if calls("evt_1") != 0 {
		t.Fatal("a rejected request reached the processor")
	}
}

func TestSequentialDuplicatesAreProcessedOnce(t *testing.T) {
	p, calls := counting()
	h := New(p, time.Hour, clock.NewFake(t0).Now)
	if code, st := deliver(t, h, "evt_1", `{"n":1}`); code != 200 || st != "processed" {
		t.Fatalf("first delivery: %d %q, want 200 processed", code, st)
	}
	for range 3 {
		if code, st := deliver(t, h, "evt_1", `{"n":1}`); code != 200 || st != "duplicate" {
			t.Fatalf("retry: %d %q, want 200 duplicate", code, st)
		}
	}
	if code, st := deliver(t, h, "evt_2", `{"n":2}`); code != 200 || st != "processed" {
		t.Fatalf("another event: %d %q, want 200 processed", code, st)
	}
	if calls("evt_1") != 1 || calls("evt_2") != 1 {
		t.Fatalf("calls evt_1=%d evt_2=%d, want 1 and 1", calls("evt_1"), calls("evt_2"))
	}
}

func TestConcurrentDuplicatesAreProcessedOnce(t *testing.T) {
	var calls atomic.Int32
	p := func(context.Context, Event) error {
		calls.Add(1)
		time.Sleep(20 * time.Millisecond) // keep the first delivery busy while the others arrive
		return nil
	}
	h := New(p, time.Hour, clock.NewFake(t0).Now)
	const n = 50
	codes := make([]int, n)
	statuses := make([]string, n)
	gate := make(chan struct{})
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			<-gate
			codes[i], statuses[i] = deliver(t, h, "evt_same", `{"n":1}`)
		})
	}
	close(gate)
	wg.Wait()
	if c := calls.Load(); c != 1 {
		t.Fatalf("processor called %d times for %d concurrent deliveries of one event, want 1", c, n)
	}
	processed := 0
	for i := range n {
		if codes[i] != 200 {
			t.Fatalf("delivery %d got %d, want 200", i, codes[i])
		}
		if statuses[i] == "processed" {
			processed++
		}
	}
	if processed != 1 {
		t.Fatalf("%d deliveries say processed, want exactly 1", processed)
	}
}

func TestDuplicateWaitsForTheDeliveryInFlight(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	p := func(context.Context, Event) error {
		entered <- struct{}{}
		<-release
		return nil
	}
	h := New(p, time.Hour, clock.NewFake(t0).Now)
	first := make(chan string, 1)
	go func() { _, st := deliver(t, h, "evt_1", `{}`); first <- st }()
	<-entered

	second := make(chan string, 1)
	go func() { _, st := deliver(t, h, "evt_1", `{}`); second <- st }()
	select {
	case st := <-second:
		t.Fatalf("duplicate answered %q while the first delivery was still processing", st)
	case <-entered:
		t.Fatal("the duplicate was processed in parallel with the first delivery")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	if st := <-first; st != "processed" {
		t.Fatalf("first delivery: %q, want processed", st)
	}
	if st := <-second; st != "duplicate" {
		t.Fatalf("second delivery: %q, want duplicate", st)
	}
}

func TestFailureIsNotRemembered(t *testing.T) {
	var calls atomic.Int32
	p := func(context.Context, Event) error {
		if calls.Add(1) == 1 {
			return errors.New("database unavailable")
		}
		return nil
	}
	h := New(p, time.Hour, clock.NewFake(t0).Now)
	if code, _ := deliver(t, h, "evt_1", `{}`); code != http.StatusInternalServerError {
		t.Fatalf("failed processing answered %d, want 500 so the sender retries", code)
	}
	if code, st := deliver(t, h, "evt_1", `{}`); code != 200 || st != "processed" {
		t.Fatalf("retry after failure: %d %q, want 200 processed", code, st)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestPanicIsAFailureNotAStuckID(t *testing.T) {
	var calls atomic.Int32
	p := func(context.Context, Event) error {
		if calls.Add(1) == 1 {
			panic("nil map in processor")
		}
		return nil
	}
	h := New(p, time.Hour, clock.NewFake(t0).Now)
	if code, _ := deliver(t, h, "evt_1", `{}`); code != http.StatusInternalServerError {
		t.Fatalf("panicking processor answered %d, want 500", code)
	}
	done := make(chan int, 1)
	go func() { code, _ := deliver(t, h, "evt_1", `{}`); done <- code }()
	select {
	case code := <-done:
		if code != 200 {
			t.Fatalf("retry after panic: %d, want 200", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("retry after a panic hangs: the id was never released")
	}
}

func TestWindowBoundsMemory(t *testing.T) {
	p, calls := counting()
	clk := clock.NewFake(t0)
	h := New(p, 10*time.Minute, clk.Now)
	deliver(t, h, "evt_old", `{}`)
	clk.Advance(5 * time.Minute)
	deliver(t, h, "evt_new", `{}`)
	clk.Advance(5*time.Minute + time.Second) // evt_old is 10m1s old, evt_new 5m1s
	if n := h.Remembered(); n != 1 {
		t.Fatalf("Remembered = %d, want 1 (only evt_new is inside the window)", n)
	}
	if _, st := deliver(t, h, "evt_new", `{}`); st != "duplicate" {
		t.Fatalf("evt_new inside the window: %q, want duplicate", st)
	}
	if _, st := deliver(t, h, "evt_old", `{}`); st != "processed" {
		t.Fatalf("evt_old after the window: %q, want processed (the window is the contract)", st)
	}
	if calls("evt_old") != 2 || calls("evt_new") != 1 {
		t.Fatalf("calls old=%d new=%d, want 2 and 1", calls("evt_old"), calls("evt_new"))
	}
}

func TestSenderHangingUpDoesNotCancelProcessing(t *testing.T) {
	gotErr := make(chan error, 1)
	p := func(ctx context.Context, _ Event) error {
		time.Sleep(20 * time.Millisecond)
		gotErr <- ctx.Err()
		return nil
	}
	h := New(p, time.Hour, clock.NewFake(t0).Now)
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/webhooks", strings.NewReader(`{}`))
	req.Header.Set(EventIDHeader, "evt_1")
	go func() { time.Sleep(5 * time.Millisecond); cancel() }()
	h.ServeHTTP(httptest.NewRecorder(), req)
	if err := <-gotErr; err != nil {
		t.Fatalf("processor saw %v: work must not stop because the sender hung up", err)
	}
	if h.Remembered() != 1 {
		t.Fatal("the processed id was not recorded")
	}
}
