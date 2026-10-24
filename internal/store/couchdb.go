package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

// EngineCouchDB is the engine name the evidence of a CouchDB query carries.
const EngineCouchDB = "couchdb"

// maxAnswer bounds how much of one CouchDB answer is read, so a runaway
// response cannot exhaust memory.
const maxAnswer = 8 << 20

// CouchDB is the evidence store that runs a Mango selector over CouchDB's
// HTTP API (ADR-0047). One registered table is one database. A query is the
// Mango form of a template, built by querytpl from validated values; no text
// a model wrote is sent (ADR-0032). The HTTP client does not follow
// redirects, so the credential cannot be sent to another address.
type CouchDB struct {
	cfg    CouchSettings
	client *http.Client
	now    func() time.Time
}

// NewCouchDB builds the store on validated settings. A nil client gets one
// with the settings' timeout and no redirects; a test may pass its own.
func NewCouchDB(cfg CouchSettings, client *http.Client) *CouchDB {
	if client == nil {
		client = &http.Client{
			Timeout: cfg.Timeout,
			// ADR-0013: a redirect would carry the Basic credential to
			// whatever address answered. The first answer is the answer.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}
	}
	return &CouchDB{cfg: cfg, client: client, now: time.Now}
}

// SetClock replaces the clock that stamps a snapshot.
func (c *CouchDB) SetClock(now func() time.Time) {
	if c == nil || now == nil {
		return
	}
	c.now = now
}

// Database is the name of the database that holds table. It is the
// configured prefix, then the table name in lower case with each `.` written
// as `-` (CouchDB database names have no dot). ValidRelation has already
// limited the table to letters, digits, `_` and one `.`, so the name needs no
// escaping in a URL path.
func (c *CouchDB) Database(table string) string {
	return c.cfg.Prefix + strings.ReplaceAll(strings.ToLower(table), ".", "-")
}

// Register creates the table's database and writes its rows as documents.
// A second register of the same table is refused: CouchDB answers 412 for a
// database that exists. Each document holds one row as text fields and has a
// zero-padded `_id` that keeps the rows in load order.
func (c *CouchDB) Register(ctx context.Context, table Table) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c == nil || c.client == nil {
		return fmt.Errorf("store: no couchdb store")
	}
	if err := checkTable(table); err != nil {
		return err
	}
	db := c.Database(table.Name)
	status, _, err := c.do(ctx, http.MethodPut, "/"+db, nil)
	if err != nil {
		return err
	}
	switch status {
	case http.StatusCreated, http.StatusAccepted:
	case http.StatusPreconditionFailed:
		return fmt.Errorf("store: table %s already exists", table.Name)
	default:
		return couchDown(status)
	}
	if len(table.Rows) == 0 {
		return nil
	}
	docs := make([]map[string]string, len(table.Rows))
	for i, row := range table.Rows {
		doc := map[string]string{"_id": fmt.Sprintf("%08d", i)}
		for j, col := range table.Columns {
			doc[col] = row[j]
		}
		docs[i] = doc
	}
	body, err := json.Marshal(map[string]any{"docs": docs})
	if err != nil {
		return err
	}
	status, _, err = c.do(ctx, http.MethodPost, "/"+db+"/_bulk_docs", body)
	if err != nil {
		return err
	}
	if status != http.StatusCreated && status != http.StatusAccepted {
		return couchDown(status)
	}
	return nil
}

// findResponse keeps only the documents of a _find answer. A field may be any
// JSON value, so it is decoded later, one cell at a time.
type findResponse struct {
	Docs []map[string]json.RawMessage `json:"docs"`
}

