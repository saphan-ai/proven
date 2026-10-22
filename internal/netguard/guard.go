// Package netguard is the sample-mode network refusal. A dial is counted and refused.
package netguard

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
)

// Guard refuses every dial and counts the attempts.
type Guard struct {
	dials  atomic.Int32
	client *http.Client
}

// Dials is how many dial attempts this guard has seen.
func (g *Guard) Dials() int { return int(g.dials.Load()) }

// Transport is an HTTP transport whose dial fails the call and increments Dials.
func (g *Guard) Transport() *http.Transport {
	return &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			g.dials.Add(1)
			return nil, errors.New("sample mode: network refused")
		},
	}
}

// Client is the HTTP client the sample review holds. Agents do not use another.
func (g *Guard) Client() *http.Client {
	return &http.Client{Transport: g.Transport()}
}

// Arm keeps that client on the guard so sample mode has one refusing transport.
func (g *Guard) Arm() {
	if g.client == nil {
		g.client = g.Client()
	}
}

// Armed reports that Arm has run.
func (g *Guard) Armed() bool { return g.client != nil }
