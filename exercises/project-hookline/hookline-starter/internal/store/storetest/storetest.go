// Package storetest is the conformance suite for store.Store. The memory store (M1) and the PostgreSQL
// store (M2) run the same tests, which is what lets the API swap one for the other without a change.
// Every test makes its own tenants, so the suite also runs against one shared database.
package storetest

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"example.com/hookline/internal/domain"
	"example.com/hookline/internal/page"
	"example.com/hookline/internal/store"
)

// Run runs every conformance test against stores made by newStore.
func Run(t *testing.T, newStore func(t *testing.T) store.Store) {
	tests := []struct {
		name string
		fn   func(t *testing.T, s store.Store)
	}{
		{"TenantByKeyHash", testTenantByKeyHash},
		{"EndpointGetIsTenantScoped", testEndpointGet},
		{"EndpointListPagesWithoutGapsOrDuplicates", testEndpointList},
		{"EndpointDisable", testEndpointDisable},
		{"Subscriptions", testSubscriptions},
		{"SubscriptionListFilterAndPages", testSubscriptionList},
		{"IngestIsIdempotent", testIngest},
		{"IngestConcurrentSameKeyStoresOneEvent", testIngestConcurrent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { tt.fn(t, newStore(t)) })
	}
}

// base is a fixed instant per test run, truncated as domain.Now truncates.
var base = time.Now().UTC().Truncate(time.Second)

// NewTenant creates a tenant with a fresh API key in s and returns it with the key.
func NewTenant(t testing.TB, s store.Tenants, name string) (domain.Tenant, string) {
	t.Helper()
	key, hash := domain.NewAPIKey()
	tn := domain.Tenant{ID: domain.NewID(base), Name: name, APIKeyHash: hash, RatePerSec: 50, Burst: 100, CreatedAt: base}
	if err := s.CreateTenant(context.Background(), tn); err != nil {
		t.Fatalf("CreateTenant: %v", err)
	}
	return tn, key
}

func newEndpoint(t *testing.T, s store.Store, tenantID string, at time.Time) domain.Endpoint {
	t.Helper()
	e := domain.Endpoint{ID: domain.NewID(at), TenantID: tenantID, URL: "https://example.com/hooks", Secret: domain.NewSigningSecret(), Enabled: true, CreatedAt: at}
	if err := s.CreateEndpoint(context.Background(), e); err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}
	return e
}

func wantErr(t *testing.T, what string, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("%s: error = %v, want %v", what, err, want)
	}
}

func testTenantByKeyHash(t *testing.T, s store.Store) {
	ctx := context.Background()
	tn, key := NewTenant(t, s, "Acme")
	got, err := s.TenantByKeyHash(ctx, domain.HashAPIKey(key))
	if err != nil || got.ID != tn.ID || got.Name != "Acme" || got.RatePerSec != 50 || got.Burst != 100 || !got.CreatedAt.Equal(base) {
		t.Fatalf("TenantByKeyHash = %+v, %v; want %+v", got, err, tn)
	}
	_, err = s.TenantByKeyHash(ctx, domain.HashAPIKey("hk_unknown"))
	wantErr(t, "unknown key", err, domain.ErrNotFound)
	dup := tn
	dup.ID = domain.NewID(base)
	wantErr(t, "second tenant with the same key hash", s.CreateTenant(ctx, dup), domain.ErrConflict)
}

func testEndpointGet(t *testing.T, s store.Store) {
	ctx := context.Background()
	a, _ := NewTenant(t, s, "A")
	b, _ := NewTenant(t, s, "B")
	e := newEndpoint(t, s, a.ID, base)
	got, err := s.GetEndpoint(ctx, a.ID, e.ID)
	if err != nil || !reflect.DeepEqual(normalize(got), normalize(e)) {
		t.Fatalf("GetEndpoint = %+v, %v; want %+v", got, err, e)
	}
	_, err = s.GetEndpoint(ctx, b.ID, e.ID)
	wantErr(t, "another tenant's endpoint", err, domain.ErrNotFound)
	_, err = s.GetEndpoint(ctx, a.ID, domain.NewID(base))
	wantErr(t, "unknown endpoint", err, domain.ErrNotFound)
}

