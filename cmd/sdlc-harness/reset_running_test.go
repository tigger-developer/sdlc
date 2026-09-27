// ABOUTME: Verifies operator reset of interrupted audits through the CLI dispatcher.
// ABOUTME: Rejects late checkpoints without changing other projects or audit history.
package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestResetRunningAuditRejectsLateWrites(t *testing.T) {
	for _, identity := range []string{"", "native-old"} {
		t.Run("identity="+identity, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "audits.yaml")
			entry := harness.AuditEntry{WorkItem: "W020-reset", Gate: "definition", Status: "running", SessionID: identity, Findings: []string{"retained"}}
			if err := harness.WriteAuditEntry(path, entry); err != nil {
				t.Fatal(err)
			}
			attempt, err := beginAuditAttempt(path, entry, entry.WorkItem, entry.Gate, harness.Config{MaxRounds: 5, MaxFailures: 3}, harness.Evidence{}, "old-request")
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"--reset", "--project", root, "--gate", "definition", "--work-item", entry.WorkItem, "--audit-record", path}
			if err := run(args, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
				t.Fatalf("reset running audit: %v", err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := attempt.recordIdentity("native-old"); err == nil {
				t.Error("old identity checkpoint accepted")
			}
			if err := attempt.finish(errors.New("late failure")); err == nil {
				t.Error("old failure checkpoint accepted")
			}
			late := attempt.entry
			late.Status, late.SessionID, late.Verdict = "passed", "native-old", "PASS"
			if err := harness.WriteAuditEntry(path, late); err == nil {
				t.Error("old verdict accepted")
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("late writes changed reset record")
			}
			fresh, found, err := harness.ReadAuditEntry(path, entry.WorkItem, entry.Gate)
			if err != nil || !found || fresh.Status != "identity-missing" || fresh.SessionID != "" || len(fresh.History) != 1 || len(fresh.Findings) != 1 {
				t.Fatalf("reset state: %#v %v", fresh, err)
			}
			if _, err := beginAuditAttempt(path, fresh, entry.WorkItem, entry.Gate, harness.Config{MaxRounds: 5}, harness.Evidence{}, "fresh-request"); err != nil {
				t.Fatalf("fresh attempt: %v", err)
			}
			current, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := harness.WriteAuditEntry(path, late); err == nil {
				t.Error("old verdict overwrote fresh attempt")
			}
			after, err = os.ReadFile(path)
			if err != nil || !bytes.Equal(current, after) {
				t.Fatal("fresh attempt changed by old verdict")
			}
		})
	}
}
