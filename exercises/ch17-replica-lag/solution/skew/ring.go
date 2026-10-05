package skew

import (
	"fmt"
	"slices"
	"sort"
)

// Ring is a consistent-hash ring. Each node is placed at vnodes points on a 64-bit circle; a
// key belongs to the first point clockwise from the key's hash. Adding a node takes over
// only the arcs in front of its own points, so only about 1/(n+1) of the keys move,
// where modulo hashing would move most of them.
type Ring struct {
	points []uint64
	owner  map[uint64]string
}

// NewRing builds a ring over nodes with vnodes points per node.
func NewRing(nodes []string, vnodes int) *Ring {
	r := &Ring{owner: make(map[uint64]string)}
	for _, n := range nodes {
		for v := range vnodes {
			h := hash64(fmt.Sprintf("%s#%d", n, v))
			r.points = append(r.points, h)
			r.owner[h] = n
		}
	}
	slices.Sort(r.points)
	return r
}

// Node returns the node that owns key.
func (r *Ring) Node(key string) string {
	h := hash64(key)
	i := sort.Search(len(r.points), func(i int) bool { return r.points[i] >= h })
	if i == len(r.points) {
		i = 0 // wrap around the circle
	}
	return r.owner[r.points[i]]
}

// Moved returns the fraction of keys whose owner differs between before and after.
func Moved(keys []string, before, after func(string) string) float64 {
	moved := 0
	for _, k := range keys {
		if before(k) != after(k) {
			moved++
		}
	}
	return float64(moved) / float64(len(keys))
}
