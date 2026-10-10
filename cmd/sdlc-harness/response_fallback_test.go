// ABOUTME: Exercises a silent primary and responsive fallback through the same Hermes interface.
// ABOUTME: Verifies response deadlines recover without premature operator-intervention messages.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHermesResponseDeadlineSelectsHermesFallback(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
model=
while [ "$#" -gt 0 ]; do
    if [ "$1" = -m ]; then shift; model="$1"; break; fi
    shift
done
cat >/dev/null
printf '%s\n' "$model" >> "$RESPONSE_FALLBACK_CALLS"
printf 'session_id: native-%s\n' "$model" >&2
if [ "$model" = primary ]; then sleep 20; exit 1; fi
printf 'sdlc_response: content\n' >&2
printf 'GATE: delivery-code\nREVISION: fixture\nVERDICT: PASS\n'
`
	// #nosec G306 -- private executable provider double never contacts a model.
	if err := os.WriteFile(filepath.Join(bin, "hermes"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	calls := filepath.Join(root, "calls")
	t.Setenv("RESPONSE_FALLBACK_CALLS", calls)
	config := filepath.Join(root, "global.yaml")
	if err := os.WriteFile(config, []byte("delivery:\n  audit:\n    harness: hermes\n    provider: fixture\n    model: primary\n    timeout: 10s\n    total_timeout: 5s\n    response_start_timeout: 1s\n    response_idle_timeout: 2s\n    fallback:\n      harness: hermes\n      provider: other-fixture\n      model: fallback\n"), 0600); err != nil {
		t.Fatal(err)
	}
	registry := filepath.Join(root, "prompts.yaml")
	if err := os.WriteFile(registry, []byte("version: 1\nsession_recovery_instructions: Recover using the supplied evidence.\ngates:\n  delivery-code:\n    prompt: audit\n"), 0600); err != nil {
		t.Fatal(err)
	}
	writeReadyDocuments(t, root)
	args := []string{"--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "delivery-code", "--audit-record", filepath.Join(root, "audits.yaml"), "--work-item", "response-fallback-fixture", "--input", filepath.Join(root, "spec.org")}
	var output, diagnostics bytes.Buffer
	started := time.Now()
	err := runFixture(args, strings.NewReader("review"), &output, &diagnostics)
	if err != nil || !strings.Contains(output.String(), "VERDICT: PASS") || time.Since(started) > 4*time.Second {
		t.Fatalf("fallback error=%v elapsed=%s output=%q diagnostics=%q", err, time.Since(started), output.String(), diagnostics.String())
	}
	body, err := os.ReadFile(calls)
	if err != nil || string(body) != "primary\nfallback\n" {
		t.Fatalf("unexpected provider launches=%q err=%v", body, err)
	}
	if strings.Contains(diagnostics.String(), "Human intervention required") {
		t.Fatal("successful recovery requested human intervention")
	}
}