func testEndpointList(t *testing.T, s store.Store) {
	ctx := context.Background()
	a, _ := NewTenant(t, s, "A")
	b, _ := NewTenant(t, s, "B")
	var want []string // newest first
	for i := range 5 {
		want = append([]string{newEndpoint(t, s, a.ID, base.Add(time.Duration(i)*time.Second)).ID}, want...)
	}
	// Two endpoints created in the same microsecond: the id breaks the tie, so the order stays total.
	tie := base.Add(10 * time.Second)
	x, y := newEndpoint(t, s, a.ID, tie), newEndpoint(t, s, a.ID, tie)
	if x.ID < y.ID {
		x, y = y, x
	}
	want = append([]string{x.ID, y.ID}, want...)
	newEndpoint(t, s, b.ID, base.Add(time.Hour)) // another tenant's: never listed for A

	var got []string
	var after *page.Cursor
	for pages := 0; ; pages++ {
		if pages > 10 {
			t.Fatal("pagination does not terminate")
		}
		items, more, err := s.ListEndpoints(ctx, a.ID, after, 3)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range items {
			got = append(got, e.ID)
		}
		if !more {
			break
		}
		last := items[len(items)-1]
		after = &page.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pages = %v\nwant    %v", got, want)
	}
	items, more, err := s.ListEndpoints(ctx, a.ID, nil, 7)
	if err != nil || len(items) != 7 || more {
		t.Fatalf("exact-size page = %d items, more=%v, %v; want 7, false", len(items), more, err)
	}
}

func testEndpointDisable(t *testing.T, s store.Store) {
	ctx := context.Background()
	a, _ := NewTenant(t, s, "A")
	b, _ := NewTenant(t, s, "B")
	e := newEndpoint(t, s, a.ID, base)
	_, err := s.DisableEndpoint(ctx, b.ID, e.ID)
	wantErr(t, "disable another tenant's endpoint", err, domain.ErrNotFound)
	for range 2 { // disabling twice is not an error
		got, err := s.DisableEndpoint(ctx, a.ID, e.ID)
		if err != nil || got.Enabled || got.ID != e.ID {
			t.Fatalf("DisableEndpoint = %+v, %v", got, err)
		}
	}
	if got, _ := s.GetEndpoint(ctx, a.ID, e.ID); got.Enabled {
		t.Fatal("endpoint still enabled after DisableEndpoint")
	}
}

func newSub(tenantID, endpointID, eventType string, at time.Time) domain.Subscription {
	return domain.Subscription{ID: domain.NewID(at), TenantID: tenantID, EndpointID: endpointID, EventType: eventType, CreatedAt: at}
}

func testSubscriptions(t *testing.T, s store.Store) {
	ctx := context.Background()
	a, _ := NewTenant(t, s, "A")
	b, _ := NewTenant(t, s, "B")
	e := newEndpoint(t, s, a.ID, base)
	sub := newSub(a.ID, e.ID, "invoice.paid", base)
	if err := s.CreateSubscription(ctx, sub); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSubscription(ctx, newSub(a.ID, e.ID, domain.AllTypes, base)); err != nil {
		t.Fatalf("wildcard next to an exact type: %v", err)
	}
	wantErr(t, "same endpoint and type twice", s.CreateSubscription(ctx, newSub(a.ID, e.ID, "invoice.paid", base)), domain.ErrConflict)
	wantErr(t, "B subscribes A's endpoint", s.CreateSubscription(ctx, newSub(b.ID, e.ID, "invoice.paid", base)), domain.ErrNotFound)
	wantErr(t, "unknown endpoint", s.CreateSubscription(ctx, newSub(a.ID, domain.NewID(base), "invoice.paid", base)), domain.ErrNotFound)

	got, err := s.GetSubscription(ctx, a.ID, sub.ID)
	if err != nil || !reflect.DeepEqual(normalize(got), normalize(sub)) {
		t.Fatalf("GetSubscription = %+v, %v; want %+v", got, err, sub)
	}
	_, err = s.GetSubscription(ctx, b.ID, sub.ID)
	wantErr(t, "B reads A's subscription", err, domain.ErrNotFound)
	wantErr(t, "B deletes A's subscription", s.DeleteSubscription(ctx, b.ID, sub.ID), domain.ErrNotFound)
	if err := s.DeleteSubscription(ctx, a.ID, sub.ID); err != nil {
		t.Fatal(err)
	}
	wantErr(t, "delete twice", s.DeleteSubscription(ctx, a.ID, sub.ID), domain.ErrNotFound)
	if err := s.CreateSubscription(ctx, newSub(a.ID, e.ID, "invoice.paid", base)); err != nil {
		t.Fatalf("subscribe again after delete: %v", err)
	}
}

