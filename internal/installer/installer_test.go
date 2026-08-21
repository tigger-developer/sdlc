package installer

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type fixture struct {
	root, source string
}

func newFixture(t *testing.T, providers ...string) fixture {
	t.Helper()
	root := t.TempDir()
	f := fixture{root: root, source: filepath.Join(root, "staging", "sdlc")}
	writeFile(t, filepath.Join(f.source, "MAIN.md"), []byte("# Main\n"))
	writeFile(t, filepath.Join(f.source, "README.md"), []byte("# Readme\n"))
	writeFile(t, filepath.Join(f.source, "commands", "build.md"), []byte("# Build\n"))
	writeFile(t, filepath.Join(f.source, "commands", "nested", "review.md"), []byte("# Review\n"))
	writeFile(t, filepath.Join(f.source, "skills", "audit-code", "SKILL.md"), []byte("# Audit\n"))
	writeFile(t, filepath.Join(f.source, "skills", "draft-design-issue", "SKILL.md"), []byte("# Draft\n"))
	writeFile(t, filepath.Join(f.source, "hooks", "agent-command-guard.sh"), []byte("#!/bin/sh\n"))
	writeFile(t, filepath.Join(f.source, "templates", "codex-sdlc.rules.example"), []byte("prefix_rule()\n"))
	writeFile(t, filepath.Join(f.source, ".git", "config"), []byte("staging metadata\n"))
	for _, provider := range providers {
		mkdirAll(t, filepath.Join(root, "."+provider))
	}
	return f
}

func TestInteractiveCopiesEveryOwnedItemToDetectedHomes(t *testing.T) {
	f := newFixture(t, "agents", "claude", "codex", "copilot", "hermes")
	claudeConfig := filepath.Join(f.root, ".claude", "settings.json")
	codexConfig := filepath.Join(f.root, ".codex", "config.toml")
	hermesConfig := filepath.Join(f.root, ".hermes", "config.yaml")
	writeFile(t, claudeConfig, []byte("{\"personal\":true}\n"))
	writeFile(t, codexConfig, []byte("personal = true\n"))
	writeFile(t, hermesConfig, []byte("personal: true\n"))

	var output bytes.Buffer
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &output); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), "[yes/no]") != 1 {
		t.Fatalf("expected one batch confirmation:\n%s", output.String())
	}

	for _, home := range []string{".agents", ".claude", ".codex", ".copilot", ".hermes"} {
		assertRegularTree(t, filepath.Join(f.root, home, "sdlc"))
		assertFile(t, filepath.Join(f.root, home, "sdlc", "MAIN.md"), "# Main\n")
		assertAbsent(t, filepath.Join(f.root, home, "sdlc", ".git"))
	}
	for _, skill := range sourceDirectories(t, filepath.Join(f.source, "skills")) {
		for _, home := range []string{".agents", ".claude", ".copilot", ".hermes"} {
			assertSameDirectory(t, filepath.Join(f.source, "skills", skill), filepath.Join(f.root, home, "skills", skill))
		}
	}
	for _, destination := range []string{
		filepath.Join(f.root, ".claude", "commands"),
		filepath.Join(f.root, ".codex", "prompts-commands"),
		filepath.Join(f.root, ".copilot", "prompts-commands"),
	} {
		assertSameDirectory(t, filepath.Join(f.source, "commands"), destination)
	}
	assertAbsent(t, filepath.Join(f.root, ".codex", "skills", "audit-code"))
	assertAbsent(t, filepath.Join(f.root, ".hermes", "commands"))
	assertFile(t, claudeConfig, "{\"personal\":true}\n")
	assertFile(t, codexConfig, "personal = true\n")
	assertFile(t, hermesConfig, "personal: true\n")

	output.Reset()
	if err := RunInteractive(f.source, f.root, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "All detected SDLC copies are current") || strings.Contains(output.String(), "[yes/no]") {
		t.Fatalf("repeated install was not current:\n%s", output.String())
	}
}

func TestDeclineWritesNothingAndWrongTypesAreBackedUp(t *testing.T) {
	f := newFixture(t, "agents", "claude")
	var output bytes.Buffer
	if err := RunInteractive(f.source, f.root, strings.NewReader("no\n"), &output); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, filepath.Join(f.root, ".agents", "sdlc"))
	assertAbsent(t, filepath.Join(f.root, ".claude", "sdlc"))

	writeFile(t, filepath.Join(f.root, ".agents", "sdlc"), []byte("wrong type\n"))
	output.Reset()
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &output); err != nil {
		t.Fatal(err)
	}
	assertRegularTree(t, filepath.Join(f.root, ".agents", "sdlc"))
	backup := assertOneBackup(t, filepath.Join(f.root, ".agents", "sdlc"))
	assertFile(t, backup, "wrong type\n")
	assertRegularTree(t, filepath.Join(f.root, ".claude", "sdlc"))
}

