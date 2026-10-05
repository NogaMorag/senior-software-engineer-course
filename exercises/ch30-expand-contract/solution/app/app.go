// Package app is the application whose schema changes underneath it: a tiny users service
// with a writer and a reader loop. The Phase it runs in decides which columns its SQL
// touches; one Phase is one deployed version of the code.
package app

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Phase is the version of the application code that is deployed.
type Phase int32

const (
	PhaseOld       Phase = iota + 1 // v1: writes and reads fullname
	PhaseDualWrite                  // v2: writes both columns, reads fullname
	PhaseReadNew                    // v3: writes both columns, reads display_name
	PhaseNewOnly                    // v4: writes and reads display_name only
)

func (p Phase) String() string {
	switch p {
	case PhaseOld:
		return "v1 old"
	case PhaseDualWrite:
		return "v2 dual-write"
	case PhaseReadNew:
		return "v3 read-new"
	case PhaseNewOnly:
		return "v4 new-only"
	}
	return fmt.Sprintf("Phase(%d)", int32(p))
}

// insertSQL takes the name as $1 and returns the new id.
func insertSQL(p Phase) string {
	switch p {
	case PhaseOld:
		return `INSERT INTO users (fullname) VALUES ($1) RETURNING id`
	case PhaseDualWrite, PhaseReadNew:
		return `INSERT INTO users (fullname, display_name) VALUES ($1, $1) RETURNING id`
	case PhaseNewOnly:
		return `INSERT INTO users (display_name) VALUES ($1) RETURNING id`
	}
	panic("unknown phase " + p.String())
}

// renameSQL takes the id as $1 and the new name as $2.
func renameSQL(p Phase) string {
	switch p {
	case PhaseOld:
		return `UPDATE users SET fullname = $2 WHERE id = $1`
	case PhaseDualWrite, PhaseReadNew:
		return `UPDATE users SET fullname = $2, display_name = $2 WHERE id = $1`
	case PhaseNewOnly:
		return `UPDATE users SET display_name = $2 WHERE id = $1`
	}
	panic("unknown phase " + p.String())
}

// readSQL takes the id as $1 and returns the name.
func readSQL(p Phase) string {
	switch p {
	case PhaseOld, PhaseDualWrite:
		return `SELECT fullname FROM users WHERE id = $1`
	case PhaseReadNew, PhaseNewOnly:
		return `SELECT display_name FROM users WHERE id = $1`
	}
	panic("unknown phase " + p.String())
}

// Store runs the SQL of a given phase.
type Store struct{ Pool *pgxpool.Pool }

// Insert creates a user and returns its id.
func (s Store) Insert(ctx context.Context, p Phase, name string) (int64, error) {
	var id int64
	err := s.Pool.QueryRow(ctx, insertSQL(p), name).Scan(&id)
	return id, err
}

// Rename changes the name of user id.
func (s Store) Rename(ctx context.Context, p Phase, id int64, name string) error {
	tag, err := s.Pool.Exec(ctx, renameSQL(p), id, name)
	if err == nil && tag.RowsAffected() != 1 {
		err = fmt.Errorf("rename %d: %d rows affected", id, tag.RowsAffected())
	}
	return err
}

// Name reads the name of user id. A NULL fails the scan, so a reader that switched to the
// new column before the backfill finished counts as a failed request.
func (s Store) Name(ctx context.Context, p Phase, id int64) (string, error) {
	var name string
	err := s.Pool.QueryRow(ctx, readSQL(p), id).Scan(&name)
	if err == nil && name == "" {
		err = errors.New("empty name")
	}
	return name, err
}

// Registry remembers every user id and the last name successfully written for it.
type Registry struct {
	mu    sync.Mutex
	ids   []int64
	names map[int64]string
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{names: map[int64]string{}} }

// Set records the latest name of id.
func (r *Registry) Set(id int64, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.names[id]; !ok {
		r.ids = append(r.ids, id)
	}
	r.names[id] = name
}

