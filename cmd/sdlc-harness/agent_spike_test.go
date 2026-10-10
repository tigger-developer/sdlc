package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentSpikeRunsStandaloneReviewInSDLCProject(t *testing.T) {
	project, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	probe := filepath.Join(bin, "prompt")
	script := `#!/bin/sh
output=
previous=
for argument in "$@"; do
  if [ "$previous" = "-o" ]; then output="$argument"; fi
  previous="$argument"
done
/bin/cat > "$SPIKE_PROBE"
printf 'AUDIT: standalone\nREVISION: candidate\nVERDICT: PASS\n' > "$output"
printf '{"type":"thread.started","thread_id":"spike-fixture"}\n'
`
	// #nosec G306 -- private executable fixture never invokes a hosted provider.
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SPIKE_PROBE", probe)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	// TestMain reports an SDLC profile for every project; the spike must still
	// select the standalone review without requiring gate or record options.
	args := []string{"--agent-spike", "--project", project, "--input", "go.mod", "--audit-prompts", filepath.Join(project, "src", "prompts", "audits.yaml"), "--global-config", filepath.Join(project, "absent-global.yaml"), "--harness", "codex", "--model", "fixture"}
	var output, diagnostics bytes.Buffer
	if err := runFixture(args, strings.NewReader("Workflow: hands-off spike"), &output, &diagnostics); err != nil {
		t.Fatalf("agent spike audit failed: %v\n%s", err, diagnostics.String())
	}
	if !strings.Contains(diagnostics.String(), "SPIKE MODE") || strings.Contains(diagnostics.String(), "STANDALONE MODE") {
		t.Fatalf("agent spike diagnostics = %q", diagnostics.String())
	}
	if !strings.Contains(output.String(), "AUDIT: standalone") {
		t.Fatalf("agent spike output = %q", output.String())
	}
	prompt, err := os.ReadFile(probe)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(prompt, []byte("SPIKE-AUDIT.md")) || !bytes.Contains(prompt, []byte("go.mod")) {
		t.Fatal("agent spike prompt omitted the spike standard or project evidence")
	}
	if bytes.Contains(prompt, []byte("Audit gate:")) {
		t.Fatal("agent spike prompt contains an SDLC gate")
	}
}

func TestAgentSpikeRejectsSDLCGateOptions(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "code.go"), []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, extra := range [][]string{
		{"--gate", "delivery-code"},
		{"--audit-record", "audits.yaml"},
		{"--work-item", "W001"},
	} {
		args := append([]string{"--agent-spike", "--project", project, "--input", "code.go"}, extra...)
		var output bytes.Buffer
		err := runFixture(args, strings.NewReader(""), &output, &output)
		if err == nil || !strings.Contains(err.Error(), "without --gate, --audit-record, --work-item or --reset") {
			t.Errorf("agent spike with %v error = %v", extra, err)
		}
	}
	if _, err := os.Stat(filepath.Join(project, "audits.yaml")); !os.IsNotExist(err) {
		t.Fatalf("refused agent spike created an audit record: %v", err)
	}
}

func TestAgentSpikeRequiresAuditPhase(t *testing.T) {
	var output bytes.Buffer
	err := runFixture([]string{"--agent-spike", "--phase", "build"}, strings.NewReader(""), &output, &output)
	if err == nil || !strings.Contains(err.Error(), "--agent-spike applies only to --phase audit") {
		t.Fatalf("agent spike outside audit phase error = %v", err)
	}
}
