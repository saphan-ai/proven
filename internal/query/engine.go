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
	// name may be schema.table, the form the evidence views use.
	if len(toks) < 4 || !strings.EqualFold(toks[1].val, "TABLE") {
		return fmt.Errorf("query: bad CREATE TABLE")
	}
	name, i, err := takeQualified(toks, 2)
	if err != nil || i >= len(toks) || toks[i].val != "(" {
		return fmt.Errorf("query: bad CREATE TABLE")
	}
	if _, ok := db.tables[name]; ok {
		return fmt.Errorf("query: table %s already exists", name)
	}
	var cols []string
	i++
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
	if len(toks) < 4 || !strings.EqualFold(toks[1].val, "INTO") {
		return fmt.Errorf("query: bad INSERT")
	}
	name, i, err := takeQualified(toks, 2)
	if err != nil || i >= len(toks) || !strings.EqualFold(toks[i].val, "VALUES") {
		return fmt.Errorf("query: bad INSERT")
	}
	tb, ok := db.tables[name]
	if !ok {
		return fmt.Errorf("query: no table %s", name)
	}
	i++
	if i >= len(toks) || toks[i].val != "(" {
		return fmt.Errorf("query: bad INSERT")
	}
	var vals []string
	for i = i + 1; i < len(toks) && toks[i].val != ")"; i++ {
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
	cols, i, err := parseSelectList(toks, 1)
	if err != nil {
		return nil, nil, err
	}
	i++ // skip FROM
	leftRef, i, err := parseTableRef(toks, i)
	if err != nil {
		return nil, nil, err
	}
	var rightRef *tableRef
	var onL, onR colRef
	if i < len(toks) && strings.EqualFold(toks[i].val, "INNER") {
		i++
		if i >= len(toks) || !strings.EqualFold(toks[i].val, "JOIN") {
			return nil, nil, fmt.Errorf("query: expected JOIN")
		}
	}
	if i < len(toks) && strings.EqualFold(toks[i].val, "JOIN") {
		i++
		r, next, err := parseTableRef(toks, i)
		if err != nil {
			return nil, nil, err
		}
		rightRef = &r
		i = next
		if i >= len(toks) || !strings.EqualFold(toks[i].val, "ON") {
			return nil, nil, fmt.Errorf("query: JOIN needs ON")
		}
		i++
		onL, i, err = takeColRef(toks, i)
		if err != nil || i >= len(toks) || toks[i].val != "=" {
			return nil, nil, fmt.Errorf("query: bad ON")
		}
		i++
		onR, i, err = takeColRef(toks, i)
		if err != nil {
			return nil, nil, fmt.Errorf("query: bad ON")
		}
	}
	var where []wherePred
	if i < len(toks) {
		if !strings.EqualFold(toks[i].val, "WHERE") {
			return nil, nil, fmt.Errorf("query: unexpected %s", toks[i].val)
		}
		i++
		for {
			c, next, err := takeColRef(toks, i)
			if err != nil || next >= len(toks) || toks[next].val != "=" || next+1 >= len(toks) || toks[next+1].kind != kindString {
				return nil, nil, fmt.Errorf("query: bad WHERE")
			}
			where = append(where, wherePred{col: c, val: toks[next+1].val})
			i = next + 2
			if i >= len(toks) {
				break
			}
			if !strings.EqualFold(toks[i].val, "AND") {
				return nil, nil, fmt.Errorf("query: only AND is supported")
			}
			i++
		}
	}
	left, err := db.bind(leftRef)
	if err != nil {
		return nil, nil, err
	}
	var right *side
	if rightRef != nil {
		s, err := db.bind(*rightRef)
		if err != nil {
			return nil, nil, err
		}
		right = &s
		if err := mustResolve(onL, &left, right); err != nil {
			return nil, nil, err
		}
		if err := mustResolve(onR, &left, right); err != nil {
			return nil, nil, err
		}
	}
	for _, c := range cols {
		if err := mustResolve(c, &left, right); err != nil {
			return nil, nil, err
		}
	}
	for _, p := range where {
		if err := mustResolve(p.col, &left, right); err != nil {
			return nil, nil, err
		}
	}
	names := make([]string, len(cols))
	for n, c := range cols {
		names[n] = c.name
	}
	var rows [][]string
	if right == nil {
		for _, row := range left.tb.rows {
			if !rowWhere(where, left, row, nil, nil) {
				continue
			}
			rows = append(rows, project(cols, left, row, nil, nil))
		}
		return names, rows, nil
	}
	for _, lr := range left.tb.rows {
		for _, rr := range right.tb.rows {
			lv, err := lookup(onL, left, lr, right, rr)
			if err != nil {
				return nil, nil, err
			}
			rv, err := lookup(onR, left, lr, right, rr)
			if err != nil {
				return nil, nil, err
			}
			if lv != rv {
				continue
			}
			if !rowWhere(where, left, lr, right, rr) {
				continue
			}
			rows = append(rows, project(cols, left, lr, right, rr))
		}
	}
	return names, rows, nil
}

type colRef struct {
	qual string
	name string
}

