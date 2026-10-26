//go:build linux && amd64

package cli

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

// A connect syscall is the common point of http.DefaultTransport, a
// Transport that dials for itself, and a raw net.Dial. Counting that
// syscall is 0 when the scan dials nothing and greater than 0 when it
// dials. A RoundTrip counter on http.DefaultTransport stays 0 in both
// of those dialling cases, so it cannot tell them apart.
//
// The numbers below are the linux/amd64 forms. Go's syscall list stops
// before __NR_seccomp, so the number is written out. PR_SET_NO_NEW_PRIVS
// is required first: without it the filter call returns EACCES.
// TSYNC and NEW_LISTENER together are refused unless TSYNC_ESRCH is set
// too; that refusal was measured on this kernel (EINVAL for flags 9,
// success for flags 25). The ioctl numbers encode the struct sizes
// SECCOMP_GET_NOTIF_SIZES reported here: notification 80, response 24.

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
)

// TestSampleModeOpensNoConnection proves a sample scan makes no TCP or UDP
// connection. The child is the test binary again, with a seccomp listener
// on connect. The listener lets the call proceed and counts it. An AF_UNIX
// connect is not a network connection and is counted apart.
func TestSampleModeOpensNoConnection(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0])
	cmd.Dir = dir
	cmd.Env = []string{
		"PROVEN_CONNECT_CHILD=1",
		"PROVEN_CONNECT_TARGET=" + target,
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + dir,
		"TMPDIR=" + dir,
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	exitCode, network, unixN, other, unread, perr := parseConnectReport(stdout.String())
	if perr != nil || runErr != nil || exitCode != 0 || network != 0 || other != 0 || unread != 0 {
		t.Fatalf("exit %d network %d unix %d other %d unread %d\nparse %v\nrun %v\nstdout %s\nstderr %s",
			exitCode, network, unixN, other, unread, perr, runErr, stdout.String(), stderr.String())
	}
}

func init() {
	connectChildMain = connectChild
}

func connectChild() int {
	var network, unixN, other, unread atomic.Int32
	if err := listenConnect(&network, &unixN, &other, &unread); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", os.Getenv("PROVEN_CONNECT_TARGET")}, emptyEnv, &stdout, &stderr, nil)
	fmt.Printf("exit=%d\nnetwork=%d\nunix=%d\nother=%d\nunread=%d\n",
		code, network.Load(), unixN.Load(), other.Load(), unread.Load())
	if code != 0 {
		fmt.Fprint(os.Stderr, stderr.String())
	}
	os.Stdout.Sync()
	return code
}

func listenConnect(network, unixN, other, unread *atomic.Int32) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, prSetNoNewPrivs, 1, 0, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("no_new_privs: %w", errno)
	}
	// Load the syscall number. If it is connect, notify; every other call runs.
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
		return fmt.Errorf("seccomp: %w", errno)
	}
	go serveConnect(int(fd), network, unixN, other, unread)
	return nil
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

func serveConnect(fd int, network, unixN, other, unread *atomic.Int32) {
	for {
		var note seccompNotif
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), seccompIoctlNotifRecv, uintptr(unsafe.Pointer(&note)))
		if errno != 0 {
			return
		}
		// The address is in the blocked thread. Count only IP families.
		// A failed read is not zero: zero would look like "no dial".
		fam, err := connectFamily(note.Pid, note.Data.Args[1])
		switch {
		case err != nil:
			unread.Add(1)
		case fam == syscall.AF_INET || fam == syscall.AF_INET6:
			network.Add(1)
		case fam == syscall.AF_UNIX:
			unixN.Add(1)
		default:
			other.Add(1)
		}
		resp := seccompNotifResp{ID: note.ID, Flags: seccompUserNotifFlagContinue}
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), seccompIoctlNotifSend, uintptr(unsafe.Pointer(&resp)))
	}
}

func connectFamily(pid uint32, addr uint64) (uint16, error) {
	if addr == 0 {
		return 0, fmt.Errorf("connect with no address")
	}
	f, err := os.Open("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/mem")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	if _, err := f.Seek(int64(addr), 0); err != nil {
		return 0, err
	}
	var fam uint16
	if err := binary.Read(f, binary.LittleEndian, &fam); err != nil {
		return 0, err
	}
	return fam, nil
}

func parseConnectReport(out string) (exitCode, network, unixN, other, unread int, err error) {
	exitCode, err = reportField(out, "exit")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	network, err = reportField(out, "network")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	unixN, err = reportField(out, "unix")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	other, err = reportField(out, "other")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	unread, err = reportField(out, "unread")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	return exitCode, network, unixN, other, unread, nil
}

func reportField(out, key string) (int, error) {
	prefix := key + "="
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strconv.Atoi(strings.TrimPrefix(line, prefix))
		}
	}
	return 0, fmt.Errorf("missing %s", key)
}
