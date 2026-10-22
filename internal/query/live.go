package query

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/service"
)

// Live runs one SELECT on Zetaris over the SQL-editor route the 2026-10-21
// measurement proved: POST {base}/sql-editor/sqls/run. The start/close routes
// under /query/sql were not measured, so this client does not call them.
// The request body is closed on every response, including a non-200 and a timeout.
type Live struct {
	cfg    Settings
	client *http.Client
	now    func() time.Time
	newID  func() (string, error)
}

// NewLive builds a live Zetaris client. A nil client uses one that does not
// follow redirects, so the API key is not sent to another host.
func NewLive(cfg Settings, client *http.Client) *Live {
	if client == nil {
		client = &http.Client{
			Timeout: cfg.Timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	return &Live{cfg: cfg, client: client, now: time.Now, newID: newRequestID}
}

type sqlRunRequest struct {
	QueryID string `json:"queryId"`
	SQL     string `json:"sql"`
	Source  string `json:"source"`
	Limit   int    `json:"limit"`
}

type sqlRunResponse struct {
	Headers []string            `json:"headers"`
	Data    [][]json.RawMessage `json:"data"`
}

// Query sends one SELECT and returns the rows. The engine name is zetaris.
// The snapshot is the UTC time of the call. A 401, a timeout, or an
// unreachable endpoint names the service zetaris and does not print the key.
func (l *Live) Query(ctx context.Context, sql string) (Result, error) {
	if l == nil || l.client == nil || l.cfg.BaseURL == "" || l.cfg.APIKey == "" || l.cfg.OrgID == "" || l.cfg.Timeout <= 0 || l.cfg.RowLimit < 1 {
		return Result{}, down("not configured")
	}
	if !oneSelect(sql) {
		return Result{}, down("not a select")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, l.cfg.Timeout)
	defer cancel()
	reqID, err := l.newID()
	if err != nil {
		return Result{}, down("not configured")
	}
	queryID, err := l.newID()
	if err != nil {
		return Result{}, down("not configured")
	}
	payload, err := json.Marshal(sqlRunRequest{
		QueryID: queryID,
		SQL:     sql,
		Source:  "SqlEditor",
		Limit:   l.cfg.RowLimit,
	})
	if err != nil {
		return Result{}, down("not configured")
	}
	endpoint := l.cfg.BaseURL + "/sql-editor/sqls/run"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return Result{}, down("not configured")
	}
	req.Header.Set("Authorization", "Bearer "+l.cfg.APIKey)
	req.Header.Set("X-Org-ID", l.cfg.OrgID)
	req.Header.Set("X-Request-ID", reqID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := l.client.Do(req)
	if err != nil {
		return Result{}, classify(parent, ctx, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		discard(resp.Body)
		return Result{}, down("unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		discard(resp.Body)
		return Result{}, down(httpReason(resp.StatusCode))
	}
	var parsed sqlRunResponse
	dec := json.NewDecoder(io.LimitReader(resp.Body, 8<<20))
	if err := dec.Decode(&parsed); err != nil {
		if parent.Err() != nil {
			return Result{}, parent.Err()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return Result{}, down("timeout")
		}
		return Result{}, down("bad response")
	}
	cols := parsed.Headers
	if cols == nil {
		cols = []string{}
	}
	rows := make([][]string, 0, len(parsed.Data))
	for _, raw := range parsed.Data {
		if len(rows) >= l.cfg.RowLimit {
			break
		}
		row := make([]string, len(raw))
		for i, cell := range raw {
			s, cerr := cellString(cell)
			if cerr != nil {
				return Result{}, down("bad response")
			}
			row[i] = s
		}
		rows = append(rows, row)
	}
	when := l.now
	if when == nil {
		when = time.Now
	}
	return Result{
		SQL:      sql,
		Engine:   "zetaris",
		Snapshot: when().UTC().Format(time.RFC3339),
		Columns:  cols,
		Rows:     rows,
	}, nil
}

func oneSelect(sql string) bool {
	stmts := splitStatements(sql)
	if len(stmts) != 1 {
		return false
	}
	fields := strings.Fields(stmts[0])
	return len(fields) > 0 && strings.EqualFold(fields[0], "SELECT")
}

func cellString(raw json.RawMessage) (string, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "", nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	var n json.Number
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&n); err == nil && n != "" {
		return n.String(), nil
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return strconv.FormatBool(b), nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", err
	}
	out, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func newRequestID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func classify(parent, ctx context.Context, err error) error {
	if parent.Err() != nil {
		return parent.Err()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || timedOut(err) {
		return down("timeout")
	}
	return down("unreachable")
}

func timedOut(err error) bool {
	var t interface{ Timeout() bool }
	return errors.As(err, &t) && t.Timeout()
}

func httpReason(code int) string {
	return "http " + strconv.Itoa(code)
}

func discard(r io.Reader) {
	_, _ = io.Copy(io.Discard, io.LimitReader(r, 4096))
}

func down(reason string) *service.Down {
	return &service.Down{Service: "zetaris", Reason: reason}
}
