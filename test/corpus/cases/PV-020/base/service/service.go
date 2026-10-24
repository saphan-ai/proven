// Package service is a small request gate with no size limit.
// Every known route accepts any body.
package service

// Route is one operation the service offers.
type Route struct {
	Method string
	Path   string
}

// Routes is the whole surface. Order is part of the case: the check
// prints one line per route in this order.
var Routes = []Route{
	{Method: "GET", Path: "/v1/health"},
	{Method: "POST", Path: "/v1/items"},
	{Method: "POST", Path: "/v1/search"},
	{Method: "PUT", Path: "/v1/items/current"},
	{Method: "POST", Path: "/v1/uploads/media"},
	{Method: "POST", Path: "/v1/uploads/bundle"},
}

// Decision is the answer for one route and one body.
type Decision struct {
	Accepted bool
	Reason   string
}

// Service answers a route and a body. It has no size limit.
type Service struct{}

// New returns the base service.
func New() *Service { return &Service{} }

// Decide accepts any body on a known route. An unknown route is refused.
// The body is part of the operation; its length is not a reason to refuse.
func (s *Service) Decide(method, path string, body []byte) Decision {
	if !known(method, path) {
		return Decision{Accepted: false, Reason: "unknown route"}
	}
	_ = body
	return Decision{Accepted: true, Reason: "accepted"}
}

func known(method, path string) bool {
	for _, rt := range Routes {
		if rt.Method == method && rt.Path == path {
			return true
		}
	}
	return false
}
