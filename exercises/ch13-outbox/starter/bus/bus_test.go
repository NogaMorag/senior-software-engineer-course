package bus_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"example.com/ch13-outbox/bus"
	"example.com/ch13-outbox/consumer"
	"example.com/ch13-outbox/testenv"
)

func produce(t *testing.T, brokers []string, topic string, recs ...*kgo.Record) {
	t.Helper()
	cl, err := bus.NewProducer(brokers)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, r := range recs {
		r.Topic = topic
	}
	if err := cl.ProduceSync(ctx, recs...).FirstErr(); err != nil {
		t.Fatal(err)
	}
}

func TestSameKeySamePartitionInOrder(t *testing.T) {
	brokers := testenv.Brokers(t)
	topic := testenv.Topic(t, brokers, 3)
	keys := []string{"alice", "bob", "carol", "dave", "erin", "frank"}
	var recs []*kgo.Record
	for i := range 60 {
		recs = append(recs, &kgo.Record{Key: []byte(keys[i%len(keys)]), Value: []byte(strconv.Itoa(i))})
	}
	produce(t, brokers, topic, recs...)

	// Read the topic back without a group, every partition from the start.
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	partitionOf := map[string]int32{}
	lastValue := map[string]int{}
	for seen := 0; seen < len(recs); {
		fetches := cl.PollFetches(ctx)
		if err := ctx.Err(); err != nil {
			t.Fatalf("read back %d of %d records: %v", seen, len(recs), err)
		}
		fetches.EachRecord(func(r *kgo.Record) {
			seen++
			k := string(r.Key)
			v, _ := strconv.Atoi(string(r.Value))
			if p, ok := partitionOf[k]; ok && p != r.Partition {
				t.Errorf("key %s on partitions %d and %d", k, p, r.Partition)
			}
			partitionOf[k] = r.Partition
			if last, ok := lastValue[k]; ok && v < last {
				t.Errorf("key %s: value %d after %d, out of order", k, v, last)
			}
			lastValue[k] = v
		})
	}
	used := map[int32]bool{}
	for _, p := range partitionOf {
		used[p] = true
	}
	t.Logf("key to partition: %v", partitionOf)
	if len(used) < 2 {
		t.Errorf("six keys all hashed to one partition (%v); suspicious for a key-hash partitioner", partitionOf)
	}
}

// A member that dies after handling a record but before committing it: the next member
// of the group gets that record again (a duplicate), and nothing is lost.
func TestUncommittedRecordsAreRedelivered(t *testing.T) {
	brokers := testenv.Brokers(t)
	topic := testenv.Topic(t, brokers, 1)
	group := testenv.Group("redeliver")
	var recs []*kgo.Record
	for i := range 10 {
		recs = append(recs, &kgo.Record{Key: []byte("k"), Value: fmt.Appendf(nil, "m-%d", i)})
	}
	produce(t, brokers, topic, recs...)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var seen []string
	crash := errors.New("killed after the effect, before the commit")
	a, err := bus.NewGroupConsumer(brokers, group, topic)
	if err != nil {
		t.Fatal(err)
	}
	err = consumer.Run(ctx, a, func(_ context.Context, r *kgo.Record) error {
		seen = append(seen, string(r.Value))
		if string(r.Value) == "m-4" {
			return crash
		}
		return nil
	})
	a.Close()
	if !errors.Is(err, crash) {
		t.Fatalf("first member: Run = %v, want the simulated crash", err)
	}

	b, err := bus.NewGroupConsumer(brokers, group, topic)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	bctx, stop := context.WithCancel(ctx)
	var second []string
	err = consumer.Run(bctx, b, func(_ context.Context, r *kgo.Record) error {
		second = append(second, string(r.Value))
		if string(r.Value) == "m-9" {
			stop()
		}
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("second member: Run = %v", err)
	}
	if ctx.Err() != nil {
		t.Fatalf("second member never reached m-9: got %v", second)
	}
	if len(second) == 0 || second[0] != "m-4" {
		t.Fatalf("second member started at %v, want m-4 (the first uncommitted record)", second)
	}
	all := append(append([]string{}, seen...), second...)
	counts := map[string]int{}
	for _, v := range all {
		counts[v]++
	}
	for i := range 10 {
		v := fmt.Sprintf("m-%d", i)
		if counts[v] == 0 {
			t.Errorf("%s lost", v)
		}
	}
	if counts["m-4"] != 2 {
		t.Errorf("m-4 handled %d times, want 2 (at least once means a duplicate here)", counts["m-4"])
	}
	t.Logf("first member: %s | second member: %s", strings.Join(seen, " "), strings.Join(second, " "))
}

// With the cooperative-sticky balancer, a second member joining takes some partitions
// from the first, and the first keeps the rest without giving them up.
func TestCooperativeRebalanceRevokesOnlyMovedPartitions(t *testing.T) {
	brokers := testenv.Brokers(t)
	topic := testenv.Topic(t, brokers, 4)
	group := testenv.Group("coop")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	type member struct {
		mu      sync.Mutex
		owned   map[int32]bool
		revoked []int32
	}
	start := func(m *member) *kgo.Client {
		m.owned = map[int32]bool{}
		cl, err := bus.NewGroupConsumer(brokers, group, topic,
			kgo.OnPartitionsAssigned(func(_ context.Context, _ *kgo.Client, p map[string][]int32) {
				m.mu.Lock()
				for _, n := range p[topic] {
					m.owned[n] = true
				}
				m.mu.Unlock()
			}),
			kgo.OnPartitionsRevoked(func(_ context.Context, _ *kgo.Client, p map[string][]int32) {
				m.mu.Lock()
				for _, n := range p[topic] {
					delete(m.owned, n)
					m.revoked = append(m.revoked, n)
				}
				m.mu.Unlock()
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		go func() { _ = consumer.Run(ctx, cl, func(context.Context, *kgo.Record) error { return nil }) }()
		return cl
	}
	owned := func(m *member) int {
		m.mu.Lock()
		defer m.mu.Unlock()
		return len(m.owned)
	}
	waitFor := func(what string, cond func() bool) {
		for !cond() {
			select {
			case <-ctx.Done():
				t.Fatalf("timed out waiting for %s", what)
			case <-time.After(50 * time.Millisecond):
			}
		}
	}

	var a, b member
	ca := start(&a)
	defer ca.Close()
	waitFor("A to own all 4 partitions", func() bool { return owned(&a) == 4 })
	cb := start(&b)
	defer cb.Close()
	waitFor("the group to settle at 2 and 2", func() bool { return owned(&a) == 2 && owned(&b) == 2 })

	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.revoked) != 2 {
		t.Errorf("A gave up %v, want exactly the 2 partitions that moved to B", a.revoked)
	}
}
