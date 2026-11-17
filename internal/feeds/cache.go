// Package feeds acquires public source data through bounded HTTP adapters.
// Models never choose addresses or interpret refresh success.
package feeds

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/saphan-ai/proven/internal/network"
)

const maxResponse = 32 << 20

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}
type HTTPStatusError struct{ Status int }

func (e *HTTPStatusError) Error() string { return fmt.Sprintf("source returned HTTP %d", e.Status) }

type cacheEntry struct {
	ETag     string          `json:"etag,omitempty"`
	Modified string          `json:"modified,omitempty"`
	Body     json.RawMessage `json:"body"`
}

// Client never falls back to a cached success after an HTTP/parse error. A 304
// explicitly validates a retained response; OSV modified IDs validate records.
type Client struct {
	HTTP               HTTPDoer
	Policy             *network.Policy
	Dir                string
	Downloaded, Reused int
	consumed           int64
	Changed            int
}

func (c *Client) cachePath(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(c.Dir, hex.EncodeToString(sum[:])+".json")
}
func (c *Client) cached(key string) (cacheEntry, bool) {
	name := c.cachePath(key)
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxResponse*2 {
		return cacheEntry{}, false
	}
	f, err := os.Open(name)
	if err != nil {
		return cacheEntry{}, false
	}
	defer f.Close()
	var entry cacheEntry
	if json.NewDecoder(io.LimitReader(f, maxResponse*2+1)).Decode(&entry) != nil || !json.Valid(entry.Body) {
		return cacheEntry{}, false
	}
	return entry, true
}
func (c *Client) write(key string, entry cacheEntry) error {
	if err := os.MkdirAll(c.Dir, 0700); err != nil {
		return err
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(c.Dir, ".refresh-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, c.cachePath(key))
}
func (c *Client) JSON(ctx context.Context, method, address string, body []byte) ([]byte, error) {
	if err := c.Policy.CheckURL("PROVEN_SOURCE_URL", address); err != nil {
		return nil, err
	}
	key := method + " " + address + " " + string(body)
	old, have := c.cached(key)
	req, err := http.NewRequestWithContext(ctx, method, address, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Proven-source-refresh")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if method == http.MethodGet && have {
		if old.ETag != "" {
			req.Header.Set("If-None-Match", old.ETag)
		}
		if old.Modified != "" {
			req.Header.Set("If-Modified-Since", old.Modified)
		}
	}
	response, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("source request failed")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotModified && have {
		if err := c.consume(len(old.Body)); err != nil {
			return nil, err
		}
		c.Reused++
		return old.Body, nil
	}
	if response.StatusCode != http.StatusOK {
		return nil, &HTTPStatusError{response.StatusCode}
	}
	b, err := io.ReadAll(io.LimitReader(response.Body, maxResponse+1))
	if err != nil || len(b) > maxResponse || !json.Valid(b) {
		return nil, fmt.Errorf("source returned invalid or oversized JSON")
	}
	c.Downloaded++
	if err := c.consume(len(b)); err != nil {
		return nil, err
	}
	if have && bytes.Equal(old.Body, b) {
		c.Reused++
	} else {
		c.Changed++
	}
	if err = c.write(key, cacheEntry{response.Header.Get("ETag"), response.Header.Get("Last-Modified"), b}); err != nil {
		return nil, fmt.Errorf("source cache could not be written")
	}
	return b, nil
}
func NewClient(policy *network.Policy, dir string) *Client {
	return &Client{HTTP: policy.NewClient(20 * time.Second), Policy: policy, Dir: dir}
}

func (c *Client) consume(size int) error {
	c.consumed += int64(size)
	if c.consumed > 128<<20 {
		return fmt.Errorf("source refresh exceeded the 128 MiB aggregate response budget")
	}
	return nil
}
