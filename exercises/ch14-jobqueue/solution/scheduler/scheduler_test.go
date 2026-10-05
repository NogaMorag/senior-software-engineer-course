package scheduler_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"example.com/ch14-jobqueue/scheduler"
	"example.com/ch14-jobqueue/testenv"
)

func TestSlot(t *testing.T) {
	at := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	tests := []struct {
		now   string
		every time.Duration
		want  string
	}{
		{"2026-10-05T02:00:00Z", time.Hour, "2026-10-05T02:00:00Z"},
		{"2026-10-05T02:59:59.999Z", time.Hour, "2026-10-05T02:00:00Z"},
		{"2026-10-05T03:00:00Z", time.Hour, "2026-10-05T03:00:00Z"},
		{"2026-10-05T05:30:00+03:00", time.Hour, "2026-10-05T02:00:00Z"}, // same instant as 02:30Z
		{"2026-10-05T02:07:00Z", 5 * time.Minute, "2026-10-05T02:05:00Z"},
	}
	for _, tt := range tests {
		if got := scheduler.Slot(at(tt.now), tt.every); !got.Equal(at(tt.want)) {
			t.Errorf("Slot(%s, %v) = %s, want %s", tt.now, tt.every, got.Format(time.RFC3339), tt.want)
		}
	}
}

// Two replicas tick in the same hour, a few seconds apart: one job, not two. The next
// hour creates the next job.
func TestTwoReplicasSameSlotEnqueueOnce(t *testing.T) {
	db := testenv.Postgres(t)
	ctx := context.Background()
	base := time.Date(2026, 10, 5, 2, 0, 0, 0, time.UTC)
	ticks := []time.Time{base.Add(time.Second), base.Add(3 * time.Second), base.Add(time.Hour + time.Second)}
	var created []int64
	for _, now := range ticks {
		id, ok, err := scheduler.EnqueueDue(ctx, db, "nightly-report", "report", nil, time.Hour, now)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			created = append(created, id)
		}
	}
	if len(created) != 2 {
		t.Errorf("created %d jobs for 2 slots and 3 ticks, want 2", len(created))
	}
}

func TestConcurrentSchedulersCreateOneJobPerSlot(t *testing.T) {
	db := testenv.Postgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 700*time.Millisecond)
	defer cancel()
	var wg sync.WaitGroup
	for range 3 { // three replicas, each with its own in-process ticker
		wg.Go(func() {
			if err := scheduler.Run(ctx, db, "heartbeat", "beat", nil, 100*time.Millisecond); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	var jobs, slots int
	if err := db.QueryRow(context.Background(), `
		SELECT count(*), count(DISTINCT unique_key) FROM jobs WHERE kind = 'beat'`).Scan(&jobs, &slots); err != nil {
		t.Fatal(err)
	}
	if jobs == 0 || jobs != slots || jobs > 9 {
		t.Errorf("%d jobs for %d slots in 0.7 s at 0.1 s per slot; want one job per slot", jobs, slots)
	}
}
