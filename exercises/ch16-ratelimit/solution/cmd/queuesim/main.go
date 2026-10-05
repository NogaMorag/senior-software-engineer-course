// Command queuesim prints how response time grows with utilization for one server with
// a 10ms mean service time, with an unbounded queue and with a bounded one.
//
//	go run ./cmd/queuesim
package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"example.com/ch16-ratelimit/queuesim"
)

func main() {
	service := flag.Float64("service", 10, "mean service time in ms")
	requests := flag.Int("n", 200_000, "requests per run")
	limit := flag.Int("limit", 20, "queue limit for the bounded runs")
	seed := flag.Uint64("seed", 1, "random seed")
	flag.Parse()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "rho\tqueue\tW mean ms\tW p99 ms\tL seen\tlambda*W\trejected %\t")
	for _, rho := range []float64{0.5, 0.7, 0.8, 0.9, 0.95, 0.99, 1.2} {
		for _, q := range []int{0, *limit} {
			r := queuesim.Run(queuesim.Config{
				ArrivalRate: rho / *service, ServiceTime: *service, Requests: *requests, QueueLimit: q, Seed: *seed,
			})
			name := "unbounded"
			if q > 0 {
				name = fmt.Sprintf("max %d", q)
			}
			fmt.Fprintf(tw, "%.2f\t%s\t%.1f\t%.1f\t%.2f\t%.2f\t%.1f\t\n", rho, name, r.MeanResponse, r.P99Response,
				r.SeenInSystem, r.Throughput*r.MeanResponse, 100*float64(r.Rejected)/float64(*requests))
		}
	}
	tw.Flush()
}
