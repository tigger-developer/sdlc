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
				// #nosec G306 -- private executable provider doubles never contact a model.
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
			if !strings.Contains(diagnostics.String(), "usage quota") || !strings.Contains(diagnostics.String(), "Trying the configured fallback") || strings.Contains(diagnostics.String(), "session ") {
				t.Fatalf("missing safe recovery status or leaked session: %q", diagnostics.String())
			}
			if _, err := os.Stat(filepath.Join(root, ".sdlc")); !os.IsNotExist(err) {
				t.Fatalf("standalone diagnostics elected SDLC: %v", err)
			}
		})
	}
}

func TestStandaloneResponseCorrectionHasAFailureBound(t *testing.T) {
	for _, outcome := range []string{"corrected", "broken"} {
		t.Run(outcome, func(t *testing.T) {
			root := t.TempDir()
			oldProfile := profilePresent
			profilePresent = func(string) (bool, error) { return false, nil }
			t.Cleanup(func() { profilePresent = oldProfile })
			t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("STANDALONE_CALLS", filepath.Join(root, "calls"))
			t.Setenv("STANDALONE_OUTCOME", outcome)
			script := `#!/bin/sh
cat >/dev/null
printf 'call\n' >> "$STANDALONE_CALLS"
resume=no
for arg in "$@"; do
 if [ "$arg" = --resume ]; then resume=yes; fi
done
if [ "$resume" = yes ] && [ "$STANDALONE_OUTCOME" = corrected ]; then
 printf '%s\n' '{"type":"result","result":"AUDIT: standalone\nREVISION: fixture\nVERDICT: PASS\n"}'
else
 printf '%s\n' '{"type":"result","result":"invalid audit response"}'
fi
`
			// #nosec G306 -- private executable provider double; no hosted calls.
			if err := os.WriteFile(filepath.Join(root, "claude"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			config, code := filepath.Join(root, "config.yaml"), filepath.Join(root, "code.go")
			if err := os.WriteFile(config, []byte("delivery:\n  audit:\n    harness: claude\n    model: fixture\n    max_failures: 2\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(code, []byte("package fixture\n"), 0600); err != nil {
				t.Fatal(err)
			}
			registry, err := filepath.Abs(filepath.Join("..", "..", "src", "prompts", "audits.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			var output, diagnostics bytes.Buffer
			err = runFixture([]string{"--project", root, "--input", code, "--global-config", config, "--audit-prompts", registry}, strings.NewReader("Review"), &output, &diagnostics)
			if (err == nil) != (outcome == "corrected") {
				t.Fatalf("outcome=%s error=%v", outcome, err)
			}
			calls, readErr := os.ReadFile(filepath.Join(root, "calls"))
			if readErr != nil || strings.Count(string(calls), "call\n") != 2 {
				t.Fatalf("unbounded or missing correction: %q %v", calls, readErr)
			}
			if outcome == "broken" && (output.Len() != 0 || !strings.Contains(diagnostics.String(), "Recovery stopped") || strings.Contains(diagnostics.String(), "operator-authorized --reset")) {
				t.Fatalf("incorrect standalone terminal state: %q %q", output.String(), diagnostics.String())
			}
		})
	}
}

func TestStandaloneToolIncidentKeepsNativeSessionPrivate(t *testing.T) {
	root := t.TempDir()
	oldProfile := profilePresent
	profilePresent = func(string) (bool, error) { return false, nil }
	t.Cleanup(func() { profilePresent = oldProfile })
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	script := "#!/bin/sh\ncat >/dev/null\nprintf '%s\\n' '{\"type\":\"thread.started\",\"thread_id\":\"private-native-session\"}' '{\"type\":\"item.completed\",\"item\":{\"type\":\"command_execution\"}}'\n"
	// #nosec G306 -- local executable double; tools never execute.
	if err := os.WriteFile(filepath.Join(root, "codex"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	config, code := filepath.Join(root, "config.yaml"), filepath.Join(root, "code.go")
	if err := os.WriteFile(config, []byte("delivery:\n  audit:\n    harness: codex\n    model: fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(code, []byte("package fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	registry, err := filepath.Abs(filepath.Join("..", "..", "src", "prompts", "audits.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var output, diagnostics bytes.Buffer
	err = runFixture([]string{"--project", root, "--input", code, "--global-config", config, "--audit-prompts", registry}, strings.NewReader("Review"), &output, &diagnostics)
	if err == nil || output.Len() != 0 {
		t.Fatal("tool incident accepted")
	}
	if strings.Contains(err.Error()+diagnostics.String(), "private-native-session") || !strings.Contains(diagnostics.String(), "prohibited native tool") {
		t.Fatalf("unsafe or missing failure reason: %v %q", err, diagnostics.String())
	}
}
