package main

import (
	"bytes"
	"strings"
	"testing"
)

// The storm table is deterministic: with a dead upstream every layer's count is exact.
func TestStormTable(t *testing.T) {
	var out bytes.Buffer
	if code := run([]string{"-scenario", "storm", "-n", "10"}, &out); code != 0 {
		t.Fatalf("exit code %d", code)
	}
	for _, want := range []string{
		"no retries                         10  10  10  10        1.0",
		"3 attempts at every layer          10  30  90  270       27.0",
		"3 attempts at the top only         10  30  30  30        3.0",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}

// Every scenario runs to the end at a small time scale.
func TestAllScenariosRun(t *testing.T) {
	if testing.Short() {
		t.Skip("timed scenarios")
	}
	var out bytes.Buffer
	if code := run([]string{"-n", "5", "-scale", "0.2"}, &out); code != 0 {
		t.Fatalf("exit code %d", code)
	}
	for _, want := range []string{"== storm", "== deadline", "== breaker", "== bulkhead", "== degrade"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q", want)
		}
	}
}
