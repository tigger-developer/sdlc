// ABOUTME: Checks permanent Hermes failures select one fallback without outer retries.
// ABOUTME: Uses local process doubles and verifies persisted lockout and verdict budgets.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestPermanentHermesFailureRecovery(t *testing.T) {
	for _, mode := range []string{"success", "quota", "no-fallback", "limit-one"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("PERMANENT_CALLS", filepath.Join(root, "calls"))
			t.Setenv("PERMANENT_MODE", mode)
			script := `#!/bin/sh
model=
while [ "$#" -gt 0 ]; do
 if [ "$1" = -m ]; then shift; model="$1"; break; fi
 shift
done
cat >/dev/null
printf '%s\n' "$model" >> "$PERMANENT_CALLS"
printf 'session_id: native-%s\n' "$model" >&2
if [ "$model" = primary ]; then
 printf 'sdlc_failure: {"version":1,"reason":"model_not_found","status":404,"retryable":false}\n' >&2
 exit 1
fi
if [ "$PERMANENT_MODE" = quota ]; then
 printf 'sdlc_failure: {"version":1,"reason":"quota_exhausted","status":403,"retryable":false}\n' >&2
 exit 1
fi
printf 'GATE: definition\nREVISION: fixture\nVERDICT: PASS\n'
`
			// #nosec G306 -- executable local fixture never calls a hosted provider.
			if err := os.WriteFile(filepath.Join(root, "hermes"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			configuration := "delivery:\n  audit:\n    harness: hermes\n    provider: fixture\n    model: primary\n"
			if mode == "limit-one" {
				configuration += "    max_failures: 1\n"
			}
			if mode != "no-fallback" {
				configuration += "    fallback:\n      harness: hermes\n      provider: fixture\n      model: fallback\n"
			}
			config, registry, record := filepath.Join(root, "config.yaml"), filepath.Join(root, "prompts.yaml"), filepath.Join(root, "audits.yaml")
			for path, body := range map[string]string{config: configuration, registry: "version: 1\nsession_recovery_instructions: Preserve findings.\ngates:\n  definition:\n    prompt: audit\n"} {
				if err := os.WriteFile(path, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"--gate", "definition", "--project", root, "--global-config", config, "--audit-prompts", registry, "--audit-record", record, "--work-item", "W-permanent", "--input", registry}
			var output, diagnostics bytes.Buffer
			err := runFixture(args, strings.NewReader("audit"), &output, &diagnostics)
			blocked := mode == "quota" || mode == "no-fallback"
			if (err != nil) != blocked {
				t.Fatalf("err=%v diagnostics=%q", err, diagnostics.String())
			}
			calls, readErr := os.ReadFile(filepath.Join(root, "calls"))
			want := "primary\nfallback\n"
			if mode == "no-fallback" {
				want = "primary\n"
			}
			if readErr != nil || string(calls) != want {
				t.Fatalf("calls=%q err=%v", calls, readErr)
			}
			entry, found, readErr := harness.ReadAuditEntry(record, "W-permanent", "definition")
			rounds := 1
			if blocked {
				rounds = 0
			}
			if readErr != nil || !found || entry.RoundsUsed() != rounds || entry.FailureBlocked() != blocked {
				t.Fatalf("entry=%+v err=%v", entry, readErr)
			}
			_ = runFixture(args, strings.NewReader("audit"), &bytes.Buffer{}, &bytes.Buffer{})
			after, readErr := os.ReadFile(filepath.Join(root, "calls"))
			if readErr != nil || !bytes.Equal(calls, after) {
				t.Fatal("cached/blocked audit launched provider")
			}
		})
	}
}
