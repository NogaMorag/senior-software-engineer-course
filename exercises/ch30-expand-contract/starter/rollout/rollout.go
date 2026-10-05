// Package rollout runs the expand and contract plan against a running fleet: one migration,
// then a rolling deploy of the next phase, then the next migration.
package rollout

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/ch30-expand-contract/app"
	"example.com/ch30-expand-contract/db"
)

// Step is one migration followed by a deploy. Deploy 0 means no deploy after the migration.
type Step struct {
	Name    string
	Migrate int64
	Deploy  app.Phase
}

// Plan renames users.fullname to users.display_name. The order is the whole point: every
// migration must work with the code that is running when it is applied, and every deploy
// must work with the schema that exists when it starts.
var Plan = []Step{
	{"expand: add display_name", 2, app.PhaseDualWrite},
	// TODO: three more steps, one per migration you write (00003, 00004, 00005). For each,
	// decide which phase can be deployed after it; the last migration has no deploy (0).
}

// Report is what happened while one step ran.
type Report struct {
	Step     string
	Ops      int64
	Failures int64
}

// Run applies plan to the fleet. After each migration and after upgrading each instance it
// waits until every instance has finished settle more requests, so every schema version
// meets every code version that runs next to it.
func Run(ctx context.Context, migrations *pgxpool.Pool, fleet []*app.Instance, plan []Step, settle int64) ([]Report, error) {
	var reports []Report
	for _, s := range plan {
		ops, fails := totals(fleet)
		if err := db.MigrateTo(ctx, migrations, s.Migrate); err != nil {
			return reports, fmt.Errorf("%s: %w", s.Name, err)
		}
		if err := waitForOps(ctx, fleet, settle); err != nil {
			return reports, fmt.Errorf("%s: %w", s.Name, err)
		}
		if s.Deploy != 0 {
			for _, in := range fleet { // a rolling deploy: one instance at a time
				in.SetPhase(s.Deploy)
				if err := waitForOps(ctx, fleet, settle); err != nil {
					return reports, fmt.Errorf("%s: deploy %v: %w", s.Name, s.Deploy, err)
				}
			}
		}
		ops2, fails2 := totals(fleet)
		reports = append(reports, Report{Step: s.Name, Ops: ops2 - ops, Failures: fails2 - fails})
	}
	return reports, nil
}

func totals(fleet []*app.Instance) (ops, failures int64) {
	for _, in := range fleet {
		ops += in.Ops()
		failures += in.Failures()
	}
	return ops, failures
}

// waitForOps returns when every instance has finished n more requests than when it was called.
func waitForOps(ctx context.Context, fleet []*app.Instance, n int64) error {
	start := make([]int64, len(fleet))
	for i, in := range fleet {
		start[i] = in.Ops()
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for i, in := range fleet {
		for in.Ops() < start[i]+n {
			select {
			case <-ctx.Done():
				return fmt.Errorf("instance %d made no progress: %w", in.ID, ctx.Err())
			case <-time.After(5 * time.Millisecond):
			}
		}
	}
	return nil
}
