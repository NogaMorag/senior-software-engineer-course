package replica

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

// pair connects to the chapter's primary and replica, or skips the test.
// The tests in this package change recovery_min_apply_delay on the replica, so they never
// run in parallel with each other.
func pair(t *testing.T) Pair {
	t.Helper()
	primaryURL, replicaURL := os.Getenv("DATABASE_URL"), os.Getenv("REPLICA_URL")
	if primaryURL == "" || replicaURL == "" {
		t.Skip("integration test: set DATABASE_URL and REPLICA_URL to the chapter 17 servers; see README")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	p, err := Connect(ctx, primaryURL, replicaURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	if err := p.Setup(ctx); err != nil {
		t.Fatal(err)
	}
	return p
}

// withApplyDelay makes the replica apply commits d late for the rest of the test.
func withApplyDelay(t *testing.T, p Pair, d time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := p.SetApplyDelay(ctx, d); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := p.SetApplyDelay(ctx, 0); err != nil {
			t.Error(err)
			return
		}
		lsn, err := p.CurrentLSN(ctx) // leave the replica caught up for the next test
		if err == nil {
			err = p.WaitForLSN(ctx, lsn, 10*time.Millisecond)
		}
		if err != nil {
			t.Error(err)
		}
	})
}

func TestSummarize(t *testing.T) {
	ms := func(n int) time.Duration { return time.Duration(n) * time.Millisecond }
	obs := func(polls int, lags ...int) []Observation {
		out := make([]Observation, len(lags))
		for i, l := range lags {
			out[i] = Observation{ID: int64(i), Lag: ms(l), Polls: polls}
		}
		return out
	}
	tests := []struct {
		name string
		in   []Observation
		want Summary
	}{
		{"empty", nil, Summary{}},
		{"one", obs(1, 7), Summary{N: 1, Min: ms(7), Median: ms(7), P95: ms(7), Max: ms(7)}},
		{"unsorted input", obs(1, 5, 1, 4, 2, 3), Summary{N: 5, Min: ms(1), Median: ms(3), P95: ms(5), Max: ms(5)}},
		{"twenty, nearest rank", obs(2, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1),
			Summary{N: 20, Min: ms(1), Median: ms(10), P95: ms(19), Max: ms(20), MissedFirstRead: 20}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Summarize(tt.in); got != tt.want {
				t.Errorf("Summarize = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestProbeSeesItsWriteOnTheReplica(t *testing.T) {
	p := pair(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var obs []Observation
	for i := range 20 {
		o, err := p.Probe(ctx, fmt.Sprintf("probe %d", i), time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		obs = append(obs, o)
	}
	s := Summarize(obs)
	t.Logf("observed lag over %d probes: min %v, median %v, p95 %v, max %v; first read missed %d times",
		s.N, s.Min, s.Median, s.P95, s.Max, s.MissedFirstRead)
	if s.Max > 5*time.Second {
		t.Errorf("max observed lag %v on an idle local replica; is it streaming?", s.Max)
	}
}

func TestApplyDelayMakesLagVisible(t *testing.T) {
	p := pair(t)
	withApplyDelay(t, p, time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	o, err := p.Probe(ctx, "delayed", 10*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("with recovery_min_apply_delay = 1s: lag %v after %d polls", o.Lag, o.Polls)
	if o.Polls < 2 || o.Lag < 800*time.Millisecond {
		t.Errorf("lag %v after %d polls, want about 1s and several polls", o.Lag, o.Polls)
	}
}

func TestReadYourWrites(t *testing.T) {
	p := pair(t)
	withApplyDelay(t, p, time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	id, token, err := p.WriteNote(ctx, "my new address")
	if err != nil {
		t.Fatal(err)
	}
	// The naive read: straight to the replica, which has not applied the commit yet.
	if _, err := p.ReadNoteFromReplica(ctx, id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("replica read right after the write: err = %v, want ErrNotFound (the stale read)", err)
	}
	// Willing to wait longer than the lag: the replica answers once it has the token.
	start := time.Now()
	note, from, err := p.ReadNote(ctx, id, token, 3*time.Second)
	if err != nil || note != "my new address" || from != FromReplica {
		t.Fatalf("ReadNote = %q, %s, %v; want the note from the replica", note, from, err)
	}
	t.Logf("token %s: replica caught up after %v", token, time.Since(start))

	// Not willing to wait: the read falls back to the primary and still sees the write.
	id2, token2, err := p.WriteNote(ctx, "my newer address")
	if err != nil {
		t.Fatal(err)
	}
	note, from, err = p.ReadNote(ctx, id2, token2, 100*time.Millisecond)
	if err != nil || note != "my newer address" || from != FromPrimary {
		t.Fatalf("ReadNote = %q, %s, %v; want the note from the primary", note, from, err)
	}
}

func TestSendersShowsTheStreamingReplica(t *testing.T) {
	p := pair(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	senders, err := p.Senders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(senders) != 1 || senders[0].State != "streaming" || senders[0].SyncState != "async" {
		t.Fatalf("pg_stat_replication = %+v, want one async streaming standby", senders)
	}
	age, err := p.ReplayAge(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("sender %+v; replay age %v", senders[0], age)
}

func TestLoadWritesUntilCancelled(t *testing.T) {
	p := pair(t)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	rows, err := p.Load(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if rows == 0 {
		t.Fatal("Load wrote no rows in 300ms")
	}
	t.Logf("load: %d rows in 300ms with 2 writers", rows)
}
