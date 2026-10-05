package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	const sample = `203.0.113.7 - - [05/Oct/2026:09:12:01 +0300] "GET /api/orders HTTP/1.1" 200 512
not a log line
192.0.2.44 - - [05/Oct/2026:09:16:20 +0300] "GET /%D7%A2%D7%92%D7%9C%D7%94 HTTP/1.1" 302 0
`
	tests := []struct {
		name     string
		args     []string
		stdin    string
		wantCode int
		wantOut  []string // substrings expected on stdout
		wantErr  string   // substring expected on stderr
	}{
		{
			name: "stdin", args: nil, stdin: sample, wantCode: 0,
			wantOut: []string{"lines      3", "malformed  1 (first at lines [2])", "2xx        1", "3xx        1", "    1  /עגלה"},
		},
		{name: "dash is stdin", args: []string{"-"}, stdin: sample, wantCode: 0, wantOut: []string{"lines      3"}},
		{name: "sample file", args: []string{"-top", "1", "testdata/access.log"}, wantCode: 0, wantOut: []string{"lines      15", "    5  /api/orders"}},
		{name: "missing file", args: []string{"no-such-file.log"}, wantCode: 1, wantErr: "no-such-file.log"},
		{name: "two files", args: []string{"a", "b"}, wantCode: 2, wantErr: "usage"},
		{name: "unknown flag", args: []string{"-nope"}, wantCode: 2, wantErr: "-nope"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, strings.NewReader(tt.stdin), &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d; stderr: %s", code, tt.wantCode, stderr.String())
			}
			for _, want := range tt.wantOut {
				if !strings.Contains(stdout.String(), want) {
					t.Errorf("stdout missing %q; got:\n%s", want, stdout.String())
				}
			}
			if !strings.Contains(stderr.String(), tt.wantErr) {
				t.Errorf("stderr missing %q; got: %s", tt.wantErr, stderr.String())
			}
		})
	}
}
