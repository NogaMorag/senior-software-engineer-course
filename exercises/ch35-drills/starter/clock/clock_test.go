package clock

import (
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

func fired(tm Timer) bool {
	select {
	case <-tm.C():
		return true
	default:
		return false
	}
}

func TestFakeTimersFireInOrderWhenDue(t *testing.T) {
	f := NewFake(t0)
	late, early := f.NewTimer(2*time.Second), f.NewTimer(time.Second)
	if f.Waiters() != 2 {
		t.Fatalf("Waiters = %d, want 2", f.Waiters())
	}
	f.Advance(999 * time.Millisecond)
	if fired(early) || fired(late) {
		t.Fatal("a timer fired early")
	}
	f.Advance(time.Millisecond)
	if !fired(early) || fired(late) {
		t.Fatal("after 1s: want only the 1s timer to have fired")
	}
	f.Advance(time.Second)
	if !fired(late) {
		t.Fatal("after 2s: the 2s timer did not fire")
	}
	if !f.Now().Equal(t0.Add(2 * time.Second)) {
		t.Fatalf("Now = %v, want t0+2s", f.Now())
	}
}

func TestFakeStopAndZeroDuration(t *testing.T) {
	f := NewFake(t0)
	tm := f.NewTimer(time.Second)
	if !tm.Stop() || tm.Stop() {
		t.Fatal("Stop must report true once, then false")
	}
	f.Advance(time.Hour)
	if fired(tm) {
		t.Fatal("a stopped timer fired")
	}
	if !fired(f.NewTimer(0)) {
		t.Fatal("a timer of 0 must have fired already")
	}
}

func TestBlockUntil(t *testing.T) {
	f := NewFake(t0)
	go func() {
		time.Sleep(10 * time.Millisecond)
		f.NewTimer(time.Minute)
	}()
	f.BlockUntil(1) // returns once the goroutine has created its timer
	if f.Waiters() != 1 {
		t.Fatalf("Waiters = %d, want 1", f.Waiters())
	}
}
