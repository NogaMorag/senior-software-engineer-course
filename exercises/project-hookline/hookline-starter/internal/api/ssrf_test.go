package api_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"example.com/hookline/internal/api"
	"example.com/hookline/internal/ssrf"
)

// TestEndpointResolvingToAPrivateAddressIsRefused: with the SSRF check of milestone M9 wired in, a URL
// whose host resolves to 10.x is refused at creation, with an answer that does not say which address.
func TestEndpointResolvingToAPrivateAddressIsRefused(t *testing.T) {
	st := newStore(t)
	check := func(ctx context.Context, raw string) error {
		if strings.Contains(raw, "internal.example") {
			return fmt.Errorf("%w: 10.1.2.3 is in 10.0.0.0/8", ssrf.ErrBlocked)
		}
		return nil
	}
	hs := &harness{t: t, store: st, h: api.New(st, api.Options{CheckEndpoint: check, Now: (&clock{t: time.Now()}).now})}
	key := hs.tenant("Acme")
	rec := hs.do(req{Method: "POST", Path: "/v1/endpoints", Key: key, Body: `{"url":"https://internal.example/hooks"}`})
	p := wantProblem(t, rec, http.StatusBadRequest, "validation")
	if strings.Contains(rec.Body.String(), "10.1.2.3") || !strings.Contains(rec.Body.String(), "public address") {
		t.Fatalf("problem = %v; it must refuse without naming the internal address", p)
	}
	if rec := hs.do(req{Method: "POST", Path: "/v1/endpoints", Key: key, Body: `{"url":"https://hooks.example.com/x"}`}); rec.Code != http.StatusCreated {
		t.Fatalf("a public URL = %d %s", rec.Code, rec.Body)
	}
}
