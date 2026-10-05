// Package saga runs an orchestrated saga: a sequence of local steps, each committed on its
// own, where a failure undoes the completed steps by running their compensating actions in
// reverse order. This is the in-process core; a production orchestrator also persists
// which steps finished, so that a crash in the middle resumes or compensates on restart.
package saga

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Step is one local transaction and the action that semantically undoes it.
type Step struct {
	Name string
	Do   func(ctx context.Context) error
	// Compensate undoes a completed Do. It must be safe to run more than once, because a
	// real orchestrator retries it after a crash. Nil means there is nothing to undo.
	Compensate func(ctx context.Context) error
}

// Error reports a saga that failed and was (perhaps partly) compensated.
type Error struct {
	Step         string   // the step that failed
	Err          error    // why it failed
	Compensated  []string // steps successfully compensated, in the order they ran
	Compensation []error  // compensations that failed: these need a human or a retry
}

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "saga: step %q failed: %v; compensated %v", e.Step, e.Err, e.Compensated)
	if len(e.Compensation) > 0 {
		fmt.Fprintf(&b, "; %d compensations failed: %v", len(e.Compensation), errors.Join(e.Compensation...))
	}
	return b.String()
}

func (e *Error) Unwrap() error { return e.Err }

// Run runs the steps in order. If a step fails, or ctx ends before a step starts, Run
// compensates every completed step in reverse order and returns an *Error. Compensations
// run with a context that is not cancelled with ctx: a caller who gives up must not leave
// the saga half done.
func Run(ctx context.Context, steps []Step) error {
	for i, s := range steps {
		err := ctx.Err()
		if err == nil {
			err = s.Do(ctx)
		}
		if err != nil {
			return compensate(context.WithoutCancel(ctx), steps[:i], &Error{Step: s.Name, Err: err})
		}
	}
	return nil
}

func compensate(ctx context.Context, done []Step, sagaErr *Error) error {
	for i := len(done) - 1; i >= 0; i-- {
		s := done[i]
		if s.Compensate == nil {
			continue
		}
		if err := s.Compensate(ctx); err != nil {
			// keep going: the other steps still need undoing
			sagaErr.Compensation = append(sagaErr.Compensation, fmt.Errorf("compensate %q: %w", s.Name, err))
			continue
		}
		sagaErr.Compensated = append(sagaErr.Compensated, s.Name)
	}
	return sagaErr
}
