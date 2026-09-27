// ABOUTME: Verifies resets at provider-return and failure-reload boundaries.
// ABOUTME: Uses existing response/diagnostic hooks and a local executable provider double.
package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tigger-developer/sdlc/internal/harness"
)

type resetDiagnosticWriter struct {
	once  sync.Once
	reset func() error
	err   error
}

func (w *resetDiagnosticWriter) Write(data []byte) (int, error) {
	if strings.HasPrefix(string(data), "Unusable provider response:") {
		w.once.Do(func() { w.err = w.reset() })
	}
	return len(data), w.err
}

func TestRecoveryStopsAfterConcurrentReset(t *testing.T) {
	for _, boundary := range []string{"provider-return", "failure-reload"} {
		t.Run(boundary, func(t *testing.T) {
			root := t.TempDir()
			record := filepath.Join(root, "audits.yaml")
			calls := filepath.Join(root, "calls")
			t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("RESET_TEST_CALLS", calls)
			script := `#!/bin/sh
cat >/dev/null
printf 'call\n' >> "$RESET_TEST_CALLS"
printf 'session_id: native-fixture\n' >&2
printf 'unusable response\n'
exit 2
`
			// #nosec G306 -- local executable provider double; no hosted service.
			if err := os.WriteFile(filepath.Join(root, "hermes"), []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			var resetBytes []byte
			resetCount := 0
			reset := func() error {
				resetCount++
				if resetCount > 1 {
					return errors.New("obsolete invocation launched another provider request")
				}
				args := []string{"--reset", "--project", root, "--gate", "definition", "--work-item", "W020-recovery", "--audit-record", record}
				if err := run(args, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
					return err
				}
				var err error
				resetBytes, err = os.ReadFile(record)
				return err
			}
			config := harness.Config{Harness: "hermes", Provider: "fixture", Model: "primary", Timeout: 5 * time.Second, MaxRounds: 5, MaxFailures: 3, CoolOffPeriod: time.Hour, Fallback: &harness.AgentConfig{Harness: "hermes", Provider: "fixture", Model: "fallback"}}
			diagnostics := &resetDiagnosticWriter{reset: func() error { return nil }}
			request := harness.Request{Directory: root, Prompt: "audit"}
			if boundary == "provider-return" {
				request.OnResponse = func([]byte) error { return reset() }
			} else {
				diagnostics.reset = reset
			}
			runner := recoveryRun{config: config, request: request, path: record, work: "W020-recovery", gate: "definition", audit: true, diagnostics: diagnostics, cooldowns: harness.CooldownStore{Directory: filepath.Join(root, ".sdlc")}, registry: auditPromptDocument{SessionRecoveryInstructions: "Preserve history."}}
			result, err := runner.execute(harness.AuditEntry{}, false)
			if !errors.Is(err, harness.ErrAuditReset) || result.Response != "" {
				t.Fatalf("reset must stop recovery without verdict: result=%#v err=%v", result, err)
			}
			if resetCount != 1 || diagnostics.err != nil {
				t.Fatalf("reset count=%d err=%v", resetCount, diagnostics.err)
			}
			got, err := os.ReadFile(calls)
			if err != nil || string(got) != "call\n" {
				t.Fatalf("obsolete run retried or fell back: %q %v", got, err)
			}
			after, err := os.ReadFile(record)
			if err != nil || !bytes.Equal(after, resetBytes) {
				t.Fatal("recovery adopted or changed reset state")
			}
			if boundary == "provider-return" {
				deadline, err := runner.cooldowns.Deadline(config.Agent())
				if err != nil || !deadline.IsZero() {
					t.Fatalf("obsolete-generation error entered provider incident handling: %v %v", deadline, err)
				}
			}
		})
	}
}
