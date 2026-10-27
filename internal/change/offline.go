package change

import (
	"fmt"
	"os"
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

// UseOffline sets the process-wide offline setting, the one the package-level
// functions read (ProcessPolicy), and returns the setting it replaced so a
// caller can restore it. A review does not use it: its git policy is a Policy
// fixed when it is opened, so no other review in the process can change it.
func UseOffline(on bool) (previous bool) {
	return offline.Swap(on)
}

// GitCommand builds a git command under this policy. The environment is the minimal one of
// ADR-0025: PATH and HOME, never Proven's own, so a key cannot reach git.
// Offline, it also cannot fetch: GIT_NO_LAZY_FETCH stops the fetch of an
// object a partial clone left out, protocol.allow=never and one
// protocol.<name>.allow=never per transport refuse every transport if a fetch
// is tried anyway, and no prompt can wait on a terminal. The -c switches come
// first and outrank the repository's config and the user's ~/.gitconfig, so a
// per-protocol policy set there cannot reopen a transport.
// Both switches are honoured by git 2.43 (measured 2026-10-24).
func (p Policy) GitCommand(args ...string) *exec.Cmd {
	argv := args
	env := childenv.Minimal()
	if !p.public {
		// log.showSignature=false: git log would otherwise verify a signed commit
		// by starting gpg.program, gpg.x509.program or gpg.ssh.program, a program
		// the commit's header chooses and Proven does not judge.
		// core.hooksPath=/dev/null: a working-tree read can refresh the index, and
		// git then starts the clone's post-index-change hook, a program the clone
		// chooses. GIT_OPTIONAL_LOCKS=0 keeps a read from writing the index at all.
		pre := []string{"-c", "protocol.allow=never", "-c", "log.showSignature=false", "-c", "core.hooksPath=/dev/null"}
		for _, name := range transports {
			pre = append(pre, "-c", "protocol."+name+".allow=never")
		}
		argv = append(pre, args...)
		env = append(env, "GIT_NO_LAZY_FETCH=1", "GIT_TERMINAL_PROMPT=0", "GIT_ATTR_NOSYSTEM=1", "GIT_OPTIONAL_LOCKS=0")
	}
	cmd := exec.Command("git", argv...)
	cmd.Env = env
	return cmd
}

// GitCommand builds a git command under the process-wide policy; see
// Policy.GitCommand.
func GitCommand(args ...string) *exec.Cmd { return ProcessPolicy().GitCommand(args...) }

// withoutAttributes returns args with the switches that make git read no
// attribute at all, when offline. A git command
// that compares the working tree with a commit converts each file it compares
// (the clean filter), and the filter driver is a program the repository's
// .gitattributes and the user's git configuration choose, not Proven (git-lfs is
// one). The attributes of the commit and of the working tree are read from an
// empty tree, the user's attributes file is /dev/null, and GitCommand sets
// GIT_ATTR_NOSYSTEM; .git/info/attributes is read by git whatever the switches
// say, so a clone that sets anything there is refused (refuseInfoAttributes).
// The empty tree's id depends on the repository's hash, so git names it. Needs git 2.40 or later; an older git refuses the switch by name.
// Online, args come back unchanged.
func (p Policy) withoutAttributes(args []string) ([]string, error) {
	if p.public {
		return args, nil
	}
	out, err := p.GitCommand("-C", repoDir(args), "hash-object", "-t", "tree", "/dev/null").Output()
	if err != nil {
		return nil, fmt.Errorf("git: cannot name the empty tree: %w", err)
	}
	if err := p.refuseInfoAttributes(repoDir(args)); err != nil {
		return nil, err
	}
	pre := []string{"--attr-source=" + strings.TrimSpace(string(out)), "-c", "core.attributesFile=/dev/null", "-c", "core.fsmonitor=false"}
	return append(pre, ignoreSubmodules(args)...), nil
}

// ignoreSubmodules adds --ignore-submodules=all after the diff subcommand of
// args. A clone's .gitmodules or .git/config can say `ignore = none` for a
// submodule, which outranks diff.ignoreSubmodules; the command-line switch
// outranks both, so git never enters a submodule however the clone is set.
func ignoreSubmodules(args []string) []string {
	for i, a := range args {
		if a == "-C" {
			continue
		}
		if i > 0 && args[i-1] == "-C" {
			continue
		}
		if a != "diff" {
			return args
		}
		out := append(append(append([]string(nil), args[:i+1]...), "--ignore-submodules=all"), args[i+1:]...)
		return out
	}
	return args
}

// refuseInfoAttributes refuses a clone whose .git/info/attributes sets any
// attribute. git reads that file whatever the switches above say, and a line in
// it can name a filter driver, a program the clone's configuration chooses and
// Proven does not judge (ADR-0034). A file that is absent, or holds only blank
// lines and comments, sets nothing and is not refused.
func (p Policy) refuseInfoAttributes(dir string) error {
	out, err := p.GitCommand("-C", dir, "rev-parse", "--path-format=absolute", "--git-path", "info/attributes").Output()
	if err != nil {
		return fmt.Errorf("git: cannot name .git/info/attributes: %w", err)
	}
	path := strings.TrimSpace(string(out))
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("PROVEN_NETWORK=local cannot read %s, which can choose a filter program git would start: %w", path, err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			return fmt.Errorf("PROVEN_NETWORK=local refuses this clone: %s sets git attributes, and one can choose a filter program that git starts while it reads the working tree; next: empty or remove that file, or review a clone without it", path)
		}
	}
	return nil
}

// offlineMessage names a refusal to fetch. git's own words for it are kept, so
// the reader can see what was missing; the sentence in front says what to do.
// cmd is the git command that failed. Some git versions print only "bad object"
// for a missing object, so the words alone are not enough: a failure that is
// not an absent path, in a partial clone, is named as the refused fetch.
func (p Policy) offlineMessage(cmd *exec.Cmd, msg string) string {
	if p.public {
		return msg
	}
	lower := strings.ToLower(msg)
	named := strings.Contains(lower, "promisor remote") || strings.Contains(lower, "lazy fetching disabled") ||
		strings.Contains(lower, "not allowed")
	absent := strings.Contains(lower, "does not exist") || strings.Contains(lower, "exists on disk, but not in")
	if !named && !absent && p.isPartialClone(cmd) {
		named = true
	}
	if named {
		return "an object of this partial clone is missing and PROVEN_NETWORK=local never fetches it; review a full clone (" + msg + ")"
	}
	return msg
}

// repoDir is the directory a git command line names with -C, or the current one.
func repoDir(args []string) string {
	for i, a := range args {
		if a == "-C" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "."
}

// isPartialClone reports whether the repository cmd ran in has a promisor
// remote, the thing that makes git fetch a missing object. cmd names the
// repository with -C; without it git ran in the current directory.
func (p Policy) isPartialClone(cmd *exec.Cmd) bool {
	probe := p.GitCommand("-C", repoDir(cmd.Args), "config", "--get-regexp", `^(extensions\.partialclone|remote\..*\.promisor)$`)
	return probe.Run() == nil
}
