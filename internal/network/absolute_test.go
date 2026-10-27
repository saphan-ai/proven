package network

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestAbsoluteNameEndsWithADot: the name a lookup asks for is absolute, so the
// host's search list adds no other name. Removing the dot turns this red.
func TestAbsoluteNameEndsWithADot(t *testing.T) {
	if got := absoluteName("service.corp.example"); got != "service.corp.example." {
		t.Fatalf("absoluteName = %q; want a trailing dot", got)
	}
}

// recordingDNS answers every A query with 10.1.2.3 and records the question
// name of each query it sees.
type recordingDNS struct {
	mu    sync.Mutex
	names []string
}

func (r *recordingDNS) serve(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 512)
	for {
		var frame [2]byte
		if _, err := io.ReadFull(conn, frame[:]); err != nil {
			return
		}
		n := int(binary.BigEndian.Uint16(frame[:]))
		if n > len(buf) || n < 18 {
			return
		}
		if _, err := io.ReadFull(conn, buf[:n]); err != nil {
			return
		}
		var labels []string
		end := 12
		for end < n && buf[end] != 0 {
			l := int(buf[end])
			labels = append(labels, string(buf[end+1:end+1+l]))
			end += l + 1
		}
		if end+5 > n {
			return
		}
		r.mu.Lock()
		r.names = append(r.names, strings.Join(labels, ".")+".")
		r.mu.Unlock()
		kind := binary.BigEndian.Uint16(buf[end+1 : end+3])
		resp := make([]byte, 12)
		copy(resp[:2], buf[:2])
		resp[2], resp[3] = 0x81, 0x80
		resp[5] = 1
		question := append([]byte(nil), buf[12:end+5]...)
		var answer []byte
		if kind == 1 {
			answer = []byte{0xc0, 0x0c, 0, 1, 0, 1, 0, 0, 0, 30, 0, 4, 10, 1, 2, 3}
			resp[7] = 1
		}
		resp = append(append(resp, question...), answer...)
		binary.BigEndian.PutUint16(frame[:], uint16(len(resp)))
		if _, err := conn.Write(append(frame[:], resp...)); err != nil {
			return
		}
	}
}

// TestLocalAsksAListedNameOnceAsItself reads the DNS question on the wire: a
// listed name is asked as that absolute name, and no name with a search suffix
// is asked. The first-label name is the case a search list would extend.
func TestLocalAsksAListedNameOnceAsItself(t *testing.T) {
	for _, name := range []string{"service.corp.example", "daemonhost"} {
		t.Run(name, func(t *testing.T) {
			rec := &recordingDNS{}
			p := localPolicyAllowing(t, name)
			p = p.WithResolver(&net.Resolver{PreferGo: true, Dial: func(context.Context, string, string) (net.Conn, error) {
				client, server := net.Pipe()
				go rec.serve(server)
				return client, nil
			}})
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if _, err := p.Lookup(ctx, name); err != nil {
				t.Fatal(err)
			}
			rec.mu.Lock()
			defer rec.mu.Unlock()
			seen := map[string]int{}
			for _, q := range rec.names {
				seen[q]++
			}
			if len(seen) != 1 || seen[name+"."] == 0 {
				t.Fatalf("questions = %v; want only %q", rec.names, name+".")
			}
		})
	}
}
