// ABOUTME: Exercises atomic audit ownership and recovery in every audit mode.
// ABOUTME: Uses local process doubles and verifies caller privacy.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestReviewSameGateCannotBeginTwiceFromStaleSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audits.yaml")
	config := harness.Config{Harness: "claude", Model: "fixture", MaxRounds: 5, MaxFailures: 3}
	stale := harness.AuditEntry{}
	if _, err := beginAuditAttempt(path, stale, "W-review", "definition", config, harness.Evidence{}, "first"); err != nil {
		t.Fatal(err)
	}
	if _, err := beginAuditAttempt(path, stale, "W-review", "definition", config, harness.Evidence{}, "second"); err == nil {
		t.Fatal("two launches accepted from the same stale snapshot for one work item and gate")
	}
}

func TestReviewStandaloneAndSpikeUseFallbackAndPrivateDiagnostics(t *testing.T) {
	for _, spike := range []bool{false, true} {
		t.Run(map[bool]string{false: "standalone", true: "spike"}[spike], func(t *testing.T) {
			root := t.TempDir()
			oldProfile := profilePresent
			profilePresent = func(string) (bool, error) { return spike, nil }
			defer func() { profilePresent = oldProfile }()
			t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
			primary := "#!/bin/sh\ncat >/dev/null\nprintf 'session_id: native\\nsdlc_failure: {\"version\":1,\"reason\":\"quota_exhausted\",\"status\":403,\"retryable\":false}\\nprivate-provider-body\\n' >&2\nexit 1\n"
			fallback := "#!/bin/sh\ncat >/dev/null\nprintf '%s\\n' '{\"type\":\"result\",\"result\":\"AUDIT: standalone\\nREVISION: fixture\\nVERDICT: PASS\\n\"}'\n"
			for name, body := range map[string]string{"hermes": primary, "claude": fallback} {
				if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0700); err != nil {
					t.Fatal(err)
				}
			}
			global := filepath.Join(root, "config.yaml")
			code := filepath.Join(root, "code.go")
			if err := os.WriteFile(global, []byte("delivery:\n  audit:\n    harness: hermes\n    provider: fixture\n    model: fixture\n    fallback:\n      harness: claude\n      model: fixture\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(code, []byte("package fixture\n"), 0600); err != nil {
				t.Fatal(err)
			}
			registry, err := filepath.Abs(filepath.Join("..", "..", "src", "prompts", "audits.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"--project", root, "--input", code, "--global-config", global, "--audit-prompts", registry}
			if spike {
				args = append(args, "--agent-spike")
			}
			var out, diagnostics bytes.Buffer
			err = runFixture(args, strings.NewReader("Review"), &out, &diagnostics)
			if err != nil {
				t.Errorf("configured fallback not used: %v", err)
			}
			if strings.Contains(diagnostics.String(), "private-provider-body") {
				t.Error("raw provider body reached caller stderr")
			}
			if !strings.Contains(out.String(), "VERDICT: PASS") {
				t.Error("fallback verdict absent")
			}
		})
	}
}
