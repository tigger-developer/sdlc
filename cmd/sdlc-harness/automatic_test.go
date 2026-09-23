// ABOUTME: Exercises automatic audit invocation and bounded malformed-response recovery.
// ABOUTME: Uses local provider processes only, never a metered service.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestAutomaticAuditRecoversWithoutSpendingVerdictBudget(t *testing.T) {
	for _, mode := range []string{"recover", "broken", "configured"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("AUTO_ROOT", root)
			t.Setenv("AUTO_MODE", mode)
			script := `#!/bin/sh
cat > "$AUTO_ROOT/prompt"
printf '%s\n' "$@" > "$AUTO_ROOT/arguments"
printf 'call\n' >> "$AUTO_ROOT/calls"
printf 'session_id: private-native-id\n' >&2
if [ "$AUTO_MODE" != recover ] || [ ! -f "$AUTO_ROOT/tried" ]; then
  touch "$AUTO_ROOT/tried"
  printf 'unusable audit body\n'
else
  printf 'GATE: definition\nREVISION: fixture\nVERDICT: FAIL\nA concrete finding.\n'
fi
`
			// #nosec G306 -- executable local provider double.
			if err := os.WriteFile(filepath.Join(root, "hermes"), []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			config := filepath.Join(root, "config.yaml")
			registry := filepath.Join(root, "prompts.yaml")
			for path, value := range map[string]string{
				config:   "delivery:\n  audit:\n    harness: hermes\n    provider: fixture\n    model: fixture\n    max_rounds: 5\n",
				registry: "version: 1\nsession_recovery_instructions: Preserve history.\ngates:\n  definition:\n    prompt: Return the audit envelope.\n",
			} {
				if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			record := filepath.Join(root, "audits.yaml")
			if mode == "configured" {
				if err := os.WriteFile(config, []byte("delivery:\n  audit:\n    harness: hermes\n    provider: fixture\n    model: fixture\n    max_failures: 2\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "definition", "--audit-record", record, "--work-item", "W019", "--input", registry}
			var out, diagnostics bytes.Buffer
			err := run(args, strings.NewReader("audit"), &out, &diagnostics)
			if mode == "recover" && err != nil {
				t.Fatal(err)
			}
			if mode != "recover" && (err == nil || !strings.Contains(err.Error(), "Human intervention required")) {
				t.Fatalf("error = %v", err)
			}
			entry, found, readErr := harness.ReadAuditEntry(record, "W019", "definition")
			if readErr != nil || !found {
				t.Fatalf("record missing: %v", readErr)
			}
			wantCalls, wantVerdicts := 2, 1
			if mode != "recover" {
				wantCalls, wantVerdicts = 3, 0
				if mode == "configured" {
					wantCalls = 2
				}
			}
			calls, readErr := os.ReadFile(filepath.Join(root, "calls"))
			if readErr != nil || strings.Count(string(calls), "call") != wantCalls || entry.RoundsUsed() != wantVerdicts {
				t.Fatalf("calls=%q verdicts=%d err=%v", calls, entry.RoundsUsed(), readErr)
			}
			if strings.Contains(diagnostics.String(), "private-native-id") || strings.Contains(diagnostics.String(), "attempt") {
				t.Fatalf("internal state leaked: %s", &diagnostics)
			}
			logs, readErr := filepath.Glob(filepath.Join(root, ".sdlc", "audit-diagnostics", "*.log"))
			if readErr != nil || len(logs) == 0 {
				t.Fatalf("diagnostics missing: %v", readErr)
			}
			body, readErr := os.ReadFile(logs[0])
			if readErr != nil || !bytes.Contains(body, []byte("unusable audit body")) {
				t.Fatalf("rejected response missing: %s %v", body, readErr)
			}
			out.Reset()
			_ = run(args, strings.NewReader("audit"), &out, &diagnostics)
			after, readErr := os.ReadFile(filepath.Join(root, "calls"))
			if readErr != nil || !bytes.Equal(calls, after) {
				t.Fatal("cached verdict or lockout invoked provider")
			}
			if mode == "recover" {
				prompt, err := os.ReadFile(filepath.Join(root, "prompt"))
				if err != nil || !bytes.Contains(prompt, []byte("previous response was unusable")) || !bytes.Contains(prompt, []byte("GATE:, REVISION:, and VERDICT:")) {
					t.Fatal("missing format clarification")
				}
			}
			reset := []string{"--reset", "--project", root, "--gate", "definition", "--audit-record", record, "--work-item", "W019"}
			if err := run(reset, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			t.Setenv("AUTO_MODE", "recover")
			if err := run(args, strings.NewReader("audit"), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			arguments, err := os.ReadFile(filepath.Join(root, "arguments"))
			if err != nil || bytes.Contains(arguments, []byte("--resume")) {
				t.Fatal("reset resumed old context")
			}
			entry, _, err = harness.ReadAuditEntry(record, "W019", "definition")
			if err != nil || entry.RoundsUsed() != 1 || entry.ConsecutiveFailures() != 0 || entry.FailureBlocked() {
				t.Fatal("reset did not clear both counts")
			}
		})
	}
}
