// Package query runs SQL for the investigator.
package query

import (
	"fmt"
	"strings"
	"unicode"
)

type tokenKind int

const (
	kindIdent tokenKind = iota
	kindString
	kindSym
)

type token struct {
	kind tokenKind
	val  string
}

// DB is an in-memory table store for the sample queries.
type DB struct {
	tables map[string]*table
}

type table struct {
	cols []string
	rows [][]string
}

// LoadSQL runs CREATE TABLE and INSERT statements.
func LoadSQL(sql string) (*DB, error) {
	db := &DB{tables: map[string]*table{}}
	for _, stmt := range splitStatements(sql) {
		toks, err := tokenize(stmt)
		if err != nil {
			return nil, err
		}
		if len(toks) == 0 {
			continue
		}
		switch strings.ToUpper(toks[0].val) {
		case "CREATE":
			if err := db.execCreate(toks); err != nil {
				return nil, err
			}
		case "INSERT":
			if err := db.execInsert(toks); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("query: %s is not allowed while loading tables", toks[0].val)
		}
	}
	return db, nil
}

// Query runs one SELECT and returns the projected rows.
func (db *DB) Query(sql string) ([]string, [][]string, error) {
	stmts := splitStatements(sql)
	if len(stmts) != 1 {
		return nil, nil, fmt.Errorf("query: want one statement")
	}
	toks, err := tokenize(stmts[0])
	if err != nil {
		return nil, nil, err
	}
	if len(toks) == 0 || !strings.EqualFold(toks[0].val, "SELECT") {
		return nil, nil, fmt.Errorf("query: want a SELECT")
	}
	return db.execSelect(toks)
}

func (db *DB) execCreate(toks []token) error {
	// CREATE TABLE name ( col type , ... )
	if len(toks) < 4 || !strings.EqualFold(toks[1].val, "TABLE") {
		return fmt.Errorf("query: bad CREATE TABLE")
	}
	name := toks[2].val
	if toks[2].kind != kindIdent || toks[3].val != "(" {
		return fmt.Errorf("query: bad CREATE TABLE")
	}
	if _, ok := db.tables[name]; ok {
		return fmt.Errorf("query: table %s already exists", name)
	}
	var cols []string
	i := 4
	for i < len(toks) && toks[i].val != ")" {
		if toks[i].kind != kindIdent {
			return fmt.Errorf("query: bad column in CREATE TABLE")
		}
		cols = append(cols, toks[i].val)
		i++
		if i < len(toks) && toks[i].kind == kindIdent {
			i++ // type name, ignored
		}
		if i < len(toks) && toks[i].val == "," {
			i++
		}
	}
	if len(cols) == 0 {
		return fmt.Errorf("query: CREATE TABLE has no columns")
	}
	db.tables[name] = &table{cols: cols}
	return nil
}

func (db *DB) execInsert(toks []token) error {
	// INSERT INTO name VALUES ( 'a' , 'b' )
	if len(toks) < 5 || !strings.EqualFold(toks[1].val, "INTO") || !strings.EqualFold(toks[3].val, "VALUES") {
		return fmt.Errorf("query: bad INSERT")
	}
	name := toks[2].val
	tb, ok := db.tables[name]
	if !ok {
		return fmt.Errorf("query: no table %s", name)
	}
	if toks[4].val != "(" {
		return fmt.Errorf("query: bad INSERT")
	}
	var vals []string
	for i := 5; i < len(toks) && toks[i].val != ")"; i++ {
		if toks[i].val == "," {
			continue
		}
		if toks[i].kind != kindString {
			return fmt.Errorf("query: INSERT values must be strings")
		}
		vals = append(vals, toks[i].val)
	}
	if len(vals) != len(tb.cols) {
		return fmt.Errorf("query: INSERT has %d values, table %s has %d columns", len(vals), name, len(tb.cols))
	}
	tb.rows = append(tb.rows, vals)
	return nil
}

