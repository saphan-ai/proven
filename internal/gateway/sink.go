package gateway

import (
	"fmt"
	"io"
)

// The stable prefixes of the second channel. The fleet's watchers grep for
// them on standard error.
const (
	FaultPrefix     = "JUDGES-GATEWAY-FAULT"
	RecoveredPrefix = "JUDGES-GATEWAY-RECOVERED"
)

// maxPending bounds the lines held while a sink is down. Calls are refused
// while it is down, so only calls already in flight add to it.
const maxPending = 1024

// sink is a destination the gateway must be able to write: the call log or
// the alarm. A write that fails is never dropped silently. The line is kept,
// the failure is named on the fault channel, and the gateway refuses calls
// until the line can be written. All use is under Gateway.logMu.
type sink struct {
	name    string
	w       io.Writer
	pending [][]byte
	lost    int  // lines dropped because pending was full
	down    bool // the last write failed
}

// put queues one line and tries to write everything queued.
func (g *Gateway) put(s *sink, line []byte) {
	if len(s.pending) >= maxPending {
		s.lost++
	} else {
		s.pending = append(s.pending, line)
	}
	g.flush(s)
}

// flush writes the queued lines in order. It reports whether the sink is
// healthy. The first failure after a healthy spell writes one fault line; the
// first success after a failure writes one recovered line.
func (g *Gateway) flush(s *sink) bool {
	for len(s.pending) > 0 {
		if _, err := s.w.Write(s.pending[0]); err != nil {
			if !s.down {
				s.down = true
				g.fault("%s sink=%s error=%q queued=%d: the gateway refuses calls until this sink writes again",
					FaultPrefix, s.name, err.Error(), len(s.pending))
			}
			return false
		}
		s.pending = s.pending[1:]
	}
	if s.down {
		s.down = false
		g.fault("%s sink=%s lost=%d unlogged_refusals=%d", RecoveredPrefix, s.name, s.lost, g.unlogged)
		s.lost, g.unlogged = 0, 0
	}
	return true
}

// fault writes a line on the second channel. If that fails too, nothing
// else can be done; the health signal still reports the sink.
func (g *Gateway) fault(format string, a ...any) {
	_, _ = fmt.Fprintf(g.cfg.Fault, format+"\n", a...)
}

// Healthy reports whether the call log and the alarm can both be written. It
// tries any lines held back, so it is also how a sink recovers when no call
// arrives.
func (g *Gateway) Healthy() bool {
	g.logMu.Lock()
	defer g.logMu.Unlock()
	return g.sinksUp()
}

// sinksUp must be called with logMu held. Both sinks are tried, so one
// failure does not hide the other.
func (g *Gateway) sinksUp() bool {
	a := g.flush(&g.logSink)
	b := g.flush(&g.alarmSink)
	return a && b
}
