//go:build !(linux && amd64)

package review

import "testing"

// TestAirGappedWitnessSeesEveryConnection is the process-wide witness of ADR-0034. It counts connect
// syscalls with seccomp user notification, which exists on linux/amd64 only, so elsewhere it skips.
func TestAirGappedWitnessSeesEveryConnection(t *testing.T) {
	t.Skip("the connect witness is a seccomp filter and runs on linux/amd64 only; the other checks of the air-gapped mode run on every platform")
}
