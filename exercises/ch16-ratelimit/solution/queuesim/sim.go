// Package queuesim simulates one server with a FIFO queue in virtual time: random
// (Poisson) arrivals, random (exponential) service times. No sleeping, so a million
// requests take a moment, and a seed makes every run repeatable.
package queuesim

import (
	"math/rand/v2"
	"slices"
)

// Config describes one run. Times are in milliseconds of virtual time.
type Config struct {
	ArrivalRate float64 // λ: requests per millisecond
	ServiceTime float64 // mean time to serve one request
	Requests    int
	QueueLimit  int // waiting requests allowed besides the one in service; 0 means unbounded
	Seed        uint64
}

// Result is what one run observed.
type Result struct {
	Utilization  float64 // ρ = λ × service time: the fraction of time the server is busy
	Served       int
	Rejected     int
	MeanResponse float64 // W: time in the system, waiting plus service, for served requests
	P99Response  float64
	SeenInSystem float64 // L: average number in the system, as seen by arriving requests
	Throughput   float64 // λ actually admitted: served requests per millisecond of the run
}

// Run simulates cfg.
func Run(cfg Config) Result {
	rnd := rand.New(rand.NewPCG(cfg.Seed, cfg.Seed^0x9e3779b97f4a7c15))
	var (
		now       float64   // virtual clock
		lastDone  float64   // when the server finishes everything accepted so far
		inSystem  []float64 // departure times of the accepted requests not yet gone, in order
		responses []float64
		seen      float64
		rejected  int
	)
	for range cfg.Requests {
		now += rnd.ExpFloat64() / cfg.ArrivalRate
		// Requests that left before this arrival are gone.
		for len(inSystem) > 0 && inSystem[0] <= now {
			inSystem = inSystem[1:]
		}
		// Poisson arrivals see time averages, so averaging what each arrival finds gives L.
		seen += float64(len(inSystem))
		if cfg.QueueLimit > 0 && len(inSystem) > cfg.QueueLimit {
			rejected++
			continue
		}
		done := max(now, lastDone) + rnd.ExpFloat64()*cfg.ServiceTime
		lastDone = done
		inSystem = append(inSystem, done)
		responses = append(responses, done-now)
	}
	r := Result{
		Utilization:  cfg.ArrivalRate * cfg.ServiceTime,
		Served:       len(responses),
		Rejected:     rejected,
		SeenInSystem: seen / float64(cfg.Requests),
	}
	if len(responses) == 0 {
		return r
	}
	var sum float64
	for _, x := range responses {
		sum += x
	}
	r.MeanResponse = sum / float64(len(responses))
	slices.Sort(responses)
	r.P99Response = responses[len(responses)*99/100]
	r.Throughput = float64(len(responses)) / now
	return r
}
