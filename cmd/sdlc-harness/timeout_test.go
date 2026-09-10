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
if [ "$PROBE_SUCCESS" = yes ]; then
    printf 'GATE: implementation\nREVISION: checked\nVERDICT: PASS\n' > "$output"
    exit 0
fi
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
	t.Setenv("PROBE_SUCCESS", "no")
	if recovery == "identity-missing" {
		t.Setenv("PROBE_IDENTITY", "no")
	}
	config := filepath.Join(root, "config.yaml")
	registry := filepath.Join(root, "prompts.yaml")
	source := filepath.Join(root, "spec.org")
	for path, value := range map[string]string{
		config:   "version: 3\ndelivery:\n  audit:\n    harness: codex\n    model: fixture\n    max_rounds: 2\n    timeout: 1s\n",
		registry: "version: 1\ntimeout_resume_instructions: CONTINUE_INTERRUPTED_FIXTURE\ntimeout_message: RETRY_SAME_SESSION_FIXTURE\ntimeout_blocked_message: STOP_FIXTURE\ngates:\n  implementation:\n    prompt: audit\n",
		source:   "unchanged requirement",
	} {
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	record := filepath.Join(root, "audits.yaml")
	args := []string{"--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "implementation", "--audit-record", record, "--work-item", "W006-recovery", "--input", source}
	for index, action := range []string{"start", "resume"} {
		if index == 1 && recovery == "success" {
			t.Setenv("PROBE_SUCCESS", "yes")
		}
		var output, diagnostics bytes.Buffer
		err := run(append([]string{action}, args...), strings.NewReader("review"), &output, &diagnostics)
		var incident *harness.Incident
		if recovery == "success" && index == 1 {
			if err != nil {
				t.Fatal(err)
			}
		} else if !errors.As(err, &incident) || incident.Kind != "timeout" {
			t.Fatalf("timeout = %v", err)
		}
		entry, found, err := harness.ReadAuditEntry(record, "W006-recovery", "implementation")
		if err != nil || !found || len(entry.History) != index+1 || entry.History[0].Incident != "timeout" || entry.History[0].Verdict != "" {
			t.Fatalf("timeout incident/history = %#v (%v)", entry, err)
		}
		if recovery == "identity-missing" {
			if err != nil || !found || entry.SessionID != "" || entry.Status != "identity-missing" || !strings.Contains(diagnostics.String(), "STOP_FIXTURE") {
				t.Fatalf("missing identity record: %#v (%v)", entry, err)
			}
			break
		}
		wantVerdict := ""
		if recovery == "success" && index == 1 {
			wantVerdict = "PASS"
		}
		if err != nil || !found || entry.SessionID != "native-timeout" || entry.ExternalRound != index+1 || entry.Verdict != wantVerdict {
			t.Fatalf("timeout record = %#v (%v)", entry, err)
		}
		if len(entry.LatestEvidence()) != 1 || (wantVerdict == "" && output.Len() != 0) {
			t.Fatal("lost evidence or emitted verdict")
		}
		if index == 0 && !strings.Contains(diagnostics.String(), "RETRY_SAME_SESSION_FIXTURE") {
			t.Fatal(diagnostics.String())
		}
		if index == 1 {
			prompt, err := os.ReadFile(filepath.Join(root, "prompt"))
			if err != nil || !bytes.Contains(prompt, []byte("CONTINUE_INTERRUPTED_FIXTURE")) || !bytes.Contains(prompt, []byte(`"change": "unchanged"`)) {
				t.Fatalf("resume prompt = %s (%v)", prompt, err)
			}
			if recovery == "exhausted" && !strings.Contains(diagnostics.String(), "STOP_FIXTURE") {
				t.Fatal(diagnostics.String())
			}
		}
	}
	before, err := os.ReadFile(filepath.Join(root, "calls"))
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"start", "resume"} {
		err := run(append([]string{action}, args...), strings.NewReader("review"), &bytes.Buffer{}, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("%s evaded exhausted attempt limit", action)
		}
	}
	after, err := os.ReadFile(filepath.Join(root, "calls"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("blocked invocation called provider: %q -> %q (%v)", before, after, err)
	}
}
