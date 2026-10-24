// Package main is a small synthetic backup tool's alert path.
// Credentials in this case are synthetic only, such as tok_synthetic_0000_not_a_secret.
package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Alert is one alert event: its kind and the fields the code emits for it.
type Alert struct {
	Kind   string
	Fields map[string]string
}

// catalogue lists the alert kinds and the fields the code emits for each.
var catalogue = map[string][]string{
	"backup_failed":   {"job", "host", "error", "time"},
	"backup_late":     {"job", "host", "expected_at", "time"},
	"restore_done":    {"job", "host", "time"},
	"storage_low":     {"volume", "percent_free", "time"},
	"delivery_failed": {"destination", "reason", "time"},
}

var destinationID = regexp.MustCompile(`^[a-z0-9-]+$`)

// ParseDestination accepts or refuses a destination setting, with the reason.
// A setting is "chat:<id>", and a chat destination id begins with "c-".
func ParseDestination(setting string) (string, error) {
	kind, id, ok := strings.Cut(setting, ":")
	if !ok || kind != "chat" {
		return "", fmt.Errorf("unknown destination kind in %q", setting)
	}
	if !strings.HasPrefix(id, "c-") {
		return "", fmt.Errorf("destination id %q must begin with c-", id)
	}
	if !destinationID.MatchString(id) {
		return "", fmt.Errorf("destination id %q has characters outside a-z, 0-9 and -", id)
	}
	return id, nil
}

// Sender delivers one alert. It reports whether the alert was delivered.
type Sender interface {
	Send(a Alert) (delivered bool)
	Reason() string
}

// Receiver is the in-memory receiver a real sender would deliver to.
type Receiver struct {
	Got []Alert
}

type nullSender struct{ reason string }

func (n nullSender) Send(Alert) bool { return false }
func (n nullSender) Reason() string  { return n.reason }

// resolveSender returns the sender for a receiver. At this version it always
// resolves to a sender that sends nothing.
func resolveSender(_ *Receiver) Sender {
	return nullSender{reason: "no transport yet"}
}

// Result counts what one emit did.
type Result struct {
	Considered int
	Delivered  int
	Reason     string
}

// Emit sends one alert event through the resolved sender.
func Emit(kind string, rcv *Receiver) (Result, error) {
	fields, ok := catalogue[kind]
	if !ok {
		return Result{}, fmt.Errorf("unknown alert %q", kind)
	}
	a := Alert{Kind: kind, Fields: map[string]string{}}
	for _, f := range fields {
		a.Fields[f] = "x"
	}
	s := resolveSender(rcv)
	res := Result{Considered: 1, Reason: s.Reason()}
	if s.Send(a) {
		rcv.Got = append(rcv.Got, a)
		res.Delivered++
	}
	return res, nil
}

// FieldsOf lists the fields an emitted alert of this kind carries.
func FieldsOf(kind string) ([]string, error) {
	if _, ok := catalogue[kind]; !ok {
		return nil, fmt.Errorf("unknown alert %q", kind)
	}
	out := append([]string(nil), catalogue[kind]...)
	sort.Strings(out)
	return out, nil
}
