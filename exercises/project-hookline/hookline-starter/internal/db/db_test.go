package db_test

import (
	"context"
	"testing"
	"time"

	"example.com/hookline/internal/db"
)

func TestConfigSetsExplicitPoolLimits(t *testing.T) {
	cfg, err := db.Config("postgres://hookline:change-me@localhost:5438/hookline")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxConns != 10 || cfg.MaxConnLifetime != time.Hour || cfg.ConnConfig.ConnectTimeout != 5*time.Second {
		t.Errorf("pool config = max %d, lifetime %v, connect timeout %v", cfg.MaxConns, cfg.MaxConnLifetime, cfg.ConnConfig.ConnectTimeout)
	}
	if _, err := db.Config("postgres://%zz"); err == nil {
		t.Error("Config accepted a malformed URL")
	}
}

func TestConnectFailsFastWithoutAServer(t *testing.T) {
	cfg, err := db.Config("postgres://hookline:change-me@127.0.0.1:1/hookline") // nothing listens on port 1
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if pool, err := db.Connect(ctx, cfg); err == nil {
		pool.Close()
		t.Fatal("Connect succeeded against a closed port")
	}
}
