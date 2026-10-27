package change

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// signedChild writes a child of HEAD whose header carries an OpenPGP signature block,
// as a commit of the reviewed branch may, and returns its id.
func signedChild(t *testing.T, repo string) string {
	t.Helper()
	raw, err := exec.Command("git", "-C", repo, "cat-file", "commit", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	head, _, _ := strings.Cut(string(raw), "\n\n")
	parent, _ := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	var lines []string
	for _, l := range strings.Split(head, "\n") {
		if strings.HasPrefix(l, "tree ") {
			lines = append(lines, l, "parent "+strings.TrimSpace(string(parent)))
		}
		if strings.HasPrefix(l, "author ") || strings.HasPrefix(l, "committer ") {
			lines = append(lines, l)
		}
	}
	body := strings.Join(lines, "\n") +
		"\ngpgsig -----BEGIN PGP SIGNATURE-----\n \n iQEzBAABCAAdFiEE\n =abcd\n -----END PGP SIGNATURE-----\n\n[probe] fixture: signed\n"
	cmd := exec.Command("git", "-C", repo, "hash-object", "-w", "-t", "commit", "--stdin")
	cmd.Stdin = strings.NewReader(body)
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

// Property: in local mode no git command Proven starts runs a program the
// configuration names. Messages runs git log, and git log verifies a commit's
// signature with gpg.program when log.showSignature is set. Mutated field: the
// gpgsig header of the head commit.
func TestLocalMessagesStartNoSignatureProgram(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "gpg-ran")
	gpg := filepath.Join(t.TempDir(), "gpg")
	write(t, gpg, "#!/bin/sh\ntouch '"+marker+"'\nexit 1\n")
	if err := os.Chmod(gpg, 0o700); err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	git(t, repo, "init")
	write(t, filepath.Join(repo, "f.txt"), "one\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "[probe] fixture: one")
	base, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	signed := signedChild(t, repo)
	git(t, repo, "config", "log.showSignature", "true")
	git(t, repo, "config", "gpg.program", gpg)

	// Control: online git with a writable temp directory does start the program,
	// so a pass below is not a program that cannot run.
	ctl := exec.Command("git", "-C", repo, "log", "--format=%B", strings.TrimSpace(string(base))+".."+signed)
	ctl.Env = append(os.Environ(), "TMPDIR="+t.TempDir())
	if out, err := ctl.CombinedOutput(); err != nil {
		t.Fatalf("control log: %v %s", err, out)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Skip("this host's git did not start gpg.program even online; the property cannot be measured here")
	}
	os.Remove(marker)

	defer UseOffline(UseOffline(true))
	if _, err := Messages(repo, strings.TrimSpace(string(base)), signed); err != nil {
		t.Logf("messages: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("local mode: git log started gpg.program to verify a signed commit (log.showSignature)")
	}
}

// Property: in local mode the user's attributes file (core.attributesFile) names
// no filter driver for the working-tree reads. Mutated field: core.attributesFile
// in ~/.gitconfig, pointing at a file that names the driver; .gitattributes names none.
func TestLocalWorkingTreeReadsIgnoreTheUserAttributesFile(t *testing.T) {
	home := t.TempDir()
	marker := filepath.Join(t.TempDir(), "driver-ran")
	script := filepath.Join(home, "driver.sh")
	write(t, script, "#!/bin/sh\ntouch '"+marker+"'\nexec cat\n")
	if err := os.Chmod(script, 0o700); err != nil {
		t.Fatal(err)
	}
	attrs := filepath.Join(home, "attributes")
	write(t, attrs, "f.txt filter=probe\n")
	write(t, filepath.Join(home, ".gitconfig"), "[filter \"probe\"]\n\tclean = "+script+"\n[core]\n\tattributesFile = "+attrs+"\n")
	t.Setenv("HOME", home)
	repo := t.TempDir()
	git(t, repo, "init")
	write(t, filepath.Join(repo, "f.txt"), "hello\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "[probe] fixture: one")
	os.Remove(marker)
	write(t, filepath.Join(repo, "f.txt"), "changed\n")
	defer UseOffline(UseOffline(true))
	if _, err := Between(repo, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("local mode: Between ran the filter driver named by core.attributesFile")
	}
	// Control: online, the same read runs the driver, so the pass above can fail.
	UseOffline(false)
	if _, err := Between(repo, "HEAD", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("control: online, Between did not run the driver named by core.attributesFile")
	}
}
