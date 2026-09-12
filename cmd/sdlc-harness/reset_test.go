package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestResetSessionPreservesHistoryAndNativeContext(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "audits.yaml")
	entry := harness.AuditEntry{WorkItem: "W015-reset", Gate: "implementation", SessionID: "native-existing", Harness: "codex", ExternalRound: 5, Status: "exhausted", Findings: []string{"retained finding"}, History: []harness.AuditRound{{Round: 5, Incident: "resume-failed", Updated: "2026-09-13T00:00:00Z"}}}
	if err := harness.WriteAuditEntry(path, entry); err != nil {
		t.Fatal(err)
	}
	other := harness.AuditEntry{WorkItem: entry.WorkItem, Gate: "definition", SessionID: "other-gate", ExternalRound: 4, Status: "active"}
	if err := harness.WriteAuditEntry(path, other); err != nil {
		t.Fatal(err)
	}
	other, _, err := harness.ReadAuditEntry(path, other.WorkItem, other.Gate)
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"resume", "--reset-session", "--project", root, "--gate", "implementation", "--work-item", entry.WorkItem, "--audit-record", "audits.yaml"}
	var output bytes.Buffer
	if err := run(args, strings.NewReader(""), &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	got, found, err := harness.ReadAuditEntry(path, entry.WorkItem, entry.Gate)
	if err != nil || !found || got.RoundsUsed() != 0 || len(got.BudgetResets) != 1 || got.SessionID != entry.SessionID || got.ExternalRound != 5 || !reflect.DeepEqual(got.History, entry.History) || !reflect.DeepEqual(got.Findings, entry.Findings) {
		t.Fatalf("reset lost history: %#v (%v)", got, err)
	}
	otherAfter, _, err := harness.ReadAuditEntry(path, other.WorkItem, other.Gate)
	if err != nil || !reflect.DeepEqual(other, otherAfter) {
		t.Fatal("reset changed another gate")
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := run(args, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("repeated reset changed record")
	}
}

func TestResetSessionRejectsAmbiguousInvocation(t *testing.T) {
	for _, args := range [][]string{
		{"start", "--reset-session", "--gate", "definition"},
		{"resume", "--reset-session", "--phase", "build"},
		{"resume", "--reset-session", "--gate", "definition", "--audit-record", "absent", "--work-item", "W015-reset", "--input", "spec.org"},
		{"resume", "--reset-session", "--gate", "definition", "--audit-record", "absent", "--work-item", "W015-reset", "--session", "native"},
	} {
		if err := run(args, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestResetSessionRejectsRunningOrMissingEntry(t *testing.T) {
	for _, status := range []string{"running", "missing"} {
		t.Run(status, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "audits.yaml")
			entry := harness.AuditEntry{WorkItem: "W015-reset", Gate: "definition", SessionID: "native", ExternalRound: 5, Status: status}
			if err := harness.WriteAuditEntry(path, entry); err != nil {
				t.Fatal(err)
			}
			work := entry.WorkItem
			if status == "missing" {
				work = "W099-absent"
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			err = run([]string{"resume", "--reset-session", "--project", root, "--gate", "definition", "--work-item", work, "--audit-record", path}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
			if err == nil {
				t.Fatal("unsafe reset accepted")
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("rejected reset mutated record")
			}
		})
	}
}
