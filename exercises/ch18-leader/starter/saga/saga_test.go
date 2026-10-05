package saga

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
)

// journal records every call so a test can check the exact order.
type journal []string

// steps builds an order saga whose step failAt fails (-1: none fails) and whose
// compensation of brokenUndo fails ("": none).
func (j *journal) steps(failAt int, brokenUndo string) []Step {
	names := []string{"reserve-stock", "charge-card", "book-courier", "send-email"}
	out := make([]Step, len(names))
	for i, name := range names {
		out[i] = Step{
			Name: name,
			Do: func(context.Context) error {
				*j = append(*j, "do "+name)
				if i == failAt {
					return fmt.Errorf("%s: injected failure", name)
				}
				return nil
			},
			Compensate: func(context.Context) error {
				*j = append(*j, "undo "+name)
				if name == brokenUndo {
					return errors.New("refund service down")
				}
				return nil
			},
		}
	}
	out[3].Compensate = nil // an email cannot be unsent; it is the last step for that reason
	return out
}

func TestCompensationRunsInReverseOrder(t *testing.T) {
	tests := []struct {
		name            string
		failAt          int
		wantJournal     []string
		wantCompensated []string
	}{
		{"no failure", -1,
			[]string{"do reserve-stock", "do charge-card", "do book-courier", "do send-email"}, nil},
		{"first step fails", 0,
			[]string{"do reserve-stock"}, nil},
		{"second step fails", 1,
			[]string{"do reserve-stock", "do charge-card", "undo reserve-stock"},
			[]string{"reserve-stock"}},
		{"third step fails", 2,
			[]string{"do reserve-stock", "do charge-card", "do book-courier", "undo charge-card", "undo reserve-stock"},
			[]string{"charge-card", "reserve-stock"}},
		{"last step fails", 3,
			[]string{"do reserve-stock", "do charge-card", "do book-courier", "do send-email",
				"undo book-courier", "undo charge-card", "undo reserve-stock"},
			[]string{"book-courier", "charge-card", "reserve-stock"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j journal
			err := Run(context.Background(), j.steps(tt.failAt, ""))
			if !slices.Equal(j, tt.wantJournal) {
				t.Errorf("calls:\n got %q\nwant %q", j, tt.wantJournal)
			}
			if tt.failAt < 0 {
				if err != nil {
					t.Fatalf("Run = %v, want nil", err)
				}
				return
			}
			var se *Error
			if !errors.As(err, &se) {
				t.Fatalf("Run = %v, want *saga.Error", err)
			}
			if !slices.Equal(se.Compensated, tt.wantCompensated) || len(se.Compensation) != 0 {
				t.Errorf("compensated %q (failures %v), want %q", se.Compensated, se.Compensation, tt.wantCompensated)
			}
		})
	}
}

func TestFailedCompensationIsReportedAndOthersStillRun(t *testing.T) {
	var j journal
	err := Run(context.Background(), j.steps(2, "charge-card"))
	want := []string{"do reserve-stock", "do charge-card", "do book-courier", "undo charge-card", "undo reserve-stock"}
	if !slices.Equal(j, want) {
		t.Errorf("calls:\n got %q\nwant %q", j, want)
	}
	var se *Error
	if !errors.As(err, &se) || se.Step != "book-courier" {
		t.Fatalf("Run = %v, want a failure at book-courier", err)
	}
	if !slices.Equal(se.Compensated, []string{"reserve-stock"}) || len(se.Compensation) != 1 {
		t.Errorf("compensated %q, failures %v; want reserve-stock compensated and one failure", se.Compensated, se.Compensation)
	}
	t.Log(err)
}

func TestCancelledCallerStillGetsCompensation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var j journal
	steps := j.steps(-1, "")
	charge := steps[1].Do
	steps[1].Do = func(ctx context.Context) error {
		err := charge(ctx)
		cancel() // the caller gives up while the card is being charged
		return err
	}
	steps[0].Compensate = func(ctx context.Context) error {
		if ctx.Err() != nil {
			return fmt.Errorf("compensation got a cancelled context: %w", ctx.Err())
		}
		j = append(j, "undo reserve-stock")
		return nil
	}
	err := Run(ctx, steps)
	want := []string{"do reserve-stock", "do charge-card", "undo charge-card", "undo reserve-stock"}
	if !slices.Equal(j, want) {
		t.Errorf("calls:\n got %q\nwant %q", j, want)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Run = %v, want it to wrap context.Canceled", err)
	}
}
