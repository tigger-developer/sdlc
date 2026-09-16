package main

import (
	"bytes"
	"errors"
	"github.com/tigger-developer/sdlc/internal/harness"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Local fixtures only; these tests never call metered providers.
const readySpec = "* Acceptance Criteria\n** AC001.1 - Expected outcome :ac:\n* Test Definitions\n** RT001.1 - Protect outcome :testdef:\n"
const readyValidation = "#+TODO: AMBER RED | GREEN\n* Results\n** GREEN RT001.1 - Protect outcome :testdef:\n:PROPERTIES:\n:REVISION: fixture\n:COMMAND: make test\n:EVIDENCE: Local fixture passed\n:END:\n"

func writeReadyDocuments(t *testing.T, root string) {
	t.Helper()
	for name, contents := range map[string]string{"spec.org": readySpec, "validation.org": readyValidation} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReadinessRefusesBeforeConfigPromptOrAuditWrites(t *testing.T) {
	for _, invalid := range []bool{false, true} {
		root := t.TempDir()
		writeReadyDocuments(t, root)
		validation := strings.Replace(readyValidation, "GREEN RT", "RT", 1)
		if invalid {
			validation += "** RT001.1 - Duplicate :testdef:\n"
		}
		if err := os.WriteFile(filepath.Join(root, "validation.org"), []byte(validation), 0o600); err != nil {
			t.Fatal(err)
		}
		var out, diagnostics bytes.Buffer
		// Invalid config and unreadable stdin would fail if preflight were too late.
		args := []string{"start", "--project", root, "--gate", "implementation", "--audit-record", "audits.yaml", "--work-item", "W001", "--global-config", root}
		err := run(args, rejectRead{}, &out, &diagnostics)
		var refusal *readinessRefusal
		if !errors.As(err, &refusal) || !strings.Contains(out.String(), "ready: false") {
			t.Fatalf("expected early YAML refusal: %v %s", err, out.String())
		}
		if _, err := os.Stat(filepath.Join(root, "audits.yaml")); !os.IsNotExist(err) {
			t.Fatal("refusal wrote audit record")
		}
	}
}

type rejectRead struct{}

func (rejectRead) Read([]byte) (int, error) { return 0, errors.New("stdin must not be read") }

func TestTwoReviewsRetainSessionAndKeepStageSpecificEvidence(t *testing.T) {
	root := t.TempDir()
	writeReadyDocuments(t, root)
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
output=
previous=
for argument in "$@"; do
    if [ "$previous" = "-o" ]; then output="$argument"; fi
    previous="$argument"
done
cat >/dev/null
printf 'GATE: implementation\nREVISION: fixture\nVERDICT: PASS\n' > "$output"
printf '{"type":"thread.started","thread_id":"retained-test-session"}\n'
`
	// #nosec G306 -- executable local double, never a metered provider.
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	registry := filepath.Join(root, "prompts.yaml")
	if err := os.WriteFile(registry, []byte("version: 1\ngates:\n  test-code:\n    prompt: Review written tests.\n  implementation:\n    prompt: Review final package.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(root, "audits.yaml")
	for i, stage := range []string{"test-code", "delivery-code", "delivery-code"} {
		action := "resume"
		if i == 0 {
			action = "start"
		}
		args := []string{action, "--project", root, "--global-config", filepath.Join(root, "absent.yaml"), "--harness", "codex", "--model", "fixture", "--gate", "implementation", "--readiness-check", stage, "--audit-record", record, "--work-item", "W001", "--audit-prompts", registry}
		if err := run(args, strings.NewReader("review"), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		entry, _, err := harness.ReadAuditEntry(record, "W001", "implementation")
		if err != nil || entry.SessionID != "retained-test-session" || entry.ReadinessCheck != stage {
			t.Fatalf("entry: %+v, %v", entry, err)
		}
		if i == 0 && entry.Status == "passed" {
			t.Fatal("test review approved delivery")
		}
		if i > 0 && (len(entry.History) != 2 || entry.History[0].ReadinessCheck != "test-code" || entry.History[1].ReadinessCheck != "delivery-code" || entry.Status != "passed") {
			t.Fatalf("lost stages, session reuse or cache: %+v", entry)
		}
	}
}
