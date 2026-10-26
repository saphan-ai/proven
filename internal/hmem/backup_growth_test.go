package hmem_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// manyBackups writes n placeholder backup files, as a store that has slept
// many times would hold, and returns the backups directory.
func manyBackups(t *testing.T, stateDir string, n int) string {
	t.Helper()
	dir := filepath.Join(stateDir, "backups")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= n; i++ {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("backup-%d.json", i)), []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestSleepDoesNotListTheBackupDirectory proves R16 for backups: with 3,000
// backups already on disk, a sleep takes its id from the persisted sequence and
// lists no directory entry. A store with backups and no sequence file yet is
// listed once, to find the highest number, and never again. A sleep that lists
// the directory per cycle (the old shape) counts one listing per sleep and
// turns this red.
func TestSleepDoesNotListTheBackupDirectory(t *testing.T) {
	const existing = 3000
	dir, logs := t.TempDir(), t.TempDir()
	manyBackups(t, dir, existing)
	s := openApprove(t, dir, logs)
	listings := s.CountDirListings()

	for i := 1; i <= 3; i++ {
		rep, err := s.SleepExecute(nil)
		if err != nil {
			t.Fatal(err)
		}
		if want := fmt.Sprintf("backup-%d", existing+i); rep.BackupID != want {
			t.Fatalf("sleep %d made %s, want %s", i, rep.BackupID, want)
		}
	}
	if got := listings(); got != 1 {
		t.Fatalf("3 sleeps listed the backup directory %d times, want 1 (the migration only)", got)
	}
}

// TestBackupSequenceIsMonotonicAndNeverOverwrites proves the allocator is
// collision-safe: with a sequence file in place no listing happens at all, ids
// keep rising across a reopen, and a backup file that already sits at the next
// number (a stale sequence) is stepped over and left as it was.
func TestBackupSequenceIsMonotonicAndNeverOverwrites(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	bdir := manyBackups(t, dir, 0)
	if err := os.WriteFile(filepath.Join(bdir, "SEQ"), []byte("5000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// A file at the number after the sequence: it is another snapshot's.
	if err := os.WriteFile(filepath.Join(bdir, "backup-5001.json"), []byte("precious\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := openApprove(t, dir, logs)
	listings := s.CountDirListings()
	rep, err := s.SleepExecute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.BackupID != "backup-5002" {
		t.Fatalf("sleep made %s, want backup-5002 (5001 is taken)", rep.BackupID)
	}
	if b, _ := os.ReadFile(filepath.Join(bdir, "backup-5001.json")); string(b) != "precious\n" {
		t.Fatalf("an existing backup was overwritten: %q", b)
	}
	s = openApprove(t, dir, logs)
	rep, err = s.SleepExecute(nil)
	if err != nil || rep.BackupID != "backup-5003" {
		t.Fatalf("after a reopen sleep made %q, err %v, want backup-5003", rep.BackupID, err)
	}
	if got := listings(); got != 0 {
		t.Fatalf("the backup directory was listed %d times with a sequence file present", got)
	}
	if err := s.Restore("backup-5003"); err != nil {
		t.Fatalf("the newest backup cannot be restored: %v", err)
	}
}
