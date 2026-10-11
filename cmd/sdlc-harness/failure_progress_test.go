// ABOUTME: Verifies callers receive safe failure reasons and recovery transitions.
// ABOUTME: Exercises local provider doubles without publishing native bodies or sessions.
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

func TestAuditFailureProgress(t *testing.T) {
	for _, outcome := range []string{"success", "quota"} {
		t.Run(outcome, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
			primary := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"result","is_error":true,"result":"Prompt is too long"}'
printf 'secret-provider-body private-session\n' >&2
exit 1
`
			fallback := "#!/bin/sh\ncat >/dev/null\nprintf 'session_id: private-native-session\\n' >&2\n"
			if outcome == "quota" {
				fallback += "printf 'sdlc_failure: {\"version\":1,\"reason\":\"quota_exhausted\",\"status\":403,\"retryable\":false}\\n' >&2\nexit 1\n"
			} else {
				fallback += "printf 'GATE: definition\\nREVISION: fixture\\nVERDICT: PASS\\n'\n"
			}
			for name, body := range map[string]string{"claude": primary, "hermes": fallback} {
				// #nosec G306 -- private executable provider doubles never contact a model.
				if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0700); err != nil {
					t.Fatal(err)
				}
			}
			config, registry, record := filepath.Join(root, "config.yaml"), filepath.Join(root, "prompts.yaml"), filepath.Join(root, "audits.yaml")
			for path, body := range map[string]string{config: "delivery:\n  audit:\n    harness: claude\n    model: fixture\n    fallback:\n      harness: hermes\n      provider: fixture\n      model: fallback\n", registry: "version: 1\nsession_recovery_instructions: Preserve findings.\ngates:\n  definition:\n    prompt: audit\n"} {
				if err := os.WriteFile(path, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"--gate", "definition", "--project", root, "--global-config", config, "--audit-prompts", registry, "--audit-record", record, "--work-item", "W-progress", "--input", registry}
			var output, diagnostics bytes.Buffer
			err := runFixture(args, strings.NewReader("review"), &output, &diagnostics)
			if (err != nil) != (outcome == "quota") {
				t.Fatalf("err=%v", err)
			}
			for _, expected := range []string{"context limit", "Trying the configured fallback"} {
				if !strings.Contains(diagnostics.String(), expected) {
					t.Fatalf("missing %q in %q", expected, diagnostics.String())
				}
			}
			if outcome == "quota" {
				for _, expected := range []string{"usage quota", "Recovery stopped", "do not troubleshoot"} {
					if !strings.Contains(diagnostics.String(), expected) {
						t.Fatalf("missing %q in %q", expected, diagnostics.String())
					}
				}
			}
			if strings.Contains(diagnostics.String(), "secret-provider-body") || strings.Contains(diagnostics.String(), "private-native-session") {
				t.Fatal("private details leaked to caller")
			}
			deadline, readErr := (harness.CooldownStore{Directory: filepath.Join(root, ".sdlc")}).Deadline(harness.AgentConfig{Harness: "claude", Model: "fixture"})
			if readErr != nil || deadline.After(time.Now()) {
				t.Fatalf("bad request cooled healthy route: %v %v", deadline, readErr)
			}
		})
	}
}