// Run builds the mango form and sends it to the table's database.
// The sql form is not read: CouchDB speaks Mango (ADR-0047). A database that
// does not exist is a down store named couchdb, never an empty answer: an
// empty answer would read as "the source holds no such row" (ADR-0002).
func (c *CouchDB) Run(ctx context.Context, req querytpl.Request, rel querytpl.Relations) (Evidence, error) {
	if c == nil || c.client == nil {
		return Evidence{}, fmt.Errorf("store: no couchdb store")
	}
	forms, err := querytpl.BuildForms(req, rel)
	if err != nil {
		return Evidence{}, err
	}
	if err := ctx.Err(); err != nil {
		return Evidence{}, err
	}
	// The body is made from the structured query, not read back out of the
	// Mango text, so what is sent and what the evidence shows are one value.
	body, err := json.Marshal(map[string]any{
		"selector": forms.Find.Selector,
		"fields":   forms.Find.Fields,
		"limit":    c.cfg.RowLimit,
	})
	if err != nil {
		return Evidence{}, couchFailed("not configured")
	}
	status, raw, err := c.do(ctx, http.MethodPost, "/"+c.Database(forms.Find.Table)+"/_find", body)
	if err != nil {
		return Evidence{}, err
	}
	switch status {
	case http.StatusOK:
	case http.StatusNotFound:
		// ADR-0047: the same named down store as a missing table on the
		// local store, so the review degrades and names couchdb.
		return Evidence{}, couchFailed("no table " + forms.Find.Table)
	default:
		return Evidence{}, couchDown(status)
	}
	var parsed findResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Evidence{}, couchFailed("bad response")
	}
	rows := make([][]string, 0, len(parsed.Docs))
	for _, doc := range parsed.Docs {
		if len(rows) >= c.cfg.RowLimit {
			break
		}
		// Rows follow the column order of the sql form: CouchDB returns
		// the fields of a document in no fixed order, and omits a field the
		// document lacks. A lacking field is an empty cell.
		row := make([]string, len(forms.Find.Fields))
		for i, f := range forms.Find.Fields {
			cell, cerr := couchCell(doc[f])
			if cerr != nil {
				return Evidence{}, couchFailed("bad response")
			}
			row[i] = cell
		}
		rows = append(rows, row)
	}
	return Evidence{
		Form:     querytpl.FormMango,
		Query:    forms.Mango,
		Engine:   EngineCouchDB,
		Snapshot: c.now().UTC().Format(time.RFC3339),
		Columns:  append([]string(nil), forms.Find.Fields...),
		Rows:     rows,
	}, nil
}

// do sends one request with the credential and returns the status and the
// answer body, read up to maxAnswer. A transport failure is a down store;
// the credential never appears in it.
func (c *CouchDB) do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	if c.cfg.BaseURL == "" || c.cfg.User == "" || c.cfg.Password == "" || c.cfg.Timeout <= 0 || c.cfg.RowLimit < 1 {
		return 0, nil, couchFailed("not configured")
	}
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	// The caller's context is kept apart from the bounded one: a cancel from
	// the caller is returned as itself, and only this call's own deadline
	// reads as a timeout.
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, c.cfg.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, rd)
	if err != nil {
		return 0, nil, couchFailed("not configured")
	}
	req.SetBasicAuth(c.cfg.User, c.cfg.Password)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, c.classify(parent, ctx, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxAnswer))
	if err != nil {
		return 0, nil, c.classify(parent, ctx, err)
	}
	return resp.StatusCode, raw, nil
}

// classify turns a transport error into the caller's own cancel, or a down
// store named couchdb. The error text is not kept: a library may echo the
// request, and the request holds the credential.
func (c *CouchDB) classify(parent, ctx context.Context, err error) error {
	if perr := parent.Err(); perr != nil {
		return perr
	}
	// A transport timeout carries Timeout() true; asking for that method
	// keeps this file to net/http alone.
	var ne interface{ Timeout() bool }
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout()) {
		return couchFailed("timeout")
	}
	return couchFailed("unreachable")
}

// couchDown names a status code as a down store. A 401 and a 403 are named as
// a refusal of the credential; the body is never read into the reason.
func couchDown(status int) error {
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return couchFailed("unauthorized")
	}
	return couchFailed("http " + strconv.Itoa(status))
}

// couchFailed is a failed CouchDB call. The review continues and names the
// service.
func couchFailed(reason string) error {
	return &service.Down{Service: couchService, Reason: reason}
}

// couchCell reads one field of a document as text. A missing field and null
// are the empty text. A string is itself, a number keeps its written form, a
// boolean is true or false. A nested value is refused: the tables Proven
// registers hold text, and a nested value means the database is not one.
func couchCell(raw json.RawMessage) (string, error) {
	t := bytes.TrimSpace(raw)
	if len(t) == 0 || bytes.Equal(t, []byte("null")) {
		return "", nil
	}
	switch t[0] {
	case '"':
		var s string
		if err := json.Unmarshal(t, &s); err != nil {
			return "", err
		}
		return s, nil
	case '{', '[':
		return "", errors.New("nested value")
	}
	var n json.Number
	dec := json.NewDecoder(bytes.NewReader(t))
	dec.UseNumber()
	if err := dec.Decode(&n); err == nil && n != "" {
		return n.String(), nil
	}
	var b bool
	if err := json.Unmarshal(t, &b); err != nil {
		return "", err
	}
	return strconv.FormatBool(b), nil
}
