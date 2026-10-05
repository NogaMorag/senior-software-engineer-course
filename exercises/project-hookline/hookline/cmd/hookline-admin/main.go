// Command hookline-admin runs operator tasks against the database named by DATABASE_URL:
//
//	hookline-admin migrate [up | down-to <version> | status]
//	hookline-admin create-tenant -name <name> [-rate <per second>] [-burst <n>]
//
// create-tenant prints the tenant's API key once; only its SHA-256 is stored.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"example.com/hookline/internal/admin"
	"example.com/hookline/internal/bus"
	"example.com/hookline/internal/db"
	"example.com/hookline/internal/outbox"
	"example.com/hookline/internal/store/postgres"
)

const usage = `usage:
  hookline-admin migrate [up | down-to <version> | status]
  hookline-admin create-tenant -name <name> [-rate <per second>] [-burst <n>]
  hookline-admin create-topic [-partitions <n>]
DATABASE_URL names the database; create-topic needs KAFKA_BROKERS instead.
`

var errUsage = errors.New("unknown or incomplete command")

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, errUsage) {
			fmt.Fprint(os.Stderr, usage)
		}
		fmt.Fprintln(os.Stderr, "hookline-admin:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errUsage
	}
	if args[0] == "create-topic" {
		return createTopic(args[1:])
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return errors.New("DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cfg, err := db.Config(url)
	if err != nil {
		return err
	}
	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	switch args[0] {
	case "migrate":
		action := "up"
		if len(args) > 1 {
			action = args[1]
		}
		switch {
		case action == "up":
			if err := db.Migrate(ctx, pool); err != nil {
				return err
			}
		case action == "down-to" && len(args) == 3:
			v, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("down-to: %q is not a version", args[2])
			}
			if err := db.MigrateDownTo(ctx, pool, v); err != nil {
				return err
			}
		case action != "status":
			return errUsage
		}
		v, err := db.Version(ctx, pool)
		if err != nil {
			return err
		}
		fmt.Println("schema version:", v)
		return nil
	case "create-tenant":
		fs := flag.NewFlagSet("create-tenant", flag.ContinueOnError)
		name := fs.String("name", "", "tenant name (required)")
		rate := fs.Int("rate", admin.DefaultRatePerSec, "ingest rate limit per second")
		burst := fs.Int("burst", admin.DefaultBurst, "ingest burst")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		t, key, err := admin.CreateTenant(ctx, postgres.New(pool), *name, *rate, *burst, time.Now)
		if err != nil {
			return err
		}
		fmt.Printf("tenant id: %s\napi key:   %s\nThe key is shown once and only its hash is stored. Keep it now.\n", t.ID, key)
		return nil
	default:
		return errUsage
	}
}

// createTopic creates hookline.events. The partition count caps how many dispatchers share the work.
func createTopic(args []string) error {
	fs := flag.NewFlagSet("create-topic", flag.ContinueOnError)
	partitions := fs.Int("partitions", 6, "partitions of "+outbox.Topic)
	if err := fs.Parse(args); err != nil {
		return err
	}
	brokers := bus.ParseBrokers(os.Getenv("KAFKA_BROKERS"))
	if brokers == nil {
		return errors.New("KAFKA_BROKERS is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := bus.CreateTopic(ctx, brokers, outbox.Topic, int32(*partitions)); err != nil {
		return err
	}
	fmt.Printf("topic %s: %d partitions\n", outbox.Topic, *partitions)
	return nil
}
