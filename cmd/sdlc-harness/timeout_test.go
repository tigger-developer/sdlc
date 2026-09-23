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

func TestTimeoutResumesRetainedEvidenceWithinLimit(t *testing.T) {
	for _, recovery := range []string{"exhausted", "success", "identity-missing"} {
		t.Run(recovery, func(t *testing.T) { checkTimeoutRecovery(t, recovery) })
	}
}

func checkTimeoutRecovery(t *testing.T, recovery string) {
	t.Helper()
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
printf x >> "$PROBE_CALLS"
output=
previous=
for argument in "$@"; do
    if [ "$previous" = "-o" ]; then output="$argument"; fi
    previous="$argument"
done
cat > "$PROBE_PROMPT"
if [ "$PROBE_IDENTITY" = yes ]; then printf '{"type":"thread.started","thread_id":"native-timeout"}\n'; fi
if [ "$PROBE_SUCCESS" = success ] && [ -f "$PROBE_PROMPT.tried" ]; then
    printf 'GATE: delivery-code\nREVISION: checked\nVERDICT: PASS\n' > "$output"
    exit 0
fi
touch "$PROBE_PROMPT.tried"
sleep 10
`
	// #nosec G306 -- executable local provider double; never a metered test.
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PROBE_PROMPT", filepath.Join(root, "prompt"))
	t.Setenv("PROBE_CALLS", filepath.Join(root, "calls"))
	t.Setenv("PROBE_IDENTITY", "yes")
	t.Setenv("PROBE_SUCCESS", recovery)
	if recovery == "identity-missing" {
		t.Setenv("PROBE_IDENTITY", "no")
	}
	config := filepath.Join(root, "config.yaml")
	registry := filepath.Join(root, "prompts.yaml")
	source := filepath.Join(root, "spec.org")
	for path, value := range map[string]string{
		config:   "version: 3\ndelivery:\n  audit:\n    harness: codex\n    model: fixture\n    max_rounds: 2\n    timeout: 1s\n",
		registry: "version: 1\ntimeout_resume_instructions: CONTINUE_INTERRUPTED_FIXTURE\ntimeout_message: RETRY_SAME_SESSION_FIXTURE\ntimeout_blocked_message: STOP_FIXTURE\ngates:\n  delivery-code:\n    prompt: audit\n",
		source:   "unchanged requirement",
	} {
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	record := filepath.Join(root, "audits.yaml")
	writeReadyDocuments(t, root)
	args := []string{"--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "delivery-code", "--audit-record", record, "--work-item", "W006-recovery", "--input", source}

	var output, diagnostics bytes.Buffer
	err := run(args, strings.NewReader("review"), &output, &diagnostics)
	var incident *harness.Incident
	wantAttempts, wantVerdicts := 3, 0
	if recovery == "success" {
		wantAttempts, wantVerdicts = 2, 1
		if err != nil {
			t.Fatal(err)
		}
	} else if !errors.As(err, &incident) || incident.Kind != "human-intervention" {
		t.Fatalf("timeout recovery = %v", err)
	}
	entry, found, err := harness.ReadAuditEntry(record, "W006-recovery", "implementation")
	if err != nil || !found || len(entry.History) != wantAttempts || entry.RoundsUsed() != wantVerdicts || entry.History[0].Incident != "timeout" || entry.History[0].Verdict != "" {
		t.Fatalf("timeout accounting = %#v (%v)", entry, err)
	}
	if len(entry.LatestEvidence()) != 2 {
		t.Fatal("lost evidence")
	}
	if recovery == "identity-missing" && entry.SessionID != "" {
		t.Fatal("invented native identity")
	}
	if recovery == "success" {
		prompt, err := os.ReadFile(filepath.Join(root, "prompt"))
		if err != nil || !bytes.Contains(prompt, []byte("CONTINUE_INTERRUPTED_FIXTURE")) || !bytes.Contains(prompt, []byte(`"change": "unchanged"`)) {
			t.Fatalf("resume prompt = %s (%v)", prompt, err)
		}
	} else if output.Len() != 0 {
		t.Fatal("incident emitted verdict")
	}
	before, err := os.ReadFile(filepath.Join(root, "calls"))
	if err != nil {
		t.Fatal(err)
	}
	_ = run(args, strings.NewReader("review"), &bytes.Buffer{}, &bytes.Buffer{})
	after, err := os.ReadFile(filepath.Join(root, "calls"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("cache or lockout relaunched provider")
	}
}
