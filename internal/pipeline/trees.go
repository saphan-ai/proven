package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/testrun"
)

// liveSnapshotName is the evidence snapshot a live review records when the
// target already has base/ and change/. A fixture's snapshot field is
// sample-only, and live mode does not read fixture.json, so the name is the
// literal "live".
const liveSnapshotName = "live"

// workArea is the run's working area. When the logs directory is named
// "logs", its parent is that area, the same place state sits. Live git trees
// are written there so they are not created inside the repository under review.
func workArea(logDir string) (string, error) {
	abs, err := filepath.Abs(logDir)
	if err != nil {
		return "", err
	}
	if filepath.Base(abs) == "logs" {
		return filepath.Dir(abs), nil
	}
	return abs, nil
}

// loadTarget is the trees the agents read. Sample mode loads and checks the
// fixture. Live mode never reads fixture.json.
func (p *Pipeline) loadTarget(req Request) (*fixture.Fixture, []string, func(), error) {
	if p.sample {
		fx, err := fixture.Load(req.Target)
		return fx, nil, nil, err
	}
	return p.loadLiveTarget(req)
}

// loadLiveTarget picks the live trees. With a baseline commit the git range
// decides them, always: the repository under review is the change, and the
// change must not choose which trees the agents read, so a base/ or change/
// directory inside it is never taken for the prepared pair. The range is copied
// into a temporary directory under the run's working area; the returned
// function removes that directory. With no baseline, a directory that already
// has base/ and change/ is used as it is, which is how the demo is laid out.
func (p *Pipeline) loadLiveTarget(req Request) (*fixture.Fixture, []string, func(), error) {
	if req.BaselineCommit != "" {
		return p.materialize(req)
	}
	// refuseBareRepository already stopped a git repository that reached here
	// with no baseline and no prepared trees. Anything else is not a pair of
	// trees, and the readers name what is missing.
	return &fixture.Fixture{Root: req.Target, Snapshot: liveSnapshotName}, nil, nil, nil
}

// CheckTarget refuses a live review whose target has nothing for the agents to read, with the sentence that
// says what to pass. ADR-0053: Scout Intent is on by default and reads the change before the pipeline runs, so
// the caller asks this first; otherwise the plan's own failed read would be the error a user sees. Sample mode
// reads a fixture and is never refused here.
func (p *Pipeline) CheckTarget(req Request) error {
	if p.sample {
		return nil
	}
	return p.refuseBareRepository(req)
}

// refuseBareRepository refuses a git target that has no baseline commit and
// no base/ and change/ directories. The sentence says what to pass.
func (p *Pipeline) refuseBareRepository(req Request) error {
	if hasPreparedTrees(req.Target) || req.BaselineCommit != "" {
		return nil
	}
	git, err := change.WorkTree(req.Target)
	if err != nil {
		return err
	}
	if !git {
		return nil
	}
	return fmt.Errorf("a live review of a git repository needs a baseline commit, or a directory with base/ and change/")
}

// hasPreparedTrees reports whether target already holds the two trees the
// agents read. The demo is this shape, and live mode leaves it unchanged.
func hasPreparedTrees(target string) bool {
	// ADR-0025: Lstat, so a base or change that is a link is not a prepared tree. Stat would follow
	// it, and the agents and the Prover would then read whatever it names.
	base, err1 := os.Lstat(filepath.Join(target, "base"))
	chg, err2 := os.Lstat(filepath.Join(target, "change"))
	return err1 == nil && err2 == nil && base.IsDir() && chg.IsDir()
}

// materialize writes base/ from BaselineCommit and change/ from HeadCommit,
// or from the working tree when HeadCommit is empty. Only tracked files are
// copied. A tracked symlink is left out of both trees and is not followed.
// skipped names each such path once, sorted. The directory is under the run's
// working area and the caller removes it when the review returns.
func (p *Pipeline) materialize(req Request) (*fixture.Fixture, []string, func(), error) {
	dir, err := os.MkdirTemp(p.workDir, "proven-trees-")
	if err != nil {
		return nil, nil, nil, err
	}
	remove := func() { _ = os.RemoveAll(dir) }
	baseSkipped, err := change.ExportCommit(req.Target, req.BaselineCommit, filepath.Join(dir, "base"))
	if err != nil {
		remove()
		return nil, nil, nil, err
	}
	changeDir := filepath.Join(dir, "change")
	var changeSkipped []string
	if req.HeadCommit == "" {
		changeSkipped, err = change.ExportWorktree(req.Target, changeDir)
	} else {
		changeSkipped, err = change.ExportCommit(req.Target, req.HeadCommit, changeDir)
	}
	if err != nil {
		remove()
		return nil, nil, nil, err
	}
	// The trees are mounted read-only into a container whose user is not their
	// owner, so others must be able to read them (ADR-0025).
	if err := testrun.MakeReadable(dir); err != nil {
		remove()
		return nil, nil, nil, err
	}
	fx := &fixture.Fixture{Root: dir, Snapshot: gitSnapshot(req.BaselineCommit, req.HeadCommit)}
	return fx, unionSorted(baseSkipped, changeSkipped), remove, nil
}

// unionSorted returns each path once, sorted. An empty result is nil.
func unionSorted(parts ...[]string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, part := range parts {
		for _, name := range part {
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	return out
}

// gitSnapshot is the evidence snapshot for trees copied from git. It names
// the commits the trees were taken from, because live mode has no fixture.json
// to take a snapshot name from. An empty head is the working tree, written
// "..worktree", because there is no second commit to name.
func gitSnapshot(baseline, head string) string {
	if head == "" {
		return baseline + "..worktree"
	}
	return baseline + ".." + head
}
