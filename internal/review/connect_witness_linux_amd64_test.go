//go:build linux && amd64

package review

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// The process-wide witness of ADR-0034: a seccomp filter with user notification on the connect syscall.
// connect is where every outbound connection of a process ends, whatever opened it: the HTTP clients,
// a Transport built by hand, a resolver sending a DNS query, and, because a filter is inherited across
// fork and exec, git and the container runtime that Proven starts. The witness records each target and
// answers it: an address the review is allowed to reach continues, any other fails with EPERM, so a
// guard that is unplugged in a test shows up as a record and nothing leaves the host. Local IPC
// continues; a local resolver's socket is refused (see connectKind).
//
// The numbers are the linux/amd64 forms, as in internal/cli/connect_linux_test.go, where the reason
// for each is measured: Go's syscall list stops before __NR_seccomp, PR_SET_NO_NEW_PRIVS must come first,
// and TSYNC with NEW_LISTENER needs TSYNC_ESRCH. The ioctl numbers encode notification size 80 and
// response size 24 (SECCOMP_GET_NOTIF_SIZES on this kernel).

const (
	sysSeccomp                   = 317
	prSetNoNewPrivs              = 38
	seccompSetModeFilter         = 1
	seccompFilterFlagTsync       = 1
	seccompFilterFlagNewListener = 8
	seccompFilterFlagTsyncESRCH  = 16
	seccompRetUserNotif          = 0x7fc00000
	seccompRetAllow              = 0x7fff0000
	seccompIoctlNotifRecv        = 0xc0502100
	seccompIoctlNotifSend        = 0xc0182101
	seccompUserNotifFlagContinue = 1
	errnoEPERM                   = 1
)

// The kinds of connect the witness tells apart. A unix socket is local IPC: it names no host, so it is
// never judged against the allowed addresses. A local name-service daemon is the one kind of IPC that
// can resolve a host name for the caller (nscd, systemd-resolved), so its sockets are listed apart and
// refused: glibc then looks the name up itself, and a lookup of that kind ends in a network connect
// (a DNS server) that the witness judges like any other.
const (
	kindNetwork  = "network"
	kindIPC      = "ipc"
	kindResolver = "resolver"
)

// resolverSockets are the suffixes of the unix socket paths of the local name-service daemons.
var resolverSockets = []string{
	"/nscd/socket",
	"/systemd/resolve/io.systemd.Resolve",
	"/systemd/resolve/io.systemd.Resolve.Monitor",
	"/org/freedesktop/resolve1",
}

// connectKind says what a target is: IPC, a local resolver's socket, or a network address.
func connectKind(target string) string {
	path, ok := strings.CutPrefix(target, "unix:")
	if !ok {
		return kindNetwork
	}
	for _, s := range resolverSockets {
		if strings.HasSuffix(path, s) {
			return kindResolver
		}
	}
	return kindIPC
}

// connectRecord is one connect the witness saw. Target is "ip:port", "[ip6]:port", "unix:path" or a
// word naming what could not be read. Pid is the process that made the call, so a record of a child
// shows the filter reached it. Kind is one of the kinds above. Allowed is true when the connect went on.
type connectRecord struct {
	Pid     uint32
	Target  string
	Kind    string
	Allowed bool
}

// witness is the filter installed on the running process and everything it starts.
type witness struct {
	allow func(target string) bool
	mu    sync.Mutex
	seen  []connectRecord
}

// records returns a copy of what was seen so far.
func (w *witness) records() []connectRecord {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]connectRecord(nil), w.seen...)
}

// installWitness puts the filter on every thread of this process. It cannot be taken off: a caller runs
// it in a child process of its own, never in the test binary that judges the result. allow decides, for
// each target, whether the connect continues.
func installWitness(allow func(target string) bool) (*witness, error) {
	if _, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, prSetNoNewPrivs, 1, 0, 0, 0, 0); errno != 0 {
		return nil, fmt.Errorf("no_new_privs: %w", errno)
	}
	// Load the syscall number; if it is connect, notify, and every other call runs.
	filters := []syscall.SockFilter{
		{Code: 0x20, K: 0},
		{Code: 0x15, Jt: 0, Jf: 1, K: syscall.SYS_CONNECT},
		{Code: 0x06, K: seccompRetUserNotif},
		{Code: 0x06, K: seccompRetAllow},
	}
	prog := syscall.SockFprog{Len: uint16(len(filters)), Filter: &filters[0]}
	flags := uintptr(seccompFilterFlagTsync | seccompFilterFlagNewListener | seccompFilterFlagTsyncESRCH)
	fd, _, errno := syscall.Syscall(sysSeccomp, seccompSetModeFilter, flags, uintptr(unsafe.Pointer(&prog)))
	if errno != 0 {
		return nil, fmt.Errorf("seccomp: %w", errno)
	}
	w := &witness{allow: allow}
	go w.serve(int(fd))
	return w, nil
}

