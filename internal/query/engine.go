// Package query runs the Investigator's SQL. It holds the one Adapter
// interface and its two implementations: Live sends a SELECT to Zetaris over
// REST (ADR-0015), and Sample runs the same SQL on a small local engine over
// sample tables, so sample mode and the tests need no network (ADR-0002,
// point 4). The EPSS and OSV steps of the pipeline query through the same
// Adapter. The Judge keeps the SQL and its rows as evidence.
package query

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// tokenKind is what a token is: a name or keyword, a quoted string, an
// integer, or one symbol.
type tokenKind int

const (
	kindIdent tokenKind = iota
	kindString
	kindSym
	kindNumber
)

// token is one lexical unit of a statement. A string's val has its quotes
// removed and each doubled quote turned back into one.
type token struct {
	kind tokenKind
	val  string
}

// DB is an in-memory table store for the sample queries.
// It reads only the small SQL the samples and the evidence queries use:
// CREATE TABLE, INSERT of quoted strings, and a SELECT with at most one inner
// join, a WHERE of equalities joined by AND, an integer in the select list,
// and LIMIT.
type DB struct {
	tables map[string]*table
}

// table keeps every cell as text, the form the live adapter also returns.
type table struct {
	cols []string
	rows [][]string
}

// LoadSQL runs CREATE TABLE and INSERT statements.
// Any other statement is refused: a sample file only declares tables and
// their rows.
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
// More than one statement, or anything but a SELECT, is refused, so a query
// cannot change the loaded tables.
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

// execCreate defines an empty table. Column types are read and ignored,
// because every cell is text.
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
	// A second CREATE of the same name is refused, so it cannot silently drop
	// the rows already loaded.
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

// execInsert appends one row of quoted values to a table created before it.
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
		// Cells are text, so only quoted values are taken.
		if toks[i].kind != kindString {
			return fmt.Errorf("query: INSERT values must be strings")
		}
		vals = append(vals, toks[i].val)
	}
	// A row with too few or too many values is refused: a missing cell would
	// shift every column after it.
	if len(vals) != len(tb.cols) {
		return fmt.Errorf("query: INSERT has %d values, table %s has %d columns", len(vals), name, len(tb.cols))
	}
	tb.rows = append(tb.rows, vals)
	return nil
}

