package main

import (
	"testing"
	"time"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    Entry
		wantErr bool
	}{
		{
			name: "ordinary GET",
			line: `203.0.113.7 - - [05/Oct/2026:09:12:01 +0300] "GET /api/orders HTTP/1.1" 200 512`,
			want: Entry{
				Host:   "203.0.113.7",
				Time:   time.Date(2026, time.October, 5, 9, 12, 1, 0, time.FixedZone("", 3*60*60)),
				Method: "GET", Path: "/api/orders", Status: 200, Bytes: 512,
			},
		},
		{
			name: "dash means zero bytes",
			line: `198.51.100.23 - alice [05/Oct/2026:09:12:09 +0300] "GET /static/app.js HTTP/1.1" 304 -`,
			want: Entry{
				Host:   "198.51.100.23",
				Time:   time.Date(2026, time.October, 5, 9, 12, 9, 0, time.FixedZone("", 3*60*60)),
				Method: "GET", Path: "/static/app.js", Status: 304, Bytes: 0,
			},
		},
		{
			name: "combined format extra fields are ignored",
			line: `192.0.2.44 - - [05/Oct/2026:09:15:30 +0300] "GET /healthz HTTP/1.1" 200 2 "-" "curl/8.5.0"`,
			want: Entry{
				Host:   "192.0.2.44",
				Time:   time.Date(2026, time.October, 5, 9, 15, 30, 0, time.FixedZone("", 3*60*60)),
				Method: "GET", Path: "/healthz", Status: 200, Bytes: 2,
			},
		},
		{name: "empty line", line: "", wantErr: true},
		{name: "free text", line: "this line is not an access log line", wantErr: true},
		{name: "status not a number", line: `203.0.113.7 - - [05/Oct/2026:09:16:00 +0300] "GET / HTTP/1.1" abc 512`, wantErr: true},
		{name: "status out of range", line: `203.0.113.7 - - [05/Oct/2026:09:16:00 +0300] "GET / HTTP/1.1" 999 512`, wantErr: true},
		{name: "bad time", line: `203.0.113.7 - - [yesterday] "GET / HTTP/1.1" 200 512`, wantErr: true},
		{name: "request with two parts", line: `203.0.113.7 - - [05/Oct/2026:09:16:00 +0300] "GET /" 200 512`, wantErr: true},
		{name: "missing bytes", line: `203.0.113.7 - - [05/Oct/2026:09:16:00 +0300] "GET / HTTP/1.1" 200`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLine(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseLine(%q) = %+v, want an error", tt.line, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseLine(%q) error: %v", tt.line, err)
			}
			if !got.Time.Equal(tt.want.Time) {
				t.Errorf("Time = %v, want %v", got.Time, tt.want.Time)
			}
			got.Time, tt.want.Time = time.Time{}, time.Time{} // compared above with Equal
			if got != tt.want {
				t.Errorf("ParseLine(%q)\n got %+v\nwant %+v", tt.line, got, tt.want)
			}
		})
	}
}

func TestStatusClass(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{100, "1xx"}, {200, "2xx"}, {204, "2xx"}, {301, "3xx"}, {404, "4xx"}, {599, "5xx"},
	}
	for _, tt := range tests {
		if got := StatusClass(tt.status); got != tt.want {
			t.Errorf("StatusClass(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}
