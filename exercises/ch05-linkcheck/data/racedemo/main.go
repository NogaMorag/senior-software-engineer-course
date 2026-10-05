// Command racedemo is the planted race of section 5.8: eight goroutines bump
// a shared "links checked" counter. Run it with and without -race, then with
// each fix.
//
//	go run .                  # wrong total, no warning
//	go run -race .            # WARNING: DATA RACE
//	go run -race . -fix mutex
//	go run -race . -fix atomic
package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

// progress is shared by every worker goroutine.
type progress struct {
	mu      sync.Mutex // guards locked
	locked  int
	racy    int          // plain int, no protection: the planted bug
	counter atomic.Int64 // atomic read-modify-write
}

func (p *progress) addLocked() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.locked++
}

func (p *progress) addRacy() {
	p.racy++ // DATA RACE: read, add one, write back, while others do the same
}

func (p *progress) addAtomic() {
	p.counter.Add(1)
}

func main() {
	fix := flag.String("fix", "none", "none, mutex or atomic")
	flag.Parse()

	var p progress
	add, total := p.addRacy, func() int64 { return int64(p.racy) }
	switch *fix {
	case "none":
	case "mutex":
		add, total = p.addLocked, func() int64 { return int64(p.locked) }
	case "atomic":
		add, total = p.addAtomic, p.counter.Load
	default:
		fmt.Fprintln(os.Stderr, "racedemo: -fix must be none, mutex or atomic")
		os.Exit(2)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 1000 {
				add()
			}
		})
	}
	wg.Wait()
	fmt.Println("checked:", total(), "of", 8*1000)
}
