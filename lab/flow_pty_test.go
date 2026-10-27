package lab

// The flow on a real terminal. Everything in flow_test.go runs with no terminal and a stand-in for one; this test
// gives up.sh a pseudo-terminal, types Ctrl-C at the token prompt, and reads the terminal's own echo flag before and
// after. It needs python3 and a host that opens pseudo-terminals, and is skipped by name where it cannot run.

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const ptyDriver = `import os, pty, select, sys, termios, time

up = sys.argv[1]
try:
    pid, fd = pty.fork()
except OSError:
    sys.exit(77)
if pid == 0:
    # The shell stays alive after up.sh, so the terminal's state can still be read from the master.
    os.execvp("sh", ["sh", "-c", 'sh "$1"; sleep 3', "sh", up])

def drain(limit):
    out = b""
    end = time.time() + limit
    while time.time() < end:
        r, _, _ = select.select([fd], [], [], 0.2)
        if not r:
            break
        try:
            d = os.read(fd, 4096)
        except OSError:
            break
        if not d:
            break
        out += d
    return out

seen = b""
end = time.time() + 20
while time.time() < end and b"as you type" not in seen:
    seen += drain(0.5)
time.sleep(0.5)
during = bool(termios.tcgetattr(fd)[3] & termios.ECHO)
os.write(fd, b"\x03")
time.sleep(1.0)
seen += drain(1.0)
after = bool(termios.tcgetattr(fd)[3] & termios.ECHO)
print("prompt=%s echo_during=%s echo_after=%s" % (b"as you type" in seen, during, after))
`

func ptyEchoAfterInterrupt(t *testing.T, tree string) string {
	t.Helper()
	needTool(t, "python3")
	script := filepath.Join(t.TempDir(), "drive.py")
	if err := os.WriteFile(script, []byte(ptyDriver), 0o600); err != nil {
		t.Fatal(err)
	}
	if tree == "" {
		tree = mustAbs(t, root)
	}
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	writeExec(t, bin, "docker", fakeDocker)
	cmd := exec.Command("python3", "-I", script, filepath.Join(tree, "lab", "scripts", "up.sh"))
	env := append([]string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "TMPDIR=" + tmp, "FAKE_DIR=" + tmp, "TERM=xterm"}, labPorts(t)...)
	cmd.Env = append(env, liveOurs("PROVEN_LAB_GATEWAY_AUTH=token", "JUDGES_GATEWAY_URL=http://127.0.0.1:1", "PROVEN_MARKOVIAN=off", "PROVEN_LAB_SCOUT_INTENT=off")...)
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 77 {
		t.Skip("this host opens no pseudo-terminal, so the terminal case is not run here; it runs wherever one opens")
	}
	if err != nil {
		t.Fatalf("the driver: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestFlowLeavesARealTerminalWithEchoOnAfterCtrlC(t *testing.T) {
	got := ptyEchoAfterInterrupt(t, "")
	if got != "prompt=True echo_during=False echo_after=True" {
		t.Errorf("on a real terminal the echo was not off during the prompt and on after Ctrl-C: %q", got)
	}
	// The negative control: the interrupt no longer restores the terminal. Mutated field: the INT trap.
	t.Run("negative control: the INT trap removed", func(t *testing.T) {
		skipControlOnDarwin(t)
		tree := copyLab(t)
		mutate(t, tree, "lab/scripts/flow.sh", "trap 'flow_restore_echo; echo; exit 130' INT", "trap - INT")
		if got := ptyEchoAfterInterrupt(t, tree); strings.HasSuffix(got, "echo_after=True") {
			t.Errorf("the control did not turn red: the terminal was restored with the INT trap removed: %q", got)
		}
	})
}

// skipControlOnDarwin skips a terminal negative control on macOS, by name. The master measured on 26 October 2026, at
// c73eca9, that on macOS the control does not turn red: the terminal is back with echo on after the session ends
// whether or not the trap ran. That is a claim about macOS this seat did not repeat. The positive half of each test
// still runs there, and the control runs on Linux.
func skipControlOnDarwin(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "darwin" {
		t.Skip("this control cannot turn red on macOS (measured by the master at c73eca9); it runs on Linux")
	}
}
