// Package skew simulates how a partition key spreads events across partitions, so you can
// see a hot partition appear before production shows it to you.
package skew

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"sort"
)

// KeyFunc returns the partition key of the next event.
type KeyFunc func() string

// Uniform draws one of tenants tenants, each equally likely.
func Uniform(r *rand.Rand, tenants int) KeyFunc {
	return func() string { return tenantName(r.IntN(tenants)) }
}

// Zipf draws tenants with a Zipf distribution: tenant k is chosen with probability
// proportional to (1+k)^(-s), so tenant 0 is the largest customer, tenant 1 the next, and
// a long tail of small tenants follows. s must be greater than 1; larger s, more skew.
func Zipf(r *rand.Rand, tenants int, s float64) KeyFunc {
	z := rand.NewZipf(r, s, 1, uint64(tenants-1))
	return func() string { return tenantName(int(z.Uint64())) }
}

// Composite appends a random event number to the tenant key, so one tenant's events spread
// over all partitions. The price: events of one tenant no longer share a partition, so the
// log no longer keeps them in order.
func Composite(r *rand.Rand, tenant KeyFunc) KeyFunc {
	return func() string { return fmt.Sprintf("%s/%d", tenant(), r.Uint32()) }
}

func tenantName(i int) string { return fmt.Sprintf("tenant-%04d", i) }

// Partitioner maps a key to a partition index.
type Partitioner interface {
	Partition(key string) int
}

// Hash is hash partitioning: a hash of the key, modulo the partition count.
type Hash struct{ N int }

// Partition implements Partitioner.
func (h Hash) Partition(key string) int { return int(hash64(key) % uint64(h.N)) }

// Range is range partitioning: partition i holds keys below Bounds[i] (and at or above the
// previous bound); the last partition holds everything from the last bound on.
type Range struct{ Bounds []string }

// Partition implements Partitioner.
func (rg Range) Partition(key string) int {
	return sort.Search(len(rg.Bounds), func(i int) bool { return key < rg.Bounds[i] })
}

// TenantRanges splits the tenant names tenant-0000 .. tenant-(tenants-1) into n ranges of
// equal width, the way a range-partitioned table is often first laid out.
func TenantRanges(tenants, n int) Range {
	bounds := make([]string, n-1)
	for i := range bounds {
		bounds[i] = tenantName((i + 1) * tenants / n)
	}
	return Range{Bounds: bounds}
}

// Load is how many events each partition received.
type Load struct{ Counts []int }

// Simulate draws events keys from next and counts them per partition.
func Simulate(next KeyFunc, p Partitioner, partitions, events int) Load {
	l := Load{Counts: make([]int, partitions)}
	for range events {
		l.Counts[p.Partition(next())]++
	}
	return l
}

// Hottest returns the busiest partition and its count.
func (l Load) Hottest() (partition, count int) {
	for i, c := range l.Counts {
		if c > count {
			partition, count = i, c
		}
	}
	return partition, count
}

// HotFactor is the busiest partition's count divided by the mean: 1.0 is perfectly even,
// and a partition with HotFactor 3 needs three times the capacity of an average one.
func (l Load) HotFactor() float64 {
	total := 0
	for _, c := range l.Counts {
		total += c
	}
	if total == 0 {
		return 0
	}
	_, hottest := l.Hottest()
	return float64(hottest) * float64(len(l.Counts)) / float64(total)
}

// hash64 is FNV-1a followed by a 64-bit finalizer (from SplitMix64). FNV alone spreads
// short, similar strings such as "node-a#1", "node-a#2" poorly in the high bits.
func hash64(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s)) // a hash.Hash never returns an error
	x := h.Sum64()
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}
