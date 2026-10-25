package store

import (
	"context"
	"fmt"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// Zetaris is the evidence store that sends a SELECT over the REST API
// (ADR-0047, ADR-0015). Behaviour of that call is the live client's.
// Register does not send a statement: on Zetaris, registering a source is
// the administrator's act and Proven's key only reads (ADR-0023, point 2).
// Proven records the table name so a source has registered it.
type Zetaris struct {
	live   *query.Live
	tables map[string]string
}

// NewZetaris builds the store on a validated live client.
// The HTTP client is the one NewLive builds, which does not follow
// redirects. This file does not import net/http: the only file that dials
// Zetaris is internal/query/live.go. The only file of this package that
// dials is couchdb.go, and it dials CouchDB alone.
func NewZetaris(cfg query.Settings) *Zetaris {
	return &Zetaris{live: query.NewLive(cfg, nil), tables: map[string]string{}}
}

// SetClock replaces the clock that stamps a snapshot.
func (z *Zetaris) SetClock(now func() time.Time) {
	if z == nil || z.live == nil {
		return
	}
	z.live.SetClock(now)
}

// Register records the table name and its snapshot.
// It does not dial. A second register of the same name is refused. A query
// still runs against whatever the administrator registered under that name:
// this record is the source's claim, not a remote CREATE.
func (z *Zetaris) Register(ctx context.Context, table Table) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if z == nil || z.live == nil {
		return fmt.Errorf("store: no zetaris store")
	}
	if err := checkTable(table); err != nil {
		return err
	}
	if _, ok := z.tables[table.Name]; ok {
		return fmt.Errorf("store: table %s already exists", table.Name)
	}
	z.tables[table.Name] = table.Snapshot
	return nil
}

// Run builds the sql form and sends that SELECT.
// The mango form is not read: Zetaris speaks SQL (ADR-0047).
func (z *Zetaris) Run(ctx context.Context, req querytpl.Request, rel querytpl.Relations) (Evidence, error) {
	if z == nil || z.live == nil {
		return Evidence{}, fmt.Errorf("store: no zetaris store")
	}
	forms, err := querytpl.BuildForms(req, rel)
	if err != nil {
		return Evidence{}, err
	}
	res, err := z.live.Query(ctx, forms.SQL)
	if err != nil {
		return Evidence{}, err
	}
	return evidenceFrom(forms, res), nil
}

// evidenceFrom copies a query result into the evidence of the form that ran.
func evidenceFrom(forms querytpl.Forms, res query.Result) Evidence {
	return Evidence{
		Form:     querytpl.FormSQL,
		Query:    forms.SQL,
		Engine:   res.Engine,
		Snapshot: res.Snapshot,
		Columns:  res.Columns,
		Rows:     res.Rows,
	}
}
