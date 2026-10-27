// Command dependabot-baseline runs Proven's dependency check over the
// repository at the commit a frozen set of Dependabot alerts is pinned to, and
// writes the alert-by-alert comparison.
//
// Usage:
//
//	dependabot-baseline [flags]
//
// The evidence store is the one Proven's review uses: PROVEN_EVIDENCE_STORE
// chooses zetaris (the default), local or couchdb, and the settings of that
// store are read from the environment as a review reads them. Changing the
// store is changing that variable; nothing here is a different code path.
//
//	-repo dir           the git repository (default .)
//	-alerts file        the frozen reference (default test/corpus/reference/dependabot-fbdf7d3/alerts.json)
//	-out dir            where proven-findings.json and REPORT.md are written (default: beside the reference)
//	-work dir           scratch for the exported manifests (default .saphan/work/dependabot-baseline)
//	-manifest-only      read the manifests and ask no store
//	-evidence-label s   the phrase every verdict line carries; required when the store is local
//	-snapshot-date d    the date of the local OSV snapshot (YYYY-MM-DD); required when the store is local
//	-advisories dir     with a local store, build PROVEN_EVIDENCE_FILE from this directory of OSV advisories first
//
// A usage error, a refused setting, a store that is down or a failed read
// exits with status 2 and writes nothing.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/saphan-ai/proven/internal/baseline"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/store"
)

// defaultAlerts is the frozen reference of this repository.
const defaultAlerts = "test/corpus/reference/dependabot-fbdf7d3/alerts.json"

// defaultWork is the scratch directory; it sits under the ignored .saphan tree.
const defaultWork = ".saphan/work/dependabot-baseline"

// runTimeout bounds a whole run: one query per package, so a lock file with
// hundreds of packages takes minutes on a live store.
const runTimeout = 30 * time.Minute

// kevColumns are the columns of the KEV view the dependency check reads.
var kevColumns = []string{"cve_id", "vendor_project", "product", "date_added", "known_ransomware"}

func main() {
	os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}

// settings are the parsed flags.
type settings struct {
	repo, alerts, out, work, label, snapshotDate, advisories string
	manifestOnly                                             bool
}

// run is main without the process, so a test can drive it.
func run(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("dependabot-baseline", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var s settings
	fs.StringVar(&s.repo, "repo", ".", "the git repository")
	fs.StringVar(&s.alerts, "alerts", defaultAlerts, "the frozen reference")
	fs.StringVar(&s.out, "out", "", "where the findings and the report are written (default: beside the reference)")
	fs.StringVar(&s.work, "work", defaultWork, "scratch for the exported manifests")
	fs.BoolVar(&s.manifestOnly, "manifest-only", false, "read the manifests and ask no store")
	fs.StringVar(&s.label, "evidence-label", "", "the phrase every verdict line carries")
	fs.StringVar(&s.snapshotDate, "snapshot-date", "", "the date of the local OSV snapshot, YYYY-MM-DD")
	fs.StringVar(&s.advisories, "advisories", "", "build the local evidence file from this directory of OSV advisories")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		return 2
	}
	if s.out == "" {
		s.out = filepath.Dir(s.alerts)
	}
	if err := execute(s, getenv, stdout); err != nil {
		fmt.Fprintln(stderr, "dependabot-baseline: "+err.Error())
		return 2
	}
	return 0
}

