package change

import (
	"os/exec"
	"strings"
	"sync/atomic"

	"github.com/saphan-ai/proven/internal/childenv"
)

// offline is set when PROVEN_NETWORK=local: git must not reach a remote (ADR-0034).
var offline atomic.Bool

// transports are the git transports a configuration file can allow by name. The
// per-protocol key wins over protocol.allow, so each one is closed on its own.
var transports = []string{"file", "git", "http", "https", "ssh", "ext"}

// UseOffline makes every git command of this package, and of GitCommand's
// callers, unable to reach a remote, and returns the setting it replaced so a
// caller can restore it. It is process-wide, like UseLimits: the review loads
// it once, before it reads the repository.
func UseOffline(on bool) (previous bool) {
	return offline.Swap(on)
}

// GitCommand builds a git command. The environment is the minimal one of
// ADR-0025: PATH and HOME, never Proven's own, so a key cannot reach git.
// Offline, it also cannot fetch: GIT_NO_LAZY_FETCH stops the fetch of an
// object a partial clone left out, protocol.allow=never and one
// protocol.<name>.allow=never per transport refuse every transport if a fetch
// is tried anyway, and no prompt can wait on a terminal. The -c switches come
// first and outrank the repository's config and the user's ~/.gitconfig, so a
// per-protocol policy set there cannot reopen a transport.
// Both switches are honoured by git 2.43 (measured 2026-10-24).
func GitCommand(args ...string) *exec.Cmd {
	argv := args
	env := childenv.Minimal()
	if offline.Load() {
		pre := []string{"-c", "protocol.allow=never"}
		for _, name := range transports {
			pre = append(pre, "-c", "protocol."+name+".allow=never")
		}
		argv = append(pre, args...)
		env = append(env, "GIT_NO_LAZY_FETCH=1", "GIT_TERMINAL_PROMPT=0")
	}
	cmd := exec.Command("git", argv...)
	cmd.Env = env
	return cmd
}

// offlineMessage names a refusal to fetch. git's own words for it are kept, so
// the reader can see what was missing; the sentence in front says what to do.
// cmd is the git command that failed. Some git versions print only "bad object"
// for a missing object, so the words alone are not enough: a failure that is
// not an absent path, in a partial clone, is named as the refused fetch.
func offlineMessage(cmd *exec.Cmd, msg string) string {
	if !offline.Load() {
		return msg
	}
	lower := strings.ToLower(msg)
	named := strings.Contains(lower, "promisor remote") || strings.Contains(lower, "lazy fetching disabled") ||
		strings.Contains(lower, "not allowed")
	absent := strings.Contains(lower, "does not exist") || strings.Contains(lower, "exists on disk, but not in")
	if !named && !absent && isPartialClone(cmd) {
		named = true
	}
	if named {
		return "an object of this partial clone is missing and PROVEN_NETWORK=local never fetches it; review a full clone (" + msg + ")"
	}
	return msg
}

// isPartialClone reports whether the repository cmd ran in has a promisor
// remote, the thing that makes git fetch a missing object. cmd names the
// repository with -C; without it git ran in the current directory.
func isPartialClone(cmd *exec.Cmd) bool {
	dir := "."
	for i, a := range cmd.Args {
		if a == "-C" && i+1 < len(cmd.Args) {
			dir = cmd.Args[i+1]
			break
		}
	}
	probe := GitCommand("-C", dir, "config", "--get-regexp", `^(extensions\.partialclone|remote\..*\.promisor)$`)
	return probe.Run() == nil
}
