package query

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/limit"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

// File is the local evidence store: one file of tables and no server
// (ADR-0047). The file is the CREATE TABLE and INSERT text the sample
// engine already loads, so the SQL dialect is the one sample mode uses.
// Nothing in this type dials a network (ADR-0034).
// Engine is the name a result carries. Sample mode passes "local-sample"
// so an existing sample review still reads that way; a configured store
// passes "local".
type File struct {
	path   string
	db     *DB
	engine string
	now    func() time.Time
	mu     sync.Mutex
}

// CreateFile opens the local store at path, creating an empty file when
// none exists. An existing file is loaded. A directory is refused. The
// result engine is "local" and the snapshot clock is the wall clock.
func CreateFile(path string) (*File, error) {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return nil, fmt.Errorf("query: %s is a directory", path)
	}
	if os.IsNotExist(err) {
		// 0600 before umask: the file holds source rows, and it is not a
		// checkout a container has to read.
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return openFile(path, "local", time.Now)
}

// openFile loads path. The file must already exist. now nil leaves the
// snapshot empty, which is what a fixture review expects.
func openFile(path, engine string, now func() time.Time) (*File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	db, err := LoadSQL(string(b))
	if err != nil {
		return nil, err
	}
	return &File{path: path, db: db, engine: engine, now: now}, nil
}

// SetClock replaces the clock that stamps a snapshot.
// A test sets one instant on the local store and on Zetaris. A nil clock
// leaves the snapshot empty.
func (f *File) SetClock(now func() time.Time) {
	if f == nil {
		return
	}
	f.mu.Lock()
	f.now = now
	f.mu.Unlock()
}

// Register writes one table into the file and reloads it.
// A second register of the same name is refused, and the file is left as
// it was. The write is a new file renamed into place, so a failed check
// never leaves a half-written statement.
func (f *File) Register(name string, cols []string, rows [][]string) error {
	if f == nil || f.db == nil {
		return fmt.Errorf("query: no local store")
	}
	stmt, err := tableSQL(name, cols, rows)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.db.tables[name]; ok {
		return fmt.Errorf("query: table %s already exists", name)
	}
	old, err := os.ReadFile(f.path)
	if err != nil {
		return err
	}
	next := string(old)
	if len(old) > 0 && !strings.HasSuffix(next, "\n") {
		next += "\n"
	}
	next += stmt
	// Check the whole file before replacing it. A bad statement must not
	// become the store.
	db, err := LoadSQL(next)
	if err != nil {
		return err
	}
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, []byte(next), 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, f.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	f.db = db
	return nil
}

// Query runs one SELECT on the loaded tables.
// The form is sql. A fixture store leaves Snapshot empty; a configured
// store stamps the clock.
func (f *File) Query(ctx context.Context, sql string) (Result, error) {
	// ADR-0001, point 3: every adapter call is bounded, the local one too.
	ctx, cancel := limit.Bound(ctx)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if f == nil || f.db == nil {
		return Result{}, fmt.Errorf("query: no local store")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	cols, rows, err := f.db.Query(sql)
	if err != nil {
		// ADR-0047: in the configured local store a missing table is a down
		// source, as it is on Zetaris, so a review names it and goes on. The
		// sample engine keeps the plain error its fixtures expect.
		var missing *NoTableError
		if f.engine == "local" && errors.As(err, &missing) {
			return Result{}, &service.Down{Service: "local", Reason: "no table " + missing.Name}
		}
		return Result{}, err
	}
	// No match is an empty list, not nil, so the evidence prints [] and not null.
	if rows == nil {
		rows = [][]string{}
	}
	snap := ""
	if f.now != nil {
		snap = f.now().UTC().Format(time.RFC3339)
	}
	return Result{
		SQL: sql, Form: querytpl.FormSQL, Engine: f.engine,
		Snapshot: snap, Columns: cols, Rows: rows,
	}, nil
}

// tableSQL renders one table as the statements LoadSQL accepts.
// Cells are quoted and a single quote inside a cell is doubled, because
// this is stored data, not a value a model supplied. A model value still
// goes through querytpl and is refused rather than escaped (ADR-0032).
func tableSQL(name string, cols []string, rows [][]string) (string, error) {
	if !querytpl.ValidRelation(name) {
		return "", fmt.Errorf("query: table name is not a plain name")
	}
	if len(cols) == 0 {
		return "", fmt.Errorf("query: table %s has no columns", name)
	}
	seen := map[string]bool{}
	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString(name)
	b.WriteString(" (")
	for i, c := range cols {
		if !plainIdent(c) {
			return "", fmt.Errorf("query: column name is not a plain name")
		}
		if seen[c] {
			return "", fmt.Errorf("query: column %s is repeated", c)
		}
		seen[c] = true
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c)
		b.WriteString(" TEXT")
	}
	b.WriteString(");\n")
	for _, row := range rows {
		if len(row) != len(cols) {
			return "", fmt.Errorf("query: a row has %d values, table %s has %d columns", len(row), name, len(cols))
		}
		b.WriteString("INSERT INTO ")
		b.WriteString(name)
		b.WriteString(" VALUES (")
		for i, cell := range row {
			lit, err := sqlText(cell)
			if err != nil {
				return "", err
			}
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(lit)
		}
		b.WriteString(");\n")
	}
	return b.String(), nil
}

// plainIdent reports whether s is one identifier, not schema.name.
func plainIdent(s string) bool {
	return querytpl.ValidRelation(s) && !strings.Contains(s, ".")
}

// sqlText quotes a cell for the sample engine. A cell that is not UTF-8,
// or that holds a NUL, is refused: the file is text.
func sqlText(s string) (string, error) {
	if !utf8.ValidString(s) || strings.IndexByte(s, 0) >= 0 {
		return "", fmt.Errorf("query: a cell is not text")
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'", nil
}
