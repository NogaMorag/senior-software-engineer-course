package api_test

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
)

func TestCreateEndpointReturnsTheSecretOnce(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	rec := hs.do(req{Method: "POST", Path: "/v1/endpoints", Key: key, Body: `{"url":"https://example.com/hooks"}`})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	created := decode(t, rec)
	id, _ := created["id"].(string)
	if secret, _ := created["secret"].(string); !strings.HasPrefix(secret, "whsec_") {
		t.Errorf("secret = %v, want a whsec_ secret in the create response", created["secret"])
	}
	if created["enabled"] != true || created["url"] != "https://example.com/hooks" {
		t.Errorf("created = %v", created)
	}
	if loc := rec.Header().Get("Location"); loc != "/v1/endpoints/"+id {
		t.Errorf("Location = %q", loc)
	}
	got := decode(t, hs.do(req{Method: "GET", Path: "/v1/endpoints/" + id, Key: key}))
	if _, ok := got["secret"]; ok || got["id"] != id || got["created_at"] != created["created_at"] {
		t.Errorf("GET = %v; want the same endpoint without its secret", got)
	}
}

func TestCreateEndpointRejectsBadInput(t *testing.T) {
	tests := []struct {
		name        string
		dev         bool
		body        string
		contentType string
		status      int
		slug        string
		pointer     string
	}{
		{"http outside dev mode", false, `{"url":"http://localhost:9000/x"}`, "", 400, "validation", "#/url"},
		{"relative url", false, `{"url":"/hooks"}`, "", 400, "validation", "#/url"},
		{"missing url", false, `{}`, "", 400, "validation", "#/url"},
		{"url of the wrong type", false, `{"url":42}`, "", 400, "validation", "#/url"},
		{"unknown field", false, `{"url":"https://example.com","secret":"mine"}`, "", 400, "validation", "#/secret"},
		{"malformed json", false, `{"url":`, "", 400, "malformed-json", ""},
		{"two objects", false, `{"url":"https://a.example"}{}`, "", 400, "malformed-json", ""},
		{"not json", false, `url=https://example.com`, "application/x-www-form-urlencoded", 415, "unsupported-media-type", ""},
		{"too large", false, `{"url":"https://example.com/` + strings.Repeat("a", 5000) + `"}`, "", 413, "body-too-large", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := newHarness(t, tt.dev)
			key := hs.tenant("Acme")
			rec := hs.do(req{Method: "POST", Path: "/v1/endpoints", Key: key, Body: tt.body, ContentType: tt.contentType})
			p := wantProblem(t, rec, tt.status, tt.slug)
			if tt.pointer != "" && !slices.Contains(pointers(p), tt.pointer) {
				t.Errorf("errors = %v, want a pointer %s", p["errors"], tt.pointer)
			}
		})
	}
}

func TestDevModeAcceptsHTTP(t *testing.T) {
	hs := newHarness(t, true)
	key := hs.tenant("Acme")
	if e := hs.createEndpoint(key, "http://sink:8000/hooks"); e["url"] != "http://sink:8000/hooks" {
		t.Fatalf("created = %v", e)
	}
}

func TestListEndpointsPagesWithACursor(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	var want []string
	for range 5 {
		want = append([]string{hs.createEndpoint(key, "https://example.com/hooks")["id"].(string)}, want...)
	}
	var got []string
	path := "/v1/endpoints?limit=2"
	for pages := 1; ; pages++ {
		rec := hs.do(req{Method: "GET", Path: path, Key: key})
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s = %d %s", path, rec.Code, rec.Body)
		}
		body := decode(t, rec)
		for _, it := range body["items"].([]any) {
			item := it.(map[string]any)
			if _, ok := item["secret"]; ok {
				t.Fatal("a listed endpoint carries its secret")
			}
			got = append(got, item["id"].(string))
		}
		next, _ := body["next_cursor"].(string)
		if next == "" {
			if pages != 3 {
				t.Errorf("%d pages for 5 items at limit 2, want 3", pages)
			}
			break
		}
		path = "/v1/endpoints?limit=2&cursor=" + url.QueryEscape(next)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("listed %v\nwant   %v (newest first, no gaps, no duplicates)", got, want)
	}
}

func TestListEndpointsRejectsBadParameters(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	p := wantProblem(t, hs.do(req{Method: "GET", Path: "/v1/endpoints?limit=500", Key: key}), 400, "validation")
	if !slices.Equal(pointers(p), []string{"query:limit"}) {
		t.Errorf("errors = %v", p["errors"])
	}
	wantProblem(t, hs.do(req{Method: "GET", Path: "/v1/endpoints?cursor=bm90LWEtY3Vyc29y", Key: key}), 400, "invalid-cursor")
}

func TestListEndpointsEmptyIsAnEmptyArray(t *testing.T) {
	hs := newHarness(t, false)
	rec := hs.do(req{Method: "GET", Path: "/v1/endpoints", Key: hs.tenant("Acme")})
	if strings.TrimSpace(rec.Body.String()) != `{"items":[]}` {
		t.Fatalf("empty list = %s, want {\"items\":[]}", rec.Body)
	}
}

func TestDisableEndpoint(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	id := hs.createEndpoint(key, "https://example.com/hooks")["id"].(string)
	for range 2 { // idempotent: disabling a disabled endpoint is fine
		rec := hs.do(req{Method: "POST", Path: "/v1/endpoints/" + id + "/disable", Key: key})
		if rec.Code != http.StatusOK || decode(t, rec)["enabled"] != false {
			t.Fatalf("disable = %d %s", rec.Code, rec.Body)
		}
	}
	if decode(t, hs.do(req{Method: "GET", Path: "/v1/endpoints/" + id, Key: key}))["enabled"] != false {
		t.Fatal("GET after disable still shows enabled")
	}
}

func TestUnknownAndMalformedIDsAre404(t *testing.T) {
	hs := newHarness(t, false)
	key := hs.tenant("Acme")
	for _, path := range []string{
		"/v1/endpoints/0199b0a2-6c1e-7a00-8000-000000000042",
		"/v1/endpoints/not-a-uuid",
		"/v1/subscriptions/not-a-uuid",
		"/v1/events/0199B0A2-6C1E-7A00-8000-000000000042", // only the canonical lowercase form is an id
	} {
		wantProblem(t, hs.do(req{Method: "GET", Path: path, Key: key}), http.StatusNotFound, "not-found")
	}
}
