package config_test

import (
	"testing"
	"time"

	"example.com/ch09-catalog-api/internal/config"
)

func TestLoad(t *testing.T) {
	env := map[string]string{"CATALOG_MAX_BODY_BYTES": "2048", "CATALOG_IDEMPOTENCY_TTL": "1h"}
	c, err := config.Load(func(k string) string { return env[k] })
	if err != nil || c.MaxBodyBytes != 2048 || c.IdempotencyTTL != time.Hour || c.HTTPAddr != ":8080" {
		t.Fatalf("Load = %+v, %v", c, err)
	}
	env["CATALOG_IDEMPOTENCY_TTL"] = "forever"
	if _, err := config.Load(func(k string) string { return env[k] }); err == nil {
		t.Fatal("an invalid TTL must be an error")
	}
}
