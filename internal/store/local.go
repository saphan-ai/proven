package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// Local is the one-file store (ADR-0047). It has no server and opens no
// network (ADR-0034). Tables are written into the file as the CREATE and
// INSERT text the sample engine loads, and a query runs on that engine.
type Local struct {
	file *query.File
}

// OpenLocal opens the store at path, creating an empty file when none exists.
func OpenLocal(path string) (*Local, error) {
	f, err := query.CreateFile(path)
	if err != nil {
		return nil, err
	}
	return &Local{file: f}, nil
}

// SetClock replaces the clock that stamps a snapshot.
// A conformance test sets the same instant on the local store and on Zetaris.
func (l *Local) SetClock(now func() time.Time) {
	if l == nil || l.file == nil {
		return
	}
	l.file.SetClock(now)
}

// Register writes the table into the file.
// A second register of the same name is refused. The call does not use the
// network.
func (l *Local) Register(ctx context.Context, table Table) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if l == nil || l.file == nil {
		return fmt.Errorf("store: no local store")
	}
	if err := checkTable(table); err != nil {
		return err
	}
	return l.file.Register(table.Name, table.Columns, table.Rows)
}

// Run builds the sql form and runs it on the file.
// The mango form is not read: this store speaks SQL (ADR-0047).
func (l *Local) Run(ctx context.Context, req querytpl.Request, rel querytpl.Relations) (Evidence, error) {
	if l == nil || l.file == nil {
		return Evidence{}, fmt.Errorf("store: no local store")
	}
	forms, err := querytpl.BuildForms(req, rel)
	if err != nil {
		return Evidence{}, err
	}
	res, err := l.file.Query(ctx, forms.SQL)
	if err != nil {
		return Evidence{}, err
	}
	return evidenceFrom(forms, res), nil
}

// checkTable refuses a name, a column or a row the file cannot store.
// The check is here so both stores refuse the same table before either
// one writes.
func checkTable(table Table) error {
	if !querytpl.ValidRelation(table.Name) {
		return fmt.Errorf("store: table name is not a plain name")
	}
	if len(table.Columns) == 0 {
		return fmt.Errorf("store: table %s has no columns", table.Name)
	}
	seen := map[string]bool{}
	for _, c := range table.Columns {
		if !querytpl.ValidRelation(c) || strings.Contains(c, ".") {
			return fmt.Errorf("store: column name is not a plain name")
		}
		if seen[c] {
			return fmt.Errorf("store: column %s is repeated", c)
		}
		seen[c] = true
	}
	for _, row := range table.Rows {
		if len(row) != len(table.Columns) {
			return fmt.Errorf("store: a row has %d values, table %s has %d columns", len(row), table.Name, len(table.Columns))
		}
	}
	return nil
}
