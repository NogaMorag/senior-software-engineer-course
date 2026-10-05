// Package visits counts how often each visitor was greeted.
package visits

import (
	"errors"
	"fmt"
	"sync"
)

// ErrUnknown is returned for a visitor that was never counted.
var ErrUnknown = errors.New("Unknown visitor.")

// Counter is safe for concurrent use. The zero value is not; use New.
type Counter struct {
	mu     sync.Mutex
	counts map[string]int
}

// New returns an empty counter.
func New() *Counter { return &Counter{counts: map[string]int{}} }

// Add counts one visit.
func (c *Counter) Add(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[name]++
}

// Lookup returns the number of visits, or an error wrapping ErrUnknown.
func (c Counter) Lookup(name string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	n, ok := c.counts[name]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknown, name)
	}
	return n, nil
}

// Seen reports whether name was ever counted.
func (c *Counter) Seen(name string) bool {
	_, err := c.Lookup(name)
	return err != ErrUnknown
}

// Report is a one-line summary for one visitor.
func (c *Counter) Report(name string) string {
	n, err := c.Lookup(name)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("%s visited %d times", n, name)
}
