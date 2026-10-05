// Command skew simulates events spread over partitions by a partition key and prints the
// load of every partition.
//
//	skew -key uniform
//	skew -key zipf
//	skew -key composite
//	skew -key zipf -partitioner range
package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strings"

	"example.com/ch17-replica-lag/skew"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "skew:", err)
		os.Exit(2)
	}
}

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("skew", flag.ContinueOnError)
	key := fs.String("key", "uniform", "partition key: uniform | zipf | composite (zipf tenant + event id)")
	partitioner := fs.String("partitioner", "hash", "hash | range")
	partitions := fs.Int("partitions", 8, "number of partitions")
	tenants := fs.Int("tenants", 1000, "number of tenants")
	events := fs.Int("events", 100_000, "number of events")
	s := fs.Float64("s", 1.1, "Zipf exponent, greater than 1; larger is more skewed")
	seed := fs.Uint64("seed", 1, "random seed; the same seed prints the same numbers")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *s <= 1 || *partitions < 1 || *tenants < 2 {
		return fmt.Errorf("need -s > 1, -partitions >= 1 and -tenants >= 2")
	}
	r := rand.New(rand.NewPCG(*seed, 0))

	var next skew.KeyFunc
	switch *key {
	case "uniform":
		next = skew.Uniform(r, *tenants)
	case "zipf":
		next = skew.Zipf(r, *tenants, *s)
	case "composite":
		next = skew.Composite(r, skew.Zipf(r, *tenants, *s))
	default:
		return fmt.Errorf("unknown -key %q", *key)
	}
	var p skew.Partitioner
	switch *partitioner {
	case "hash":
		p = skew.Hash{N: *partitions}
	case "range":
		p = skew.TenantRanges(*tenants, *partitions)
	default:
		return fmt.Errorf("unknown -partitioner %q", *partitioner)
	}

	load := skew.Simulate(next, p, *partitions, *events)
	_, hottest := load.Hottest()
	fmt.Fprintf(out, "key=%s partitioner=%s partitions=%d tenants=%d events=%d s=%.2f\n",
		*key, *partitioner, *partitions, *tenants, *events, *s)
	for i, c := range load.Counts {
		bar := strings.Repeat("#", c*40/max(hottest, 1))
		fmt.Fprintf(out, "p%-3d %8d  %s\n", i, c, bar)
	}
	fmt.Fprintf(out, "hot factor (max / mean): %.2f\n", load.HotFactor())
	return nil
}
