package ratelimit_test

import (
	"testing"
	"time"

	"example.com/hookline/internal/ratelimit"
)

func TestRetryAfterSeconds(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "1"}, {time.Millisecond, "1"}, {time.Second, "1"}, {1001 * time.Millisecond, "2"}, {90 * time.Second, "90"},
	}
	for _, tt := range tests {
		if got := ratelimit.RetryAfterSeconds(tt.d); got != tt.want {
			t.Errorf("RetryAfterSeconds(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
