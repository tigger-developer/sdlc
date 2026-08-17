package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunParsesUserFacingFlags(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "sdlc")
	agentHome := filepath.Join(root, ".codex")
	if err := os.MkdirAll(source, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(source, "templates"), 0o700); err != nil {
		t.Fatalf("MkdirAll(templates) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(source, "commands"), 0o700); err != nil {
		t.Fatalf("MkdirAll(commands) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(source, "skills", "audit-code"), 0o700); err != nil {
		t.Fatalf("MkdirAll(skills) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "MAIN.md"), []byte("# SDLC\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(MAIN.md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("# Quickstart\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(README.md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "templates", "codex-sdlc.rules.example"), []byte("prefix_rule()\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(codex rules template) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "commands", "build.md"), []byte("# Build\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(build command) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "skills", "audit-code", "SKILL.md"), []byte("# Audit\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(audit skill) error = %v", err)
	}
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
