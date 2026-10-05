package httpapi_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/ch09-catalog-api/internal/catalog"
	"example.com/ch09-catalog-api/internal/httpapi"
	"example.com/ch09-catalog-api/internal/idem"
)

type problemBody struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
	Errors   []struct {
		Detail  string `json:"detail"`
		Pointer string `json:"pointer"`
	} `json:"errors"`
}

type client struct {
	t   *testing.T
	srv *httptest.Server
	n   int
}

func newClient(t *testing.T) *client {
	t.Helper()
	h := httpapi.New(catalog.NewMemory(nil), idem.NewStore(time.Hour, nil), httpapi.Options{MaxBodyBytes: 1024})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &client{t: t, srv: srv}
}

func (c *client) do(method, path, contentType, body string, header ...string) (*http.Response, []byte) {
	c.t.Helper()
	req, err := http.NewRequest(method, c.srv.URL+path, strings.NewReader(body))
	if err != nil {
		c.t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for i := 0; i+1 < len(header); i += 2 {
		req.Header.Set(header[i], header[i+1])
	}
	resp, err := c.srv.Client().Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, b
}

// create posts an item with a fresh idempotency key.
func (c *client) create(body string) (*http.Response, []byte) {
	c.n++
	return c.do("POST", "/v1/items", "application/json", body, "Idempotency-Key", fmt.Sprintf("test-%d", c.n))
}

func mustProblem(t *testing.T, resp *http.Response, body []byte, status int, slug string) problemBody {
	t.Helper()
	if resp.StatusCode != status {
		t.Fatalf("status = %d, want %d; body %s", resp.StatusCode, status, body)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json; body %s", ct, body)
	}
	var p problemBody
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatalf("problem body is not JSON: %v", err)
	}
	if want := "https://catalog.example.com/problems/" + slug; p.Type != want || p.Status != status || p.Title == "" {
		t.Fatalf("problem = %+v, want type %s, status %d and a title", p, want, status)
	}
	return p
}

func pointers(p problemBody) string {
	var out []string
	for _, e := range p.Errors {
		out = append(out, e.Pointer)
	}
	return strings.Join(out, ",")
}

