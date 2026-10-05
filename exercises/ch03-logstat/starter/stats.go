package main

import (
	"bufio"
	"io"
	"net/url"
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
	// TODO(day 4): parse the line with ParseLine and count its status class in s.Classes.
	// TODO(day 5): on a parse error, count it in s.Malformed and keep the line number
	// (s.Lines) in s.BadLines, at most maxBadLines of them; then return.
	// TODO(day 5): count the path without its query string in s.Paths.
}

// Collect reads r line by line. Only a read error is returned; bad lines are counted.
func Collect(r io.Reader) (*Stats, error) {
	s := NewStats()
	sc := bufio.NewScanner(r)
	// TODO(day 5): bufio.Scanner stops at 64 KiB per line by default. Raise the limit
	// to 1 MiB with sc.Buffer so TestCollectLongLine passes.
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
	// TODO(day 5): copy the map into a []PathCount, sort it with slices.SortFunc
	// (count descending, then path ascending, using cmp.Compare), and return
	// at most n rows. A negative n returns no rows.
	return nil
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
	// TODO(day 5): count runes with utf8.RuneCountInString; if the string is longer
	// than limit, keep limit-1 runes (convert to []rune) and add "…". limit < 1 returns "".
	return s
}
