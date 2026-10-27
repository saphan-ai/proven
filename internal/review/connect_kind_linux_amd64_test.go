//go:build linux && amd64

package review

import "testing"

// TestConnectKindTellsIPCFromTheNetwork: a unix socket is local IPC and is never judged as a network
// address; the socket of a local name-service daemon is listed apart, because it can resolve a host
// name for its caller; an IP address, v4 or v6, is the network.
// Mutated field: the unix: branch of connectKind (round 2 reported every connect as network, so the
// socket of nscd read as "off the allowed set").
func TestConnectKindTellsIPCFromTheNetwork(t *testing.T) {
	for target, want := range map[string]string{
		"unix:/var/run/nscd/socket":                    kindResolver,
		"unix:/run/nscd/socket":                        kindResolver,
		"unix:/run/systemd/resolve/io.systemd.Resolve": kindResolver,
		"unix:/var/run/docker.sock":                    kindIPC,
		"unix:/tmp/ipc.sock":                           kindIPC,
		"unix:":                                        kindIPC,
		"127.0.0.1:8088":                               kindNetwork,
		"8.8.8.8:53":                                   kindNetwork,
		"[::1]:9":                                      kindNetwork,
		"unreadable: address family 16":                kindNetwork,
	} {
		if got := connectKind(target); got != want {
			t.Errorf("connectKind(%q) = %q, want %q", target, got, want)
		}
	}
}
