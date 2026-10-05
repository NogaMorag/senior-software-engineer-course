package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func clock() time.Time { return time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC) }

// TestSession runs one CLI session step by step against a single ledger file.
func TestSession(t *testing.T) {
	file := filepath.Join(t.TempDir(), "ledger.json")
	steps := []struct {
		args     string
		wantCode int
		wantOut  string // substring of stdout
		wantErr  string // substring of stderr
	}{
		{"open alice Alice", exitOK, "opened alice for Alice", ""},
		{"open alice Alice", exitError, "", "account already exists"},
		{"deposit alice 120 משכורת", exitOK, "alice balance 120.00", ""},
		{"withdraw alice 45.50 groceries", exitOK, "alice balance 74.50", ""},
		{"withdraw alice 100", exitInsufficient, "", "short by 25.50"},
		{"balance alice", exitOK, "74.50", ""},
		{"deposit alice 1", exitOK, "", ""},
		{"deposit alice 2", exitOK, "", ""},
		{"history -limit 2 alice", exitOK, "more: ledger history -cursor 2 -limit 2 alice", ""},
		{"history -cursor 2 -limit 2 alice", exitOK, "#3   2026-10-05 09:00  deposit          1.00", ""},
		{"history -cursor 9 alice", exitError, "", "bad cursor"},
		{"balance bob", exitNotFound, "", "account not found"},
		{"deposit alice 1.234", exitUsage, "", "at most two decimals"},
		{"fly alice", exitUsage, "", "unknown command"},
		{"", exitUsage, "", "usage:"},
	}
	for _, step := range steps {
		t.Run(step.args, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			args := append([]string{"-file", file}, strings.Fields(step.args)...)
			code := run(args, &stdout, &stderr, clock)
			if code != step.wantCode {
				t.Fatalf("exit code = %d, want %d\nstdout: %s\nstderr: %s", code, step.wantCode, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), step.wantOut) {
				t.Errorf("stdout missing %q; got:\n%s", step.wantOut, stdout.String())
			}
			if !strings.Contains(stderr.String(), step.wantErr) {
				t.Errorf("stderr missing %q; got:\n%s", step.wantErr, stderr.String())
			}
		})
	}
}

func TestHistoryShowsHebrewMemo(t *testing.T) {
	file := filepath.Join(t.TempDir(), "ledger.json")
	var out, errOut bytes.Buffer
	for _, args := range [][]string{{"open", "dana", "Dana"}, {"deposit", "dana", "10", "החזר הוצאות"}} {
		if code := run(append([]string{"-file", file}, args...), &out, &errOut, clock); code != exitOK {
			t.Fatalf("%v: exit %d: %s", args, code, errOut.String())
		}
	}
	out.Reset()
	if code := run([]string{"-file", file, "history", "dana"}, &out, &errOut, clock); code != exitOK {
		t.Fatalf("history: exit %d: %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "החזר הוצאות") {
		t.Errorf("history output lost the memo:\n%s", out.String())
	}
}
