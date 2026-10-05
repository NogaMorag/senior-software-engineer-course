package greet_test

import (
	"testing"

	"example.com/ch27-review/internal/greet"
)

func TestWord(t *testing.T) {
	tests := []struct{ header, want string }{
		{"he-IL,he;q=0.9,en;q=0.8", "shalom"},
		{"fr-CA", "bonjour"},
		{"en-GB", "hello"},
		{"", "hello"},
		{"de-DE", "hello"},
		{"not a ;; header", "hello"},
	}
	for _, tc := range tests {
		t.Run(tc.header, func(t *testing.T) {
			if got := greet.Word(tc.header); got != tc.want {
				t.Fatalf("Word(%q) = %q; want %q", tc.header, got, tc.want)
			}
		})
	}
}

func TestGreet(t *testing.T) {
	tests := []struct{ name, header, want string }{
		{"  noga ", "he", "Shalom, Noga"},
		{"", "en", "Hello, Friend"},
		{"dana", "fr", "Bonjour, Dana"},
	}
	for _, tc := range tests {
		if got := greet.Greet(tc.name, tc.header); got != tc.want {
			t.Errorf("Greet(%q, %q) = %q; want %q", tc.name, tc.header, got, tc.want)
		}
	}
}
