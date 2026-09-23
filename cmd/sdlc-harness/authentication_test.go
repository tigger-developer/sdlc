// ABOUTME: Verifies immediate authentication fallback and terminal dual failure.
// ABOUTME: Exercises only local command doubles, with no provider calls.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthenticationFallbackIsImmediateAndBounded(t *testing.T) {
	for _, mode := range []string{"success", "authentication", "malformed", "success-limit-one", "malformed-limit-one", "no-fallback"} {
		t.Run(mode, func(t *testing.T) {
			outcome := strings.TrimSuffix(mode, "-limit-one")
			root := t.TempDir()
			t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("AUTH_ROOT", root)
			primary := "#!/bin/sh\ncat >/dev/null\nprintf 'primary\\n' >> \"$AUTH_ROOT/calls\"\nprintf 'Not logged in\\n' >&2\nexit 1\n"
			fallback := "#!/bin/sh\ncat >/dev/null\nprintf 'fallback\\n' >> \"$AUTH_ROOT/calls\"\nprintf 'session_id: fixture\\n' >&2\n"
			switch outcome {
			case "success":
				fallback += "printf 'GATE: definition\\nREVISION: fixture\\nVERDICT: PASS\\n'\n"
			case "authentication":
				fallback += "printf 'Invalid API key\\n' >&2\nexit 1\n"
			case "malformed":
				fallback += "printf 'not an envelope\\n'\n"
			}
			for name, body := range map[string]string{"claude": primary, "hermes": fallback} {
				// #nosec G306 -- local executable test double.
				if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			config := filepath.Join(root, "config.yaml")
			registry := filepath.Join(root, "prompts.yaml")
			configuration := "delivery:\n  audit:\n    harness: claude\n    model: fixture\n"
			if strings.HasSuffix(mode, "-limit-one") {
				configuration += "    max_failures: 1\n"
			}
			if outcome != "no-fallback" {
				configuration += "    fallback:\n      harness: hermes\n      provider: fixture\n      model: fixture\n"
			}
			for path, body := range map[string]string{config: configuration, registry: "version: 1\nsession_recovery_instructions: Preserve findings.\ngates:\n  definition:\n    prompt: audit\n"} {
				if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"--gate", "definition", "--project", root, "--global-config", config, "--audit-prompts", registry, "--audit-record", filepath.Join(root, "audits.yaml"), "--work-item", "W019", "--input", registry}
			err := run(args, strings.NewReader("audit"), &bytes.Buffer{}, &bytes.Buffer{})
			if (err != nil) != (outcome != "success") {
				t.Fatalf("result=%v", err)
			}
			calls, err := os.ReadFile(filepath.Join(root, "calls"))
			want := "primary\nfallback\n"
			if outcome == "no-fallback" {
				want = "primary\n"
			}
			if err != nil || string(calls) != want {
				t.Fatalf("calls=%q %v", calls, err)
			}
			_ = run(args, strings.NewReader("audit"), &bytes.Buffer{}, &bytes.Buffer{})
			after, err := os.ReadFile(filepath.Join(root, "calls"))
			if err != nil || !bytes.Equal(calls, after) {
				t.Fatal("lockout/cache called provider")
			}
		})
	}
}
