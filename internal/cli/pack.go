package cli

import (
	"bytes"
	"crypto/ed25519"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/evidence"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/settingspage"
)

// loadPackKey reads the signing key a run with --evidence-pack needs. It runs
// before the review, so a missing key costs nothing; the refusal is exit 7,
// a setting refused at load (ADR-0020 §2). The refusal names the setting and
// the path, never the key (ADR-0035 point 2).
func loadPackKey(getenv func(string) string) (ed25519.PrivateKey, error) {
	key, err := evidence.LoadPrivateKey(evidence.EnvPackKey, getenv(evidence.EnvPackKey))
	if err != nil {
		return nil, &cliError{code: exitConfig, msg: err.Error()}
	}
	return key, nil
}

// fileSize is the size of path, or zero when it does not exist. A trace file
// is appended to by every run in the directory, so the size before the review
// marks where this run's lines begin.
func fileSize(path string) int64 {
	st, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return st.Size()
}

// readFrom returns the bytes of path after offset, or nothing when the file
// is absent or shorter than offset.
func readFrom(path string, offset int64) []byte {
	body, err := os.ReadFile(path)
	if err != nil || int64(len(body)) < offset {
		return nil
	}
	return body[offset:]
}

// writePack makes the signed evidence pack of one review and writes it to
// cfg.evidencePack (ADR-0035 points 1 and 2).
func writePack(cfg config, getenv func(string) string, key ed25519.PrivateKey, result review.Result, traceStart int64) error {
	redactor := redact.FromEnv(getenv)
	in := evidence.Input{
		Result:   result,
		Trace:    readFrom(result.TracePath, traceStart),
		Commits:  resolveCommits(cfg),
		Settings: packSettings(cfg, getenv, redactor),
		Redactor: redactor,
		Key:      key,
		Now:      time.Now(),
	}
	pack, err := evidence.Build(in)
	if err != nil {
		return err
	}
	return evidence.WriteFile(cfg.evidencePack, pack)
}

// resolveCommits names the commits the review was asked about, and what git
// resolves each name to. A name git cannot resolve stays as given with an
// empty resolution; an empty head is the working tree and resolves to nothing.
func resolveCommits(cfg config) evidence.Commits {
	return evidence.Commits{
		Target:           cfg.target,
		Baseline:         cfg.baseline,
		BaselineResolved: change.ResolveCommit(cfg.target, cfg.baseline),
		Head:             cfg.head,
		HeadResolved:     change.ResolveCommit(cfg.target, cfg.head),
	}
}

// packSettings is the settings the run used, and never a secret. A secret
// setting is written as "set" or "unset". Every other value is read from the
// documented settings only, never from the whole environment, and passes
// through the redactor besides. Sample mode reads no environment setting, so
// none is recorded.
func packSettings(cfg config, getenv func(string) string, r *redact.Redactor) map[string]string {
	out := map[string]string{
		"command":  cfg.command,
		"target":   cfg.target,
		"format":   cfg.format,
		"baseline": cfg.baseline,
		"head":     cfg.head,
		"include":  strings.Join(cfg.include, ","),
		"exclude":  strings.Join(cfg.exclude, ","),
		"severity": string(cfg.severity),
		"timeout":  strconv.Itoa(cfg.timeoutSec),
		"sample":   strconv.FormatBool(cfg.sample),
		"config":   cfg.configPath,
		"error":    strconv.FormatBool(cfg.errorOnFindings),
		"pack-key": state(getenv(evidence.EnvPackKey)),
	}
	if !cfg.sample {
		for _, sec := range settingspage.Catalog() {
			for _, row := range sec.Rows {
				v := getenv(row.Name)
				if row.Secret {
					out[row.Name] = state(v)
				} else {
					out[row.Name] = v
				}
			}
		}
		for _, name := range redact.Names {
			out[name] = state(getenv(name))
		}
	}
	for k, v := range out {
		out[k] = r.Text(v)
	}
	return out
}

func state(v string) string {
	if v == "" {
		return "unset"
	}
	return "set"
}