func TestStaleLinksBecomeCopiesAndDestinationOnlyFilesSurvive(t *testing.T) {
	f := newFixture(t, "agents", "codex")
	stale := filepath.Join(f.root, "stale")
	mkdirAll(t, stale)
	symlink(t, stale, filepath.Join(f.root, ".agents", "skills", "audit-code"))
	symlink(t, filepath.Join(f.source, "commands"), filepath.Join(f.root, ".codex", "prompts-commands"))
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	assertSameDirectory(t, filepath.Join(f.source, "skills", "audit-code"), filepath.Join(f.root, ".agents", "skills", "audit-code"))
	assertSameDirectory(t, filepath.Join(f.source, "commands"), filepath.Join(f.root, ".codex", "prompts-commands"))
	assertOneBackup(t, filepath.Join(f.root, ".agents", "skills", "audit-code"))
	assertOneBackup(t, filepath.Join(f.root, ".codex", "prompts-commands"))

	extra := filepath.Join(f.root, ".agents", "skills", "audit-code", "runtime.txt")
	unowned := filepath.Join(f.root, ".agents", "skills", "personal", "SKILL.md")
	writeFile(t, extra, []byte("runtime\n"))
	writeFile(t, unowned, []byte("personal\n"))
	writeFile(t, filepath.Join(f.source, "skills", "audit-code", "SKILL.md"), []byte("# Changed\n"))
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	assertFile(t, filepath.Join(f.root, ".agents", "skills", "audit-code", "SKILL.md"), "# Changed\n")
	skillBackup := assertOneBackup(t, filepath.Join(f.root, ".agents", "skills", "audit-code", "SKILL.md"))
	assertFile(t, skillBackup, "# Audit\n")
	assertFile(t, extra, "runtime\n")
	assertFile(t, unowned, "personal\n")
}

func TestExplicitAgentLimitsProviderCopies(t *testing.T) {
	f := newFixture(t, "claude", "hermes")
	if err := Run(Options{Agent: "claude", AgentHome: filepath.Join(f.root, ".claude"), Source: f.source, Apply: true, Output: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	assertRegularTree(t, filepath.Join(f.root, ".claude", "sdlc"))
	assertSameDirectory(t, filepath.Join(f.source, "commands"), filepath.Join(f.root, ".claude", "commands"))
	assertAbsent(t, filepath.Join(f.root, ".hermes", "sdlc"))
}

func TestDryRunReportsDriftWithoutWriting(t *testing.T) {
	f := newFixture(t, "codex")
	var output bytes.Buffer
	if err := Run(Options{Agent: "codex", AgentHome: filepath.Join(f.root, ".codex"), Source: f.source, Output: &output}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "would synchronize") {
		t.Fatalf("dry run omitted drift:\n%s", output.String())
	}
	assertAbsent(t, filepath.Join(f.root, ".codex", "sdlc"))
}

func TestProviderConfigurationIsBackedUpBesideTheLiveFile(t *testing.T) {
	f := newFixture(t, "claude")
	settings := filepath.Join(f.root, ".claude", "settings.json")
	original := "{\"personal\":true,\"permissions\":{\"allow\":[\"Bash(rm:*)\"],\"deny\":[]}}\n"
	writeFile(t, settings, []byte(original))
	if err := Run(Options{
		Agent: "claude", AgentHome: filepath.Join(f.root, ".claude"), Source: f.source,
		Configure: true, Input: strings.NewReader("yes\n"), Output: &bytes.Buffer{},
	}); err != nil {
		t.Fatal(err)
	}
	backup := assertOneBackup(t, settings)
	assertFile(t, backup, original)
	updated := string(mustReadFile(t, settings))
	if !strings.Contains(updated, "\"personal\": true") || !strings.Contains(updated, "Bash(rm:*)") {
		t.Fatalf("updated settings lost owned or personal values:\n%s", updated)
	}
}

func sourceDirectories(t *testing.T, path string) []string {
	t.Helper()
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

func assertRegularTree(t *testing.T, path string) {
	t.Helper()
	err := filepath.Walk(path, func(current string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Errorf("deployed path is a symlink: %s", current)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertSameDirectory(t *testing.T, source, destination string) {
	t.Helper()
	assertRegularTree(t, destination)
	output, err := exec.Command("diff", "-qr", source, destination).CombinedOutput()
	if err != nil {
		t.Fatalf("%s differs from %s: %v\n%s", destination, source, err, output)
	}
}

func assertFile(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != expected {
		t.Fatalf("%s = %q, want %q", path, content, expected)
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("%s is not a regular file: %v, %#v", path, err, info)
	}
}

func assertAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("%s exists or cannot be inspected: %v", path, err)
	}
}

func assertOneBackup(t *testing.T, path string) string {
	t.Helper()
	matches, err := filepath.Glob(path + ".*.bak")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("backup matches for %s = %v, want one", path, matches)
	}
	return matches[0]
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	mkdirAll(t, filepath.Dir(path))
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func symlink(t *testing.T, target, path string) {
	t.Helper()
	mkdirAll(t, filepath.Dir(path))
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
}