func TestCreateAndGet(t *testing.T) {
	c := newClient(t)
	resp, body := c.create(`{"sku":"MUG-001","name":"  Blue mug ","price_cents":4500}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d %s", resp.StatusCode, body)
	}
	var it catalog.Item
	if err := json.Unmarshal(body, &it); err != nil {
		t.Fatal(err)
	}
	if it.Name != "Blue mug" || it.SKU != "MUG-001" || it.PriceCents != 4500 || it.ID == "" {
		t.Fatalf("created item = %+v", it)
	}
	loc := resp.Header.Get("Location")
	if loc != "/v1/items/"+it.ID {
		t.Fatalf("Location = %q", loc)
	}
	resp, body = c.do("GET", loc, "", "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), it.ID) {
		t.Fatalf("get: %d %s", resp.StatusCode, body)
	}
}

func TestErrorsAreProblems(t *testing.T) {
	c := newClient(t)
	c.create(`{"sku":"MUG-001","name":"Mug","price_cents":1}`)
	tests := []struct {
		name         string
		do           func() (*http.Response, []byte)
		status       int
		slug         string
		wantPointers string
	}{
		{"unknown item", func() (*http.Response, []byte) { return c.do("GET", "/v1/items/itm_99999999", "", "") }, 404, "not-found", ""},
		{"unknown route", func() (*http.Response, []byte) { return c.do("GET", "/v2/everything", "", "") }, 404, "not-found", ""},
		{"duplicate sku", func() (*http.Response, []byte) { return c.create(`{"sku":"MUG-001","name":"Mug","price_cents":1}`) }, 409, "conflict", ""},
		{"all fields missing", func() (*http.Response, []byte) { return c.create(`{}`) }, 400, "validation", "#/sku,#/name,#/price_cents"},
		{"fields invalid", func() (*http.Response, []byte) {
			return c.create(`{"sku":"mug 1","name":"` + strings.Repeat("x", 201) + `","price_cents":-1}`)
		}, 400, "validation", "#/sku,#/name,#/price_cents"},
		{"blank name", func() (*http.Response, []byte) { return c.create(`{"sku":"MUG-002","name":"   ","price_cents":1}`) }, 400, "validation", "#/name"},
		{"unknown field", func() (*http.Response, []byte) { return c.create(`{"sku":"MUG-002","name":"Mug","prise_cents":1}`) }, 400, "validation", "#/prise_cents"},
		{"wrong type", func() (*http.Response, []byte) { return c.create(`{"sku":"MUG-002","name":"Mug","price_cents":"12"}`) }, 400, "validation", "#/price_cents"},
		{"malformed json", func() (*http.Response, []byte) { return c.create(`{"sku":`) }, 400, "malformed-json", ""},
		{"trailing data", func() (*http.Response, []byte) {
			return c.create(`{"sku":"MUG-002","name":"Mug","price_cents":1} {"sku":"X"}`)
		}, 400, "malformed-json", ""},
		{"wrong content type", func() (*http.Response, []byte) {
			return c.do("POST", "/v1/items", "text/plain", `{}`, "Idempotency-Key", "ct-1")
		}, 415, "unsupported-media-type", ""},
		{"no idempotency key", func() (*http.Response, []byte) {
			return c.do("POST", "/v1/items", "application/json", `{"sku":"MUG-003","name":"Mug","price_cents":1}`)
		}, 400, "idempotency-key-missing", ""},
		{"body too large", func() (*http.Response, []byte) {
			return c.create(`{"sku":"MUG-002","name":"` + strings.Repeat("x", 2000) + `","price_cents":1}`)
		}, 413, "body-too-large", ""},
		{"limit too large", func() (*http.Response, []byte) { return c.do("GET", "/v1/items?limit=1000", "", "") }, 400, "validation", "query:limit"},
		{"forged cursor", func() (*http.Response, []byte) { return c.do("GET", "/v1/items?cursor=bm90LWEtY3Vyc29y", "", "") }, 400, "invalid-cursor", ""},
		{"method not allowed", func() (*http.Response, []byte) { return c.do("DELETE", "/v1/items/itm_00000001", "", "") }, 405, "method-not-allowed", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := tt.do()
			p := mustProblem(t, resp, body, tt.status, tt.slug)
			if p.Instance == "" {
				t.Errorf("instance is empty")
			}
			if got := pointers(p); got != tt.wantPointers {
				t.Errorf("error pointers = %q, want %q", got, tt.wantPointers)
			}
		})
	}
}

func TestMethodNotAllowedListsAllow(t *testing.T) {
	c := newClient(t)
	resp, _ := c.do("PUT", "/v1/items", "application/json", `{}`)
	if resp.StatusCode != 405 || resp.Header.Get("Allow") != "GET, POST" {
		t.Fatalf("PUT /v1/items: %d, Allow %q", resp.StatusCode, resp.Header.Get("Allow"))
	}
}

func TestListPagesThroughEverything(t *testing.T) {
	c := newClient(t)
	for i := range 7 {
		c.create(fmt.Sprintf(`{"sku":"SKU-%03d","name":"n","price_cents":1}`, i))
	}
	var ids []string
	path := "/v1/items?limit=3"
	for pages := 0; ; pages++ {
		resp, body := c.do("GET", path, "", "")
		if resp.StatusCode != 200 {
			t.Fatalf("list: %d %s", resp.StatusCode, body)
		}
		var pg struct {
			Items      []catalog.Item `json:"items"`
			NextCursor *string        `json:"next_cursor"`
		}
		if err := json.Unmarshal(body, &pg); err != nil {
			t.Fatal(err)
		}
		for _, it := range pg.Items {
			ids = append(ids, it.ID)
		}
		if pg.NextCursor == nil {
			if pages != 2 {
				t.Fatalf("7 items in pages of 3 should take 3 pages, took %d", pages+1)
			}
			break
		}
		path = "/v1/items?limit=3&cursor=" + *pg.NextCursor
	}
	if len(ids) != 7 || ids[0] <= ids[6] {
		t.Fatalf("ids = %v, want 7, newest first", ids)
	}
}

func TestEmptyListIsArray(t *testing.T) {
	c := newClient(t)
	_, body := c.do("GET", "/v1/items", "", "")
	if strings.TrimSpace(string(body)) != `{"items":[]}` {
		t.Fatalf("empty list = %s, want {\"items\":[]}", body)
	}
}

func TestIdempotentCreateThroughTheStack(t *testing.T) {
	c := newClient(t)
	body := `{"sku":"MUG-001","name":"Mug","price_cents":1}`
	first, b1 := c.do("POST", "/v1/items", "application/json", body, "Idempotency-Key", "abc")
	again, b2 := c.do("POST", "/v1/items", "application/json", body, "Idempotency-Key", "abc")
	if first.StatusCode != 201 || again.StatusCode != 201 || string(b1) != string(b2) {
		t.Fatalf("retry with the same key: %d %s then %d %s; want the same 201 twice", first.StatusCode, b1, again.StatusCode, b2)
	}
	if again.Header.Get("Idempotent-Replayed") != "true" {
		t.Fatalf("the replay should say so")
	}
	resp, b := c.do("POST", "/v1/items", "application/json", `{"sku":"MUG-002","name":"Mug","price_cents":1}`, "Idempotency-Key", "abc")
	mustProblem(t, resp, b, 422, "idempotency-key-reused")
}

func TestDeprecatedAlias(t *testing.T) {
	c := newClient(t)
	_, body := c.create(`{"sku":"MUG-001","name":"Mug","price_cents":1}`)
	var it catalog.Item
	_ = json.Unmarshal(body, &it)
	resp, b := c.do("GET", "/v1/products/"+it.ID, "", "")
	if resp.StatusCode != 200 || !strings.Contains(string(b), it.ID) {
		t.Fatalf("alias: %d %s", resp.StatusCode, b)
	}
	if !strings.HasPrefix(resp.Header.Get("Deprecation"), "@") {
		t.Errorf("Deprecation = %q, want @<unix seconds>", resp.Header.Get("Deprecation"))
	}
	if _, err := http.ParseTime(resp.Header.Get("Sunset")); err != nil {
		t.Errorf("Sunset = %q is not an HTTP date", resp.Header.Get("Sunset"))
	}
	if want := `</v1/items/` + it.ID + `>; rel="successor-version"`; resp.Header.Get("Link") != want {
		t.Errorf("Link = %q, want %q", resp.Header.Get("Link"), want)
	}
	if resp, _ := c.do("GET", "/v1/items/"+it.ID, "", ""); resp.Header.Get("Deprecation") != "" {
		t.Errorf("the current route must not be marked deprecated")
	}
}
