package testrun

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saphan-ai/proven/internal/service"
)

// Edit is one change the Fixer proposes to a file of the change: the one place
// where Old stands is replaced by New. A model writes edits, never a command and
// never a file name outside the checkout (ADR-0042, ADR-0025).
type Edit struct {
	Path string
	Old  string
	New  string
}

// The labels of the runs that prove a fix (ADR-0042 point 2), in the order they run.
const (
	RunBase    = "base"
	RunChange  = "change"
	RunPatched = "change+patch"
	RunPackage = "package on change+patch"
)

// The ways one run ends. Only ResultPassed and ResultFailed are test results:
// the others mean the run proves nothing.
const (
	ResultPassed   = "passed"
	ResultFailed   = "failed"
	ResultNotRun   = "could not run"
	ResultTimedOut = "timed out"
)

// Refusal reasons for a patch. The trace prefixes each one with "refused: ".
const (
	ReasonNoEdits       = "the patch has no edit"
	ReasonTooManyEdits  = "the patch has too many edits"
	ReasonEditsATest    = "a patch never edits a test file; the test that proves the fix stays as it was"
	ReasonEditNoOld     = "an edit names no text to replace"
	ReasonEditPath      = "an edit names a path that is not a file inside the checkout"
	ReasonEditMissing   = "an edit names a file the change does not have"
	ReasonEditNotFound  = "an edit's old text is not in the file"
	ReasonEditAmbiguous = "an edit's old text is in the file more than once"
	ReasonPatchTooLarge = "the patch is over the size limit"
)

// maxEdits bounds the edits of one patch, so a patch stays a suggestion a person
// can read in one sitting.
const maxEdits = 8

// FixRun is one run of the proof of a fix: what was run, in which tree, and how it ended.
// ExitCode is -1 when the command ended without one (a timeout, a runtime that did not start).
type FixRun struct {
	Label    string
	Command  string
	Result   string
	ExitCode int
	Output   string
}

// FixRequest asks for the runs that prove one patch (ADR-0042 point 2).
// Path, Name and Content name the proving test as the Prover recorded it; Content is
// empty when the test already exists in the trees. Attempt numbers the try, from 1.
type FixRequest struct {
	FindingID string
	Path      string
	Name      string
	Content   string
	BaseDir   string
	ChangeDir string
	Edits     []Edit
	Attempt   int
}

// FixOutcome is the record of one attempt. Runs is empty when the patch was
// refused, and Refused then says why; no code ran in that case.
type FixOutcome struct {
	Runs    []FixRun
	Refused string
}

// FixRunner runs the proof of a fix: the proving test on the base, on the change,
// and on the change with the patch applied, then the other tests of the package on
// the patched change. Container runs them in the Prover's container; Sample
// replays recorded runs. Every adapter that can prove a finding with a test can
// also prove a fix for it, so the Adapter implementations all carry it.
type FixRunner interface {
	RunFix(ctx context.Context, req FixRequest) (FixOutcome, error)
}

// RefuseEdits reports why a patch must not be applied to the change in changeDir,
// or "" when every edit applies. It changes nothing on disk.
func RefuseEdits(changeDir string, edits []Edit, maxBytes int) string {
	_, reason := planEdits(changeDir, edits, maxBytes)
	return reason
}

// planEdits applies the edits in memory, in order, and returns the new text of each
// file they touch. A later edit to the same file sees the result of an earlier one.
// The checks come before any byte is written (ADR-0042): a patch that cannot apply
// exactly as written is refused whole, never half applied.
func planEdits(dir string, edits []Edit, maxBytes int) (map[string]string, string) {
	if len(edits) == 0 {
		return nil, ReasonNoEdits
	}
	if len(edits) > maxEdits {
		return nil, ReasonTooManyEdits
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxWrittenBytes
	}
	size := 0
	next := map[string]string{}
	for _, e := range edits {
		size += len(e.Old) + len(e.New)
		if size > maxBytes {
			return nil, ReasonPatchTooLarge
		}
		if reason := refuseEditPath(e.Path); reason != "" {
			return nil, reason
		}
		if e.Old == "" {
			return nil, ReasonEditNoOld
		}
		text, ok := next[e.Path]
		if !ok {
			b, reason := readRegular(dir, e.Path)
			if reason != "" {
				return nil, reason
			}
			text = b
		}
		switch strings.Count(text, e.Old) {
		case 0:
			return nil, ReasonEditNotFound
		case 1:
			next[e.Path] = strings.Replace(text, e.Old, e.New, 1)
		default:
			return nil, ReasonEditAmbiguous
		}
	}
	return next, ""
}

// refuseEditPath checks the path of one edit before any file is opened.
// A test file is refused: the fix is proved by the test that showed the problem,
// so a patch that could rewrite that test would prove nothing (ADR-0042 point 8).
func refuseEditPath(p string) string {
	if p == "" || filepath.IsAbs(p) || pathLeaves(p) {
		return ReasonEditPath
	}
	if isGoTestFile(p) {
		return ReasonEditsATest
	}
	for _, seg := range strings.Split(filepath.ToSlash(p), "/") {
		if !segmentPattern.MatchString(seg) || strings.Contains(seg, "..") {
			return ReasonEditPath
		}
	}
	return ""
}

