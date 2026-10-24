package pipeline

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/pathguard"
	"github.com/saphan-ai/proven/internal/trace"
	"github.com/saphan-ai/proven/internal/world"
)

// changeKey names the change a review read. Two reviews of the same target
// and the same commits share it, so a stored refutation can match.
func changeKey(req Request) string {
	sum := sha256.Sum256([]byte(req.Target + "\n" + req.BaselineCommit + "\n" + req.HeadCommit))
	return hex.EncodeToString(sum[:])[:16]
}

// defaultStateDir sits beside a logs directory named "logs", and inside any
// other log directory. It is not the repository root.
func defaultStateDir(logDir string) (string, error) {
	abs, err := filepath.Abs(logDir)
	if err != nil {
		return "", err
	}
	if filepath.Base(abs) == "logs" {
		return filepath.Join(filepath.Dir(abs), "state"), nil
	}
	return filepath.Join(abs, "state"), nil
}

// moduleRoot is the proven module that contains the working directory.
// An empty string means the check cannot run.
func moduleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// begin opens the trace and the two stores, after refusing a logs, state or work
// directory that lies inside the reviewed target. The comparison is made on
// resolved paths, so a directory reached through a link is refused too, and it is
// made before any file is written: a refused review leaves nothing in the target.
// The state is also refused inside Proven's own module, as before. An empty
// target names no tree to protect; the readers refuse it later.
func (p *Pipeline) begin(target string) error {
	if p.trace != nil {
		return nil
	}
	if target != "" {
		for _, dir := range []struct{ what, path string }{
			{"logs", p.logDir}, {"state", p.stateDir}, {"work", p.workDir},
		} {
			inside, err := pathguard.Inside(dir.path, target)
			if err != nil {
				return fmt.Errorf("pipeline: the %s directory %s cannot be placed against the reviewed target %s: %w", dir.what, dir.path, target, err)
			}
			if inside {
				return fmt.Errorf("pipeline: the %s directory %s is inside the reviewed target %s; choose a %s directory outside it", dir.what, dir.path, target, dir.what)
			}
		}
	}
	if mod := moduleRoot(); mod != "" {
		inside, err := pathguard.Inside(p.stateDir, mod)
		if err != nil {
			return fmt.Errorf("pipeline: the state directory %s cannot be placed against the proven module: %w", p.stateDir, err)
		}
		if inside {
			return fmt.Errorf("pipeline: the state directory %s is inside the proven module", p.stateDir)
		}
	}
	tw, err := trace.Open(p.logDir)
	if err != nil {
		return err
	}
	tw.SetRedactor(p.redactor)
	mem, err := hmem.Open(hmem.Options{
		Dir: filepath.Join(p.stateDir, "hmem"), LogDir: p.logDir, RepoRoot: target,
	})
	if err != nil {
		tw.Close()
		return err
	}
	wm, err := world.Open(world.Options{
		Dir: filepath.Join(p.stateDir, "world"), LogDir: p.logDir, RepoRoot: target, Memory: mem,
	})
	if err != nil {
		tw.Close()
		return err
	}
	p.trace, p.mem, p.world = tw, mem, wm
	return nil
}

// fold reads the trace the review just wrote and applies any new events to the
// World Model, which records each landed refutation in H-MEM.
// A second pass over the same file changes nothing.
func (p *Pipeline) fold() error {
	f, err := os.Open(p.trace.Path())
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = p.world.Ingest(f, 0)
	return err
}
