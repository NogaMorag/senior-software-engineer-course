// Package db opens connection pools and applies the embedded migrations one version at a time.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"example.com/ch30-expand-contract/migrations"
)

// LockTimeout bounds how long one migration statement may wait for a lock. A DDL statement
// that waits in the lock queue blocks every query that arrives after it, so failing fast and
// retrying later is better than queueing behind a long transaction.
const LockTimeout = 2 * time.Second

// AppConfig is the pool the application instances use.
func AppConfig(url string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 8
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	return cfg, nil
}

// MigrationConfig is a separate, small pool for goose. lock_timeout is a startup parameter of
// every connection in it, so it applies to every statement of every migration, including
// NO TRANSACTION ones, without each file having to remember a SET (which would also outlive
// the file on that session).
func MigrationConfig(url string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 2
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	cfg.ConnConfig.RuntimeParams["lock_timeout"] = fmt.Sprint(LockTimeout.Milliseconds())
	cfg.ConnConfig.RuntimeParams["application_name"] = "migrator"
	return cfg, nil
}

// Connect opens a pool and checks that the database answers.
func Connect(ctx context.Context, cfg *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func withMigrator(pool *pgxpool.Pool, fn func(*goose.Provider) error) error {
	sqlDB := stdlib.OpenDBFromPool(pool) // closing it does not close the pool
	defer sqlDB.Close()
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS, goose.WithSessionLocker(locker))
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}
	return fn(provider)
}

// MigrateTo applies pending migrations up to and including version (goose up-to). In
// production each version is its own deploy; the rollout test calls it once per step.
func MigrateTo(ctx context.Context, pool *pgxpool.Pool, version int64) error {
	return withMigrator(pool, func(p *goose.Provider) error {
		if _, err := p.UpTo(ctx, version); err != nil {
			return fmt.Errorf("migrate up to %d: %w", version, err)
		}
		return nil
	})
}

// Version returns the current schema version (0 for an empty database).
func Version(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	var v int64
	err := withMigrator(pool, func(p *goose.Provider) error {
		var err error
		v, err = p.GetDBVersion(ctx)
		return err
	})
	return v, err
}
