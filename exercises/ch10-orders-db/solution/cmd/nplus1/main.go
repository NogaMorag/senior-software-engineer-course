// Command nplus1 runs the N+1 and the batched versions of "orders with items" against
// the seeded chapter 10 database and prints how many queries each sent and how long it
// took. Usage:
//
//	DATABASE_URL=postgres://course:change-me@localhost:5410/orders go run ./cmd/nplus1 -customer 1 -limit 200
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	orders "example.com/ch10-orders-db"
)

func main() {
	customer := flag.Int64("customer", 1, "customer id")
	limit := flag.Int("limit", 200, "number of orders to load")
	runs := flag.Int("runs", 5, "repetitions; the median is printed")
	flag.Parse()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("set DATABASE_URL; see the README")
	}
	ctx := context.Background()
	cfg, err := orders.Config(url)
	if err != nil {
		log.Fatal(err)
	}
	counter := &orders.QueryCounter{}
	cfg.ConnConfig.Tracer = counter
	pool, err := orders.Connect(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	repo := orders.New(pool)

	variants := []struct {
		name string
		load func(context.Context, int64, int) ([]orders.Order, error)
	}{
		{"N+1", repo.OrdersWithItemsNPlusOne},
		{"batched", repo.OrdersWithItems},
	}
	for _, v := range variants {
		var times []time.Duration
		var queries int64
		var n int
		for range *runs {
			counter.Reset()
			start := time.Now()
			got, err := v.load(ctx, *customer, *limit)
			if err != nil {
				log.Fatalf("%s: %v", v.name, err)
			}
			times = append(times, time.Since(start))
			queries, n = counter.Count(), len(got)
		}
		slices.Sort(times)
		fmt.Printf("%-8s orders=%d queries=%d median=%v\n", v.name, n, queries, times[len(times)/2].Round(10*time.Microsecond))
	}
}
