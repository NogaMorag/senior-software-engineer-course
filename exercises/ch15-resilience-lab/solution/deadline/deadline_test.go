package deadline_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"example.com/ch15-resilience-lab/deadline"
)

func TestMiddlewareTurnsTheHeaderIntoADeadline(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   time.Duration // the budget the handler should see, roughly
	}{
		{"header below max", "300", 300 * time.Millisecond},
		{"header above max is clamped", "60000", 2 * time.Second},
		{"missing header gets max", "", 2 * time.Second},
		{"garbage gets max", "soon", 2 * time.Second},
		{"zero gets max", "0", 2 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var left time.Duration
			h := deadline.Middleware(2*time.Second, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				dl, ok := r.Context().Deadline()
				if !ok {
					t.Fatal("handler context has no deadline")
				}
				left = time.Until(dl)
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set(deadline.Header, tt.header)
			}
			h.ServeHTTP(httptest.NewRecorder(), req)
			if left > tt.want || left < tt.want-100*time.Millisecond {
				t.Fatalf("handler saw %v left, want about %v", left, tt.want)
			}
		})
	}
}

func TestDownstreamKeepsAReserve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	down, stop, err := deadline.Downstream(ctx, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	req, _ := http.NewRequestWithContext(down, http.MethodGet, "http://example.invalid/", nil)
	deadline.Inject(req)
	ms, err := strconv.Atoi(req.Header.Get(deadline.Header))
	if err != nil || ms > 250 || ms < 150 {
		t.Fatalf("injected header = %q, want about 250 (300 minus the 50ms reserve)", req.Header.Get(deadline.Header))
	}
}

func TestDownstreamRefusesWhenTheBudgetIsSpent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, _, err := deadline.Downstream(ctx, 50*time.Millisecond); !errors.Is(err, deadline.ErrNoBudget) {
		t.Fatalf("Downstream with 20ms left and a 50ms reserve = %v, want ErrNoBudget", err)
	}
}

func TestNoDeadlineNoHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	deadline.Inject(req)
	if v := req.Header.Get(deadline.Header); v != "" {
		t.Fatalf("header = %q without a deadline, want none", v)
	}
}