// runVerify is `proven verify <pack>`: it checks the pack against the public
// key and prints what was reviewed, when, by which models and on what
// evidence. Exit 0 verified; 2 refused, naming the check and the file; 7 a
// key setting refused.
func runVerify(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	keyPath := fs.String("public-key", "", "path to the public key file")
	// Flags may sit after the pack, as they may after a scan target.
	var flagArgs, pos []string
	for i := 0; i < len(args); i++ {
		name, ok := flagToken(args[i])
		if !ok {
			pos = append(pos, args[i])
			continue
		}
		flagArgs = append(flagArgs, args[i])
		if name == "public-key" && !strings.Contains(args[i], "=") && i+1 < len(args) {
			i++
			flagArgs = append(flagArgs, args[i])
		}
	}
	if err := fs.Parse(flagArgs); err != nil {
		fmt.Fprintf(stderr, "%s\nknown flags: --public-key\n", err.Error())
		return exitFatal
	}
	if len(pos) != 1 {
		fmt.Fprintln(stderr, "verify takes exactly one pack file")
		return exitFatal
	}
	path := *keyPath
	setting := "--public-key"
	if path == "" {
		path, setting = getenv(evidence.EnvPackPublicKey), evidence.EnvPackPublicKey
	}
	pub, err := evidence.LoadPublicKey(setting, path)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return exitConfig
	}
	rep, err := evidence.VerifyFile(pos[0], pub)
	if err != nil {
		fmt.Fprintln(stderr, visible(err.Error()))
		return exitFatal
	}
	fmt.Fprint(stdout, renderReport(pos[0], rep))
	return exitOK
}

// renderReport is the text `proven verify` prints for a pack it accepted.
// Every string that comes from the pack, or from its path, goes through
// visible: a signed pack is the signer's text, not the reader's.
func renderReport(path string, rep *evidence.Report) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "evidence pack verified: %s\n", visible(path))
	fmt.Fprintf(&b, "signed with key: %s\n", visible(rep.KeyID))
	fmt.Fprintf(&b, "created: %s\n", visible(rep.Created))
	fmt.Fprintf(&b, "reviewed: %s\n", describeCommits(rep.Commits))
	fmt.Fprintf(&b, "models: %d use(s), %d with a provider receipt found in the trace, %d without\n", len(rep.Models), rep.Receipted, rep.Unreceipted)
	for _, m := range rep.Models {
		fmt.Fprintf(&b, "  %s: %s via %s, receipt %s\n", visible(m.Agent), visible(orNone(m.Model)), visible(orNone(m.Route)), visible(orNone(m.ReceiptID)))
	}
	fmt.Fprintf(&b, "evidence: %d finding(s), %d test or tool run(s), %d query(ies), %d trace event(s)\n", rep.Findings, rep.Tests, rep.Queries, rep.TraceEvents)
	levels := make([]string, 0, len(rep.ProofCounts))
	for level := range rep.ProofCounts {
		levels = append(levels, level)
	}
	sort.Strings(levels)
	for _, level := range levels {
		fmt.Fprintf(&b, "  %s: %d\n", visible(level), rep.ProofCounts[level])
	}
	if len(rep.FailedSvcs) > 0 {
		fmt.Fprintf(&b, "degraded: %s did not answer\n", visible(strings.Join(rep.FailedSvcs, ", ")))
	}
	fmt.Fprintf(&b, "checked: %s (%d files)\n", strings.Join(rep.ChecksPassed, "; "), len(rep.Files))
	for _, n := range rep.NotChecked {
		fmt.Fprintf(&b, "not checked: %s\n", n)
	}
	return b.String()
}

func describeCommits(c evidence.Commits) string {
	base := visible(orNone(c.Baseline))
	if c.BaselineResolved != "" {
		base += " (" + visible(c.BaselineResolved) + ")"
	}
	head := visible(c.Head)
	if c.Head == "" {
		head = "the working tree"
	} else if c.HeadResolved != "" {
		head += " (" + visible(c.HeadResolved) + ")"
	}
	return fmt.Sprintf("%s, baseline %s, head %s", visible(c.Target), base, head)
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