type wherePred struct {
	col colRef
	val string
}

type tableRef struct {
	name  string
	alias string
}

type side struct {
	ref tableRef
	tb  *table
	idx map[string]int
}

func parseSelectList(toks []token, i int) ([]colRef, int, error) {
	var cols []colRef
	for i < len(toks) && !strings.EqualFold(toks[i].val, "FROM") {
		if toks[i].val == "," {
			i++
			continue
		}
		c, next, err := takeColRef(toks, i)
		if err != nil {
			return nil, i, fmt.Errorf("query: bad select list")
		}
		cols = append(cols, c)
		i = next
	}
	if len(cols) == 0 || i >= len(toks) {
		return nil, i, fmt.Errorf("query: SELECT needs FROM")
	}
	return cols, i, nil
}

func parseTableRef(toks []token, i int) (tableRef, int, error) {
	name, i, err := takeQualified(toks, i)
	if err != nil {
		return tableRef{}, i, fmt.Errorf("query: SELECT needs a table")
	}
	ref := tableRef{name: name}
	if i < len(toks) && toks[i].kind == kindIdent && !isKeyword(toks[i].val) {
		ref.alias = toks[i].val
		i++
	}
	return ref, i, nil
}

func takeQualified(toks []token, i int) (string, int, error) {
	if i >= len(toks) || toks[i].kind != kindIdent {
		return "", i, fmt.Errorf("query: expected a name")
	}
	name := toks[i].val
	i++
	for i+1 < len(toks) && toks[i].val == "." && toks[i+1].kind == kindIdent {
		name += "." + toks[i+1].val
		i += 2
	}
	return name, i, nil
}

func takeColRef(toks []token, i int) (colRef, int, error) {
	name, i, err := takeQualified(toks, i)
	if err != nil {
		return colRef{}, i, err
	}
	parts := strings.Split(name, ".")
	switch len(parts) {
	case 1:
		return colRef{name: parts[0]}, i, nil
	case 2:
		return colRef{qual: parts[0], name: parts[1]}, i, nil
	default:
		return colRef{}, i, fmt.Errorf("query: bad column %s", name)
	}
}

func isKeyword(s string) bool {
	switch strings.ToUpper(s) {
	case "WHERE", "JOIN", "INNER", "ON", "AND", "SELECT", "FROM", "ORDER", "LIMIT", "GROUP", "LEFT", "RIGHT":
		return true
	default:
		return false
	}
}

func (db *DB) bind(ref tableRef) (side, error) {
	tb, ok := db.tables[ref.name]
	if !ok {
		return side{}, fmt.Errorf("query: no table %s", ref.name)
	}
	idx := make(map[string]int, len(tb.cols))
	for n, c := range tb.cols {
		idx[c] = n
	}
	return side{ref: ref, tb: tb, idx: idx}, nil
}

func (s side) matches(qual string) bool {
	if qual == "" {
		return false
	}
	if s.ref.alias != "" && s.ref.alias == qual {
		return true
	}
	if s.ref.name == qual {
		return true
	}
	if i := strings.LastIndex(s.ref.name, "."); i >= 0 && s.ref.name[i+1:] == qual {
		return true
	}
	return false
}

func (s side) has(ref colRef) bool {
	if ref.qual != "" && !s.matches(ref.qual) {
		return false
	}
	_, ok := s.idx[ref.name]
	return ok
}

func mustResolve(ref colRef, left, right *side) error {
	lok := left.has(ref)
	rok := right != nil && right.has(ref)
	if lok && rok {
		return fmt.Errorf("query: ambiguous column %s", ref.name)
	}
	if !lok && !rok {
		label := ref.name
		if ref.qual != "" {
			label = ref.qual + "." + ref.name
		}
		return fmt.Errorf("query: no column %s", label)
	}
	return nil
}

func lookup(ref colRef, left side, leftRow []string, right *side, rightRow []string) (string, error) {
	var lv, rv string
	lok, rok := false, false
	if left.has(ref) {
		lv = leftRow[left.idx[ref.name]]
		lok = true
	}
	if right != nil && right.has(ref) {
		rv = rightRow[right.idx[ref.name]]
		rok = true
	}
	if lok && rok {
		return "", fmt.Errorf("query: ambiguous column %s", ref.name)
	}
	if lok {
		return lv, nil
	}
	if rok {
		return rv, nil
	}
	return "", fmt.Errorf("query: no column %s", ref.name)
}

func rowWhere(where []wherePred, left side, leftRow []string, right *side, rightRow []string) bool {
	for _, p := range where {
		got, err := lookup(p.col, left, leftRow, right, rightRow)
		if err != nil || got != p.val {
			return false
		}
	}
	return true
}

func project(cols []colRef, left side, leftRow []string, right *side, rightRow []string) []string {
	out := make([]string, len(cols))
	for n, c := range cols {
		v, err := lookup(c, left, leftRow, right, rightRow)
		if err != nil {
			out[n] = ""
			continue
		}
		out[n] = v
	}
	return out
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
		case '(', ')', ',', '=', '*', '.':
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
