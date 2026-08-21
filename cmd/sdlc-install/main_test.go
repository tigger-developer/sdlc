package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunParsesUserFacingFlags(t *testing.T) {
	root, source := newCLIFixture(t)
	agentHome := filepath.Join(root, ".codex")
	var output bytes.Buffer

	err := run([]string{
		"--agent", "codex",
		"--agent-home", agentHome,
		"--source", source,
	}, strings.NewReader(""), &output)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(output.String(), "DRY RUN") {
		t.Fatalf("output = %q, want dry-run plan", output.String())
	}
}

func TestDefaultRunDetectsInstalledProviderSubset_RT4_4(t *testing.T) {
	root, source := newCLIFixture(t)
	for _, name := range []string{".codex", ".hermes"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", root)
	var output bytes.Buffer
	if err := run([]string{"--source", source}, strings.NewReader("no\n"), &output); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if got := output.String(); !strings.Contains(got, "Detected agents: codex, hermes") || strings.Contains(got, ".claude") || strings.Contains(got, ".copilot") {
		t.Fatalf("provider detection output = %q", got)
	}
}

func TestRunHelpIsSuccessful(t *testing.T) {
	var output bytes.Buffer

	err := run([]string{"--help"}, strings.NewReader(""), &output)
	if err != nil {
		t.Fatalf("run(--help) error = %v", err)
	}
	for _, flagName := range []string{"-agent", "-agent-home", "-source", "-apply", "-configure"} {
		if !strings.Contains(output.String(), flagName) {
			t.Errorf("help output missing %q: %q", flagName, output.String())
		}
	}
}

func newCLIFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "sdlc")
	for path, content := range map[string]string{
		"MAIN.md":                            "# SDLC\n",
		"README.md":                          "# Quickstart\n",
		"templates/codex-sdlc.rules.example": "prefix_rule()\n",
		"commands/build.md":                  "# Build\n",
		"skills/audit-code/SKILL.md":         "# Audit\n",
	} {
		fullPath := filepath.Join(source, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root, source
}
