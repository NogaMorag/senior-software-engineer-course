package replica

import (
	"testing"
	"time"
)

func TestRouterSticksToPrimaryAfterAWrite(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	r := &Router{Window: 2 * time.Second, Now: func() time.Time { return now }}

	steps := []struct {
		advance time.Duration
		write   string // user who writes at this step, if any
		reader  string
		want    Source
	}{
		{0, "", "dana", FromReplica},     // never wrote
		{0, "dana", "dana", FromPrimary}, // just wrote
		{0, "", "noa", FromReplica},      // someone else's write does not matter
		{1999 * time.Millisecond, "", "dana", FromPrimary},
		{time.Millisecond, "", "dana", FromReplica}, // window over
		{0, "dana", "dana", FromPrimary},            // a new write opens a new window
	}
	for i, s := range steps {
		now = now.Add(s.advance)
		if s.write != "" {
			r.RecordWrite(s.write)
		}
		if got := r.ForRead(s.reader); got != s.want {
			t.Errorf("step %d: ForRead(%q) = %s, want %s", i, s.reader, got, s.want)
		}
	}
}
