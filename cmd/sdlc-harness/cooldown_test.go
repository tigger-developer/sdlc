// ABOUTME: Verifies project-local cooldown routing and retained audit safeguards.
// ABOUTME: All providers in this regression package remain local command doubles.
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

func TestCooldownRouting(t *testing.T) {
	for _, mode := range []string{"active", "resume-primary", "no-state-retained", "expired", "no-fallback", "corrupt", "exhausted", "other-model", "cache"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
			for name, output := range map[string]string{
				"claude": `printf '%s\n' '{"type":"result","result":"GATE: definition\nREVISION: fixture\nVERDICT: PASS"}'`,
				"hermes": "printf 'session_id: fallback-session\\n' >&2\nprintf 'GATE: definition\\nREVISION: fixture\\nVERDICT: PASS\\n'",
			} {
				body := fmt.Sprintf("#!/bin/sh\ncat >/dev/null\nprintf '%%s\\n' %q >> %q\n%s\n", name, filepath.Join(root, "calls"), output)
				body += fmt.Sprintf("for arg; do if [ \"$arg\" = --resume ]; then printf 'resume\\n' >> %q; fi; done\n", filepath.Join(root, "resumed"))
				// #nosec G306 -- executable local provider double.
				if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			config := "delivery:\n  audit:\n    harness: claude\n    model: fixture\n    max_rounds: 5\n"
			if mode != "no-fallback" {
				config += "    fallback:\n      harness: hermes\n      provider: nous\n      model: fallback-fixture\n"
			}
			for name, body := range map[string]string{"config.yaml": config, "prompts.yaml": "version: 1\nsession_recovery_instructions: Preserve history.\ngates:\n  definition:\n    prompt: Review.\n"} {
				if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			agent := harness.AgentConfig{Harness: "claude", Model: "fixture"}
			store := harness.CooldownStore{Directory: filepath.Join(root, ".sdlc")}
			deadline := time.Now().Add(time.Hour)
			if mode == "expired" {
				deadline = time.Now().Add(-time.Hour)
			}
			cooled := agent
			if mode == "other-model" {
				cooled.Model = "other"
			}
			if mode != "no-state-retained" {
				if err := store.Record(cooled, deadline); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "corrupt" {
				paths, err := filepath.Glob(filepath.Join(store.Directory, "cooldowns", "*.json"))
				if err != nil || len(paths) != 1 {
					t.Fatalf("state files=%v err=%v", paths, err)
				}
				if err := os.WriteFile(paths[0], []byte("{"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			record := filepath.Join(root, "audits.yaml")
			args := []string{"start", "--project", root, "--global-config", filepath.Join(root, "config.yaml"), "--audit-prompts", filepath.Join(root, "prompts.yaml"), "--gate", "definition", "--audit-record", record, "--work-item", "W017-cooldown"}
			args = append(args, "--input", filepath.Join(root, "prompts.yaml"))
			if mode == "expired" || mode == "exhausted" || mode == "resume-primary" || mode == "no-state-retained" {
				entry := harness.AuditEntry{WorkItem: "W017-cooldown", Gate: "definition", SessionID: "fallback-session", Harness: "hermes", Provider: "nous", Model: "fallback-fixture", Configured: &agent, Status: "incident", ExternalRound: 2}
				if mode == "resume-primary" {
					entry.Harness, entry.Provider, entry.Model, entry.SessionID = "claude", "", "fixture", "primary-session"
				}
				if mode == "exhausted" {
					entry.ExternalRound = 5
					for i := 1; i <= 5; i++ {
						entry.History = append(entry.History, harness.AuditRound{Round: i, Verdict: "FAIL"})
					}
				}
				if err := harness.WriteAuditEntry(record, entry); err != nil {
					t.Fatal(err)
				}
				args[0] = "resume"
			}
			var diagnostics bytes.Buffer
			err := run(args, strings.NewReader("audit"), &bytes.Buffer{}, &diagnostics)
			blocked := mode == "no-fallback" || mode == "corrupt" || mode == "exhausted"
			if (err != nil) != blocked {
				t.Fatalf("error=%v diagnostics=%s", err, diagnostics.String())
			}
			calls, readErr := os.ReadFile(filepath.Join(root, "calls"))
			if blocked {
				if !os.IsNotExist(readErr) {
					t.Fatalf("blocked invocation launched provider: %s %v", calls, readErr)
				}
				return
			}
			want := "hermes\n"
			if mode == "expired" || mode == "other-model" || mode == "no-state-retained" {
				want = "claude\n"
			}
			if readErr != nil || string(calls) != want {
				t.Fatalf("calls=%q want=%q err=%v", calls, want, readErr)
			}
			if mode == "resume-primary" {
				if _, err := os.Stat(filepath.Join(root, "resumed")); !os.IsNotExist(err) {
					t.Fatalf("passed primary session to fallback: %v", err)
				}
			}
			entry, _, err := harness.ReadAuditEntry(record, "W017-cooldown", "definition")
			if err != nil {
				t.Fatal(err)
			}
			if mode == "expired" && (entry.ExternalRound != 3 || len(entry.Sessions) != 1 || entry.Sessions[0].Reason != "primary-cooldown-expired") {
				t.Fatalf("lost resume provenance: %#v", entry)
			}
			if mode == "resume-primary" && (entry.ExternalRound != 3 || len(entry.Sessions) != 1 || entry.Sessions[0].Reason != "primary-cooldown") {
				t.Fatalf("lost cooled primary provenance: %#v", entry)
			}
			if mode == "cache" {
				args[0] = "resume"
				if err := run(args, strings.NewReader("audit"), &bytes.Buffer{}, &diagnostics); err != nil {
					t.Fatal(err)
				}
				calls, err = os.ReadFile(filepath.Join(root, "calls"))
				if err != nil || string(calls) != want || !strings.Contains(diagnostics.String(), "CACHE HIT") {
					t.Fatalf("cache relaunched provider: %s %v", calls, err)
				}
			}
		})
	}
}
