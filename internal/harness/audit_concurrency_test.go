// ABOUTME: Exercises audit-record isolation and serialized read-modify-write operations.
// ABOUTME: Uses local records only, with no external provider calls.
package harness

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

func TestResetRejectsStaleSnapshotAndIsolatesProject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audits.yaml")
	other := filepath.Join(t.TempDir(), "audits.yaml")
	entry := AuditEntry{WorkItem: "W020", Gate: "definition", Status: "active", SessionID: "old"}
	for _, p := range []string{path, other} {
		if err := WriteAuditEntry(p, entry); err != nil {
			t.Fatal(err)
		}
	}
	before, err := os.ReadFile(other)
	if err != nil {
		t.Fatal(err)
	}
	if err := ResetAuditBudget(path, entry.WorkItem, entry.Gate); err != nil {
		t.Fatal(err)
	}
	if err := WriteAuditEntry(path, entry); err == nil {
		t.Fatal("stale snapshot accepted after reset")
	}
	after, err := os.ReadFile(other)
	if err != nil || string(before) != string(after) {
		t.Fatal("another project changed")
	}
	if err := WriteAuditEntry(other, entry); err != nil {
		t.Fatalf("another project blocked: %v", err)
	}
}

func TestConcurrentAuditEntriesSurvive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audits.yaml")
	start := make(chan struct{})
	errs := make(chan error, 32)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			errs <- WriteAuditEntry(path, AuditEntry{WorkItem: fmt.Sprintf("W%03d", i), Gate: "definition", SessionID: "native", Status: "active"})
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 32; i++ {
		if _, found, err := ReadAuditEntry(path, fmt.Sprintf("W%03d", i), "definition"); err != nil || !found {
			t.Fatalf("concurrent entry %d lost: %v", i, err)
		}
	}
}

func TestAuditRecordWriterProcess(t *testing.T) {
	path := os.Getenv("SDLC_TEST_AUDIT_RECORD")
	if path == "" {
		return
	}
	if err := WriteAuditEntry(path, AuditEntry{WorkItem: os.Getenv("SDLC_TEST_AUDIT_WORK"), Gate: "definition", SessionID: "native", Status: "active"}); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentAuditProcessesSurvive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audits.yaml")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmds := make([]*exec.Cmd, 16)
	for i := range cmds {
		cmd := exec.Command(exe, "-test.run=^TestAuditRecordWriterProcess$")
		cmd.Env = append(os.Environ(), "SDLC_TEST_AUDIT_RECORD="+path, fmt.Sprintf("SDLC_TEST_AUDIT_WORK=W%03d", i))
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		cmds[i] = cmd
	}
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	for i := range cmds {
		if _, found, err := ReadAuditEntry(path, fmt.Sprintf("W%03d", i), "definition"); err != nil || !found {
			t.Fatalf("process entry %d lost: %v", i, err)
		}
	}
}

func TestResetThroughAliasFencesOriginalWriter(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "audits.yaml")
	alias := filepath.Join(root, "alias.yaml")
	entry := AuditEntry{WorkItem: "W020", Gate: "definition", Status: "running"}
	if err := WriteAuditEntry(path, entry); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(path, alias); err != nil {
		t.Fatal(err)
	}
	if err := ResetAuditBudget(alias, entry.WorkItem, entry.Gate); err != nil {
		t.Fatal(err)
	}
	if err := WriteAuditEntry(path, entry); err == nil {
		t.Fatal("alias reset did not fence original writer")
	}
}
