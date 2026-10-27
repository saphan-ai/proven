package network_test

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/network"
)

// dnsAnswering serves an in-memory DNS over a pipe: every A query is answered
// with addr, every other query with an empty answer. No socket is opened.
func dnsAnswering(addr [4]byte) func(conn net.Conn) {
	return func(conn net.Conn) {
		defer conn.Close()
		buf := make([]byte, 512)
		for {
			var frame [2]byte
			if _, err := io.ReadFull(conn, frame[:]); err != nil {
				return
			}
			n := int(binary.BigEndian.Uint16(frame[:]))
			if n > len(buf) {
				return
			}
			if _, err := io.ReadFull(conn, buf[:n]); err != nil {
				return
			}
			if n < 18 {
				continue
			}
			end := 12
			for end < n && buf[end] != 0 {
				end += int(buf[end]) + 1
			}
			if end+5 > n {
				continue
			}
			question := append([]byte(nil), buf[12:end+5]...)
			kind := binary.BigEndian.Uint16(buf[end+1 : end+3])
			var answer []byte
			if kind == 1 {
				answer = append([]byte{0xc0, 0x0c, 0, 1, 0, 1, 0, 0, 0, 30, 0, 4}, addr[:]...)
			}
			resp := make([]byte, 12)
			copy(resp[:2], buf[:2])
			resp[2], resp[3] = 0x81, 0x80
			resp[5] = 1
			if len(answer) > 0 {
				resp[7] = 1
			}
			resp = append(resp, question...)
			resp = append(resp, answer...)
			binary.BigEndian.PutUint16(frame[:], uint16(len(resp)))
			_, _ = conn.Write(append(frame[:], resp...))
		}
	}
}

func policyFrom(t *testing.T, env map[string]string) *network.Policy {
	t.Helper()
	p, err := network.Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// The dial of an allowed name whose answer is outside names the host and the
// address, and the refusal comes before a connection.
func TestLocalDialRefusesAnOutsideAnswerByNameAndAddress(t *testing.T) {
	p := policyFrom(t, map[string]string{"PROVEN_NETWORK": "local", "PROVEN_NETWORK_ALLOW": "service.corp.example"})
	p = p.WithResolver(&net.Resolver{PreferGo: true, Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		client, server := net.Pipe()
		go dnsAnswering([4]byte{203, 0, 113, 9})(server)
		return client, nil
	}})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := p.DialContext(ctx, "tcp", "service.corp.example:443")
	if conn != nil {
		conn.Close()
		t.Fatal("a connection was opened for an outside answer")
	}
	if err == nil || !strings.Contains(err.Error(), "service.corp.example") || !strings.Contains(err.Error(), "203.0.113.9") {
		t.Fatalf("refusal does not name the host and the address: %v", err)
	}
}

// A private answer for an allowed name is used: the check refuses outside
// addresses and nothing else.
func TestLocalAllowedNameKeepsAPrivateDNSAnswer(t *testing.T) {
	p := policyFrom(t, map[string]string{"PROVEN_NETWORK": "local", "PROVEN_NETWORK_ALLOW": "service.corp.example"})
	p = p.WithResolver(&net.Resolver{PreferGo: true, Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		client, server := net.Pipe()
		go dnsAnswering([4]byte{10, 1, 2, 3})(server)
		return client, nil
	}})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ips, err := p.Lookup(ctx, "service.corp.example")
	if err != nil || len(ips) != 1 || ips[0] != netip.MustParseAddr("10.1.2.3") {
		t.Fatalf("ips = %v, err = %v; want [10.1.2.3]", ips, err)
	}
}

// An allowed name must not turn an outside address into an allowed connection.
func TestLocalAllowedNameRejectsOutsideDNSAnswer(t *testing.T) {
	p := policyFrom(t, map[string]string{"PROVEN_NETWORK": "local", "PROVEN_NETWORK_ALLOW": "service.corp.example"})
	res := &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		client, server := net.Pipe()
		go dnsAnswering([4]byte{203, 0, 113, 9})(server)
		return client, nil
	}}
	p = p.WithResolver(res)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// The control: the same resolver does return the outside address, so a
	// refusal below is the policy's and not an empty answer.
	raw, err := res.LookupNetIP(ctx, "ip", "service.corp.example.")
	if err != nil || len(raw) == 0 || raw[0] != netip.MustParseAddr("203.0.113.9") {
		t.Fatalf("resolver did not return the control answer: %v %v", raw, err)
	}
	ips, err := p.Lookup(ctx, "service.corp.example")
	for _, ip := range ips {
		if ip == netip.MustParseAddr("203.0.113.9") {
			t.Fatalf("local mode accepted outside DNS answer %s; Lookup must refuse it before DialContext uses it", ip)
		}
	}
	if err == nil || !strings.Contains(err.Error(), "203.0.113.9") {
		t.Fatalf("Lookup did not refuse the outside answer by address: ips=%v err=%v", ips, err)
	}
}
