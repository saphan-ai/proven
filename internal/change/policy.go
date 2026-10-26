package change

import (
	"context"
	"os/exec"
)

// Policy is how one review reads a repository: whether its git may reach a
// remote, and the size bounds of its reads. It is a value fixed when the review
// is opened, and every git command and every read of that review takes it from
// there, so another review opened in the same process, at the same time or
// later, cannot change it (ADR-0034). The zero Policy is the public mode with
// the default bounds.
type Policy struct {
	offline bool
	limits  Limits
}

// NewPolicy is the policy of a review. offline is the air-gapped mode; a field
// of l that is zero or below keeps its default, so a partly filled Limits is
// never an unbounded one.
func NewPolicy(offline bool, l Limits) Policy {
	return Policy{offline: offline, limits: fillLimits(l)}
}

// Offline reports whether git, under this policy, cannot reach a remote.
func (p Policy) Offline() bool { return p.offline }

// Bounds returns the size bounds this policy applies to a read.
func (p Policy) Bounds() Limits { return fillLimits(p.limits) }

// ProcessPolicy is the policy the package-level functions use: the process-wide
// settings of UseOffline and UseLimits. A review does not read it; a command
// line step that runs before any review is opened, and a test, may.
func ProcessPolicy() Policy {
	return Policy{offline: offline.Load(), limits: processBounds()}
}

type policyKey struct{}

// WithPolicy returns a context that carries p. The pipeline's calls into the
// agents take a context and no policy, so a review hands its policy down this
// way, as it hands down the KEV view and the adapter bound.
func WithPolicy(ctx context.Context, p Policy) context.Context {
	return context.WithValue(ctx, policyKey{}, p)
}

// PolicyFrom is the policy a review put in ctx. A context that carries none
// gets the process-wide policy, which is how a reader outside a review (a test,
// a command line step) behaves.
func PolicyFrom(ctx context.Context) Policy {
	if p, ok := ctx.Value(policyKey{}).(Policy); ok {
		return p
	}
	return ProcessPolicy()
}

// GitCommand builds a git command under the process-wide policy; see
// Policy.GitCommand.
func GitCommand(args ...string) *exec.Cmd { return ProcessPolicy().GitCommand(args...) }

// Read reads a tree under the process-wide policy; see Policy.Read.
func Read(root string) ([]File, error) { return ProcessPolicy().Read(root) }

// WorkTree reports, under the process-wide policy, whether target is a git work tree.
func WorkTree(target string) (bool, error) { return ProcessPolicy().WorkTree(target) }

// TopLevel is Policy.TopLevel under the process-wide policy.
func TopLevel(target string) (string, bool, error) { return ProcessPolicy().TopLevel(target) }

// Between is Policy.Between under the process-wide policy.
func Between(target, baseline, head string) ([]string, error) {
	return ProcessPolicy().Between(target, baseline, head)
}

// StatusBetween is Policy.StatusBetween under the process-wide policy.
func StatusBetween(target, baseline, head string) ([]Status, error) {
	return ProcessPolicy().StatusBetween(target, baseline, head)
}

// Blob is Policy.Blob under the process-wide policy.
func Blob(target, rev, rel string) ([]byte, bool, error) {
	return ProcessPolicy().Blob(target, rev, rel)
}

// Messages is Policy.Messages under the process-wide policy.
func Messages(target, baseline, head string) ([]string, error) {
	return ProcessPolicy().Messages(target, baseline, head)
}

// ResolveCommit is Policy.ResolveCommit under the process-wide policy.
func ResolveCommit(target, rev string) string { return ProcessPolicy().ResolveCommit(target, rev) }

// ExportCommit is Policy.ExportCommit under the process-wide policy.
func ExportCommit(target, rev, dst string) ([]string, error) {
	return ProcessPolicy().ExportCommit(target, rev, dst)
}

// ExportWorktree is Policy.ExportWorktree under the process-wide policy.
func ExportWorktree(target, dst string) ([]string, error) {
	return ProcessPolicy().ExportWorktree(target, dst)
}

// FactsBetween is Policy.FactsBetween under the process-wide policy.
func FactsBetween(target, baseline, head string) (map[string]Facts, error) {
	return ProcessPolicy().FactsBetween(target, baseline, head)
}
