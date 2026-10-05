package main

import "testing"

func TestGreeting(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain name", "Noga", "Hello, Noga!"},
		{"two words", "Noga M", "Hello, Noga M!"},
		{"surrounding spaces", "  Noga ", "Hello, Noga!"},
		{"empty", "", "Hello, world!"},
		{"only spaces", "   ", "Hello, world!"},
		{"hebrew", "נגה", "Hello, נגה!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := greeting(tt.in); got != tt.want {
				t.Errorf("greeting(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