// readRegular reads rel under dir when it is a regular file. A link is refused,
// because it would lead out of the checkout.
func readRegular(dir, rel string) (string, string) {
	if dir == "" {
		return "", ReasonEditMissing
	}
	full := filepath.Join(dir, filepath.FromSlash(rel))
	info, err := os.Lstat(full)
	if err != nil {
		return "", ReasonEditMissing
	}
	if !info.Mode().IsRegular() {
		return "", ReasonEditPath
	}
	b, err := os.ReadFile(full)
	if err != nil {
		return "", ReasonEditMissing
	}
	return string(b), ""
}

// stagePatched copies src into a new directory and applies the edits to the copy.
// The caller's tree is never changed. The caller removes the returned directory.
func stagePatched(src string, edits []Edit, maxBytes int) (string, error) {
	files, reason := planEdits(src, edits, maxBytes)
	if reason != "" {
		return "", fmt.Errorf("testrun: patch refused: %s", reason)
	}
	dst, err := os.MkdirTemp("", "proven-patched-")
	if err != nil {
		return "", err
	}
	if err := copyTree(src, dst); err != nil {
		os.RemoveAll(dst)
		return "", err
	}
	for rel, text := range files {
		if err := writeInside(dst, rel, text); err != nil {
			os.RemoveAll(dst)
			return "", err
		}
	}
	// The container user is not the owner of the copy and must be able to read it.
	if err := MakeReadable(dst); err != nil {
		os.RemoveAll(dst)
		return "", err
	}
	return dst, nil
}

// PackageArgv is the command that runs every test of the package that holds the
// proving test: go test ./<package>, built from the same checked path as TestArgv.
// A broken neighbour test is how a patch that fixes one thing breaks another.
func PackageArgv(testPath, name string) ([]string, string) {
	argv, reason := TestArgv(testPath, name)
	if reason != "" {
		return nil, reason
	}
	// The package is the last word of the test command (see TestArgv).
	return []string{"go", "test", argv[len(argv)-1]}, ""
}

// resultOf names how a side ended, in the words of the FixRun record.
func resultOf(k runKind) string {
	switch k {
	case kindPass:
		return ResultPassed
	case kindFail:
		return ResultFailed
	case kindTimeout:
		return ResultTimedOut
	default:
		return ResultNotRun
	}
}

// RunFix runs the proof of a fix in the container (ADR-0042 point 2): the proving
// test on the base, on the change and on the change with the patch applied, then
// the package's tests on the patched change. The patch is applied to a scratch copy
// only, and only the copy is mounted; the review's own trees are never written.
// The package run is skipped when the patched test did not pass, since the attempt
// has already failed and the run would only cost time.
func (c *Container) RunFix(ctx context.Context, req FixRequest) (FixOutcome, error) {
	if c == nil || c.runtime == "" {
		return FixOutcome{}, &service.Down{Service: "prover", Reason: "Prover off: no container runtime"}
	}
	if err := ctx.Err(); err != nil {
		return FixOutcome{}, err
	}
	// The same checks as Run, before anything is copied or started (ADR-0025).
	if req.Content != "" {
		if reason := RefuseWritten(req.Path, req.Content, req.Name, req.BaseDir, req.ChangeDir, c.cfg.MaxWrittenBytes); reason != "" {
			return FixOutcome{Refused: reason}, nil
		}
	}
	testArgv, reason := TestArgv(req.Path, req.Name)
	if reason != "" {
		return FixOutcome{Refused: reason}, nil
	}
	pkgArgv, _ := PackageArgv(req.Path, req.Name)
	if reason := RefuseEdits(req.ChangeDir, req.Edits, c.cfg.MaxWrittenBytes); reason != "" {
		return FixOutcome{Refused: reason}, nil
	}
	if req.BaseDir == "" || req.ChangeDir == "" {
		return FixOutcome{Refused: "base directory and change directory are required"}, nil
	}
	if linkedCheckout(req.BaseDir, req.ChangeDir) {
		return FixOutcome{Refused: "a checkout is a symbolic link and is not mounted"}, nil
	}
	baseDir, changeDir := req.BaseDir, req.ChangeDir
	if req.Content != "" {
		// Both sides run the same test text, in copies, as Run does.
		cc, err := stageWritten(req.ChangeDir, req.Path, req.Content)
		if err != nil {
			return FixOutcome{}, err
		}
		defer os.RemoveAll(cc)
		bc, err := stageWritten(req.BaseDir, req.Path, req.Content)
		if err != nil {
			return FixOutcome{}, err
		}
		defer os.RemoveAll(bc)
		baseDir, changeDir = bc, cc
	}
	patched, err := stagePatched(changeDir, req.Edits, c.cfg.MaxWrittenBytes)
	if err != nil {
		return FixOutcome{}, err
	}
	defer os.RemoveAll(patched)

	var out FixOutcome
	steps := []struct {
		label string
		dir   string
		argv  []string
	}{
		{RunBase, baseDir, testArgv},
		{RunChange, changeDir, testArgv},
		{RunPatched, patched, testArgv},
		{RunPackage, patched, pkgArgv},
	}
	for _, s := range steps {
		// The package tests are worth running only for a patch that fixed the test.
		if s.label == RunPackage && (len(out.Runs) < 3 || out.Runs[2].Result != ResultPassed) {
			break
		}
		text, kind, code := c.sideExit(ctx, s.dir, s.argv)
		if kind == kindCancelled {
			return FixOutcome{}, withRemoval(ctx.Err(), text)
		}
		if down := c.absentImage(text, kind); down != nil {
			return FixOutcome{}, down
		}
		out.Runs = append(out.Runs, FixRun{
			Label: s.label, Command: displayCommand(s.argv), Result: resultOf(kind),
			ExitCode: code, Output: trimOutput(text),
		})
	}
	return out, nil
}