// Random returns a known id, or false when there is none yet.
func (r *Registry) Random() (int64, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.ids) == 0 {
		return 0, false
	}
	return r.ids[rand.IntN(len(r.ids))], true
}

// Expected returns a copy of id to last written name.
func (r *Registry) Expected() map[int64]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[int64]string, len(r.names))
	for id, n := range r.names {
		out[id] = n
	}
	return out
}

// Instance is one running copy of the service: a writer loop and a reader loop that share
// a phase. Upgrading an instance is SetPhase; a rolling deploy upgrades instances one by one.
type Instance struct {
	ID    int
	Store Store
	Reg   *Registry
	Pause time.Duration // between operations of each loop, to keep the load small

	phase    atomic.Int32
	ops      atomic.Int64
	failures atomic.Int64
	mu       sync.Mutex
	errs     []string
	own      []int64 // ids this instance inserted; only the writer loop touches it
}

// NewInstance returns an instance in phase p.
func NewInstance(id int, store Store, reg *Registry, p Phase) *Instance {
	in := &Instance{ID: id, Store: store, Reg: reg, Pause: 2 * time.Millisecond}
	in.phase.Store(int32(p))
	return in
}

// Phase returns the deployed version.
func (in *Instance) Phase() Phase { return Phase(in.phase.Load()) }

// SetPhase deploys another version of the code to this instance.
func (in *Instance) SetPhase(p Phase) { in.phase.Store(int32(p)) }

// Ops is the number of finished requests, successful or not.
func (in *Instance) Ops() int64 { return in.ops.Load() }

// Failures is the number of failed requests.
func (in *Instance) Failures() int64 { return in.failures.Load() }

// Errors returns the first few error messages, for the test log.
func (in *Instance) Errors() []string {
	in.mu.Lock()
	defer in.mu.Unlock()
	return append([]string(nil), in.errs...)
}

// Run starts both loops and returns when ctx is canceled and both have stopped.
func (in *Instance) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { in.loop(ctx, in.write) })
	wg.Go(func() { in.loop(ctx, in.read) })
	wg.Wait()
}

// loop runs op until ctx is canceled. A request in flight finishes first: its context is
// detached from ctx, so stopping never cuts a write short after the database committed it.
func (in *Instance) loop(ctx context.Context, op func(context.Context, Phase) error) {
	for ctx.Err() == nil {
		opCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		err := op(opCtx, in.Phase())
		cancel()
		in.ops.Add(1)
		if err != nil {
			in.fail(err)
		}
		time.Sleep(in.Pause)
	}
}

func (in *Instance) fail(err error) {
	in.failures.Add(1)
	in.mu.Lock()
	defer in.mu.Unlock()
	if len(in.errs) < 5 {
		in.errs = append(in.errs, err.Error())
	}
}

// write inserts a user or renames one this instance created. Each id has a single writer,
// so the registry's last name is the database's last name.
func (in *Instance) write(ctx context.Context, p Phase) error {
	if len(in.own) == 0 || rand.IntN(2) == 0 {
		name := fmt.Sprintf("user %d-%d", in.ID, len(in.own))
		id, err := in.Store.Insert(ctx, p, name)
		if err != nil {
			return fmt.Errorf("%v insert: %w", p, err)
		}
		in.own = append(in.own, id)
		in.Reg.Set(id, name)
		return nil
	}
	id := in.own[rand.IntN(len(in.own))]
	name := fmt.Sprintf("user %d renamed at %d", id, time.Now().UnixNano())
	if err := in.Store.Rename(ctx, p, id, name); err != nil {
		return fmt.Errorf("%v rename: %w", p, err)
	}
	in.Reg.Set(id, name)
	return nil
}

func (in *Instance) read(ctx context.Context, p Phase) error {
	id, ok := in.Reg.Random()
	if !ok {
		return nil
	}
	if _, err := in.Store.Name(ctx, p, id); err != nil {
		return fmt.Errorf("%v read %d: %w", p, id, err)
	}
	return nil
}
