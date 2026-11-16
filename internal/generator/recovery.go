package generator

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// ErrReviewBudget stops model work while leaving time to write a partial result.
var ErrReviewBudget = errors.New("insufficient review time before the finalization reserve")

type CallPolicy struct {
	RecoveryRetries                 int
	AttemptTimeout, DeadlineReserve time.Duration
	RepairContext                   bool
	RepairMaxBytes                  int
}

type PolicyProvider interface{ GenerationPolicy() CallPolicy }

func PolicyOf(g Generator) CallPolicy {
	if p, ok := g.(PolicyProvider); ok {
		return p.GenerationPolicy()
	}
	return CallPolicy{RecoveryRetries: 1, RepairMaxBytes: 64 << 10}
}

func (l *Live) options() CallOptions {
	if l.cfg.Options != nil {
		return *l.cfg.Options
	}
	return DefaultCallOptions()
}

func (l *Live) GenerationPolicy() CallPolicy {
	o := l.options()
	return CallPolicy{RecoveryRetries: o.RecoveryRetries, AttemptTimeout: l.cfg.Timeout,
		DeadlineReserve: o.DeadlineReserve, RepairContext: o.RepairContext, RepairMaxBytes: o.RepairMaxBytes}
}

func (s *Second) GenerationPolicy() CallPolicy  { return s.live.GenerationPolicy() }
func (c *Command) GenerationPolicy() CallPolicy { return (&Live{cfg: c.cfg}).GenerationPolicy() }

type stepKey struct{}
type reserveKey struct{}
type reviewWindowKey struct{}
type reviewWindow struct {
	cap     time.Duration
	limited bool
}
type stepBudget struct{ remaining atomic.Int32 }

func ReviewContext(ctx context.Context) context.Context {
	w := reviewWindow{}
	if deadline, ok := ctx.Deadline(); ok {
		w.cap, w.limited = time.Until(deadline)/10, true
	}
	return context.WithValue(ctx, reviewWindowKey{}, w)
}

func NewStep(ctx context.Context, policy CallPolicy) context.Context {
	w, ok := ctx.Value(reviewWindowKey{}).(reviewWindow)
	if !ok {
		ctx = ReviewContext(ctx)
		w, _ = ctx.Value(reviewWindowKey{}).(reviewWindow)
	}
	reserve := policy.DeadlineReserve
	if w.limited && w.cap < reserve {
		reserve = w.cap
	}
	if reserve < 0 {
		reserve = 0
	}
	ctx = context.WithValue(ctx, reserveKey{}, reserve)
	b := &stepBudget{}
	b.remaining.Store(int32(policy.RecoveryRetries))
	return context.WithValue(ctx, stepKey{}, b)
}

func ensureStep(ctx context.Context, policy CallPolicy) context.Context {
	if ctx.Value(stepKey{}) != nil {
		return ctx
	}
	return NewStep(ctx, policy)
}

func reserveFor(ctx context.Context) time.Duration {
	r, _ := ctx.Value(reserveKey{}).(time.Duration)
	return r
}

func timeForAttempt(ctx context.Context, timeout time.Duration, recovery bool) bool {
	if ctx.Err() != nil {
		return false
	}
	if deadline, ok := ctx.Deadline(); ok {
		left := time.Until(deadline) - reserveFor(ctx)
		if left <= 0 || ((recovery || reserveFor(ctx) > 0) && left < timeout) {
			return false
		}
	}
	return true
}

// TakeRecovery shares one counter across the transport and schema decoder.
func TakeRecovery(ctx context.Context, timeout time.Duration) (bool, error) {
	b, _ := ctx.Value(stepKey{}).(*stepBudget)
	if b == nil {
		return false, nil
	}
	for n := b.remaining.Load(); n > 0; n = b.remaining.Load() {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if !timeForAttempt(ctx, timeout, true) {
			return false, ErrReviewBudget
		}
		if b.remaining.CompareAndSwap(n, n-1) {
			return true, nil
		}
	}
	return false, nil
}
