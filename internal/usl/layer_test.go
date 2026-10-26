// Package usl holds the tests of Proven's semantic layer in Zetaris
// (lab/zetaris/usl, docs/ZETARIS_USL.md). The layer is SQL that an administrator
// applies, so this package has no code of its own: it parses the definitions the
// administrator applies and proves them against sample mode.
package usl_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// root is the repository root, from internal/usl.
var root = filepath.Join("..", "..")

func uslPath(name string) string { return filepath.Join(root, "lab", "zetaris", "usl", name) }

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// statement is one statement of layer.sql, with the comment lines that stand
// directly above it.
type statement struct {
	SQL      string
	Comments []string
}

// parseStatements reads a definitions file the way apply.sh does: a whole-line
// comment, a statement that ends with a semicolon at the end of a line.
func parseStatements(t *testing.T, text string) []statement {
	t.Helper()
	var out []statement
	var comments []string
	var buf []string
	for _, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trim, "--"):
			if len(buf) > 0 {
				t.Fatalf("a comment inside a statement: %q", trim)
			}
			comments = append(comments, trim)
		case trim == "":
			if len(buf) == 0 {
				comments = nil
			}
		default:
			if strings.Contains(trim, "--") {
				t.Fatalf("a comment inside a statement line: %q", trim)
			}
			buf = append(buf, trim)
			if strings.HasSuffix(trim, ";") {
				sql := strings.TrimSuffix(strings.Join(buf, " "), ";")
				out = append(out, statement{SQL: sql, Comments: comments})
				buf = nil
				// The comments above a DROP describe the view its next statement creates.
				if !strings.HasPrefix(sql, "DROP ") {
					comments = nil
				}
			}
		}
	}
	if len(buf) > 0 {
		t.Fatalf("an unterminated statement: %q", strings.Join(buf, " "))
	}
	return out
}

var viewRe = regexp.MustCompile(`^CREATE SCHEMASTORE VIEW (\w+) WITH CONTAINER (\w+) AS (SELECT .+)$`)

// entity is a layer view that is a plain projection of one source table.
type entity struct {
	Name    string // px_usl.px_package
	Source  string
	Columns []string
}

var projectionRe = regexp.MustCompile(`^SELECT (.+) FROM ([\w.]+)$`)

// entities returns the projection views of the layer, in file order, and the
// views that join, which the local engine is not asked to evaluate.
func entities(t *testing.T) (ents []entity, joins []string) {
	t.Helper()
	for _, st := range parseStatements(t, read(t, uslPath("layer.sql"))) {
		m := viewRe.FindStringSubmatch(st.SQL)
		if m == nil {
			continue
		}
		name := m[2] + "." + m[1]
		if strings.Contains(m[3], " JOIN ") {
			joins = append(joins, name)
			continue
		}
		p := projectionRe.FindStringSubmatch(m[3])
		if p == nil {
			t.Fatalf("view %s is neither a plain projection nor a join: %s", name, m[3])
		}
		var cols []string
		for _, c := range strings.Split(p[1], ",") {
			c = strings.TrimSpace(c)
			if !regexp.MustCompile(`^\w+$`).MatchString(c) {
				t.Fatalf("view %s selects %q: an entity view keeps its source's column names", name, c)
			}
			cols = append(cols, c)
		}
		ents = append(ents, entity{Name: name, Source: p[2], Columns: cols})
	}
	return ents, joins
}

// The four entities and the one relation view the order asks for. A fifth
// view, or a missing one, is a change to the layer that this list must follow.
func TestLayerDefinesTheFourEntitiesAndTheRelations(t *testing.T) {
	ents, joins := entities(t)
	var got []string
	for _, e := range ents {
		got = append(got, e.Name+"<-"+e.Source)
	}
	want := []string{
		"px_usl.px_package<-px_osvfile.px_osv_pkg",
		"px_usl.px_advisory<-px_evidence.px_osv_id",
		"px_usl.px_kev_cve<-px_evidence.px_kev_vulns",
		"px_usl.px_epss_score<-px_evidence.px_epss_cve_rows",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("entity views:\n got %v\nwant %v", got, want)
	}
	if len(joins) != 1 || joins[0] != "px_usl.px_package_evidence" {
		t.Fatalf("relation views: got %v, want px_usl.px_package_evidence", joins)
	}
}

// Each entity and relation is commented with its source table and key.
func TestEveryViewNamesItsSourceAndKey(t *testing.T) {
	for _, st := range parseStatements(t, read(t, uslPath("layer.sql"))) {
		m := viewRe.FindStringSubmatch(st.SQL)
		if m == nil {
			continue
		}
		text := strings.Join(st.Comments, "\n")
		if strings.HasPrefix(m[1], "px_package_evidence") {
			for _, rel := range []string{
				"px_package.id = px_advisory.id",
				"px_advisory.alias = cve id",
				"cve id = px_kev_cve.cve_id",
				"cve id = px_epss_score.cve_id",
			} {
				if !strings.Contains(text, rel) {
					t.Errorf("the relation view does not name the join key %q", rel)
				}
			}
			continue
		}
		if !strings.Contains(text, "Source: ") || !strings.Contains(text, "Key: ") {
			t.Errorf("view %s has no Source and Key comment above it", m[1])
		}
	}
}

// A join is an inner join by default (ADR-0023), which drops a CVE with no KEV
// row and reads as "not exploited". The relation view must keep every row.
func TestRelationViewKeepsRowsWithNoMatch(t *testing.T) {
	for _, st := range parseStatements(t, read(t, uslPath("layer.sql"))) {
		if !strings.Contains(st.SQL, "px_package_evidence WITH") {
			continue
		}
		joins := strings.Count(st.SQL, " JOIN ")
		left := strings.Count(st.SQL, " LEFT JOIN ")
		if joins != 3 || left != 3 {
			t.Fatalf("want three joins, all LEFT: %d joins, %d left: %s", joins, left, st.SQL)
		}
		return
	}
	t.Fatal("no px_package_evidence view")
}
