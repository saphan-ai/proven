package change

import (
	"os/exec"
	"strings"
	"sync/atomic"

	"github.com/saphan-ai/proven/internal/childenv"
)

// offline is set when PROVEN_NETWORK=local: git must not reach a remote (ADR-0034).
var offline atomic.Bool

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
// object a partial clone left out, protocol.allow=never refuses every
// transport if a fetch is tried anyway, and no prompt can wait on a terminal.
// Both switches are honoured by git 2.43 (measured 2026-10-24).
func GitCommand(args ...string) *exec.Cmd {
	argv := args
	env := childenv.Minimal()
	if offline.Load() {
		argv = append([]string{"-c", "protocol.allow=never"}, args...)
		env = append(env, "GIT_NO_LAZY_FETCH=1", "GIT_TERMINAL_PROMPT=0")
	}
	cmd := exec.Command("git", argv...)
	cmd.Env = env
	return cmd
}

// offlineMessage names a refusal to fetch. git's own words for it are kept, so
// the reader can see what was missing; the sentence in front says what to do.
func offlineMessage(msg string) string {
	if !offline.Load() {
		return msg
	}
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "promisor remote") || strings.Contains(lower, "lazy fetching disabled") || strings.Contains(lower, "not allowed") {
		return "an object of this partial clone is missing and PROVEN_NETWORK=local never fetches it; review a full clone (" + msg + ")"
	}
	return msg
}
