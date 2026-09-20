package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tigger-developer/sdlc/internal/harness"
)

// RT014.4: exercise real process, output validation and audit persistence using
// local doubles only. No metered provider is part of this regression test.
func TestFallbackForUnusableAuditResult(t *testing.T) {
	for _, mode := range []string{"limit", "reset-limit", "launch-error", "codex-failure", "custom-period", "timeout", "malformed", "wrong-gate", "empty", "PASS", "FAIL", "PROVISIONAL PASS", "fallback-fails", "exhausted", "evidence-changed"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("SDLC_HARNESS_STATE_DIR", filepath.Join(root, "state"))
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
launch-error|codex-failure|custom-period) exit 2;;
timeout) sleep 10; exit 1;;
malformed) printf '{"type":"result","result":"not a verdict"}\n'; exit 0;;
wrong-gate) printf '%s\n' '{"type":"result","result":"GATE: definition\nREVISION: fixture\nVERDICT: PASS"}'; exit 0;;
empty) printf '{"type":"result","result":""}\n'; exit 0;;
evidence-changed) printf 'changed' > "$FALLBACK_SOURCE";;
esac
verdict="$FALLBACK_MODE"
if [ "$verdict" = evidence-changed ]; then verdict=PASS; fi
printf '{"type":"result","result":"GATE: delivery-code\\nREVISION: fixture\\nVERDICT: %s"}\n' "$verdict"
`
			fallback := `#!/bin/sh
cat >/dev/null
printf 'fallback\n' >> "$FALLBACK_CALLS"
printf 'session_id: fallback-session\n' >&2
if [ "$FALLBACK_MODE" = fallback-fails ]; then exit 3; fi
printf 'GATE: delivery-code\nREVISION: fixture\nVERDICT: PASS\n'
`
			primaryName := "claude"
			if mode == "codex-failure" {
				primaryName = "codex"
			}
			for name, body := range map[string]string{primaryName: primary, "hermes": fallback} {
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
			period := time.Hour
			periodConfig := ""
			if mode == "custom-period" {
				period = 2 * time.Hour
				periodConfig = "      cool_off_period: 2h\n"
			}
			registry := filepath.Join(root, "prompts.yaml")
			for path, value := range map[string]string{
				source:   "requirement",
				config:   fmt.Sprintf("version: 3\ndelivery:\n  audit:\n    harness: %s\n    model: fixture\n    timeout: 1s\n    max_rounds: %d\n    fallback:\n      harness: hermes\n      provider: nous\n      model: fallback-fixture\n%s", primaryName, limit, periodConfig),
				registry: "version: 1\nsession_recovery_instructions: Use full evidence and preserve history.\ngates:\n  delivery-code:\n    prompt: audit\n",
			} {
				if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			record := filepath.Join(root, "audits.yaml")
			writeReadyDocuments(t, root)
			args := []string{"start", "--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "delivery-code", "--audit-record", record, "--work-item", "W015-fallback", "--input", source}
			roundBase := 0
			if mode == "reset-limit" {
				roundBase = 5
				prior := harness.AuditEntry{WorkItem: "W015-fallback", Gate: "implementation", SessionID: "old-claude", Harness: "claude", Model: "fixture", ExternalRound: 5, Status: "exhausted", History: []harness.AuditRound{{Round: 5, Incident: "resume-failed"}}}
				if err := harness.WriteAuditEntry(record, prior); err != nil {
					t.Fatal(err)
				}
				reset := []string{"resume", "--reset-session", "--project", root, "--gate", "delivery-code", "--audit-record", record, "--work-item", "W015-fallback"}
				if err := run(reset, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
					t.Fatal(err)
				}
				args[0] = "resume"
			}
			var output, diagnostics bytes.Buffer
			started := time.Now()
			err := run(args, strings.NewReader("audit"), &output, &diagnostics)
			blocked := mode == "exhausted" || mode == "fallback-fails" || mode == "evidence-changed"
			if (err != nil) != blocked {
				t.Fatalf("error = %v\n%s", err, diagnostics.String())
			}
			valid := mode == "PASS" || mode == "FAIL" || mode == "PROVISIONAL PASS"
			deadline, stateErr := (harness.CooldownStore{Directory: filepath.Join(root, ".sdlc")}).Deadline(harness.AgentConfig{Harness: primaryName, Model: "fixture"})
			if stateErr != nil {
				t.Fatal(stateErr)
			}
			if !valid && mode != "evidence-changed" {
				if deadline.Before(started.Add(period)) || deadline.After(time.Now().Add(period)) {
					t.Fatalf("cooldown %s is not %s after the failed attempt", deadline, period)
				}
			} else if !deadline.IsZero() {
				t.Fatal("usable verdict or local failure created cooldown")
			}
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
				if err := os.WriteFile(source, []byte(readySpec+"updated requirement\n"), 0o600); err != nil {
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
				// A different project must have its own cooldown and audit budget.
				other := t.TempDir()
				writeReadyDocuments(t, other)
				otherArgs := []string{"start", "--project", other, "--global-config", config, "--audit-prompts", registry, "--gate", "delivery-code", "--audit-record", filepath.Join(other, "audits.yaml"), "--work-item", "W017-cooldown"}
				var nextDiagnostics bytes.Buffer
				if err := run(otherArgs, strings.NewReader("audit"), &bytes.Buffer{}, &nextDiagnostics); err != nil {
					t.Fatal(err)
				}
				calls, readErr = os.ReadFile(filepath.Join(root, "calls"))
				if readErr != nil || string(calls) != "primary\nfallback\nfallback\nprimary\nfallback\n" {
					t.Fatalf("cross-project calls=%q err=%v diagnostics=%s", calls, readErr, nextDiagnostics.String())
				}
				otherEntry, _, readErr := harness.ReadAuditEntry(filepath.Join(other, "audits.yaml"), "W017-cooldown", "implementation")
				if readErr != nil || otherEntry.ExternalRound != 2 {
					t.Fatalf("skip consumed a round: %#v %v", otherEntry, readErr)
				}
			}
		})
	}
}
