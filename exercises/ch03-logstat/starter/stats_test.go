package main

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestCollectSampleLog(t *testing.T) {
	f, err := os.Open("testdata/access.log")
	if err != nil {
		t.Fatalf("open sample log: %v", err)
	}
	defer f.Close()

	s, err := Collect(f)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if s.Lines != 15 || s.Malformed != 2 {
		t.Errorf("Lines, Malformed = %d, %d; want 15, 2", s.Lines, s.Malformed)
	}
	if want := []int{7, 14}; !slices.Equal(s.BadLines, want) {
		t.Errorf("BadLines = %v, want %v", s.BadLines, want)
	}
	if want := [6]int{0, 0, 8, 2, 2, 1}; s.Classes != want {
		t.Errorf("Classes = %v, want %v", s.Classes, want)
	}
	if got := s.Paths["/api/orders"]; got != 5 {
		t.Errorf(`Paths["/api/orders"] = %d, want 5 (the query string must not split the path)`, got)
	}
}

func TestCollectKeepsOnlyFirstBadLines(t *testing.T) {
	s, err := Collect(strings.NewReader(strings.Repeat("garbage\n", 8)))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if s.Malformed != 8 {
		t.Errorf("Malformed = %d, want 8", s.Malformed)
	}
	if want := []int{1, 2, 3, 4, 5}; !slices.Equal(s.BadLines, want) {
		t.Errorf("BadLines = %v, want %v", s.BadLines, want)
	}
}

func TestCollectLongLine(t *testing.T) {
	long := `203.0.113.7 - - [05/Oct/2026:09:12:01 +0300] "GET /` + strings.Repeat("a", 100_000) + ` HTTP/1.1" 200 1`
	s, err := Collect(strings.NewReader(long + "\n"))
	if err != nil {
		t.Fatalf("Collect on a 100 KB line: %v (did you raise the scanner buffer?)", err)
	}
	if s.Lines != 1 || s.Malformed != 0 {
		t.Errorf("Lines, Malformed = %d, %d; want 1, 0", s.Lines, s.Malformed)
	}
}

func TestTopPaths(t *testing.T) {
	paths := map[string]int{"/b": 3, "/a": 3, "/c": 1, "/d": 7}
	tests := []struct {
		name string
		n    int
		want []PathCount
	}{
		{"top three, ties broken by path", 3, []PathCount{{"/d", 7}, {"/a", 3}, {"/b", 3}}},
		{"n larger than the map", 10, []PathCount{{"/d", 7}, {"/a", 3}, {"/b", 3}, {"/c", 1}}},
		{"zero", 0, []PathCount{}},
		{"negative is treated as zero", -1, []PathCount{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run several times: map order changes between iterations, the result must not.
			for range 20 {
				if got := TopPaths(paths, tt.n); !slices.Equal(got, tt.want) {
					t.Fatalf("TopPaths(n=%d) = %v, want %v", tt.n, got, tt.want)
				}
			}
		})
	}
}

func TestDisplayPathAndTruncate(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		width int
		want  string
	}{
		{"ascii unchanged", "/api/orders", 40, "/api/orders"},
		{"hebrew decoded", "/%D7%A2%D7%92%D7%9C%D7%94", 40, "/עגלה"},
		{"hebrew cut by runes, not bytes", "/%D7%9E%D7%95%D7%A6%D7%A8%D7%99%D7%9D/%D7%A0%D7%A2%D7%9C%D7%99%D7%99%D7%9D", 8, "/מוצרים…"},
		{"invalid escape kept as is", "/100%", 40, "/100%"},
		{"exact fit is not cut", "/abc", 4, "/abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(DisplayPath(tt.path), tt.width); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
