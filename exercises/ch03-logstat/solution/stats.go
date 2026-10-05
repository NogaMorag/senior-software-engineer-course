package main

import (
	"bufio"
	"cmp"
	"io"
	"net/url"
	"slices"
	"strings"
	"unicode/utf8"
)

// maxBadLines is how many malformed line numbers a report keeps.
const maxBadLines = 5

// Stats is what logstat counts. Its zero value is not ready: use NewStats.
type Stats struct {
	Lines     int
	Malformed int
	BadLines  []int  // line numbers of the first maxBadLines malformed lines
	Classes   [6]int // Classes[2] counts 2xx, Classes[5] counts 5xx; index 0 is unused
	Paths     map[string]int
}

// NewStats returns a Stats with its map allocated.
func NewStats() *Stats {
	return &Stats{Paths: make(map[string]int)}
}

// Add counts one line. A malformed line is counted, never fatal.
func (s *Stats) Add(line string) {
	s.Lines++
	e, err := ParseLine(line)
	if err != nil {
		s.Malformed++
		if len(s.BadLines) < maxBadLines {
			s.BadLines = append(s.BadLines, s.Lines)
		}
		return
	}
	s.Classes[e.Status/100]++
	path, _, _ := strings.Cut(e.Path, "?") // count /a?x=1 and /a?x=2 as /a
	s.Paths[path]++
}

// Collect reads r line by line. Only a read error is returned; bad lines are counted.
func Collect(r io.Reader) (*Stats, error) {
	s := NewStats()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024) // allow lines up to 1 MiB
	for sc.Scan() {
		s.Add(sc.Text())
	}
	return s, sc.Err()
}

// PathCount is one row of the top-paths table.
type PathCount struct {
	Path  string
	Count int
}

// TopPaths returns the n most requested paths, most frequent first.
// Ties are broken by path so the output is the same on every run,
// even though map iteration order is not.
func TopPaths(paths map[string]int, n int) []PathCount {
	rows := make([]PathCount, 0, len(paths))
	for p, c := range paths {
		rows = append(rows, PathCount{Path: p, Count: c})
	}
	slices.SortFunc(rows, func(a, b PathCount) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return cmp.Compare(a.Path, b.Path)
	})
	return rows[:max(0, min(n, len(rows)))] // clamp n into 0..len(rows)
}

// DisplayPath decodes percent-encoding (%D7%A2 and friends) so Hebrew paths are readable.
// If the path is not valid percent-encoding it is returned unchanged.
func DisplayPath(p string) string {
	decoded, err := url.PathUnescape(p)
	if err != nil {
		return p
	}
	return decoded
}

// Truncate shortens s to at most limit runes (not bytes), marking the cut with "…".
func Truncate(s string, limit int) string {
	if limit < 1 {
		return ""
	}
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit-1]) + "…"
}
