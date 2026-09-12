package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

// RT014.4: exercise real process, output validation and audit persistence using
// local doubles only. No metered provider is part of this regression test.
func TestFallbackForUnusableAuditResult(t *testing.T) {
	for _, mode := range []string{"limit", "reset-limit", "launch-error", "timeout", "malformed", "wrong-gate", "empty", "PASS", "FAIL", "PROVISIONAL PASS", "fallback-fails", "exhausted", "evidence-changed"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			bin := filepath.Join(root, "bin")
			if err := os.Mkdir(bin, 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("FALLBACK_MODE", mode)
			if mode == "reset-limit" {
				t.Setenv("FALLBACK_MODE", "limit")
			}
			t.Setenv("FALLBACK_CALLS", filepath.Join(root, "calls"))
			source := filepath.Join(root, "spec.org")
			t.Setenv("FALLBACK_SOURCE", source)
			primary := `#!/bin/sh
cat >/dev/null
printf 'primary\n' >> "$FALLBACK_CALLS"
case "$FALLBACK_MODE" in
limit|fallback-fails|exhausted) printf '%s\n' "You've hit your session limit · resets 4:40am (Europe/Dublin)" >&2; exit 1;;
launch-error) exit 2;;
timeout) sleep 10; exit 1;;
malformed) printf '{"type":"result","result":"not a verdict"}\n'; exit 0;;
wrong-gate) printf '%s\n' '{"type":"result","result":"GATE: definition\nREVISION: fixture\nVERDICT: PASS"}'; exit 0;;
empty) printf '{"type":"result","result":""}\n'; exit 0;;
evidence-changed) printf 'changed' > "$FALLBACK_SOURCE";;
esac
verdict="$FALLBACK_MODE"
if [ "$verdict" = evidence-changed ]; then verdict=PASS; fi
printf '{"type":"result","result":"GATE: implementation\\nREVISION: fixture\\nVERDICT: %s"}\n' "$verdict"
`
			fallback := `#!/bin/sh
cat >/dev/null
printf 'fallback\n' >> "$FALLBACK_CALLS"
printf 'session_id: fallback-session\n' >&2
if [ "$FALLBACK_MODE" = fallback-fails ]; then exit 3; fi
printf 'GATE: implementation\nREVISION: fixture\nVERDICT: PASS\n'
`
			for name, body := range map[string]string{"claude": primary, "hermes": fallback} {
				// #nosec G306 -- local executable command double.
				if err := os.WriteFile(filepath.Join(bin, name), []byte(body), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			limit := 5
			if mode == "exhausted" {
				limit = 1
			}
			config := filepath.Join(root, "config.yaml")
			registry := filepath.Join(root, "prompts.yaml")
			for path, value := range map[string]string{
				source:   "requirement",
				config:   fmt.Sprintf("version: 3\ndelivery:\n  audit:\n    harness: claude\n    model: fixture\n    timeout: 1s\n    max_rounds: %d\n    fallback:\n      harness: hermes\n      provider: nous\n      model: fallback-fixture\n", limit),
				registry: "version: 1\nsession_recovery_instructions: Use full evidence and preserve history.\ngates:\n  implementation:\n    prompt: audit\n",
			} {
				if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			record := filepath.Join(root, "audits.yaml")
			args := []string{"start", "--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "implementation", "--audit-record", record, "--work-item", "W015-fallback", "--input", source}
			roundBase := 0
			if mode == "reset-limit" {
				roundBase = 5
				prior := harness.AuditEntry{WorkItem: "W015-fallback", Gate: "implementation", SessionID: "old-claude", Harness: "claude", Model: "fixture", ExternalRound: 5, Status: "exhausted", History: []harness.AuditRound{{Round: 5, Incident: "resume-failed"}}}
				if err := harness.WriteAuditEntry(record, prior); err != nil {
					t.Fatal(err)
				}
				reset := []string{"resume", "--reset-session", "--project", root, "--gate", "implementation", "--audit-record", record, "--work-item", "W015-fallback"}
				if err := run(reset, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
					t.Fatal(err)
				}
				args[0] = "resume"
			}
			var output, diagnostics bytes.Buffer
			err := run(args, strings.NewReader("audit"), &output, &diagnostics)
			blocked := mode == "exhausted" || mode == "fallback-fails" || mode == "evidence-changed"
			if (err != nil) != blocked {
				t.Fatalf("error = %v\n%s", err, diagnostics.String())
			}
			valid := mode == "PASS" || mode == "FAIL" || mode == "PROVISIONAL PASS"
			wantCalls := "primary\n"
			if !valid && mode != "exhausted" && mode != "evidence-changed" {
				wantCalls += "fallback\n"
			}
			calls, readErr := os.ReadFile(filepath.Join(root, "calls"))
			if readErr != nil || string(calls) != wantCalls {
				t.Fatalf("calls = %q (%v), want %q", calls, readErr, wantCalls)
			}
			entry, found, readErr := harness.ReadAuditEntry(record, "W015-fallback", "implementation")
			if readErr != nil || !found || entry.ExternalRound != roundBase+strings.Count(wantCalls, "\n") || entry.RoundsUsed() != strings.Count(wantCalls, "\n") {
				t.Fatalf("record = %#v (%v)", entry, readErr)
			}
			if valid && entry.Verdict != mode {
				t.Fatalf("verdict = %q", entry.Verdict)
			}
			if !valid && entry.History[0].Incident == "" {
				t.Fatal("lost primary incident")
			}
			if strings.Contains(wantCalls, "fallback") && (entry.Harness != "hermes" || entry.Provider != "nous" || entry.SessionID != "fallback-session" || len(entry.Sessions) != 1) {
				t.Fatalf("fallback provenance = %#v", entry)
			}
			if !valid && !blocked && entry.Verdict != "PASS" {
				t.Fatalf("fallback verdict = %q", entry.Verdict)
			}
			if blocked && output.Len() != 0 {
				t.Fatalf("incident emitted a verdict: %s", output.String())
			}
			// Changed evidence avoids a cache hit; the next request resumes Hermes,
			// without retrying the primary or consuming a fresh-context slot.
			if mode == "limit" {
				if err := os.WriteFile(source, []byte("updated requirement"), 0o600); err != nil {
					t.Fatal(err)
				}
				args[0] = "resume"
				if err := run(args, strings.NewReader("audit"), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
					t.Fatal(err)
				}
				calls, _ = os.ReadFile(filepath.Join(root, "calls"))
				if string(calls) != "primary\nfallback\nfallback\n" {
					t.Fatalf("resume calls = %q", calls)
				}
			}
		})
	}
}
