package linkcheck_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	linkcheck "example.com/ch05-linkcheck"
)

// impls runs every behavioural test against both implementations.
var impls = []struct {
	name string
	run  func(*linkcheck.Checker, context.Context, []string) ([]linkcheck.Result, error)
}{
	{"Check", (*linkcheck.Checker).Check},
	{"CheckPool", (*linkcheck.Checker).CheckPool},
}

// testServer serves a few paths with known behaviour and counts the requests
// that are in flight at the same time.
type testServer struct {
	*httptest.Server
	inFlight, maxInFlight atomic.Int64
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	ts := &testServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ok", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("GET /missing", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	// /slow answers only after 10 seconds, or never if the client gives up first.
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(10 * time.Second):
		case <-r.Context().Done():
		}
	})
	// /work?ms=N takes N milliseconds and records how many requests overlap.
	mux.HandleFunc("GET /work", func(w http.ResponseWriter, r *http.Request) {
		n := ts.inFlight.Add(1)
		defer ts.inFlight.Add(-1)
		for {
			old := ts.maxInFlight.Load()
			if n <= old || ts.maxInFlight.CompareAndSwap(old, n) {
				break
			}
		}
		ms, _ := strconv.Atoi(r.URL.Query().Get("ms"))
		select {
		case <-time.After(time.Duration(ms) * time.Millisecond):
		case <-r.Context().Done():
		}
	})
	ts.Server = httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func (ts *testServer) checker(workers int, timeout time.Duration) *linkcheck.Checker {
	return &linkcheck.Checker{Client: ts.Client(), Workers: workers, Timeout: timeout}
}

