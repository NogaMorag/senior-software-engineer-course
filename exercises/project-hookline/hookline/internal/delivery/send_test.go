package delivery_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/hookline/internal/delivery"
)

var fixedNow = time.Unix(1791201600, 0)

func job(url string) delivery.Job {
	return delivery.Job{
		DeliveryID: "0199b0a2-6c1e-7a00-8000-0000000000d1", EventID: "0199b0a2-6c1e-7a00-8000-0000000000e1",
		Attempt: 3, URL: url, Secret: "whsec_change-me", Enabled: true, EventType: "invoice.paid",
		EventAt: time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC), Payload: json.RawMessage(`{"amount": 4200}`),
	}
}

func TestSendSignsAndSetsTheHeaders(t *testing.T) {
	var got *http.Request
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, body = r, must(io.ReadAll(r.Body))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	j := job(srv.URL + "/hooks/a")
	res := delivery.Sender{Now: func() time.Time { return fixedNow }}.Send(context.Background(), j)
	if !res.OK() || res.StatusCode != 204 {
		t.Fatalf("Send = %+v, want delivered with 204", res)
	}
	for header, want := range map[string]string{
		delivery.HeaderEventID: j.EventID, delivery.HeaderDeliveryID: j.DeliveryID, delivery.HeaderAttempt: "3",
		delivery.HeaderSignature: delivery.Sign(j.Secret, fixedNow.Unix(), body), "Content-Type": "application/json",
	} {
		if v := got.Header.Get(header); v != want {
			t.Errorf("%s = %q, want %q", header, v, want)
		}
	}
	if string(body) != string(j.Body()) {
		t.Errorf("body = %s, want exactly the signed bytes %s", body, j.Body())
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil || parsed["id"] != j.EventID || parsed["type"] != "invoice.paid" {
		t.Errorf("body = %s", body)
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		status    int
		ok, retry bool
	}{
		{200, true, false}, {201, true, false}, {204, true, false},
		{408, false, true}, {429, false, true}, {500, false, true}, {502, false, true}, {503, false, true},
		{400, false, false}, {401, false, false}, {404, false, false}, {410, false, false}, {422, false, false},
		{301, false, false}, {302, false, false},
	}
	for _, tt := range tests {
		retry, err := delivery.Classify(tt.status)
		if (err == nil) != tt.ok || retry != tt.retry {
			t.Errorf("Classify(%d) = retry %v, err %v; want ok=%v retry=%v", tt.status, retry, err, tt.ok, tt.retry)
		}
	}
}

func TestSendStatusesAndRetryAfter(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		retryAfter string
		retry      bool
		wantAfter  time.Duration
	}{
		{"server error", 503, "", true, 0},
		{"rate limited with Retry-After", 429, "7", true, 7 * time.Second},
		{"bad request is final", 400, "", false, 0},
		{"redirect is not followed and is final", 302, "", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/elsewhere" {
					t.Error("the redirect was followed")
				}
				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}
				if tt.status == 302 {
					w.Header().Set("Location", "/elsewhere")
				}
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()
			res := delivery.Sender{}.Send(context.Background(), job(srv.URL))
			if res.OK() || res.StatusCode != tt.status || res.Retryable != tt.retry || res.RetryAfter != tt.wantAfter {
				t.Fatalf("Send = %+v", res)
			}
		})
	}
}

func TestSendTimeoutAndRefusedAreRetryable(t *testing.T) {
	block := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-block:
		case <-r.Context().Done():
		}
	}))
	defer slow.Close()
	defer close(block)
	start := time.Now()
	res := delivery.Sender{Timeout: 100 * time.Millisecond}.Send(context.Background(), job(slow.URL))
	if res.OK() || !res.Retryable || res.StatusCode != 0 || !strings.Contains(res.Err.Error(), "no response within") {
		t.Fatalf("timeout: %+v", res)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("the attempt took %v; the per-attempt timeout did not apply", time.Since(start))
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close() // nothing listens there now
	res = delivery.Sender{Timeout: time.Second}.Send(context.Background(), job("http://"+addr+"/hooks"))
	if res.OK() || !res.Retryable || res.StatusCode != 0 {
		t.Fatalf("connection refused: %+v", res)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
