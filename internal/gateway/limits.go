package gateway

import (
	"fmt"
	"sync"
	"time"
)

// The refusal codes of the limiter.
const (
	CodeRateMinute   = "rate_limit_minute"
	CodeAllowanceDay = "allowance_day"
)

// AlarmPrefix starts every alarm line. The fleet's watchers grep for it.
const AlarmPrefix = "JUDGES-GATEWAY-ALARM"

// alarmPercent is the share of a limit at which a token's use raises an alarm.
const alarmPercent = 80

// window is one fixed window of use.
type window struct {
	start   time.Time
	n       int
	alarmed bool
}

type usage struct{ minute, day window }

// Limiter counts calls per token in a fixed minute and a fixed UTC day. The
// counts live in memory: a restart of the gateway starts them again, which
// the ADR names.
type Limiter struct {
	mu    sync.Mutex
	state map[string]*usage
}

// NewLimiter returns an empty limiter.
func NewLimiter() *Limiter { return &Limiter{state: map[string]*usage{}} }

// Allow counts one call for the token. It returns a refusal when the minute or
// the day is full, and, when this call brings a window to 80% of its limit, one
// alarm line. A refused call is not counted. Each window alarms once.
func (l *Limiter) Allow(rec Record, now time.Time) (alarms []string, refused *Refusal) {
	return l.AllowKey("token", rec.ID, rec.PerMinute, rec.PerDay, now)
}

// AllowKey is Allow for any caller: kind is "token" or "identity", name the
// token id or the identity, and the limits those of its row.
func (l *Limiter) AllowKey(kind, name string, perMinute, perDay int, now time.Time) (alarms []string, refused *Refusal) {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := kind + ":" + name
	u := l.state[key]
	if u == nil {
		u = &usage{}
		l.state[key] = u
	}
	now = now.UTC()
	roll(&u.minute, now.Truncate(time.Minute))
	roll(&u.day, now.Truncate(24*time.Hour))
	if u.day.n >= perDay {
		return nil, &Refusal{Status: 429, Code: CodeAllowanceDay, Detail: fmt.Sprintf("this %s's allowance of %d calls a day is used; it renews at 00:00 UTC", kind, perDay)}
	}
	if u.minute.n >= perMinute {
		return nil, &Refusal{Status: 429, Code: CodeRateMinute, Detail: fmt.Sprintf("this %s may make %d calls a minute; wait for the next minute", kind, perMinute)}
	}
	u.minute.n++
	u.day.n++
	if a := alarm(&u.minute, "minute", kind, name, perMinute); a != "" {
		alarms = append(alarms, a)
	}
	if a := alarm(&u.day, "day", kind, name, perDay); a != "" {
		alarms = append(alarms, a)
	}
	return alarms, nil
}

func roll(w *window, start time.Time) {
	if !w.start.Equal(start) {
		*w = window{start: start}
	}
}

func alarm(w *window, name, kind, id string, limit int) string {
	threshold := (limit*alarmPercent + 99) / 100
	if w.alarmed || w.n < threshold {
		return ""
	}
	w.alarmed = true
	return fmt.Sprintf("%s %s=%s window=%s used=%d limit=%d", AlarmPrefix, kind, id, name, w.n, limit)
}
