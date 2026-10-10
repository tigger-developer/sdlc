// ABOUTME: Verifies the overall audit deadline stops recovery and preserves lockout evidence.
// ABOUTME: Exercises the CLI boundary with a local, silent provider process.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestOverallDeadlineStopsBeforeAnotherAttempt(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	// #nosec G306 -- fixture-owned executable never contacts an external model.
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte("#!/bin/sh\ncat >/dev/null\nprintf x >> \"$PROBE_CALLS\"\nsleep 20\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	calls := filepath.Join(root, "calls")
	t.Setenv("PROBE_CALLS", calls)
	config := filepath.Join(root, "global.yaml")
	registry := filepath.Join(root, "prompts.yaml")
	if err := os.WriteFile(config, []byte("delivery:\n  audit:\n    harness: codex\n    model: fixture\n    timeout: 10s\n    total_timeout: 1s\n    max_failures: 9\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(registry, []byte("version: 1\nsession_recovery_instructions: recover\ngates:\n  delivery-code:\n    prompt: audit\n"), 0600); err != nil {
		t.Fatal(err)
	}
	writeReadyDocuments(t, root)
	record := filepath.Join(root, "audits.yaml")
	args := []string{"--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "delivery-code", "--audit-record", record, "--work-item", "W016-overall-deadline", "--input", filepath.Join(root, "spec.org")}
	started := time.Now()
	var output, diagnostics bytes.Buffer
	err := runFixture(args, strings.NewReader("audit"), &output, &diagnostics)
	if err == nil || !strings.Contains(err.Error(), "overall audit deadline") || !strings.Contains(err.Error(), "Human intervention required") || time.Since(started) > 5*time.Second {
		t.Fatalf("deadline error=%v elapsed=%s", err, time.Since(started))
	}
	entry, found, readErr := harness.ReadAuditEntry(record, "W016-overall-deadline", "implementation")
	if readErr != nil || !found || !entry.FailureBlocked() || entry.RoundsUsed() != 0 || output.Len() != 0 {
		t.Fatalf("deadline evidence=%#v err=%v output=%q", entry, readErr, output.String())
	}
	body, readErr := os.ReadFile(calls)
	if readErr != nil || string(body) != "x" {
		t.Fatalf("provider launches=%q err=%v", body, readErr)
	}
	if err := runFixture(args, strings.NewReader("audit"), &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("deadline lockout allowed another audit")
	}
	body, readErr = os.ReadFile(calls)
	if readErr != nil || string(body) != "x" {
		t.Fatal("locked audit relaunched provider")
	}
}
