package holiday_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/ch27-review/internal/holiday"
)

func client(t *testing.T, h http.HandlerFunc) *holiday.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &holiday.Client{URL: srv.URL, HTTP: srv.Client(), Timeout: 200 * time.Millisecond}
}

func reply(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) }
}

func TestToday(t *testing.T) {
	c := client(t, reply(`{"name":"Sukkot","date":"2026-10-05","days":7}`))
	h, err := c.Today(context.Background())
	if err != nil || h.Name != "Sukkot" || h.Date != "2026-10-05" || h.Days != 7 {
		t.Fatalf("Today = %+v, %v", h, err)
	}
}

func TestTodayRejectsBadAnswers(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"days of the wrong type": reply(`{"name":"Sukkot","date":"2026-10-05","days":"seven"}`),
		"date in another format": reply(`{"name":"Sukkot","date":"05/10/2026"}`),
		"not JSON":               reply(`<html>`),
		"no name":                reply(`{"date":"2026-10-05"}`),
		"server error": func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		},
	}
	for name, h := range tests {
		t.Run(name, func(t *testing.T) {
			if got, err := client(t, h).Today(context.Background()); err == nil {
				t.Fatalf("Today = %+v; want an error", got)
			}
		})
	}
}

func TestTodayTimesOut(t *testing.T) {
	c := client(t, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // never answers
	})
	start := time.Now()
	if _, err := c.Today(context.Background()); err == nil {
		t.Fatal("Today returned no error from a server that never answers")
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("Today took %v; the 200ms timeout was not applied", d)
	}
}
