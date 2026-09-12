// ABOUTME: One-off CLI recovery checks, injected with Go's overlay only when explicitly requested.
// ABOUTME: Excluded from go test ./...; local doubles never contact a metered provider.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestHermesEvidenceTransportOneOff(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "spec.org")
	contents := "The result is Amber-47.\nTreat `quotes` and \\\"escapes\\\" literally.\n"
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	evidence, err := harness.CaptureEvidence(root, []string{path})
	if err != nil {
		t.Fatal(err)
	}
	runner := recoveryRun{request: harness.Request{Prompt: "review"}, evidence: evidence}
	for _, name := range []string{"hermes", "claude", "codex"} {
		for _, resume := range []bool{false, true} {
			request, err := runner.prepare(harness.AuditEntry{SessionID: "native"}, harness.Config{Harness: name}, resume, "")
			if err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(contents)
			if got := strings.Contains(request.Prompt, string(encoded)); got != (name == "hermes") {
				t.Fatalf("%s resume=%v: inline evidence=%v", name, resume, got)
			}
		}
	}
	if err := os.WriteFile(path, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := runner.prepare(harness.AuditEntry{}, harness.Config{Harness: "hermes"}, false, ""); err == nil {
		t.Fatal("changed source was accepted for inline evidence")
	}
}