func testSubscriptionList(t *testing.T, s store.Store) {
	ctx := context.Background()
	a, _ := NewTenant(t, s, "A")
	e1 := newEndpoint(t, s, a.ID, base)
	e2 := newEndpoint(t, s, a.ID, base)
	for i, typ := range []string{"a.one", "a.two", "a.three"} {
		at := base.Add(time.Duration(i) * time.Second)
		for _, e := range []domain.Endpoint{e1, e2} {
			if err := s.CreateSubscription(ctx, newSub(a.ID, e.ID, typ, at)); err != nil {
				t.Fatal(err)
			}
		}
	}
	all, more, err := s.ListSubscriptions(ctx, a.ID, "", nil, 100)
	if err != nil || len(all) != 6 || more {
		t.Fatalf("all = %d, more=%v, %v; want 6", len(all), more, err)
	}
	first, more, err := s.ListSubscriptions(ctx, a.ID, e1.ID, nil, 2)
	if err != nil || len(first) != 2 || !more || first[0].EventType != "a.three" || first[1].EventType != "a.two" {
		t.Fatalf("first page for e1 = %+v, more=%v, %v", first, more, err)
	}
	last := first[1]
	rest, more, err := s.ListSubscriptions(ctx, a.ID, e1.ID, &page.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}, 2)
	if err != nil || len(rest) != 1 || more || rest[0].EventType != "a.one" || rest[0].EndpointID != e1.ID {
		t.Fatalf("second page for e1 = %+v, more=%v, %v", rest, more, err)
	}
}