func (db *DB) execSelect(toks []token) ([]string, [][]string, error) {
	var cols []string
	i := 1
	for i < len(toks) && !strings.EqualFold(toks[i].val, "FROM") {
		if toks[i].val == "," {
			i++
			continue
		}
		if toks[i].kind != kindIdent {
			return nil, nil, fmt.Errorf("query: bad select list")
		}
		cols = append(cols, toks[i].val)
		i++
	}
	if len(cols) == 0 || i >= len(toks) {
		return nil, nil, fmt.Errorf("query: SELECT needs FROM")
	}
	i++ // skip FROM
	if i >= len(toks) || toks[i].kind != kindIdent {
		return nil, nil, fmt.Errorf("query: SELECT needs a table")
	}
	tb, ok := db.tables[toks[i].val]
	if !ok {
		return nil, nil, fmt.Errorf("query: no table %s", toks[i].val)
	}
	i++
	var where []pred
	if i < len(toks) {
		if !strings.EqualFold(toks[i].val, "WHERE") {
			return nil, nil, fmt.Errorf("query: unexpected %s", toks[i].val)
		}
		i++
		for {
			if i+2 >= len(toks) || toks[i].kind != kindIdent || toks[i+1].val != "=" || toks[i+2].kind != kindString {
				return nil, nil, fmt.Errorf("query: bad WHERE")
			}
			where = append(where, pred{toks[i].val, toks[i+2].val})
			i += 3
			if i >= len(toks) {
				break
			}
			if !strings.EqualFold(toks[i].val, "AND") {
				return nil, nil, fmt.Errorf("query: only AND is supported")
			}
			i++
		}
	}
	colIndex := map[string]int{}
	for n, c := range tb.cols {
		colIndex[c] = n
	}
	proj := make([]int, len(cols))
	for n, c := range cols {
		idx, ok := colIndex[c]
		if !ok {
			return nil, nil, fmt.Errorf("query: no column %s", c)
		}
		proj[n] = idx
	}
	for _, p := range where {
		if _, ok := colIndex[p.col]; !ok {
			return nil, nil, fmt.Errorf("query: no column %s", p.col)
		}
	}
	var rows [][]string
	for _, row := range tb.rows {
		if !match(row, colIndex, where) {
			continue
		}
		out := make([]string, len(proj))
		for n, idx := range proj {
			out[n] = row[idx]
		}
		rows = append(rows, out)
	}
	return append([]string(nil), cols...), rows, nil
}

type pred struct{ col, val string }

func match(row []string, colIndex map[string]int, where []pred) bool {
	for _, p := range where {
		if row[colIndex[p.col]] != p.val {
			return false
		}
	}
	return true
}

func splitStatements(sql string) []string {
	var out []string
	var b strings.Builder
	in := false
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		if c == '\'' {
			if in && i+1 < len(sql) && sql[i+1] == '\'' {
				b.WriteByte('\'')
				b.WriteByte('\'')
				i++
				continue
			}
			in = !in
			b.WriteByte(c)
			continue
		}
		if c == ';' && !in {
			if stmt := strings.TrimSpace(b.String()); stmt != "" {
				out = append(out, stmt)
			}
			b.Reset()
			continue
		}
		b.WriteByte(c)
	}
	if stmt := strings.TrimSpace(b.String()); stmt != "" {
		out = append(out, stmt)
	}
	return out
}

func tokenize(s string) ([]token, error) {
	var out []token
	i := 0
	for i < len(s) {
		r := rune(s[i])
		if unicode.IsSpace(r) {
			i++
			continue
		}
		if s[i] == '\'' {
			var b strings.Builder
			j := i + 1
			closed := false
			for j < len(s) {
				if s[j] == '\'' {
					if j+1 < len(s) && s[j+1] == '\'' {
						b.WriteByte('\'')
						j += 2
						continue
					}
					closed = true
					j++
					break
				}
				b.WriteByte(s[j])
				j++
			}
			if !closed {
				return nil, fmt.Errorf("query: unterminated string")
			}
			out = append(out, token{kind: kindString, val: b.String()})
			i = j
			continue
		}
		if isIdent(s[i]) && (len(out) == 0 || !unicode.IsDigit(r)) {
			j := i + 1
			for j < len(s) && (isIdent(s[j]) || unicode.IsDigit(rune(s[j]))) {
				j++
			}
			out = append(out, token{kind: kindIdent, val: s[i:j]})
			i = j
			continue
		}
		switch s[i] {
		case '(', ')', ',', '=', '*':
			out = append(out, token{kind: kindSym, val: s[i : i+1]})
			i++
		default:
			return nil, fmt.Errorf("query: bad character %q", s[i])
		}
	}
	return out, nil
}

func isIdent(b byte) bool {
	return b == '_' || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
