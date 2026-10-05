package page_test

import (
	"encoding/base64"
	"testing"
	"time"

	"example.com/hookline/internal/page"
)

func TestCursorRoundTrip(t *testing.T) {
	c := page.Cursor{CreatedAt: time.Date(2026, 10, 5, 12, 0, 0, 123456789, time.UTC), ID: "0199b0a2-6c1e-7a00-8000-000000000042"}
	s := page.Encode(c)
	for _, ch := range s {
		if ch == '+' || ch == '/' || ch == '=' {
			t.Fatalf("cursor %q is not URL-safe", s)
		}
	}
	got, err := page.Decode(s)
	if err != nil {
		t.Fatalf("Decode(Encode(c)) error: %v", err)
	}
	if !got.CreatedAt.Equal(c.CreatedAt) || got.ID != c.ID {
		t.Fatalf("round trip = %+v, want %+v", got, c)
	}
}

func TestDecodeRejects(t *testing.T) {
	b64 := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	tests := map[string]string{
		"empty":         "",
		"not base64":    "%%%",
		"not json":      b64("hello"),
		"wrong version": b64(`{"v":2,"t":1,"id":"itm_1"}`),
		"missing id":    b64(`{"v":1,"t":1}`),
		"missing time":  b64(`{"v":1,"id":"itm_1"}`),
	}
	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := page.Decode(in); err != page.ErrInvalid {
				t.Fatalf("Decode(%q) error = %v, want ErrInvalid", in, err)
			}
		})
	}
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"", page.DefaultLimit, false},
		{"1", 1, false},
		{"100", 100, false},
		{"0", 0, true},
		{"101", 0, true},
		{"-5", 0, true},
		{"ten", 0, true},
	}
	for _, tt := range tests {
		got, err := page.ParseLimit(tt.in)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("ParseLimit(%q) = %d, %v; want %d, error %v", tt.in, got, err, tt.want, tt.wantErr)
		}
	}
}