// ActiveSubscriptions checks store.Subscriptions.ActiveSubscriptions (milestone M3). It is separate from
// Run so that the M1 and M2 commands can skip it by name until milestone M3.
func ActiveSubscriptions(t *testing.T, s store.Store) {
	ctx := context.Background()
	a, _ := NewTenant(t, s, "A")
	b, _ := NewTenant(t, s, "B")
	on := newEndpoint(t, s, a.ID, base)
	off := newEndpoint(t, s, a.ID, base)
	other := newEndpoint(t, s, b.ID, base)
	for _, sub := range []domain.Subscription{
		newSub(a.ID, on.ID, "invoice.paid", base), newSub(a.ID, on.ID, domain.AllTypes, base),
		newSub(a.ID, off.ID, "invoice.paid", base), newSub(b.ID, other.ID, "invoice.paid", base),
	} {
		if err := s.CreateSubscription(ctx, sub); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.DisableEndpoint(ctx, a.ID, off.ID); err != nil {
		t.Fatal(err)
	}
	subs, err := s.ActiveSubscriptions(ctx, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 2 {
		t.Fatalf("active = %+v, want the 2 subscriptions of the enabled endpoint", subs)
	}
	for _, sub := range subs {
		if sub.EndpointID != on.ID || sub.TenantID != a.ID {
			t.Fatalf("active subscription %+v belongs to a disabled endpoint or another tenant", sub)
		}
	}
}

// NewEvent builds an event for ingest with the body hash of body.
func NewEvent(tenantID, key, typ, payload string, at time.Time) domain.Event {
	body := fmt.Sprintf(`{"type":%q,"payload":%s}`, typ, payload)
	return domain.Event{ID: domain.NewID(at), TenantID: tenantID, Type: typ, Payload: json.RawMessage(payload),
		IdempotencyKey: key, BodyHash: sha256.Sum256([]byte(body)), CreatedAt: at}
}

func testIngest(t *testing.T, s store.Store) {
	ctx := context.Background()
	a, _ := NewTenant(t, s, "A")
	b, _ := NewTenant(t, s, "B")
	first := NewEvent(a.ID, "key-1", "invoice.paid", `{"invoice": "inv_1", "amount": 4200}`, base)
	stored, replayed, err := s.IngestEvent(ctx, first)
	if err != nil || replayed || stored.ID != first.ID {
		t.Fatalf("first ingest = %+v, replayed=%v, %v", stored, replayed, err)
	}

	retry := NewEvent(a.ID, "key-1", "invoice.paid", `{"invoice": "inv_1", "amount": 4200}`, base.Add(time.Second))
	again, replayed, err := s.IngestEvent(ctx, retry)
	if err != nil || !replayed || again.ID != first.ID || !again.CreatedAt.Equal(first.CreatedAt) {
		t.Fatalf("retry = %+v, replayed=%v, %v; want the original event %s", again, replayed, err, first.ID)
	}

	changed := NewEvent(a.ID, "key-1", "invoice.paid", `{"invoice": "inv_1", "amount": 9900}`, base)
	_, _, err = s.IngestEvent(ctx, changed)
	wantErr(t, "same key, different body", err, domain.ErrIdempotencyMismatch)

	other := NewEvent(b.ID, "key-1", "invoice.paid", `{"invoice": "inv_1", "amount": 4200}`, base)
	if got, replayed, err := s.IngestEvent(ctx, other); err != nil || replayed || got.ID != other.ID {
		t.Fatalf("the same key is a new event for another tenant: %+v, replayed=%v, %v", got, replayed, err)
	}

	got, err := s.GetEvent(ctx, a.ID, first.ID)
	if err != nil || got.Type != "invoice.paid" || !jsonEqual(got.Payload, first.Payload) || !got.CreatedAt.Equal(base) {
		t.Fatalf("GetEvent = %+v, %v", got, err)
	}
	_, err = s.GetEvent(ctx, b.ID, first.ID)
	wantErr(t, "B reads A's event", err, domain.ErrNotFound)
}

func testIngestConcurrent(t *testing.T, s store.Store) {
	ctx := context.Background()
	a, _ := NewTenant(t, s, "A")
	const n = 16
	type result struct {
		id       string
		replayed bool
		err      error
	}
	results := make(chan result, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			<-start // release all at once, so the inserts really race
			e := NewEvent(a.ID, "same-key", "invoice.paid", `{"n": 1}`, base)
			got, replayed, err := s.IngestEvent(ctx, e)
			results <- result{got.ID, replayed, err}
		})
	}
	close(start)
	wg.Wait()
	close(results)
	ids, fresh := map[string]bool{}, 0
	for r := range results {
		if r.err != nil {
			t.Fatalf("concurrent ingest: %v", r.err)
		}
		ids[r.id] = true
		if !r.replayed {
			fresh++
		}
	}
	if len(ids) != 1 || fresh != 1 {
		t.Fatalf("%d concurrent ingests with one key: %d distinct ids, %d not replayed; want 1 and 1", n, len(ids), fresh)
	}
}

// normalize drops what a store need not return byte for byte (time zone pointer).
func normalize[T any](v T) T {
	switch x := any(&v).(type) {
	case *domain.Endpoint:
		x.CreatedAt = x.CreatedAt.UTC()
	case *domain.Subscription:
		x.CreatedAt = x.CreatedAt.UTC()
	}
	return v
}

func jsonEqual(a, b []byte) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
