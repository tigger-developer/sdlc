package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// RT014.5: a failed primary's response file cannot become the fallback verdict.
func TestFallbackCannotReuseFailedPrimaryOutput(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
model=
output=
previous=
for argument in "$@"; do
    if [ "$previous" = "-o" ]; then output="$argument"; fi
    if [ "$previous" = "-m" ]; then model="$argument"; fi
    previous="$argument"
done
cat >/dev/null
printf '{"type":"thread.started","thread_id":"%s-session"}\n' "$model"
if [ "$model" = primary ]; then
    printf 'GATE: implementation\nREVISION: stale\nVERDICT: PASS\n' > "$output"
    exit 1
fi
exit 0
`
	// #nosec G306 -- local executable double, never a metered invocation.
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	config := filepath.Join(root, "config.yaml")
	registry := filepath.Join(root, "prompts.yaml")
	source := filepath.Join(root, "spec.org")
	for path, value := range map[string]string{
		config:   "version: 3\ndelivery:\n  audit:\n    harness: codex\n    model: primary\n    timeout: 5s\n    max_rounds: 5\n    fallback:\n      harness: codex\n      model: fallback\n",
		registry: "version: 1\nsession_recovery_instructions: Preserve history.\ngates:\n  implementation:\n    prompt: audit\n",
		source:   "requirement",
	} {
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var output, diagnostics bytes.Buffer
	err := run([]string{"start", "--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "implementation", "--audit-record", filepath.Join(root, "audits.yaml"), "--work-item", "W014-fallback-output", "--input", source}, strings.NewReader("audit"), &output, &diagnostics)
	if err == nil || !strings.Contains(err.Error(), "response-empty") || output.Len() != 0 {
		t.Fatalf("stale result accepted: error=%v output=%q diagnostics=%s", err, output.String(), diagnostics.String())
	}
}
