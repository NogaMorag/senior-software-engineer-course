package queuesim_test

import (
	"math"
	"testing"

	"example.com/ch16-ratelimit/queuesim"
)

// Little's law, L = λW, holds for any stable queue. Here L is counted by the arriving
// requests and W is timed per request: two different measurements that must agree.
func TestLittlesLaw(t *testing.T) {
	for _, cfg := range []queuesim.Config{
		{ArrivalRate: 0.05, ServiceTime: 10, Requests: 200_000, Seed: 1},                // ρ = 0.5
		{ArrivalRate: 0.09, ServiceTime: 10, Requests: 200_000, Seed: 2},                // ρ = 0.9
		{ArrivalRate: 0.12, ServiceTime: 10, Requests: 200_000, QueueLimit: 5, Seed: 3}, // overloaded, bounded
	} {
		r := queuesim.Run(cfg)
		lw := r.Throughput * r.MeanResponse
		if math.Abs(r.SeenInSystem-lw)/lw > 0.05 {
			t.Errorf("ρ=%.2f limit=%d: L=%.3f but λW=%.3f", r.Utilization, cfg.QueueLimit, r.SeenInSystem, lw)
		}
	}
}

// For this queue (M/M/1) theory says W = S / (1 - ρ): at ρ = 0.9 a 10ms request takes 100ms.
func TestResponseTimeExplodesNearSaturation(t *testing.T) {
	var prev float64
	for _, rho := range []float64{0.5, 0.8, 0.9, 0.95} {
		r := queuesim.Run(queuesim.Config{ArrivalRate: rho / 10, ServiceTime: 10, Requests: 400_000, Seed: 4})
		want := 10 / (1 - rho)
		if math.Abs(r.MeanResponse-want)/want > 0.15 {
			t.Errorf("ρ=%.2f: W=%.1fms, theory %.1fms", rho, r.MeanResponse, want)
		}
		if r.MeanResponse <= prev {
			t.Errorf("ρ=%.2f: W=%.1fms did not grow", rho, r.MeanResponse)
		}
		prev = r.MeanResponse
	}
}

// A bounded queue trades rejections for a bounded response time.
func TestBoundedQueueBoundsResponseTime(t *testing.T) {
	over := queuesim.Config{ArrivalRate: 0.15, ServiceTime: 10, Requests: 100_000, Seed: 5} // ρ = 1.5
	unbounded := queuesim.Run(over)
	over.QueueLimit = 10
	bounded := queuesim.Run(over)
	if bounded.Rejected == 0 || unbounded.Rejected != 0 {
		t.Fatalf("rejected: bounded %d, unbounded %d", bounded.Rejected, unbounded.Rejected)
	}
	if bounded.P99Response > 300 || unbounded.P99Response < 10*bounded.P99Response {
		t.Fatalf("p99: bounded %.0fms, unbounded %.0fms; want the bounded one small and far below", bounded.P99Response, unbounded.P99Response)
	}
}
