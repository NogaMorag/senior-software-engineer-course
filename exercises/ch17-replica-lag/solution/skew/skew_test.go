package skew

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

func newRand() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) } // fixed seed: same numbers every run

func TestRangePartition(t *testing.T) {
	rg := TenantRanges(1000, 4) // bounds tenant-0250, tenant-0500, tenant-0750
	tests := []struct {
		key  string
		want int
	}{
		{"tenant-0000", 0},
		{"tenant-0249", 0},
		{"tenant-0250", 1},
		{"tenant-0749", 2},
		{"tenant-0750", 3},
		{"tenant-0999", 3},
		{"zzz", 3},
	}
	for _, tt := range tests {
		if got := rg.Partition(tt.key); got != tt.want {
			t.Errorf("Range.Partition(%q) = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestHashIsStableAndInRange(t *testing.T) {
	h := Hash{N: 8}
	for i := range 1000 {
		key := fmt.Sprintf("k%d", i)
		p := h.Partition(key)
		if p < 0 || p >= 8 {
			t.Fatalf("Hash.Partition(%q) = %d, out of range", key, p)
		}
		if again := h.Partition(key); again != p {
			t.Fatalf("Hash.Partition(%q) = %d then %d", key, p, again)
		}
	}
}

func TestSimulateShowsHotPartitions(t *testing.T) {
	const tenants, partitions, events = 1000, 8, 200_000
	tests := []struct {
		name        string
		key         func(r *rand.Rand) KeyFunc
		partitioner Partitioner
		minHot      float64 // the hot factor must be at least this
		maxHot      float64 // and at most this
	}{
		{"uniform tenants, hash", func(r *rand.Rand) KeyFunc { return Uniform(r, tenants) }, Hash{N: partitions}, 1.0, 1.2},
		{"zipf tenants, hash", func(r *rand.Rand) KeyFunc { return Zipf(r, tenants, 1.1) }, Hash{N: partitions}, 1.8, 8},
		{"zipf tenants, range", func(r *rand.Rand) KeyFunc { return Zipf(r, tenants, 1.1) }, TenantRanges(tenants, partitions), 4, 8},
		{"zipf tenants plus event id, hash", func(r *rand.Rand) KeyFunc { return Composite(r, Zipf(r, tenants, 1.1)) }, Hash{N: partitions}, 1.0, 1.2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			load := Simulate(tt.key(newRand()), tt.partitioner, partitions, events)
			hot := load.HotFactor()
			t.Logf("counts %v, hot factor %.2f", load.Counts, hot)
			if hot < tt.minHot || hot > tt.maxHot {
				t.Errorf("hot factor %.2f, want between %.1f and %.1f", hot, tt.minHot, tt.maxHot)
			}
		})
	}
}

func TestHotFactor(t *testing.T) {
	tests := []struct {
		counts []int
		want   float64
	}{
		{[]int{10, 10, 10, 10}, 1},
		{[]int{40, 0, 0, 0}, 4},
		{[]int{30, 10}, 1.5},
		{[]int{0, 0}, 0},
	}
	for _, tt := range tests {
		if got := (Load{Counts: tt.counts}).HotFactor(); got != tt.want {
			t.Errorf("HotFactor(%v) = %v, want %v", tt.counts, got, tt.want)
		}
	}
}

func TestRingMovesFewKeysWhenANodeJoins(t *testing.T) {
	keys := make([]string, 20_000)
	for i := range keys {
		keys[i] = fmt.Sprintf("tenant-%05d", i)
	}
	four := []string{"node-a", "node-b", "node-c", "node-d"}
	five := append(four[:4:4], "node-e")

	ringBefore, ringAfter := NewRing(four, 200), NewRing(five, 200)
	ringMoved := Moved(keys, ringBefore.Node, ringAfter.Node)

	modBefore, modAfter := Hash{N: 4}, Hash{N: 5}
	modMoved := Moved(keys,
		func(k string) string { return four[modBefore.Partition(k)] },
		func(k string) string { return five[modAfter.Partition(k)] })

	t.Logf("keys moved: ring %.1f%%, modulo %.1f%%", 100*ringMoved, 100*modMoved)
	if ringMoved > 0.3 {
		t.Errorf("ring moved %.2f of the keys, want about 0.2", ringMoved)
	}
	if modMoved < 0.7 {
		t.Errorf("modulo moved %.2f of the keys, want about 0.8", modMoved)
	}

	share := map[string]int{}
	for _, k := range keys {
		share[ringAfter.Node(k)]++
	}
	for node, n := range share {
		if f := float64(n) / float64(len(keys)); f < 0.12 || f > 0.28 {
			t.Errorf("node %s owns %.2f of the keys, want about 0.2", node, f)
		}
	}
}
