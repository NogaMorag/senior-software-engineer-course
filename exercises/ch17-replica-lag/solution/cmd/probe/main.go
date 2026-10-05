// Command probe measures replication lag: it writes rows on the primary and polls the
// replica until each one is visible, then prints what it observed.
//
//	probe -n 50                 # an idle primary
//	probe -n 50 -load 4         # with four writers loading the primary
//	probe -n 10 -delay 2s       # with recovery_min_apply_delay set on the replica
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"example.com/ch17-replica-lag/replica"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "probe:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("probe", flag.ContinueOnError)
	n := fs.Int("n", 20, "number of probes")
	interval := fs.Duration("interval", time.Millisecond, "how often to poll the replica")
	load := fs.Int("load", 0, "writers loading the primary during the probes")
	delay := fs.Duration("delay", 0, "set recovery_min_apply_delay on the replica for the run")
	if err := fs.Parse(args); err != nil {
		return err
	}
	primaryURL, replicaURL := os.Getenv("DATABASE_URL"), os.Getenv("REPLICA_URL")
	if primaryURL == "" || replicaURL == "" {
		return errors.New("set DATABASE_URL (primary) and REPLICA_URL (replica); see the README")
	}
	p, err := replica.Connect(ctx, primaryURL, replicaURL)
	if err != nil {
		return err
	}
	defer p.Close()
	if err := p.Setup(ctx); err != nil {
		return err
	}
	if *delay > 0 {
		if err := p.SetApplyDelay(ctx, *delay); err != nil {
			return err
		}
		defer func() { // reset even after Ctrl+C, with a fresh context
			resetCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := p.SetApplyDelay(resetCtx, 0); err != nil {
				fmt.Fprintln(os.Stderr, "probe: could not reset recovery_min_apply_delay:", err)
			}
		}()
	}

	loadCtx, stopLoad := context.WithCancel(ctx)
	loadDone := make(chan int64, 1)
	if *load > 0 {
		go func() {
			rows, err := p.Load(loadCtx, *load)
			if err != nil {
				fmt.Fprintln(os.Stderr, "probe: load:", err)
			}
			loadDone <- rows
		}()
		time.Sleep(500 * time.Millisecond) // let the load build up before the first probe
	}

	var obs []replica.Observation
	for i := range *n {
		o, err := p.Probe(ctx, fmt.Sprintf("probe %d", i), *interval)
		if err != nil {
			stopLoad()
			return err
		}
		obs = append(obs, o)
		fmt.Fprintf(out, "probe %3d  row %-8d lag %-12v polls %d\n", i, o.ID, o.Lag.Round(time.Microsecond), o.Polls)
	}
	senders, err := p.Senders(ctx) // read while the load is still running
	stopLoad()
	if err != nil {
		return err
	}
	if *load > 0 {
		fmt.Fprintf(out, "load: %d rows written by %d writers\n", <-loadDone, *load)
	}

	s := replica.Summarize(obs)
	fmt.Fprintf(out, "summary: n=%d min=%v median=%v p95=%v max=%v first-read-missed=%d\n",
		s.N, s.Min.Round(time.Microsecond), s.Median.Round(time.Microsecond),
		s.P95.Round(time.Microsecond), s.Max.Round(time.Microsecond), s.MissedFirstRead)
	for _, sn := range senders {
		lag := "n/a"
		if sn.ReplayLag != nil {
			lag = sn.ReplayLag.Round(time.Microsecond).String()
		}
		fmt.Fprintf(out, "pg_stat_replication: %s state=%s sync=%s replay_lag=%s behind=%d bytes\n",
			sn.Name, sn.State, sn.SyncState, lag, sn.ReplayLagBytes)
	}
	return nil
}
