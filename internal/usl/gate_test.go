package usl_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// The gate is a SELECT that Zetaris runs, so these tests cannot run its text.
// They run its meaning: violations is the same rule, written in Go from the
// numbers and the pattern that gate.sql itself holds, over the sample rows. What
// Zetaris's own dialect makes of the text (RLIKE, CAST, isnan) is not measured
// here; docs/ZETARIS_USL.md says so.

var gatePatternRe = regexp.MustCompile(`RLIKE '([^']+)'`)

// gatePattern is the cve_id pattern written in gate.sql.
func gatePattern(t *testing.T) *regexp.Regexp {
	t.Helper()
	m := gatePatternRe.FindStringSubmatch(read(t, uslPath("gate.sql")))
	if m == nil {
		t.Fatal("gate.sql has no RLIKE pattern")
	}
	return regexp.MustCompile(m[1])
}

// inUnit reports whether s is a number from 0 to 1. Anything else, including NaN
// and a value that does not parse, is outside, as the gate's CAST makes it.
func inUnit(s string) bool {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return err == nil && !math.IsNaN(f) && f >= 0 && f <= 1
}

// violations returns the rows of px_epss_score that break the rule.
func violations(t *testing.T, rows [][]string) [][]string {
	t.Helper()
	cve := gatePattern(t)
	var out [][]string
	for _, r := range rows {
		if !cve.MatchString(r[0]) || !inUnit(r[1]) || !inUnit(r[2]) {
			out = append(out, r)
		}
	}
	return out
}

