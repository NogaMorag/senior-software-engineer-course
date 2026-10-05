// Package config reads the service configuration from environment variables, as in chapter 6.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config is everything the binary needs.
type Config struct {
	HTTPAddr       string
	GRPCAddr       string
	MaxBodyBytes   int64
	IdempotencyTTL time.Duration
}

// Load reads CATALOG_HTTP_ADDR, CATALOG_GRPC_ADDR, CATALOG_MAX_BODY_BYTES and CATALOG_IDEMPOTENCY_TTL.
func Load(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	c := Config{HTTPAddr: ":8080", GRPCAddr: ":9090", MaxBodyBytes: 16 << 10, IdempotencyTTL: 24 * time.Hour}
	if v := getenv("CATALOG_HTTP_ADDR"); v != "" {
		c.HTTPAddr = v
	}
	if v := getenv("CATALOG_GRPC_ADDR"); v != "" {
		c.GRPCAddr = v
	}
	if v := getenv("CATALOG_MAX_BODY_BYTES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("CATALOG_MAX_BODY_BYTES: want a positive integer, got %q", v)
		}
		c.MaxBodyBytes = n
	}
	if v := getenv("CATALOG_IDEMPOTENCY_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("CATALOG_IDEMPOTENCY_TTL: want a positive duration such as 24h, got %q", v)
		}
		c.IdempotencyTTL = d
	}
	return c, nil
}
