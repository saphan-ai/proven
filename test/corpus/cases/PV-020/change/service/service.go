// Package service is a small request gate.
// Production refuses a body larger than DocumentedLimit on every route
// except the documented upload routes, and accepts that body on those routes.
// The limit has no bypass.
package service

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// DocumentedLimit is the maximum body, in bytes, on a route that is not
// an exempt upload. A body of this size is accepted. One byte more is refused.
const DocumentedLimit = 2048

// Route is one operation the service offers.
type Route struct {
	Method string
	Path   string
	Exempt bool
}

// DocumentedExempt is the upload list the service publishes.
// It is a separate value from the Exempt flags on Routes, so a control can
// see the two drift apart. They agree in this case.
var DocumentedExempt = []string{
	"POST /v1/uploads/media",
	"POST /v1/uploads/bundle",
}

// Routes is the whole surface. Order is part of the case: the check
// prints one line per route in this order.
var Routes = []Route{
	{Method: "GET", Path: "/v1/health", Exempt: false},
	{Method: "POST", Path: "/v1/items", Exempt: false},
	{Method: "POST", Path: "/v1/search", Exempt: false},
	{Method: "PUT", Path: "/v1/items/current", Exempt: false},
	{Method: "POST", Path: "/v1/uploads/media", Exempt: true},
	{Method: "POST", Path: "/v1/uploads/bundle", Exempt: true},
}

// Decision is the answer for one route and one body.
type Decision struct {
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason"`
}

// Service answers a route and a body.
// maxBody == 0 means the limit is not applied. Production does not use that.
type Service struct {
	maxBody int
}

// Production returns the service this case ships: the limit is on.
func Production() *Service { return &Service{maxBody: DocumentedLimit} }

// WithoutLimit returns a service that does not apply the limit.
// The in-process controls use it to show they notice the limit's absence.
// It is not the service the case ships.
func WithoutLimit() *Service { return &Service{maxBody: 0} }

// MaxBody returns the applied limit. Zero means the limit is not applied.
func (s *Service) MaxBody() int { return s.maxBody }

// Decide reports accepted or refused, with a reason.
// A body above the limit is refused on every route that is not exempt,
// including a route the service does not offer. An exempt route accepts it.
func (s *Service) Decide(method, path string, body []byte) Decision {
	return s.decideSize(method, path, int64(len(body)))
}

func (s *Service) decideSize(method, path string, size int64) Decision {
	route, ok := lookup(method, path)
	exempt := ok && route.Exempt
	if s.maxBody > 0 && !exempt && size > int64(s.maxBody) {
		return Decision{Accepted: false, Reason: "body exceeds limit"}
	}
	if !ok {
		return Decision{Accepted: false, Reason: "unknown route"}
	}
	if exempt && s.maxBody > 0 && size > int64(s.maxBody) {
		return Decision{Accepted: true, Reason: "exempt upload"}
	}
	return Decision{Accepted: true, Reason: "accepted"}
}

func lookup(method, path string) (Route, bool) {
	for _, rt := range Routes {
		if rt.Method == method && rt.Path == path {
			return rt, true
		}
	}
	return Route{}, false
}

// Handler serves the same rule as Decide over HTTP.
// On a route the limit covers, the body is counted as it arrives.
// Bodies are counted without buffering them.
func (s *Service) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size, err := s.readBodySize(w, r)
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) && s.limitApplies(r.Method, r.URL.Path) {
				writeJSON(w, http.StatusRequestEntityTooLarge, Decision{Accepted: false, Reason: "body exceeds limit"})
				return
			}
			writeJSON(w, http.StatusBadRequest, Decision{Accepted: false, Reason: "body unreadable"})
			return
		}
		d := s.decideSize(r.Method, r.URL.Path, size)
		code := http.StatusOK
		if !d.Accepted {
			if d.Reason == "unknown route" {
				code = http.StatusNotFound
			} else {
				code = http.StatusRequestEntityTooLarge
			}
		}
		writeJSON(w, code, d)
	})
}

func (s *Service) readBodySize(w http.ResponseWriter, r *http.Request) (int64, error) {
	rc := r.Body
	if s.limitApplies(r.Method, r.URL.Path) {
		rc = http.MaxBytesReader(w, rc, int64(s.maxBody))
	}
	return io.Copy(io.Discard, rc)
}

func (s *Service) limitApplies(method, path string) bool {
	if s.maxBody <= 0 {
		return false
	}
	route, ok := lookup(method, path)
	if !ok {
		return true
	}
	return !route.Exempt
}

func writeJSON(w http.ResponseWriter, code int, d Decision) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(d)
}