// execute runs one comparison and writes its two files.
func execute(s settings, getenv func(string) string, stdout io.Writer) error {
	ref, err := baseline.LoadReference(s.alerts)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.work, 0o755); err != nil {
		return err
	}
	tree, err := os.MkdirTemp(s.work, "tree-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tree) // scratch copy of the manifests; a failed removal leaves only scratch files
	manifests, err := baseline.Export(s.repo, ref.PinnedCommit, tree)
	if err != nil {
		return err
	}
	opt := baseline.Options{Commit: ref.PinnedCommit}
	var runner query.Runner
	if s.manifestOnly {
		opt.Evidence = "none: manifest-only run, no evidence store was asked"
	} else {
		runner, opt, err = openEvidence(s, getenv, opt)
		if err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()
	f, err := baseline.Analyze(ctx, tree, manifests, ref, runner, opt)
	if err != nil {
		return err
	}
	return write(s.out, ref, f, stdout)
}

// openEvidence opens the store the environment names and fills in the
// evidence phrase every verdict line will carry.
func openEvidence(s settings, getenv func(string) string, opt baseline.Options) (query.Runner, baseline.Options, error) {
	choice, err := store.Load(getenv)
	if err != nil {
		return nil, opt, err
	}
	opt.Store = choice.Name
	table, err := osv.LoadTable(getenv)
	if err != nil {
		return nil, opt, err
	}
	kev, err := osv.LoadView(getenv)
	if err != nil {
		return nil, opt, err
	}
	opt.Table, opt.KEVView = table, kev
	switch choice.Name {
	case store.NameLocal:
		return openLocal(s, choice.File, table, kev, opt)
	case store.NameCouchDB:
		cfg, err := store.LoadCouch(getenv)
		if err != nil {
			return nil, opt, err
		}
		opt.Evidence = labelOr(s.label, "couchdb live evidence")
		return store.Runner(store.NewCouchDB(cfg, nil)), opt, nil
	}
	zset, err := query.Load(getenv)
	if err != nil {
		if d, ok := service.As(err); ok {
			return nil, opt, fmt.Errorf("the Zetaris store is not configured: %s; set its variables or choose PROVEN_EVIDENCE_STORE=local, or pass -manifest-only", d.Reason)
		}
		return nil, opt, err
	}
	opt.Evidence = labelOr(s.label, "zetaris live evidence (OSV, KEV)")
	return store.Runner(store.NewZetaris(zset)), opt, nil
}

// openLocal opens the local evidence file, building it from a directory of
// OSV advisories first when one is named. The snapshot date and the label are
// required: a local file has no clock of its own worth quoting.
func openLocal(s settings, file, table, kev string, opt baseline.Options) (query.Runner, baseline.Options, error) {
	if s.label == "" || s.snapshotDate == "" {
		return nil, opt, fmt.Errorf("the local store needs -evidence-label and -snapshot-date, so that every verdict line says which OSV snapshot it rests on")
	}
	day, err := time.Parse("2006-01-02", s.snapshotDate)
	if err != nil {
		return nil, opt, fmt.Errorf("-snapshot-date %q is not YYYY-MM-DD", s.snapshotDate)
	}
	opt.Evidence = s.label
	if s.advisories != "" {
		if err := buildLocal(file, s.advisories, table, kev); err != nil {
			return nil, opt, err
		}
		opt.Evidence += "; KEV view: empty, none loaded"
	}
	loc, err := store.OpenLocal(file)
	if err != nil {
		return nil, opt, err
	}
	// A fixed clock, so the same snapshot gives the same findings on every run.
	loc.SetClock(func() time.Time { return day })
	return store.Runner(loc), opt, nil
}

// buildLocal writes the local evidence file: the OSV table from the
// advisories, and an empty KEV view so the KEV query is answered, not down.
// An existing file is replaced, so the file is always the one directory.
func buildLocal(file, advisories, table, kev string) error {
	rows, err := osv.ReadDir(advisories)
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		return err
	}
	loc, err := store.OpenLocal(file)
	if err != nil {
		return err
	}
	cells := make([][]string, 0, len(rows))
	for _, r := range rows {
		cells = append(cells, r.Fields())
	}
	ctx := context.Background()
	if err := loc.Register(ctx, store.Table{Name: table, Columns: querytpl.OSVPackageColumns, Rows: cells}); err != nil {
		return err
	}
	return loc.Register(ctx, store.Table{Name: kev, Columns: kevColumns, Rows: [][]string{}})
}

func labelOr(label, fallback string) string {
	if label != "" {
		return label
	}
	return fallback
}

// write puts the findings and the report in dir, and says so on stdout.
func write(dir string, ref baseline.Reference, f baseline.Findings, stdout io.Writer) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, baseline.FindingsFile), append(b, '\n'), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, baseline.ReportFile), []byte(baseline.Report(ref, f)), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "wrote %s and %s in %s (run: %s; evidence: %s)\n", baseline.FindingsFile, baseline.ReportFile, dir, f.Mode, f.Evidence)
	return nil
}
