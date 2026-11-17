package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/feeds"
	"github.com/saphan-ai/proven/internal/manifest"
	"github.com/saphan-ai/proven/internal/store"
)

func runDataSync(args []string, env func(string) string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprintln(stdout, "usage: proven data sync [--baseline-commit SHA] [TARGET]")
		return exitOK
	}
	cfg, err := resolve(append([]string{"scan"}, args...), env, fileExists)
	if err != nil {
		printErr(stderr, err)
		return errorCode(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	result, err := refreshData(ctx, cfg, env)
	if len(result.Sources) != 0 {
		if e := json.NewEncoder(stdout).Encode(result); e != nil {
			return exitFatal
		}
	}
	if err != nil {
		printErr(stderr, err)
		if len(result.Sources) != 0 {
			return exitIncomplete
		}
		return errorCode(err)
	}
	return exitOK
}

// refreshData exports inert Git content, uses the existing manifest readers and
// writes through the selected SQL adapter. It never builds target code.
func refreshData(ctx context.Context, cfg config, env func(string) string) (feeds.Result, error) {
	var result feeds.Result
	if cfg.sample {
		return result, &cliError{code: exitConfig, msg: "--refresh-data is unavailable in recorded sample mode"}
	}
	settings, err := feeds.Load(env)
	if err != nil {
		return result, &cliError{code: exitConfig, msg: err.Error()}
	}
	choice, err := store.Load(env)
	if err != nil {
		return result, &cliError{code: exitConfig, msg: err.Error()}
	}
	if choice.Name != store.NameSQLite && choice.Name != store.NamePostgres {
		return result, &cliError{code: exitConfig, msg: "automatic source import uses sqlite or postgres; other stores retain their native import adapters"}
	}
	rd := change.NewPolicy(settings.Network.Local(), change.DefaultLimits())
	root := cfg.target
	if top, ok, err := rd.TopLevel(root); err != nil {
		return result, err
	} else if ok {
		root = top
	}
	for _, path := range []string{settings.Cache, choice.File} {
		if path != "" {
			if err := outsideRefreshRoot(root, path); err != nil {
				return result, err
			}
		}
	}
	trees, err := os.MkdirTemp("", "proven-source-trees-")
	if err != nil {
		return result, fmt.Errorf("source refresh scratch could not be created")
	}
	defer os.RemoveAll(trees)
	base, head := filepath.Join(trees, "base"), filepath.Join(trees, "change")
	if cfg.baseline != "" {
		if _, err = rd.ExportCommit(cfg.target, cfg.baseline, base); err != nil {
			return result, err
		}
	} else if err = os.Mkdir(base, 0700); err != nil {
		return result, err
	}
	if cfg.head != "" {
		_, err = rd.ExportCommit(cfg.target, cfg.head, head)
	} else {
		_, err = rd.ExportWorktree(cfg.target, head)
	}
	if err != nil {
		return result, err
	}
	files, err := rd.Read(trees)
	if err != nil {
		return result, err
	}
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Path
	}
	selected, _, err := change.Select(names, cfg.include, cfg.exclude)
	if err != nil {
		return result, err
	}
	allowed := map[string]bool{}
	for _, path := range selected {
		allowed[path] = true
	}
	var scope []change.File
	for _, f := range files {
		if allowed[f.Path] {
			scope = append(scope, f)
		}
	}
	hits, notes, err := manifest.Collect(scope, manifest.Readers())
	if err != nil {
		return result, err
	}
	packages, unresolved := feeds.Packages(hits)
	var db *store.SQLStore
	if choice.Name == store.NameSQLite {
		db, err = store.OpenSQLite(choice.File)
	} else {
		db, err = store.OpenPostgres(env(store.EnvPostgresURL), settings.Network)
	}
	if err != nil {
		return result, fmt.Errorf("source refresh database could not be opened")
	}
	defer db.Close()
	limit, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	result, err = feeds.Sync(limit, settings, packages, nil, db.Replace)
	result.Notes = append(append(notes, unresolved...), result.Notes...)
	if len(unresolved) > 0 && err == nil {
		err = fmt.Errorf("source refresh did not resolve %d package versions", len(unresolved))
	}
	return result, err
}

// Resolve existing ancestors before checking containment: a cache/database
// symlink must not turn the refresh into a write to the reviewed checkout.
func outsideRefreshRoot(root, path string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	name, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	var suffix []string
	for {
		if _, err := os.Lstat(name); err == nil {
			break
		}
		parent := filepath.Dir(name)
		if parent == name {
			return fmt.Errorf("source refresh path has no existing parent")
		}
		suffix = append(suffix, filepath.Base(name))
		name = parent
	}
	name, err = filepath.EvalSymlinks(name)
	if err != nil {
		return fmt.Errorf("source refresh path could not be resolved")
	}
	for i := len(suffix) - 1; i >= 0; i-- {
		name = filepath.Join(name, suffix[i])
	}
	rel, err := filepath.Rel(root, name)
	if err != nil || rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return &cliError{code: exitConfig, msg: "source cache and evidence database must be outside the reviewed checkout"}
	}
	return nil
}
