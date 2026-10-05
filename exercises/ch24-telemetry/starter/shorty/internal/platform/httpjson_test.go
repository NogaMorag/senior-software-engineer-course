package platform_test

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/ch24-telemetry/internal/platform"
)

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int // 0 means success
		want       point
	}{
		{"valid", `{"x": 1, "y": 2}`, 0, point{1, 2}},
		{"empty body", ``, http.StatusBadRequest, point{}},
		{"not JSON", `x=1`, http.StatusBadRequest, point{}},
		{"wrong type", `{"x": "one"}`, http.StatusBadRequest, point{}},
		{"unknown field", `{"x": 1, "z": 3}`, http.StatusBadRequest, point{}},
		{"two values", `{"x": 1} {"x": 2}`, http.StatusBadRequest, point{}},
		{"too large", `{"x": 1, "y": 2, "padding": "` + strings.Repeat("a", 100) + `"}`, http.StatusRequestEntityTooLarge, point{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			got, err := platform.DecodeJSON[point](httptest.NewRecorder(), req, 64)
			if tt.wantStatus == 0 {
				if err != nil || got != tt.want {
					t.Fatalf("DecodeJSON = %+v, %v; want %+v, nil", got, err, tt.want)
				}
				return
			}
			var he *platform.HTTPError
			if !errors.As(err, &he) || he.Status != tt.wantStatus {
				t.Fatalf("DecodeJSON error = %v, want *HTTPError with status %d", err, tt.wantStatus)
			}
		})
	}
}

func TestRespondErrorHidesInternalErrors(t *testing.T) {
	logs := &syncBuffer{}
	log := platform.NewLogger(logs, slog.LevelInfo)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	platform.RespondError(rec, req, log, errors.New("dial tcp 10.0.0.7:5432: connection refused"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "10.0.0.7") {
		t.Errorf("response leaks the internal error: %s", rec.Body.String())
	}
	if lines := logs.lines(t); len(lines) != 1 || !strings.Contains(lines[0]["error"].(string), "10.0.0.7") {
		t.Errorf("the internal error must be logged once, got %v", lines)
	}
}
