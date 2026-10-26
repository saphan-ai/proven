package testrun

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The trust boundary of the runtime (ADR-0034). The container runtime is a
// program the operator installs and points to, so it is the operator's trusted
// base: Proven does not judge what it does. Proven checks two things about it
// and says what it is: the name must not be a relative path, which depends on
// the directory Proven happens to run in, and the file, and the directory that
// holds it, must not be writable by every user, because anyone could then
// replace the program Proven starts.

// versionBound is the longest `<runtime> --version` may run when the start line
// is built. A runtime that does not answer in time is named "version unknown".
const versionBound = 5 * time.Second

// refuseRelativeRuntime refuses a runtime given as a relative path: a name with
// a separator that is not absolute. A bare name is not refused; it is resolved
// through PATH once, by NewContainer.
func refuseRelativeRuntime(runtime string) error {
	if !strings.ContainsRune(runtime, '/') || filepath.IsAbs(runtime) {
		return nil
	}
	return refuse(envRuntimeSetting, clip(runtime),
		"PROVEN_NETWORK=local found a relative path, which names a different program in each directory, so the runtime is not run",
		"a bare name such as docker or podman, found through PATH, or an absolute path such as /usr/bin/docker")
}

// vetRuntimeFile refuses a resolved runtime that any user could replace: the
// file is world-writable, or a directory above it is world-writable without
// the sticky bit. The link is followed, so the file that runs is the one judged;
// every directory on the way to it, links included, is judged as well.
func vetRuntimeFile(path string) error {
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return refuse(envRuntimeSetting, clip(path), "PROVEN_NETWORK=local cannot read the runtime's file: "+err.Error(),
			"an absolute path to a runtime file Proven can read")
	}
	info, err := os.Stat(real)
	if err != nil {
		return refuse(envRuntimeSetting, clip(path), "PROVEN_NETWORK=local cannot read the runtime's file: "+err.Error(),
			"an absolute path to a runtime file Proven can read")
	}
	if info.Mode().Perm()&0o002 != 0 {
		return refuse(envRuntimeSetting, clip(real),
			fmt.Sprintf("the runtime file is writable by every user (mode %04o), so any user could replace the program Proven starts; it is not run", info.Mode().Perm()),
			"a runtime that only its owner can change: chmod o-w on the file, or another path")
	}
	dirs, err := linkDirs(path)
	if err != nil {
		return refuse(envRuntimeSetting, clip(path), "PROVEN_NETWORK=local cannot follow the runtime's links: "+err.Error(),
			"an absolute path to a runtime file Proven can read")
	}
	for _, dir := range dirs {
		di, err := os.Stat(dir)
		if err != nil {
			return refuse(envRuntimeSetting, clip(dir), "PROVEN_NETWORK=local cannot read the runtime's directory: "+err.Error(),
				"an absolute path to a runtime in a directory Proven can read")
		}
		if di.Mode().Perm()&0o002 != 0 && di.Mode()&os.ModeSticky == 0 {
			return refuse(envRuntimeSetting, clip(dir),
				fmt.Sprintf("a directory above the runtime is writable by every user and has no sticky bit (mode %04o), so any user could replace the program under it; it is not run", di.Mode().Perm()),
				"a runtime in a directory that only its owner can change: chmod o-w on the directory, or another path")
		}
	}
	return nil
}

// maxLinks bounds the links linkDirs follows, as the kernel bounds a lookup.
const maxLinks = 40

// linkDirs lists every directory that holds a name on the way from path to the
// file it names, once: each directory from the root down, and, for every link
// met on the way (of the file or of a directory), each directory of its target.
// A directory any user can write without the sticky bit lets that user replace
// the name in it, so each one is judged.
func linkDirs(path string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	links := 0
	var walk func(p string) (string, error)
	walk = func(p string) (string, error) {
		cur := string(filepath.Separator)
		for _, part := range strings.Split(filepath.Clean(p), string(filepath.Separator)) {
			if part == "" {
				continue
			}
			if !seen[cur] {
				seen[cur] = true
				out = append(out, cur)
			}
			next := filepath.Join(cur, part)
			fi, err := os.Lstat(next)
			if err != nil {
				return "", err
			}
			if fi.Mode()&os.ModeSymlink == 0 {
				cur = next
				continue
			}
			if links++; links > maxLinks {
				return "", fmt.Errorf("more than %d links", maxLinks)
			}
			target, err := os.Readlink(next)
			if err != nil {
				return "", err
			}
			if !filepath.IsAbs(target) {
				target = filepath.Join(cur, target)
			}
			if cur, err = walk(target); err != nil {
				return "", err
			}
		}
		return cur, nil
	}
	if _, err := walk(path); err != nil {
		return nil, err
	}
	return out, nil
}

// Describe is the start line of local mode: the absolute path of the runtime
// and the first line of its --version output, so the operator sees what runs.
// A version that cannot be read is "version unknown", never a refusal. The
// call is bounded by versionBound and by ctx.
func (c *Container) Describe(ctx context.Context) string {
	version := "version unknown"
	ctx, cancel := context.WithTimeout(ctx, versionBound)
	defer cancel()
	if out, err := c.invoke(ctx, []string{c.runtime, "--version"}); err == nil {
		for _, line := range strings.Split(out, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				version = clip(line)
				break
			}
		}
	}
	return "container runtime " + c.runtime + ": " + version
}

// Local reports whether the runtime was loaded in local mode, the mode in
// which Describe is printed and logged.
func (c *Container) Local() bool { return c != nil && c.cfg.local }
