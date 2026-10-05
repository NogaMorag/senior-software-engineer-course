package platform_test

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"example.com/ch24-telemetry/internal/platform"
)

func lookupFrom(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestEnvDefaultsAndValues(t *testing.T) {
	env := platform.NewEnv(lookupFrom(map[string]string{
		"S":     "hello",
		"EMPTY": "", // an empty value counts as not set
		"N":     "42",
		"D":     "250ms",
		"L":     "warn",
	}))
	if got := env.String("S", "x"); got != "hello" {
		t.Errorf("String = %q", got)
	}
	if got := env.String("EMPTY", "def"); got != "def" {
		t.Errorf("String(empty) = %q, want the default", got)
	}
	if got := env.Int64("N", 1); got != 42 {
		t.Errorf("Int64 = %d", got)
	}
	if got := env.Duration("D", time.Second); got != 250*time.Millisecond {
		t.Errorf("Duration = %v", got)
	}
	if got := env.Level("L", slog.LevelInfo); got != slog.LevelWarn {
		t.Errorf("Level = %v", got)
	}
	if got := env.Duration("MISSING", 3*time.Second); got != 3*time.Second {
		t.Errorf("Duration(missing) = %v", got)
	}
	if err := env.Err(); err != nil {
		t.Errorf("Err = %v, want nil", err)
	}
}

func TestEnvCollectsEveryProblem(t *testing.T) {
	env := platform.NewEnv(lookupFrom(map[string]string{
		"N": "many",
		"D": "ten",
		"L": "loud",
	}))
	env.Required("NEEDED")
	env.Int64("N", 1)
	env.Duration("D", time.Second)
	env.Level("L", slog.LevelInfo)

	err := env.Err()
	if err == nil {
		t.Fatal("Err = nil, want four problems")
	}
	for _, want := range []string{
		"NEEDED: required, not set",
		`N: "many" is not an integer`,
		`D: "ten" is not a duration`,
		`L: "loud" is not a log level`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}
