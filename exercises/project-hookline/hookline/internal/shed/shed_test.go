package shed_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"example.com/hookline/internal/shed"
)

// blocking is a handler that holds every request until release is closed.
func blocking() (http.Handler, chan struct{}, *sync.WaitGroup) {
	release := make(chan struct{})
	var entered sync.WaitGroup
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered.Done()
		<-release
		w.WriteHeader(http.StatusOK)
	}), release, &entered
}

func send(h http.Handler, high bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/events", nil)
	if high {
		req.Header.Set("X-Priority", "high")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestShedsAboveTheLimitImmediately(t *testing.T) {
	next, release, entered := blocking()
	s := shed.New(shed.Options{Limit: 2, RetryAfter: 2 * time.Second})
	h := s.Wrap(next)

	var done sync.WaitGroup
	entered.Add(2)
	for range 2 {
		done.Go(func() {
			if rec := send(h, false); rec.Code != http.StatusOK {
				t.Errorf("admitted request: status %d", rec.Code)
			}
		})
	}
	entered.Wait() // two requests are in flight now

	start := time.Now()
	rec := send(h, false)
	if rec.Code != http.StatusServiceUnavailable || rec.Header().Get("Retry-After") != "2" {
		t.Fatalf("third request: status %d, Retry-After %q; want 503 and 2", rec.Code, rec.Header().Get("Retry-After"))
	}
	if el := time.Since(start); el > 50*time.Millisecond {
		t.Fatalf("shedding took %v: a shed request must not wait", el)
	}

	close(release)
	done.Wait()
	entered.Add(1)
	if rec := send(h, false); rec.Code != http.StatusOK {
		t.Fatalf("after the load drained: status %d, want 200", rec.Code)
	}
	if s.Served() != 3 || s.Shed(shed.Low) != 1 || s.InFlight() != 0 {
		t.Fatalf("served %d, shed %d, in flight %d; want 3, 1, 0", s.Served(), s.Shed(shed.Low), s.InFlight())
	}
}

func TestHighPriorityKeepsItsReserve(t *testing.T) {
	next, release, entered := blocking()
	s := shed.New(shed.Options{Limit: 4, ReservedForHigh: 2, Priority: shed.ByHeader("X-Priority", "high")})
	h := s.Wrap(next)
	var done sync.WaitGroup
	hold := func(high bool) {
		entered.Add(1)
		done.Go(func() { send(h, high) })
		entered.Wait()
	}
	hold(false)
	hold(false) // low priority has used its two slots
	if rec := send(h, false); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("third low-priority request: status %d, want 503", rec.Code)
	}
	hold(true)
	hold(true) // high priority gets the reserved two
	if rec := send(h, true); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("fifth request in flight: status %d, want 503 at the total limit", rec.Code)
	}
	close(release)
	done.Wait()
	if s.Shed(shed.Low) != 1 || s.Shed(shed.High) != 1 {
		t.Fatalf("shed low %d, high %d; want 1 and 1", s.Shed(shed.Low), s.Shed(shed.High))
	}
}

func TestNeverMoreThanTheLimitInFlight(t *testing.T) {
	var mu sync.Mutex
	inside, peak := 0, 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		inside++
		peak = max(peak, inside)
		mu.Unlock()
		time.Sleep(time.Millisecond)
		mu.Lock()
		inside--
		mu.Unlock()
	})
	h := shed.New(shed.Options{Limit: 5}).Wrap(next)
	var wg sync.WaitGroup
	for range 200 {
		wg.Go(func() { send(h, false) })
	}
	wg.Wait()
	if peak > 5 {
		t.Fatalf("peak in flight = %d, want at most 5", peak)
	}
}
