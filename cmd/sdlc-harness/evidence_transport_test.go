package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestAuditEvidenceTransportMatchesProviderFileAccess(t *testing.T) {
	project := t.TempDir()
	path := filepath.Join(project, "spec.org")
	if err := os.WriteFile(path, []byte("UNIQUE_AUDIT_EVIDENCE_MARKER\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	evidence, err := harness.CaptureEvidence(project, []string{path})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"codex", "claude", "copilot", "hermes"} {
		t.Run(name, func(t *testing.T) {
			runner := recoveryRun{request: harness.Request{Prompt: "Audit this project.", Evidence: evidence.Files}, evidence: evidence}
			request, err := runner.prepare(harness.AuditEntry{}, harness.Config{Harness: name}, false, "")
			if err != nil {
				t.Fatal(err)
			}
			inlined := strings.Contains(request.Prompt, "UNIQUE_AUDIT_EVIDENCE_MARKER")
			if inlined != (name == "copilot" || name == "hermes") {
				t.Fatalf("%s inline evidence = %t", name, inlined)
			}
			if !strings.Contains(request.Prompt, path) {
				t.Fatalf("%s prompt omitted original evidence path", name)
			}
		})
	}
}
