// Package service is the shared shape of an adapter that failed.
package service

import "errors"

// Down is a failed outside service. The review continues and names it.
type Down struct {
	Service string
	Reason  string
}

func (d *Down) Error() string {
	if d == nil {
		return "service down"
	}
	return d.Service + ": " + d.Reason
}

// As reports a Down error.
func As(err error) (*Down, bool) {
	var d *Down
	if errors.As(err, &d) {
		return d, true
	}
	return nil, false
}