func epssRows(t *testing.T, tablesSQL string) [][]string {
	t.Helper()
	db, err := query.LoadSQL(tablesSQL)
	if err != nil {
		t.Fatal(err)
	}
	_, rows, err := db.Query("SELECT cve_id, epss, percentile, score_date FROM px_usl.px_epss_score")
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func fixture(t *testing.T, name string) string {
	return read(t, filepath.Join(root, "testdata", "zetaris-usl", name))
}

// The pattern in gate.sql is Proven's cve_id format.
func TestGatePatternIsProvensCVEFormat(t *testing.T) {
	re := gatePattern(t)
	for _, id := range []string{
		"CVE-2021-44228", "CVE-2026-0001", "CVE-2026-123456789012", "CVE-2026-000", "cve-2026-0001", "CVE-2026-0001 ",
		"CVE-26-0001", "", "CVE-2026-1234567890123", "xCVE-2026-0001", "CVE-2026-0001\n",
	} {
		if re.MatchString(id) != querytpl.ValidCVE(id) {
			t.Errorf("%q: gate.sql says %v, querytpl.ValidCVE says %v", id, re.MatchString(id), querytpl.ValidCVE(id))
		}
	}
}

func TestGateCleanFixturePasses(t *testing.T) {
	if v := violations(t, epssRows(t, fixture(t, "epss-clean.sql"))); len(v) != 0 {
		t.Fatalf("the clean fixture returned %v", v)
	}
}

// Law #27. The violating fixture differs from the clean one in one field: the
// epss of CVE-2026-0001, 0.50000 in the clean fixture and 42 here.
func TestGateReturnsTheViolatingRow(t *testing.T) {
	clean, bad := fixture(t, "epss-clean.sql"), fixture(t, "epss-violating.sql")
	if strings.Replace(clean, "'0.50000'", "'42'", 1) != bad {
		t.Fatal("the violating fixture is not the clean fixture with the epss of CVE-2026-0001 changed to 42")
	}
	v := violations(t, epssRows(t, bad))
	if len(v) != 1 || v[0][0] != "CVE-2026-0001" || v[0][1] != "42" {
		t.Fatalf("want exactly the row CVE-2026-0001 with epss 42, got %v", v)
	}
}

// Each field the rule guards, mutated alone on the clean fixture, makes the gate
// return exactly that one row.
func TestGateSeesEachGuardedFieldAlone(t *testing.T) {
	clean := fixture(t, "epss-clean.sql")
	row := "('CVE-2026-0001', '0.50000', '0.90000', '2026-10-20')"
	if !strings.Contains(clean, row) {
		t.Fatal("the clean fixture lost its row")
	}
	for _, c := range []struct{ field, row string }{
		{"cve_id: wrong shape", "('CVE-26-1', '0.50000', '0.90000', '2026-10-20')"},
		{"cve_id: empty", "('', '0.50000', '0.90000', '2026-10-20')"},
		{"epss: above 1", "('CVE-2026-0001', '1.01', '0.90000', '2026-10-20')"},
		{"epss: below 0", "('CVE-2026-0001', '-0.1', '0.90000', '2026-10-20')"},
		{"epss: not a number", "('CVE-2026-0001', 'NaN', '0.90000', '2026-10-20')"},
		{"epss: empty", "('CVE-2026-0001', '', '0.90000', '2026-10-20')"},
		{"percentile: above 1", "('CVE-2026-0001', '0.50000', '2', '2026-10-20')"},
		{"percentile: not a number", "('CVE-2026-0001', '0.50000', 'high', '2026-10-20')"},
	} {
		v := violations(t, epssRows(t, strings.Replace(clean, row, c.row, 1)))
		if len(v) != 1 {
			t.Errorf("%s: the gate returned %d rows, want 1", c.field, len(v))
		}
	}
}

// stub is a directory with a curl that records every call and answers from files.
type stub struct{ dir string }

const stubCurl = `#!/bin/sh
printf 'called\n' >> "$STUB_DIR/calls"
printf '%s\n' "$*" >> "$STUB_DIR/argv"
out=
data=
while [ $# -gt 0 ]; do
  case $1 in
    -o) out=$2; shift ;;
    --data-binary) data=${2#@}; shift ;;
  esac
  shift
done
cat > "$STUB_DIR/config"
cat "$data" >> "$STUB_DIR/requests"
printf '\n' >> "$STUB_DIR/requests"
if [ -f "$STUB_DIR/transport_error" ]; then
  cat "$STUB_DIR/transport_error" >&2
  exit 7
fi
body=$(cat "$data")
case $body in
  *'COUNT(*) AS violations'*)
    cp "$STUB_DIR/count.json" "$out"
    if [ -f "$STUB_DIR/gate_status" ]; then cat "$STUB_DIR/gate_status"; else printf 200; fi
    exit 0 ;;
  *'LIMIT 20'*) cp "$STUB_DIR/rows.json" "$out"; printf 200; exit 0 ;;
esac
if [ -f "$STUB_DIR/refuse_on" ] && printf '%s' "$body" | grep -qF -- "$(cat "$STUB_DIR/refuse_on")"; then
  cp "$STUB_DIR/refusal.json" "$out"
  printf 400
  exit 0
fi
printf '{"headers":[],"data":[]}' > "$out"
printf 200
`

const stubFails = `#!/bin/sh
printf 'called\n' >> "$STUB_DIR/calls"
exit 99
`

func newStub(t *testing.T, script string) stub {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh on the path")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "curl"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	s := stub{dir: dir}
	s.set(t, "count.json", `{"headers":["violations"],"data":[[0]]}`)
	s.set(t, "rows.json", `{"headers":[],"data":[]}`)
	return s
}

func (s stub) set(t *testing.T, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(s.dir, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (s stub) file(name string) string {
	b, _ := os.ReadFile(filepath.Join(s.dir, name))
	return string(b)
}

func (s stub) calls() int { return strings.Count(s.file("calls"), "called") }

const testKey = "k3y-SECRET-0123456789"

var allSettings = map[string]string{
	query.EnvRESTURL: "http://zetaris.invalid:8888/api/v1.0",
	query.EnvOrgID:   "123",
	query.EnvAPIKey:  testKey,
}

// apply runs apply.sh with exactly the environment given: the developer's own
// Zetaris settings are never passed down.
func (s stub) apply(t *testing.T, env map[string]string) (stdout, stderr string, code int) {
	t.Helper()
	script, err := filepath.Abs(uslPath("apply.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", script)
	cmd.Dir = t.TempDir()
	cmd.Env = []string{"PATH=" + s.dir + string(os.PathListSeparator) + os.Getenv("PATH"), "TMPDIR=" + t.TempDir(), "STUB_DIR=" + s.dir}
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	err = cmd.Run()
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		code = ee.ExitCode()
	default:
		t.Fatal(err)
	}
	return o.String(), e.String(), code
}

func countOf(t *testing.T, rows [][]string) string {
	b, err := json.Marshal(map[string]any{"headers": []string{"violations"}, "data": [][]int{{len(rows)}}})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const settingsLines = "PROVEN_OSV_TABLE=px_usl.px_package\nPROVEN_KEV_VIEW=px_usl.px_kev_cve\nPROVEN_EPSS_VIEW=px_usl.px_epss_score\nPROVEN_OSV_ID_VIEW=px_usl.px_advisory\n"

// Law #27. The gate returns the violating row, and apply.sh refuses naming the
// rule and the count, prints no settings, and does not leak the key.
// Mutated field: the epss of CVE-2026-0001 (epss-violating.sql).
func TestApplyRefusesWhenTheGateReturnsARow(t *testing.T) {
	s := newStub(t, stubCurl)
	rows := violations(t, epssRows(t, fixture(t, "epss-violating.sql")))
	s.set(t, "count.json", countOf(t, rows))
	body, _ := json.Marshal(map[string]any{"headers": []string{"cve_id", "epss", "percentile", "score_date"}, "data": rows})
	s.set(t, "rows.json", string(body))
	stdout, stderr, code := s.apply(t, allSettings)
	if code != 4 {
		t.Fatalf("exit %d, want 4\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	for _, want := range []string{"usl-epss-score-valid", "1 violating row", "CVE-2026-0001"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr does not say %q:\n%s", want, stderr)
		}
	}
	if strings.Contains(stdout, "PROVEN_") {
		t.Errorf("the settings were printed after a refusal:\n%s", stdout)
	}
	if strings.Contains(stdout+stderr, testKey) {
		t.Error("the key was printed")
	}
}

// The clean fixture passes: every statement is sent, in order, then the gate,
// and the settings are printed.
func TestApplyAppliesTheLayerWhenTheGatePasses(t *testing.T) {
	s := newStub(t, stubCurl)
	s.set(t, "count.json", countOf(t, violations(t, epssRows(t, fixture(t, "epss-clean.sql")))))
	stdout, stderr, code := s.apply(t, allSettings)
	if code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	stmts := parseStatements(t, read(t, uslPath("layer.sql")))
	if got, want := s.calls(), len(stmts)+1; got != want {
		t.Errorf("%d calls, want %d statements and the gate", got, want)
	}
	reqs := s.file("requests")
	last := -1
	for _, st := range stmts {
		i := strings.Index(reqs, st.SQL)
		if i < 0 || i < last {
			t.Fatalf("statement not sent, or sent out of order: %s", st.SQL)
		}
		last = i
	}
	if !strings.Contains(reqs[last:], "COUNT(*) AS violations") {
		t.Error("the gate was not the last call")
	}
	if !strings.Contains(stdout, settingsLines) {
		t.Errorf("the settings are not printed:\n%s", stdout)
	}
	if strings.Contains(stdout+stderr, testKey) {
		t.Error("the key was printed")
	}
}

// The key reaches curl on its config pipe, never on its command line.
func TestApplyKeepsTheKeyOffTheCommandLine(t *testing.T) {
	s := newStub(t, stubCurl)
	s.apply(t, allSettings)
	if strings.Contains(s.file("argv"), testKey) || strings.Contains(s.file("argv"), "zetaris.invalid") {
		t.Errorf("the key or the address is on curl's command line:\n%s", s.file("argv"))
	}
	if !strings.Contains(s.file("config"), "Bearer "+testKey) {
		t.Error("the key did not reach curl's config")
	}
}

// Only a statement marked "tolerate" may be refused, and only for the word named.
func TestApplyToleratesOnlyTheMarkedRefusals(t *testing.T) {
	s := newStub(t, stubCurl)
	s.set(t, "refuse_on", "DROP SCHEMASTORE VIEW")
	s.set(t, "refusal.json", `{"msg":"view does not exist"}`)
	if _, stderr, code := s.apply(t, allSettings); code != 0 {
		t.Fatalf("a refused DROP of a missing view stopped the apply: exit %d\n%s", code, stderr)
	}

	// Mutated field: the statement refused. A CREATE is not marked, so its refusal stops the apply,
	// prints Zetaris's refusal and not the key it echoed.
	s = newStub(t, stubCurl)
	s.set(t, "refuse_on", "CREATE SCHEMASTORE VIEW px_epss_score")
	s.set(t, "refusal.json", `{"msg":"User is not permitted to run this command (key `+testKey+`)"}`)
	stdout, stderr, code := s.apply(t, allSettings)
	if code != 3 {
		t.Fatalf("exit %d, want 3\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "not permitted") || !strings.Contains(stderr, "[key]") {
		t.Errorf("Zetaris's refusal is not shown, redacted:\n%s", stderr)
	}
	if strings.Contains(stdout+stderr, testKey) {
		t.Error("the key was printed")
	}
	if strings.Contains(s.file("requests"), "COUNT(*) AS violations") {
		t.Error("the gate ran after a refused statement")
	}
}

// An answer that cannot be read is not a pass.
func TestApplyRefusesAnUnreadableGateAnswer(t *testing.T) {
	for _, body := range []string{
		`{}`, `{"headers":["violations"],"data":[]}`, `not json`, `{"data":[["many"]]}`,
		// Law #27, one field each from the valid answer {"headers":["violations"],"data":[[0]]}.
		`{"headers":["violations"],"data":[[0],[7]]}`, // data: a second row
		`{"headers":["violations"],"data":[[0,7]]}`,   // data: a second cell
		`{"headers":["violations"],"data":[[-1]]}`,    // data: a negative count
		`{"headers":["violations"],"data":[[1.5]]}`,   // data: not a whole number
		`{"headers":["violations"],"data":[[true]]}`,  // data: not a number
		`{"headers":["violations"],"data":[[null]]}`,  // data: no value
		`{"data":[[0]]}`, // headers: absent
		`{"headers":["something_else"],"data":[[0]]}`,                      // headers: another name
		`{"headers":["violations","extra"],"data":[[0]]}`,                  // headers: two names
		`{"message":"partial","headers":["violations"],"data":[[0]]}`,      // an added message field
		`{"error":"boom","headers":["violations"],"data":[[0]]}`,           // an added error field
		`{"errors":["boom"],"headers":["violations"],"data":[[0]]}`,        // an added errors field
		`{"headers":["violations"],"data":[[0]]} {"headers":[],"data":[]}`, // a second document
		`{"headers":["violations"],"data":[[7]],"data":[[0]]}`,             // data: named twice, the last one zero
		`{"headers":["violations"],"headers":["violations"],"data":[[0]]}`, // headers: named twice
	} {
		s := newStub(t, stubCurl)
		s.set(t, "count.json", body)
		stdout, stderr, code := s.apply(t, allSettings)
		if code != 5 {
			t.Errorf("%s: exit %d, want 5", body, code)
		}
		if strings.Contains(stdout, "PROVEN_") {
			t.Errorf("%s: the settings were printed", body)
		}
		if !strings.Contains(stderr, "usl-epss-score-valid") {
			t.Errorf("%s: the refusal does not name the rule:\n%s", body, stderr)
		}
	}
}

// A readable answer passes whether the count is a number or a string of digits.
func TestApplyReadsAWellFormedCount(t *testing.T) {
	for _, body := range []string{
		`{"headers":["violations"],"data":[[0]]}`,
		`{"headers":["VIOLATIONS"],"data":[["0"]]}`,
		"{\n \"headers\": [\"violations\"],\n \"data\": [ [ 0 ] ]\n}",
	} {
		s := newStub(t, stubCurl)
		s.set(t, "count.json", body)
		if _, stderr, code := s.apply(t, allSettings); code != 0 {
			t.Errorf("%s: exit %d\n%s", body, code, stderr)
		}
	}
}

// No diagnostic prints the organisation id. Mutated field: the id echoed by the
// remote, in a refusal of a statement, in a refusal of the gate query and in a
// transport error. The key is echoed in each too.
func TestApplyRedactsTheOrganisationIdEverywhere(t *testing.T) {
	const org = "7391825"
	env := map[string]string{}
	for k, v := range allSettings {
		env[k] = v
	}
	env[query.EnvOrgID] = org
	echo := `{"msg":"org ` + org + ` is not allowed (key ` + testKey + `, X-Org-ID: ` + org + `)"}`
	for _, c := range []struct {
		name string
		prep func(s stub)
		code int
	}{
		{"statement refusal", func(s stub) {
			s.set(t, "refuse_on", "CREATE SCHEMASTORE VIEW px_epss_score")
			s.set(t, "refusal.json", echo)
		}, 3},
		{"transport error", func(s stub) {
			s.set(t, "transport_error", "curl: (7) failed to connect, X-Org-ID: "+org+", Bearer "+testKey+"\n")
		}, 3},
	} {
		s := newStub(t, stubCurl)
		c.prep(s)
		stdout, stderr, code := s.apply(t, env)
		if code != c.code {
			t.Errorf("%s: exit %d, want %d\n%s", c.name, code, c.code, stderr)
		}
		if strings.Contains(stdout+stderr, org) || strings.Contains(stdout+stderr, testKey) {
			t.Errorf("%s: the organisation id or the key reached the output:\n%s\n%s", c.name, stdout, stderr)
		}
		if !strings.Contains(stderr, "[org]") {
			t.Errorf("%s: the refusal is not shown with the id redacted:\n%s", c.name, stderr)
		}
	}

	// The gate query's own refusal, and a refusal body the gate prints with the violating rows.
	s := newStub(t, stubCurl)
	s.set(t, "gate_status", "500")
	s.set(t, "count.json", echo)
	stdout, stderr, code := s.apply(t, env)
	if code != 3 || strings.Contains(stdout+stderr, org) || strings.Contains(stdout+stderr, testKey) {
		t.Errorf("gate refusal: exit %d, id or key in the output:\n%s\n%s", code, stdout, stderr)
	}
	s = newStub(t, stubCurl)
	s.set(t, "count.json", `{"headers":["violations"],"data":[[1]]}`)
	s.set(t, "rows.json", `{"headers":["cve_id"],"data":[["`+org+`"]]}`)
	stdout, stderr, code = s.apply(t, env)
	if code != 4 || strings.Contains(stdout+stderr, org) {
		t.Errorf("rows of a refused gate: exit %d, id in the output:\n%s\n%s", code, stdout, stderr)
	}
}

// A setting that is missing, and each that is malformed, stops the script before any call.
func TestApplyNeedsEverySetting(t *testing.T) {
	for _, drop := range []string{query.EnvRESTURL, query.EnvOrgID, query.EnvAPIKey} {
		s := newStub(t, stubCurl)
		env := map[string]string{}
		for k, v := range allSettings {
			if k != drop {
				env[k] = v
			}
		}
		_, stderr, code := s.apply(t, env)
		if code != 2 || !strings.Contains(stderr, drop) {
			t.Errorf("without %s: exit %d, stderr %q", drop, code, stderr)
		}
		if s.calls() != 0 {
			t.Errorf("without %s: curl was called", drop)
		}
	}
	for _, bad := range []struct{ name, value string }{
		{query.EnvOrgID, "12 3"}, {query.EnvOrgID, "abc"}, {query.EnvRESTURL, "ftp://zetaris.invalid"},
		{query.EnvAPIKey, `abc"def`}, {query.EnvAPIKey, "abc def"},
	} {
		s := newStub(t, stubCurl)
		env := map[string]string{}
		for k, v := range allSettings {
			env[k] = v
		}
		env[bad.name] = bad.value
		if _, _, code := s.apply(t, env); code != 2 || s.calls() != 0 {
			t.Errorf("%s=%q: exit %d, %d calls", bad.name, bad.value, code, s.calls())
		}
	}
}

// DRY=1 prints the statements and the gate and sends nothing. The stub fails the
// test if curl is called, whether or not the settings are present.
func TestDryRunSendsNothing(t *testing.T) {
	for _, env := range []map[string]string{{"DRY": "1"}, {"DRY": "1", query.EnvRESTURL: allSettings[query.EnvRESTURL], query.EnvOrgID: "123", query.EnvAPIKey: testKey}} {
		s := newStub(t, stubFails)
		stdout, stderr, code := s.apply(t, env)
		if code != 0 {
			t.Fatalf("exit %d\n%s", code, stderr)
		}
		if s.calls() != 0 {
			t.Fatal("DRY=1 called curl")
		}
		stmts := parseStatements(t, read(t, uslPath("layer.sql")))
		if n := strings.Count(stdout, "DRY statement "); n != len(stmts) {
			t.Errorf("%d statements printed, want %d", n, len(stmts))
		}
		for _, st := range stmts {
			if !strings.Contains(stdout, st.SQL) {
				t.Errorf("statement not printed: %s", st.SQL)
			}
		}
		if !strings.Contains(stdout, "DRY gate usl-epss-score-valid: SELECT COUNT(*) AS violations FROM (") {
			t.Errorf("the gate is not printed:\n%s", stdout)
		}
		if strings.Contains(stdout+stderr, testKey) || strings.Contains(stdout+stderr, "zetaris.invalid") {
			t.Error("DRY=1 printed a setting")
		}
	}
}
