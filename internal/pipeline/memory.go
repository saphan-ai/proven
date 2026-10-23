package pipeline

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
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