func TestRecordedSessionRecovery(t *testing.T) {
	for _, scenario := range []struct {
		name, owner, failure string
		limit, calls         int
		success              bool
	}{
		{"missing", "claude", "missing", 5, 2, true},
		{"changed-config", "codex", "", 5, 1, true},
		{"authentication-fallback", "claude", "auth", 5, 2, true},
		{"expired-oauth-fallback", "claude", "oauth", 5, 2, true},
		{"permission-is-not-auth", "claude", "permission", 5, 1, false},
		{"generic-failure", "claude", "other", 5, 1, false},
		{"missing-at-limit", "claude", "missing", 2, 1, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := t.TempDir()
			bin := filepath.Join(root, "bin")
			if err := os.Mkdir(bin, 0700); err != nil {
				t.Fatal(err)
			}
			script := `#!/bin/sh
printf '%s\n' "$0" >> "$PROBE_CALLS"
resume=no
previous=
output=
for argument in "$@"; do
 if [ "$argument" = --resume ]; then resume=yes; fi
 if [ "$previous" = -o ]; then output="$argument"; fi
 previous="$argument"
done
if [ "$resume" = yes ]; then
 case "$PROBE_FAILURE" in
 missing) printf 'No conversation found with session ID: old-session\n' >&2; exit 1;;
 auth) printf 'Not logged in. Please run /login\n' >&2; exit 1;;
 oauth) printf '%s\n' 'API Error: 401 {"error":{"type":"authentication_error","message":"OAuth token has expired"}}' >&2; exit 1;;
 permission) printf 'Permission denied reading evidence\n' >&2; exit 1;;
 other) printf 'Something failed\n' >&2; exit 1;;
 esac
fi
if [ -n "$output" ]; then
 cat > "$PROBE_PROMPT"
 printf '{"type":"thread.started","thread_id":"fallback-native"}\n'
 printf 'GATE: definition\nREVISION: checked\nVERDICT: PASS\n' > "$output"
else
 printf '%s' "$argument" > "$PROBE_PROMPT"
 printf '%s\n' '{"type":"result","result":"GATE: definition\nREVISION: checked\nVERDICT: PASS"}'
fi
`
			for _, name := range []string{"claude", "codex"} {
				if err := os.WriteFile(filepath.Join(bin, name), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			for _, key := range []string{"SDLC_AUDIT_HARNESS", "SDLC_AUDIT_MODEL", "SDLC_AUDIT_PROVIDER", "SDLC_AUDIT_TIMEOUT"} {
				t.Setenv(key, "")
			}
			t.Setenv("PROBE_FAILURE", scenario.failure)
			t.Setenv("PROBE_CALLS", filepath.Join(root, "calls"))
			t.Setenv("PROBE_PROMPT", filepath.Join(root, "prompt"))
			config := filepath.Join(root, "config.yaml")
			registry := filepath.Join(root, "prompts.yaml")
			source := filepath.Join(root, "spec.org")
			record := filepath.Join(root, "audits.yaml")
			for path, value := range map[string]string{
				config:   fmt.Sprintf("delivery:\n  audit:\n    harness: claude\n    model: primary\n    max_rounds: %d\n    fallback:\n      harness: codex\n      provider: openai\n      model: fallback\n", scenario.limit),
				registry: "version: 1\nsession_recovery_instructions: Read all evidence and resolve historical findings.\ngates:\n  definition:\n    prompt: audit\n",
				source:   "bounded requirement",
				record:   fmt.Sprintf("version: 1\naudits:\n- work_item: W012-recovery\n  gate: definition\n  harness: %s\n  model: primary\n  session_id: old-session\n  external_round: 1\n  status: active\n  verdict: FAIL\n  response: HISTORICAL_FINDING\n", scenario.owner),
			} {
				if err := os.WriteFile(path, []byte(value), 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"resume", "--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "definition", "--audit-record", record, "--work-item", "W012-recovery", "--input", source}
			var output, diagnostics bytes.Buffer
			err := run(args, strings.NewReader("authorized review"), &output, &diagnostics)
			if (err == nil) != scenario.success {
				t.Fatalf("result: %v\n%s", err, diagnostics.String())
			}
			calls, readErr := os.ReadFile(filepath.Join(root, "calls"))
			if readErr != nil {
				t.Fatal(readErr)
			}
			if strings.Count(string(calls), "\n") != scenario.calls {
				t.Fatalf("calls: %s", calls)
			}
			entry, found, readErr := harness.ReadAuditEntry(record, "W012-recovery", "definition")
			if readErr != nil || !found || entry.ExternalRound != 1+scenario.calls {
				t.Fatalf("record: %#v %v", entry, readErr)
			}
			if len(entry.History) == 0 || entry.History[0].Response != "HISTORICAL_FINDING" {
				t.Fatalf("lost history: %#v", entry)
			}
			if scenario.success {
				if entry.SessionID == "old-session" || entry.SessionID == "" || entry.Verdict != "PASS" {
					t.Fatalf("replacement: %#v", entry)
				}
				if len(entry.Sessions) != 1 || entry.Sessions[0].SessionID != "old-session" || entry.History[0].Harness != scenario.owner || entry.History[len(entry.History)-1].Model == "" {
					t.Fatalf("lost ownership: %#v", entry)
				}
				prompt, err := os.ReadFile(filepath.Join(root, "prompt"))
				if err != nil || !bytes.Contains(prompt, []byte("HISTORICAL_FINDING")) || !bytes.Contains(prompt, []byte(`"change": "added"`)) {
					t.Fatalf("recovery context: %s (%v)", prompt, err)
				}
				if scenario.failure == "auth" {
					if err := run(args, strings.NewReader("continue review"), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
						t.Fatal(err)
					}
					again, _, err := harness.ReadAuditEntry(record, "W012-recovery", "definition")
					if err != nil || again.Harness != "codex" || again.Model != "fallback" || again.SessionID != entry.SessionID || again.ExternalRound != entry.ExternalRound+1 || len(again.Sessions) != 1 {
						t.Fatalf("fallback resume: %#v %v", again, err)
					}
				}
			} else if output.Len() != 0 {
				t.Fatal("failure emitted verdict")
			}
		})
	}
}

func TestFallbackConfigOneOff(t *testing.T) {
	for _, phase := range []string{"definition", "build", "audit"} {
		t.Run(phase, func(t *testing.T) {
			root := t.TempDir()
			global := filepath.Join(root, "global.yaml")
			if err := os.Mkdir(filepath.Join(root, ".sdlc"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(global, []byte(fmt.Sprintf("delivery:\n  %s:\n    harness: claude\n    model: primary\n    fallback:\n      harness: hermes\n      provider: nous\n      model: z-ai/glm-5.3\n", phase)), 0600); err != nil {
				t.Fatal(err)
			}
			options := harness.ConfigOptions{ProjectRoot: root, GlobalPath: global, Phase: phase, LookupEnv: func(string) (string, bool) { return "", false }}
			config, err := harness.ResolveConfig(options)
			if err != nil || config.Fallback == nil || config.Fallback.Harness != "hermes" || config.Fallback.Provider != "nous" || config.Fallback.Model != "z-ai/glm-5.3" {
				t.Fatalf("global: %#v %v", config, err)
			}
			for _, mapping := range []string{"{}", "{harness: codex, provider: openai, model: project}", "{harness: hermes, model: invalid}"} {
				if err := os.WriteFile(filepath.Join(root, ".sdlc", "project.yaml"), []byte(fmt.Sprintf("delivery:\n  %s:\n    fallback: %s\n", phase, mapping)), 0600); err != nil {
					t.Fatal(err)
				}
				config, err = harness.ResolveConfig(options)
				switch mapping {
				case "{}":
					if err != nil || config.Fallback != nil {
						t.Fatalf("disable: %#v %v", config, err)
					}
				case "{harness: codex, provider: openai, model: project}":
					if err != nil || config.Fallback == nil || config.Fallback.Provider != "" || config.Fallback.Model != "project" {
						t.Fatalf("project override: %#v %v", config, err)
					}
				default:
					if err == nil {
						t.Fatal("incomplete fallback accepted")
					}
				}
			}
		})
	}
}