// checkNoLeaks fails the test if the runtime finds goroutines blocked forever
// on channels or locks that nothing else can reach (section 5.9). A goroutine
// that is about to block needs a moment to get there, so it looks several
// times over a short window.
func checkNoLeaks(t *testing.T) {
	t.Helper()
	for range 10 {
		var buf bytes.Buffer
		if err := pprof.Lookup("goroutineleak").WriteTo(&buf, 1); err != nil {
			t.Fatalf("goroutineleak profile: %v", err)
		}
		if !strings.HasPrefix(buf.String(), "goroutineleak profile: total 0\n") {
			t.Errorf("leaked goroutines:\n%s", buf.String())
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func verdicts(results []linkcheck.Result) []string {
	out := make([]string, len(results))
	for i, r := range results {
		out[i] = r.Verdict()
	}
	return out
}

// Day 8: one result per URL, classified correctly, and a slow URL times out
// without holding up the run.
func TestCheckClassifies(t *testing.T) {
	for _, impl := range impls {
		t.Run(impl.name, func(t *testing.T) {
			ts := newTestServer(t)
			closed := httptest.NewServer(http.NotFoundHandler())
			closed.Close() // nothing listens on this address any more

			urls := []string{
				ts.URL + "/ok",
				ts.URL + "/missing",
				ts.URL + "/slow",
				"://not-a-url",
				closed.URL + "/gone",
			}
			start := time.Now()
			results, err := impl.run(ts.checker(len(urls), 200*time.Millisecond), context.Background(), urls)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("Check returned %v, want nil", err)
			}
			want := []string{"ok", "broken", "timeout", "error", "error"}
			if got := verdicts(results); fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("verdicts = %v, want %v", got, want)
			}
			if results[0].Status != http.StatusOK || results[1].Status != http.StatusNotFound {
				t.Errorf("statuses = %d, %d; want 200, 404", results[0].Status, results[1].Status)
			}
			if !errors.Is(results[2].Err, context.DeadlineExceeded) {
				t.Errorf("slow URL error = %v, want context.DeadlineExceeded", results[2].Err)
			}
			if elapsed > 2*time.Second {
				t.Errorf("run took %v; the slow handler must not be waited for", elapsed)
			}
			checkNoLeaks(t)
		})
	}
}

// Day 9: never more than Workers requests in flight.
func TestCheckBoundsConcurrency(t *testing.T) {
	for _, impl := range impls {
		t.Run(impl.name, func(t *testing.T) {
			ts := newTestServer(t)
			const workers = 3
			urls := make([]string, 20)
			for i := range urls {
				urls[i] = ts.URL + "/work?ms=20"
			}
			results, err := impl.run(ts.checker(workers, 5*time.Second), context.Background(), urls)
			if err != nil {
				t.Fatalf("Check returned %v", err)
			}
			for i, r := range results {
				if r.Err != nil {
					t.Errorf("result %d: %v", i, r.Err)
				}
			}
			got := ts.maxInFlight.Load()
			if got > workers {
				t.Errorf("max in flight = %d, want at most %d", got, workers)
			}
			if got < 2 {
				t.Errorf("max in flight = %d; the requests did not run concurrently", got)
			}
			checkNoLeaks(t)
		})
	}
}

// Day 9: results come back in input order although they finish in reverse.
func TestCheckKeepsInputOrder(t *testing.T) {
	for _, impl := range impls {
		t.Run(impl.name, func(t *testing.T) {
			ts := newTestServer(t)
			var urls []string
			for ms := 100; ms >= 0; ms -= 20 {
				urls = append(urls, fmt.Sprintf("%s/work?ms=%d", ts.URL, ms))
			}
			results, err := impl.run(ts.checker(len(urls), 5*time.Second), context.Background(), urls)
			if err != nil {
				t.Fatalf("Check returned %v", err)
			}
			for i, r := range results {
				if r.URL != urls[i] {
					t.Errorf("results[%d].URL = %s, want %s", i, r.URL, urls[i])
				}
			}
			checkNoLeaks(t)
		})
	}
}

// Day 9: with FailFast, the first failure cancels everything else.
func TestCheckFailFast(t *testing.T) {
	for _, impl := range impls {
		t.Run(impl.name, func(t *testing.T) {
			ts := newTestServer(t)
			urls := []string{ts.URL + "/missing"}
			for range 10 {
				urls = append(urls, ts.URL+"/slow")
			}
			c := ts.checker(2, 5*time.Second)
			c.FailFast = true

			start := time.Now()
			results, err := impl.run(c, context.Background(), urls)
			if elapsed := time.Since(start); elapsed > 2*time.Second {
				t.Errorf("run took %v; fail-fast must cancel the slow requests", elapsed)
			}
			if !errors.Is(err, linkcheck.ErrBroken) {
				t.Fatalf("Check returned %v, want an error wrapping ErrBroken", err)
			}
			if results[0].Verdict() != "broken" {
				t.Errorf("first verdict = %s, want broken", results[0].Verdict())
			}
			for i, r := range results[1:] {
				if v := r.Verdict(); v != "cancelled" && v != "skipped" {
					t.Errorf("results[%d] = %s (%v), want cancelled or skipped", i+1, v, r.Err)
				}
			}
			checkNoLeaks(t)
		})
	}
}

// Day 9: cancelling the parent context stops the run quickly.
func TestCheckParentCancel(t *testing.T) {
	for _, impl := range impls {
		t.Run(impl.name, func(t *testing.T) {
			ts := newTestServer(t)
			urls := make([]string, 6)
			for i := range urls {
				urls[i] = ts.URL + "/slow"
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			start := time.Now()
			_, err := impl.run(ts.checker(2, 5*time.Second), ctx, urls)
			if elapsed := time.Since(start); elapsed > 2*time.Second {
				t.Errorf("run took %v after the parent was cancelled", elapsed)
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("Check returned %v, want context.DeadlineExceeded", err)
			}
			checkNoLeaks(t)
		})
	}
}

func TestWriteReport(t *testing.T) {
	results := []linkcheck.Result{
		{URL: "http://a/ok", Status: 200},
		{URL: "http://a/missing", Status: 404, Err: fmt.Errorf("%w: status 404", linkcheck.ErrBroken)},
		{URL: "http://a/slow", Err: context.DeadlineExceeded},
		{URL: "http://a/later", Err: linkcheck.ErrSkipped},
	}
	var buf bytes.Buffer
	notOK, err := linkcheck.WriteReport(&buf, results)
	if err != nil {
		t.Fatal(err)
	}
	want := "" +
		"ok        200  http://a/ok\n" +
		"broken    404  http://a/missing\n" +
		"timeout     -  http://a/slow\n" +
		"skipped     -  http://a/later\n" +
		"checked 4, ok 1, not ok 3\n"
	if buf.String() != want {
		t.Errorf("report:\n%s\nwant:\n%s", buf.String(), want)
	}
	if notOK != 3 {
		t.Errorf("notOK = %d, want 3", notOK)
	}
}
