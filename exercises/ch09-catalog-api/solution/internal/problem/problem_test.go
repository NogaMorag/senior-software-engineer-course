package problem_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/ch09-catalog-api/internal/problem"
)

func TestWrite(t *testing.T) {
	tests := []struct {
		name       string
		p          *problem.Problem
		wantStatus int
		wantType   string
		wantKeys   []string
		absentKeys []string
	}{
		{
			name:       "not found has the five standard members",
			p:          problem.NotFound.New("no item with id itm_1"),
			wantStatus: 404,
			wantType:   problem.TypeBase + "not-found",
			wantKeys:   []string{"type", "title", "status", "detail", "instance"},
			absentKeys: []string{"errors"},
		},
		{
			name: "validation carries the errors extension",
			p: func() *problem.Problem {
				p := problem.Validation.New("bad")
				p.Errors = []problem.FieldError{{Detail: "is required", Pointer: "#/sku"}}
				return p
			}(),
			wantStatus: 400,
			wantType:   problem.TypeBase + "validation",
			wantKeys:   []string{"type", "title", "status", "errors"},
		},
		{
			name:       "empty detail is omitted",
			p:          problem.Internal.New(""),
			wantStatus: 500,
			wantType:   problem.TypeBase + "internal",
			absentKeys: []string{"detail", "errors"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			problem.Write(rec, httptest.NewRequest(http.MethodGet, "/v1/items/itm_1", nil), tt.p)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
				t.Fatalf("Content-Type = %q, want application/problem+json", ct)
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not JSON: %v", err)
			}
			if body["type"] != tt.wantType {
				t.Errorf("type = %v, want %s", body["type"], tt.wantType)
			}
			if body["status"] != float64(tt.wantStatus) {
				t.Errorf("status member = %v, want %d", body["status"], tt.wantStatus)
			}
			if body["instance"] != "/v1/items/itm_1" {
				t.Errorf("instance = %v, want the request path", body["instance"])
			}
			for _, k := range tt.wantKeys {
				if _, ok := body[k]; !ok {
					t.Errorf("member %q missing in %s", k, rec.Body)
				}
			}
			for _, k := range tt.absentKeys {
				if _, ok := body[k]; ok {
					t.Errorf("member %q should be omitted in %s", k, rec.Body)
				}
			}
		})
	}
}

func TestTitleIsFixedPerType(t *testing.T) {
	a, b := problem.Conflict.New("sku A"), problem.Conflict.New("sku B")
	if a.Title != b.Title || a.Type != b.Type || a.Detail == b.Detail {
		t.Fatalf("title and type must be fixed per type and detail per occurrence: %+v %+v", a, b)
	}
}
