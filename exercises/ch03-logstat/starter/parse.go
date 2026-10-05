package main

import (
	"fmt"
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
	// TODO(day 4): split the line with strings.Cut and strings.Fields, in this order:
	//   1. host: everything before the first space
	//   2. time: between "[" and `] "`; parse it with time.Parse(clfTime, ...)
	//   3. request: up to the closing `" `; exactly three parts: method, path, protocol
	//   4. status (100..599, strconv.Atoi) and bytes ("-" means 0, strconv.ParseInt)
	// Return a descriptive error for every malformed input; never panic.
	return Entry{}, fmt.Errorf("TODO: ParseLine is not implemented yet")
}

// StatusClass returns "2xx" for 200..299 and so on.
func StatusClass(status int) string {
	return fmt.Sprintf("%dxx", status/100)
}