// execSelect runs
// SELECT cols FROM t [a] [[INNER] JOIN u [b] ON x = y] [WHERE c = 'v' [AND ...]] [LIMIT n].
// A select item may be a column or an integer literal.
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
	// limit < 0 means the query set no LIMIT. Zero is a real limit: no row.
	limit := -1
	if i < len(toks) && strings.EqualFold(toks[i].val, "WHERE") {
		i++
		for {
			c, next, err := takeColRef(toks, i)
			if err != nil || next >= len(toks) || toks[next].val != "=" || next+1 >= len(toks) || toks[next+1].kind != kindString {
				return nil, nil, fmt.Errorf("query: bad WHERE")
			}
			where = append(where, wherePred{col: c, val: toks[next+1].val})
			i = next + 2
			if i >= len(toks) || strings.EqualFold(toks[i].val, "LIMIT") {
				break
			}
			// Only AND: a predicate this engine cannot read is refused, not half-read.
			if !strings.EqualFold(toks[i].val, "AND") {
				return nil, nil, fmt.Errorf("query: only AND is supported")
			}
			i++
		}
	}
	if i < len(toks) && strings.EqualFold(toks[i].val, "LIMIT") {
		// LIMIT is applied, not ignored: a query must not return rows it did
		// not ask for, and it must be able to ask for one.
		i++
		if i >= len(toks) || toks[i].kind != kindNumber {
			return nil, nil, fmt.Errorf("query: LIMIT needs an integer")
		}
		n, nerr := strconv.Atoi(toks[i].val)
		if nerr != nil {
			return nil, nil, fmt.Errorf("query: LIMIT needs an integer")
		}
		limit = n
		i++
	}
	// ORDER BY and anything else are refused rather than ignored, so a query
	// never returns rows it did not ask for.
	if i < len(toks) {
		return nil, nil, fmt.Errorf("query: unexpected %s", toks[i].val)
	}
	// Every column is resolved before any row is read. A wrong or ambiguous name
	// is an error, never an empty result, because an empty result would read as
	// "no evidence".
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
			if limit >= 0 && len(rows) >= limit {
				break
			}
			if !rowWhere(where, left, row, nil, nil) {
				continue
			}
			rows = append(rows, project(cols, left, row, nil, nil))
		}
		return names, rows, nil
	}
	// A nested-loop inner join. The sample tables are small, so no index is
	// built.
	for _, lr := range left.tb.rows {
		if limit >= 0 && len(rows) >= limit {
			break
		}
		for _, rr := range right.tb.rows {
			if limit >= 0 && len(rows) >= limit {
				break
			}
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

// colRef is a column, optionally qualified by a table name or an alias.
// literal is set when the select item is an integer. The cell is the digits,
// and the item is not resolved against a table.
type colRef struct {
	qual    string
	name    string
	literal bool
}

// wherePred is one column = 'value' test of a WHERE clause.
type wherePred struct {
	col colRef
	val string
}

// tableRef is a table named in FROM or JOIN, with its optional alias.
type tableRef struct {
	name  string
	alias string
}

// side is a table bound for one query, with its column positions by name.
type side struct {
	ref tableRef
	tb  *table
	idx map[string]int
}

// parseSelectList reads the column list up to FROM. A star or an expression
// is refused: every output column is named.
func parseSelectList(toks []token, i int) ([]colRef, int, error) {
	var cols []colRef
	for i < len(toks) && !strings.EqualFold(toks[i].val, "FROM") {
		if toks[i].val == "," {
			i++
			continue
		}
		// An integer is a constant column. It is not a name, so it is not
		// resolved against the table: SELECT 1 is a probe, not a column.
		if toks[i].kind == kindNumber {
			cols = append(cols, colRef{name: toks[i].val, literal: true})
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

// parseTableRef reads a table name and an optional alias. A keyword after the
// name is not taken as an alias.
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

// takeQualified reads a dotted name such as schema.table or alias.column.
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

// takeColRef splits a dotted name into a qualifier and a column. A name of
// more than two parts is refused.
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

// isKeyword lists the words that end a table reference, so they are never
// read as an alias. The parser then refuses those it does not support.
func isKeyword(s string) bool {
	switch strings.ToUpper(s) {
	case "WHERE", "JOIN", "INNER", "ON", "AND", "SELECT", "FROM", "ORDER", "LIMIT", "GROUP", "LEFT", "RIGHT":
		return true
	default:
		return false
	}
}

// bind looks up a table by its full name and indexes its columns.
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

// matches reports whether qual names this side: its alias, its full name, or
// the last part of a schema.table name.
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

// has reports whether the column is on this side and its qualifier, if any,
// names this side.
func (s side) has(ref colRef) bool {
	if ref.qual != "" && !s.matches(ref.qual) {
		return false
	}
	_, ok := s.idx[ref.name]
	return ok
}

// mustResolve requires a column to exist on exactly one side of the join.
func mustResolve(ref colRef, left, right *side) error {
	if ref.literal {
		return nil
	}
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

// lookup reads a column's value from the current pair of rows.
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

// rowWhere reports whether a pair of rows passes every WHERE test.
func rowWhere(where []wherePred, left side, leftRow []string, right *side, rightRow []string) bool {
	for _, p := range where {
		got, err := lookup(p.col, left, leftRow, right, rightRow)
		if err != nil || got != p.val {
			return false
		}
	}
	return true
}

// project builds one output row in select-list order. The columns were
// resolved before the scan; a cell that still does not resolve is left empty.
func project(cols []colRef, left side, leftRow []string, right *side, rightRow []string) []string {
	out := make([]string, len(cols))
	for n, c := range cols {
		if c.literal {
			out[n] = c.name
			continue
		}
		v, err := lookup(c, left, leftRow, right, rightRow)
		if err != nil {
			out[n] = ""
			continue
		}
		out[n] = v
	}
	return out
}

// splitStatements splits on semicolons outside quoted strings. A doubled
// quote inside a string is kept as written, for the tokenizer to read.
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

// tokenize turns a statement into tokens. A character the small grammar does
// not use is refused, so an unexpected statement fails loudly instead of
// running in part. A digit is an integer literal: SELECT 1 and LIMIT need one.
func tokenize(s string) ([]token, error) {
	var out []token
	i := 0
	for i < len(s) {
		r := rune(s[i])
		if unicode.IsSpace(r) {
			i++
			continue
		}
		// A digit is an integer literal, not a name. SELECT 1 and LIMIT both
		// need one, and a name cannot start with a digit.
		if s[i] >= '0' && s[i] <= '9' {
			j := i + 1
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}
			out = append(out, token{kind: kindNumber, val: s[i:j]})
			i = j
			continue
		}
		if s[i] == '\'' {
			var b strings.Builder
			j := i + 1
			closed := false
			for j < len(s) {
				if s[j] == '\'' {
					// Two quotes inside a string are one quote, as in standard SQL.
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
		if isIdent(s[i]) {
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

// isIdent reports whether b can start a name: an ASCII letter or an
// underscore.
func isIdent(b byte) bool {
	return b == '_' || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