type seccompData struct {
	Nr                 int32
	Arch               uint32
	InstructionPointer uint64
	Args               [6]uint64
}

type seccompNotif struct {
	ID    uint64
	Pid   uint32
	Flags uint32
	Data  seccompData
}

type seccompNotifResp struct {
	ID    uint64
	Val   int64
	Error int32
	Flags uint32
}

func (w *witness) serve(fd int) {
	for {
		var note seccompNotif
		if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), seccompIoctlNotifRecv, uintptr(unsafe.Pointer(&note))); errno != 0 {
			return
		}
		// The address is in the memory of the blocked thread. A failed read is a record of its own,
		// never a silent zero, and it is refused: an address nobody could read is not an allowed one.
		target, err := connectTarget(note.Pid, note.Data.Args[1], note.Data.Args[2])
		if err != nil {
			target = "unreadable: " + err.Error()
		}
		kind := connectKind(target)
		var ok bool
		switch {
		case err != nil || kind == kindResolver:
			ok = false
		case kind == kindIPC:
			ok = true
		default:
			ok = w.allow(target)
		}
		w.mu.Lock()
		w.seen = append(w.seen, connectRecord{Pid: processOf(note.Pid), Target: target, Kind: kind, Allowed: ok})
		w.mu.Unlock()
		resp := seccompNotifResp{ID: note.ID}
		if ok {
			resp.Flags = seccompUserNotifFlagContinue
		} else {
			resp.Error = -errnoEPERM
		}
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), seccompIoctlNotifSend, uintptr(unsafe.Pointer(&resp)))
	}
}

// processOf turns the thread id the notification carries into the id of its process, so a thread of the
// review and a process it started are told apart. A thread that has gone keeps its own id.
func processOf(tid uint32) uint32 {
	b, err := os.ReadFile("/proc/" + strconv.FormatUint(uint64(tid), 10) + "/status")
	if err != nil {
		return tid
	}
	for _, line := range strings.Split(string(b), "\n") {
		if rest, ok := strings.CutPrefix(line, "Tgid:"); ok {
			if n, err := strconv.ParseUint(strings.TrimSpace(rest), 10, 32); err == nil {
				return uint32(n)
			}
		}
	}
	return tid
}

// connectTarget reads the sockaddr a blocked connect was given and names where it points.
func connectTarget(pid uint32, addr, size uint64) (string, error) {
	if addr == 0 || size < 2 {
		return "", fmt.Errorf("connect with no address")
	}
	if size > 128 {
		size = 128
	}
	f, err := os.Open("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/mem")
	if err != nil {
		return "", err
	}
	defer f.Close()
	buf := make([]byte, size)
	n, err := f.ReadAt(buf, int64(addr))
	if n < 2 {
		return "", fmt.Errorf("read of the address: %v", err)
	}
	buf = buf[:n]
	switch fam := binary.LittleEndian.Uint16(buf); fam {
	case syscall.AF_INET:
		if n < 8 {
			return "", fmt.Errorf("short AF_INET address")
		}
		return net.JoinHostPort(net.IP(buf[4:8]).String(), strconv.Itoa(int(binary.BigEndian.Uint16(buf[2:4])))), nil
	case syscall.AF_INET6:
		if n < 24 {
			return "", fmt.Errorf("short AF_INET6 address")
		}
		return net.JoinHostPort(net.IP(buf[8:24]).String(), strconv.Itoa(int(binary.BigEndian.Uint16(buf[2:4])))), nil
	case syscall.AF_UNIX:
		path := buf[2:]
		for i, b := range path {
			if b == 0 {
				path = path[:i]
				break
			}
		}
		return "unix:" + string(path), nil
	default:
		return "", fmt.Errorf("address family %d", fam)
	}
}
