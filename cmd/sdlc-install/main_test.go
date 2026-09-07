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

// RT-7.2
func TestDefaultRunDetectsInstalledProviderSubset_RT4_4(t *testing.T) {
	root, source := newCLIFixture(t)
	for _, name := range []string{".codex", ".hermes"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".hermes", "config.yaml"), []byte("model:\n  provider: test\n"), 0o600); err != nil {
		t.Fatal(err)
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
	for _, flagName := range []string{"-agent", "-agent-home", "-source", "-apply", "-configure", "-version"} {
		if !strings.Contains(output.String(), flagName) {
			t.Errorf("help output missing %q: %q", flagName, output.String())
		}
	}
	for _, explanation := range []string{"repository-internal", "make install", "does not initialize a software project"} {
		if !strings.Contains(output.String(), explanation) {
			t.Errorf("help output missing %q: %q", explanation, output.String())
		}
	}
}

func TestRunVersionReportsEmbeddedRelease(t *testing.T) {
	previous := buildRelease
	buildRelease = "v3.0.0"
	t.Cleanup(func() { buildRelease = previous })
	var output bytes.Buffer

	if err := run([]string{"--version"}, strings.NewReader(""), &output); err != nil {
		t.Fatalf("run(--version) error = %v", err)
	}
	if output.String() != "sdlc-install v3.0.0\n" {
		t.Fatalf("version output = %q", output.String())
	}
}

func newCLIFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "sdlc")
	for path, content := range map[string]string{
		"src/MAIN.md":                        "# SDLC\n",
		"README.md":                          "# Quickstart\n",
		"templates/codex-sdlc.rules.example": "prefix_rule()\n",
		"commands/build.md":                  "# Build\n",
		"skills/audit-code/SKILL.md":         "# Audit\n",
		"hooks/agent-command-guard.sh":       "#!/bin/sh\n",
		"bin/sdlc-init":                      "initializer\n",
		"bin/sdlc-preview":                   "previewer\n",
		"bin/sdlc-merge-legacy-acs":          "ledger merger\n",
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
