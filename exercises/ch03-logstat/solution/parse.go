package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Entry is one parsed access-log line in Common Log Format:
//
//	host ident user [time] "METHOD path PROTO" status bytes
type Entry struct {
	Host   string
	Time   time.Time
	Method string
	Path   string
	Status int
	Bytes  int64
}

// clfTime is Go's reference time (Mon Jan 2 15:04:05 MST 2006) written in the log's layout.
const clfTime = "02/Jan/2006:15:04:05 -0700"

// ParseLine parses one line. It never panics: every malformed input becomes an error.
func ParseLine(line string) (Entry, error) {
	var e Entry
	host, rest, ok := strings.Cut(line, " ")
	if !ok || host == "" {
		return Entry{}, fmt.Errorf("no host field")
	}
	e.Host = host

	_, rest, ok = strings.Cut(rest, "[")
	if !ok {
		return Entry{}, fmt.Errorf("no [time] field")
	}
	stamp, rest, ok := strings.Cut(rest, `] "`)
	if !ok {
		return Entry{}, fmt.Errorf("no request after time")
	}
	t, err := time.Parse(clfTime, stamp)
	if err != nil {
		return Entry{}, fmt.Errorf("bad time %q: %v", stamp, err)
	}
	e.Time = t

	request, rest, ok := strings.Cut(rest, `" `)
	if !ok {
		return Entry{}, fmt.Errorf("unterminated request")
	}
	parts := strings.Fields(request)
	if len(parts) != 3 {
		return Entry{}, fmt.Errorf("request %q: want 3 parts, got %d", request, len(parts))
	}
	e.Method, e.Path = parts[0], parts[1]

	fields := strings.Fields(rest)
	if len(fields) < 2 {
		return Entry{}, fmt.Errorf("missing status or bytes")
	}
	e.Status, err = strconv.Atoi(fields[0])
	if err != nil || e.Status < 100 || e.Status > 599 {
		return Entry{}, fmt.Errorf("bad status %q", fields[0])
	}
	if fields[1] != "-" { // "-" means no body was sent
		e.Bytes, err = strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return Entry{}, fmt.Errorf("bad bytes %q", fields[1])
		}
	}
	return e, nil
}

// StatusClass returns "2xx" for 200..299 and so on.
func StatusClass(status int) string {
	return fmt.Sprintf("%dxx", status/100)
}
