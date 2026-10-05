// Package db opens the connection pool and applies the embedded migrations (from chapter 11).
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"example.com/hookline/migrations"
)

// Config parses a URL such as postgres://hookline:change-me@localhost:5438/hookline and sets
// explicit pool limits instead of relying on defaults. Callers may adjust the result.
func Config(url string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	// Upper bound on connections this process opens. Every replica of a service has
	// its own pool, so replicas times MaxConns must stay below the server's
	// max_connections, with room left for migrations, psql and monitoring.
	cfg.MaxConns = 10
	cfg.MinConns = 0
	cfg.MaxConnLifetime = time.Hour       // recycle, so server-side memory does not grow forever
	cfg.MaxConnIdleTime = 5 * time.Minute // give idle connections back to the server
	cfg.HealthCheckPeriod = 30 * time.Second
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	return cfg, nil
}

// Connect opens the pool and checks that the database answers.
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

// withMigrator runs fn with a goose provider over the embedded migrations. The provider
// takes a PostgreSQL advisory lock for the duration, so two processes that start at
// the same moment do not both run the same migration.
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

// Migrate applies every pending migration (goose up).
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return withMigrator(pool, func(p *goose.Provider) error {
		if _, err := p.Up(ctx); err != nil {
			return fmt.Errorf("migrate up: %w", err)
		}
		return nil
	})
}

// MigrateDownTo rolls back every migration above version; 0 means all of them.
func MigrateDownTo(ctx context.Context, pool *pgxpool.Pool, version int64) error {
	return withMigrator(pool, func(p *goose.Provider) error {
		if _, err := p.DownTo(ctx, version); err != nil {
			return fmt.Errorf("migrate down to %d: %w", version, err)
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
