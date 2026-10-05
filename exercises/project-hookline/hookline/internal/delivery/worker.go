package delivery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"example.com/hookline/internal/backoff"
	"example.com/hookline/internal/metrics"
)

// Pool is a set of delivery workers sharing one database pool (chapter 14's worker, on deliveries).
type Pool struct {
	DB          *pgxpool.Pool
	Sender      Sender
	Workers     int            // goroutines; 0 means 1
	Lease       time.Duration  // how long a claim lasts; must exceed Sender.Timeout
	Poll        time.Duration  // idle wait when nothing is due
	MaxAttempts int            // per round; a replay starts a new round
	Backoff     backoff.Policy // delay before retry n of a round
	Rand        backoff.Rand   // nil means backoff.Global
	Log         *slog.Logger   // nil discards
	// Guards are the per-endpoint breakers and bulkheads (milestone M6). nil means none.
	Guards *Guards
	// Metrics records lag, attempts and postponements (milestone M8). nil means none.
	Metrics *metrics.Metrics
}

// Run starts the workers and blocks until ctx is cancelled and every worker has finished and recorded
// its current attempt (drain). It returns nil after a clean drain.
func (p *Pool) Run(ctx context.Context) error {
	if p.Lease <= p.timeout() {
		return fmt.Errorf("lease %s must be longer than the attempt timeout %s", p.Lease, p.timeout())
	}
	var wg sync.WaitGroup
	for i := range max(p.Workers, 1) {
		wg.Go(func() { p.loop(ctx, i) })
	}
	wg.Wait()
	return nil
}

func (p *Pool) timeout() time.Duration {
	if p.Sender.Timeout <= 0 {
		return 10 * time.Second
	}
	return p.Sender.Timeout
}

func (p *Pool) loop(ctx context.Context, worker int) {
	for ctx.Err() == nil { // stop claiming as soon as shutdown starts
		j, err := Claim(ctx, p.DB, p.Lease, p.MaxAttempts)
		if err != nil && ctx.Err() == nil {
			p.log().Error("claim failed", slog.Int("worker", worker), slog.String("error", err.Error()))
		}
		if j == nil {
			select {
			case <-ctx.Done():
			case <-time.After(p.Poll):
			}
			continue
		}
		p.RunOne(ctx, j)
	}
}

// RunOne sends one claimed delivery and records the result, even if shutdown starts meanwhile: the
// attempt's context is detached from ctx but bounded by the lease, because after the lease another
// worker may take the delivery and our result no longer counts.
func (p *Pool) RunOne(ctx context.Context, j *Job) {
	// One span per attempt, a child of the ingest request's span: the trace that started at
	// POST /v1/events ends here, or in the receiver when it traces too (milestone M8).
	ctx, span := otel.Tracer(metrics.Scope).Start(withTraceparent(ctx, j.TraceContext), "hookline.deliver",
		trace.WithAttributes(attribute.String("hookline.delivery.id", j.DeliveryID),
			attribute.String("hookline.event.id", j.EventID), attribute.String("hookline.endpoint.id", j.EndpointID),
			attribute.Int("hookline.attempt", j.Attempt)))
	defer span.End()
	log := p.log().With(slog.String("delivery_id", j.DeliveryID), slog.String("event_id", j.EventID), slog.Int("attempt", j.Attempt))
	rec, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	var res Result
	if j.Enabled {
		var sent bool
		res, sent = p.guardedSend(ctx, j)
		if !sent {
			return // postponed: no request, no attempt
		}
	} else {
		res = Result{Started: time.Now(), Err: errors.New("endpoint disabled")} // not retryable: dead
	}

	if res.StatusCode != 0 {
		span.SetAttributes(attribute.Int("http.response.status_code", res.StatusCode))
	}
	if res.OK() {
		if err := Complete(rec, p.DB, j, res); err != nil {
			log.ErrorContext(ctx, "recording success failed", slog.String("error", err.Error()))
			return
		}
		p.metrics(func(m *metrics.Metrics) { m.Delivered(ctx, j.EventAt, time.Now()) })
		log.DebugContext(ctx, "delivered", slog.Int("status", res.StatusCode), slog.Duration("duration", res.Duration))
		return
	}
	span.SetStatus(codes.Error, res.Err.Error())
	delay := p.Backoff.Delay(j.Round, p.rand())
	if res.RetryAfter > delay { // the receiver asked for more time; honour it up to the cap
		delay = min(res.RetryAfter, p.Backoff.Cap)
	}
	status, err := Fail(rec, p.DB, j, res, p.MaxAttempts, delay)
	if err != nil {
		log.ErrorContext(ctx, "recording failure failed", slog.String("error", err.Error()), slog.String("cause", res.Err.Error()))
		return
	}
	outcome := "retry"
	if status == "dead" {
		outcome = "dead"
	}
	p.metrics(func(m *metrics.Metrics) { m.Attempt(ctx, outcome) })
	log.InfoContext(ctx, "delivery attempt failed", slog.String("error", res.Err.Error()), slog.Int("status", res.StatusCode),
		slog.String("next", status), slog.Duration("retry_in", delay))
}

// guardedSend sends j through its endpoint's bulkhead and breaker. When either refuses, the delivery is
// postponed without spending an attempt, and sent is false: a slow or dead endpoint costs this worker a
// database round trip, not a worker blocked for the whole timeout.
func (p *Pool) guardedSend(ctx context.Context, j *Job) (res Result, sent bool) {
	attempt := func() Result {
		attemptCtx, cancel := context.WithDeadline(context.WithoutCancel(ctx), j.LeaseExpiry)
		defer cancel()
		return p.Sender.Send(attemptCtx, *j)
	}
	if p.Guards == nil {
		return attempt(), true
	}
	g := p.Guards.get(j.EndpointID)
	if g.bulkhead != nil {
		if err := g.bulkhead.Acquire(ctx); err != nil {
			p.postpone(ctx, j, busyWait(), "bulkhead", err)
			return Result{}, false
		}
		defer g.bulkhead.Release()
	}
	done, err := g.breaker.Allow()
	if err != nil {
		p.postpone(ctx, j, p.Guards.openWait(), "breaker", err)
		return Result{}, false
	}
	res = attempt()
	done(breakerError(res))
	return res, true
}

func (p *Pool) postpone(ctx context.Context, j *Job, after time.Duration, reason string, why error) {
	p.metrics(func(m *metrics.Metrics) { m.Postponed(ctx, reason) })
	rec, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if err := Postpone(rec, p.DB, j, after); err != nil {
		p.log().Error("postpone failed", slog.String("delivery_id", j.DeliveryID), slog.String("error", err.Error()))
		return
	}
	p.log().Debug("delivery postponed", slog.String("delivery_id", j.DeliveryID), slog.String("endpoint_id", j.EndpointID),
		slog.String("reason", why.Error()), slog.Duration("retry_in", after))
}

// metrics runs record when the pool has Metrics.
func (p *Pool) metrics(record func(*metrics.Metrics)) {
	if p.Metrics != nil {
		record(p.Metrics)
	}
}

// withTraceparent returns ctx with the remote span context of traceparent, or ctx unchanged.
func withTraceparent(ctx context.Context, traceparent string) context.Context {
	if traceparent == "" {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier{"traceparent": traceparent})
}

func (p *Pool) rand() backoff.Rand {
	if p.Rand == nil {
		return backoff.Global
	}
	return p.Rand
}

func (p *Pool) log() *slog.Logger {
	if p.Log == nil {
		return slog.New(slog.DiscardHandler)
	}
	return p.Log
}
