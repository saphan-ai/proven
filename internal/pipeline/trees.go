package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/fixture"
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
func (p *Pipeline) loadTarget(req Request) (*fixture.Fixture, func(), error) {
	if p.sample {
		fx, err := fixture.Load(req.Target)
		return fx, nil, err
	}
	return p.loadLiveTarget(req)
}

// loadLiveTarget picks the live trees. A directory that already has base/ and
// change/ is used as it is, which is how the demo is laid out. A git
// repository is copied into a temporary directory under the run's working
// area; the returned function removes that directory.
func (p *Pipeline) loadLiveTarget(req Request) (*fixture.Fixture, func(), error) {
	if hasPreparedTrees(req.Target) {
		return &fixture.Fixture{Root: req.Target, Snapshot: liveSnapshotName}, nil, nil
	}
	if req.BaselineCommit == "" {
		// refuseBareRepository already stopped a git repository that reached
		// here with no baseline. Anything else is not a pair of trees.
		return &fixture.Fixture{Root: req.Target, Snapshot: liveSnapshotName}, nil, nil
	}
	return p.materialize(req)
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
	base, err1 := os.Stat(filepath.Join(target, "base"))
	chg, err2 := os.Stat(filepath.Join(target, "change"))
	return err1 == nil && err2 == nil && base.IsDir() && chg.IsDir()
}

// materialize writes base/ from BaselineCommit and change/ from HeadCommit,
// or from the working tree when HeadCommit is empty. Only tracked files are
// copied. The directory is under the run's working area and the caller
// removes it when the review returns.
func (p *Pipeline) materialize(req Request) (*fixture.Fixture, func(), error) {
	dir, err := os.MkdirTemp(p.workDir, "proven-trees-")
	if err != nil {
		return nil, nil, err
	}
	remove := func() { _ = os.RemoveAll(dir) }
	if err := change.ExportCommit(req.Target, req.BaselineCommit, filepath.Join(dir, "base")); err != nil {
		remove()
		return nil, nil, err
	}
	changeDir := filepath.Join(dir, "change")
	if req.HeadCommit == "" {
		err = change.ExportWorktree(req.Target, changeDir)
	} else {
		err = change.ExportCommit(req.Target, req.HeadCommit, changeDir)
	}
	if err != nil {
		remove()
		return nil, nil, err
	}
	return &fixture.Fixture{Root: dir, Snapshot: gitSnapshot(req.BaselineCommit, req.HeadCommit)}, remove, nil
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
