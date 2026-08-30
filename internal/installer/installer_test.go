package installer

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fixture struct {
	root, source string
}

func newFixture(t *testing.T, providers ...string) fixture {
	t.Helper()
	root := t.TempDir()
	f := fixture{root: root, source: filepath.Join(root, "staging", "sdlc")}
	writeFile(t, filepath.Join(f.source, "src", "MAIN.md"), []byte("# Main\n"))
	writeFile(t, filepath.Join(f.source, "src", "technologies", "GO.md"), []byte("# Go\n"))
	writeFile(t, filepath.Join(f.source, "src", "guides", "nested.md"), []byte("# Nested\n"))
	writeFile(t, filepath.Join(f.source, "src", "templates", "project-init", "legacy-acs-header.md"), []byte("# Legacy\n***\n"))
	writeFile(t, filepath.Join(f.source, "src", "prompts", "project-init", "constitution.md.tmpl"), []byte("Create {{.TemplatePath}}\n"))
	writeFile(t, filepath.Join(f.source, "README.md"), []byte("# Readme\n"))
	writeFile(t, filepath.Join(f.source, "CHANGELOG.md"), []byte("# Changelog\n"))
	writeFile(t, filepath.Join(f.source, "LEARNINGS.md"), []byte("# Learnings\n"))
	writeFile(t, filepath.Join(f.source, "docs", "ACs.md"), []byte("# Acceptance criteria\n"))
	writeFile(t, filepath.Join(f.source, "cmd", "sdlc-install", "main.go"), []byte("package main\n"))
	writeFile(t, filepath.Join(f.source, "internal", "installer", "installer.go"), []byte("package installer\n"))
	writeFile(t, filepath.Join(f.source, "go.mod"), []byte("module example.test/sdlc\n"))
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

// RT-6.1, RT-6.3, RT-7.1, RT-8.1, RT-8.4
func TestInteractiveInstallsOneCanonicalTreeAndProviderAdapters(t *testing.T) {
	f := newFixture(t, "claude", "codex", "copilot", "hermes")
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

	assertRegularTree(t, filepath.Join(f.root, ".agents", "sdlc"))
	assertFile(t, filepath.Join(f.root, ".agents", "sdlc", "MAIN.md"), "# Main\n")
	assertFile(t, filepath.Join(f.root, ".agents", "sdlc", "guides", "nested.md"), "# Nested\n")
	assertFile(t, filepath.Join(f.root, ".agents", "sdlc", "templates", "project-init", "legacy-acs-header.md"), "# Legacy\n***\n")
	assertFile(t, filepath.Join(f.root, ".agents", "sdlc", "prompts", "project-init", "constitution.md.tmpl"), "Create {{.TemplatePath}}\n")
	assertAbsent(t, filepath.Join(f.root, ".agents", "sdlc", "templates", "codex-sdlc.rules.example"))
	assertAbsent(t, filepath.Join(f.root, ".agents", "sdlc", ".git"))
	for _, relative := range []string{"src", "README.md", "CHANGELOG.md", "LEARNINGS.md", "cmd", "docs", "internal", "go.mod"} {
		assertAbsent(t, filepath.Join(f.root, ".agents", "sdlc", relative))
	}
	for _, home := range []string{".claude", ".codex", ".copilot", ".hermes"} {
		assertAbsent(t, filepath.Join(f.root, home, "sdlc"))
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
	if claude := string(mustReadFile(t, claudeConfig)); !strings.Contains(claude, "\"personal\": true") || !strings.Contains(claude, "Bash(sed:*)") {
		t.Fatalf("interactive install did not preserve and configure Claude:\n%s", claude)
	}
	assertFile(t, codexConfig, "personal = true\n")
	if hermes := string(mustReadFile(t, hermesConfig)); !strings.Contains(hermes, "personal: true") || !strings.Contains(hermes, "agent-command-guard.sh") {
		t.Fatalf("interactive install did not preserve and configure Hermes:\n%s", hermes)
	}

	output.Reset()
	if err := RunInteractive(f.source, f.root, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "All detected SDLC copies are current") || strings.Contains(output.String(), "[yes/no]") {
		t.Fatalf("repeated install was not current:\n%s", output.String())
	}
}

// RT-5.1
func TestInteractiveIgnoresTimestampOnlySourceChanges(t *testing.T) {
	f := newFixture(t, "agents", "codex")
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	mainPath := filepath.Join(f.source, "src", "MAIN.md")
	newTime := time.Now().Add(time.Minute)
	if err := os.Chtimes(mainPath, newTime, newTime); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := RunInteractive(f.source, f.root, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "All detected SDLC copies are current") || strings.Contains(output.String(), "[yes/no]") {
		t.Fatalf("timestamp-only change created deployment variance:\n%s", output.String())
	}
}

// RT-5.2, RT-8.3
func TestInteractiveDefaultOutputListsOnlyVariances(t *testing.T) {
	f := newFixture(t, "agents", "codex")
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(f.source, "src", "MAIN.md"), []byte("# Changed main\n"))

	var output bytes.Buffer
	if err := RunInteractive(f.source, f.root, strings.NewReader("no\n"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "would back up and replace differing file") {
		t.Fatalf("default output omitted variance:\n%s", output.String())
	}
	if !strings.Contains(output.String(), filepath.Join(f.root, ".agents", "sdlc", "MAIN.md")) {
		t.Fatalf("default output omitted the exact differing file:\n%s", output.String())
	}
	if strings.Contains(output.String(), "Installation: current") {
		t.Fatalf("default output included matching destinations:\n%s", output.String())
	}
}

// RT-5.3, RT-8.3
func TestInteractiveVerboseOutputIncludesMatchingDestinations(t *testing.T) {
	t.Setenv("VERBOSE", "1")
	f := newFixture(t, "agents", "codex")
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(f.source, "src", "MAIN.md"), []byte("# Changed main\n"))

	var output bytes.Buffer
	if err := RunInteractive(f.source, f.root, strings.NewReader("no\n"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "would back up and replace differing file") || !strings.Contains(output.String(), "Installation: current") {
		t.Fatalf("verbose output did not include variances and matching destinations:\n%s", output.String())
	}
}

// RT-8.2
func TestInteractiveIgnoresRepositoryOnlyChanges(t *testing.T) {
	f := newFixture(t, "agents", "codex")
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(f.source, "README.md"), []byte("# Changed readme\n"))
	writeFile(t, filepath.Join(f.source, "internal", "installer", "installer.go"), []byte("package changed\n"))

	var output bytes.Buffer
	if err := RunInteractive(f.source, f.root, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "All detected SDLC copies are current") || strings.Contains(output.String(), "[yes/no]") {
		t.Fatalf("repository-only changes created a deployment variance:\n%s", output.String())
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
	assertAbsent(t, filepath.Join(f.root, ".claude", "sdlc"))
}

func TestStaleLinksBecomeCopiesAndDestinationOnlyFilesSurvive(t *testing.T) {
	f := newFixture(t, "agents", "codex")
	stale := filepath.Join(f.root, "stale")
	mkdirAll(t, stale)
	symlink(t, stale, filepath.Join(f.root, ".agents", "skills", "audit-code"))
	symlink(t, filepath.Join(f.source, "commands"), filepath.Join(f.root, ".codex", "prompts-commands"))
	var output bytes.Buffer
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &output); err != nil {
		t.Fatalf("%v\n%s", err, output.String())
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

func TestInteractiveRetiresKnownLegacyArtifactsAndPreservesUnownedFiles(t *testing.T) {
	f := newFixture(t, "claude", "codex", "copilot", "hermes")
	removePath(t, filepath.Join(f.source, "commands"))
	removePath(t, filepath.Join(f.source, "skills", "draft-design-issue"))
	writeFile(t, filepath.Join(f.root, ".claude", "settings.json"), []byte("{\"personal\":true}\n"))
	writeFile(t, filepath.Join(f.root, ".codex", "config.toml"), []byte("personal = true\n"))
	writeFile(t, filepath.Join(f.root, ".hermes", "config.yaml"), []byte("personal: true\n"))

	legacyPaths := seedLegacyRuntimeArtifacts(t, f.root)
	unownedPaths := []string{
		filepath.Join(f.root, ".agents", "sdlc", "commands", "personal.md"),
		filepath.Join(f.root, ".agents", "skills", "personal", "SKILL.md"),
		filepath.Join(f.root, ".claude", "commands", "personal.md"),
	}
	for _, path := range unownedPaths {
		writeFile(t, path, []byte("personal\n"))
	}

	var output bytes.Buffer
	if err := RunInteractive(f.source, f.root, strings.NewReader("no\n"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "would back up and retire legacy SDLC artefact") {
		t.Fatalf("retirement variances were not listed:\n%s", output.String())
	}
	for _, path := range legacyPaths {
		assertFile(t, path, "legacy\n")
		assertNoBackup(t, path)
	}

	output.Reset()
	if err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &output); err != nil {
		t.Fatalf("%v\n%s", err, output.String())
	}
	for _, path := range legacyPaths {
		assertAbsent(t, path)
		backup := assertOneBackup(t, path)
		assertFile(t, backup, "legacy\n")
	}
	for _, path := range unownedPaths {
		assertFile(t, path, "personal\n")
	}

	output.Reset()
	if err := RunInteractive(f.source, f.root, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "All detected SDLC copies are current") || strings.Contains(output.String(), "[yes/no]") {
		t.Fatalf("repeated retirement was not idempotent:\n%s", output.String())
	}
}

func TestInstallAllowsRetiredOptionalCommandsDirectoryToBeAbsent(t *testing.T) {
	f := newFixture(t)
	removePath(t, filepath.Join(f.source, "commands"))

	var output bytes.Buffer
	if err := Run(Options{
		Agent:     "custom",
		AgentHome: filepath.Join(f.root, ".custom"),
		Source:    f.source,
		Apply:     true,
		Output:    &output,
	}); err != nil {
		t.Fatalf("install with no commands directory failed: %v\n%s", err, output.String())
	}
	assertFile(t, filepath.Join(f.root, ".agents", "sdlc", "MAIN.md"), "# Main\n")
}

// RT-6.2
func TestExplicitAgentInstallsCanonicalTreeAndSelectedProviderAdapters(t *testing.T) {
	f := newFixture(t, "claude", "hermes")
	if err := Run(Options{Agent: "claude", AgentHome: filepath.Join(f.root, ".claude"), Source: f.source, Apply: true, Output: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	assertRegularTree(t, filepath.Join(f.root, ".agents", "sdlc"))
	assertAbsent(t, filepath.Join(f.root, ".claude", "sdlc"))
	assertSameDirectory(t, filepath.Join(f.source, "commands"), filepath.Join(f.root, ".claude", "commands"))
	assertAbsent(t, filepath.Join(f.root, ".hermes", "sdlc"))
}

func TestDryRunReportsDriftWithoutWriting(t *testing.T) {
	f := newFixture(t, "codex")
	var output bytes.Buffer
	if err := Run(Options{Agent: "codex", AgentHome: filepath.Join(f.root, ".codex"), Source: f.source, Output: &output}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "would install missing file") {
		t.Fatalf("dry run omitted drift:\n%s", output.String())
	}
	assertAbsent(t, filepath.Join(f.root, ".agents", "sdlc"))
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

func seedLegacyRuntimeArtifacts(t *testing.T, root string) []string {
	t.Helper()
	commandNames := []string{
		"build.md",
		"end-discovery.md",
		"implement.md",
		"migrate-acs.md",
		"review.md",
		"start-discovery.md",
		"write-tests.md",
	}
	commandRoots := []string{
		filepath.Join(root, ".agents", "sdlc", "commands"),
		filepath.Join(root, ".claude", "commands"),
		filepath.Join(root, ".codex", "prompts-commands"),
		filepath.Join(root, ".copilot", "prompts-commands"),
	}
	skillNames := []string{
		"audit-acs",
		"design-solution",
		"draft-bug-fix",
		"draft-design-issue",
		"draft-issue",
	}
	skillRoots := []string{
		filepath.Join(root, ".agents", "sdlc", "skills"),
		filepath.Join(root, ".agents", "skills"),
		filepath.Join(root, ".claude", "skills"),
		filepath.Join(root, ".copilot", "skills"),
		filepath.Join(root, ".hermes", "skills"),
	}

	var paths []string
	for _, relative := range retiredSharedFiles {
		path := filepath.Join(root, ".agents", "sdlc", relative)
		writeFile(t, path, []byte("legacy\n"))
		paths = append(paths, path)
	}
	for _, root := range commandRoots {
		for _, name := range commandNames {
			path := filepath.Join(root, name)
			writeFile(t, path, []byte("legacy\n"))
			paths = append(paths, path)
		}
	}
	for _, root := range skillRoots {
		for _, name := range skillNames {
			path := filepath.Join(root, name, "SKILL.md")
			writeFile(t, path, []byte("legacy\n"))
			paths = append(paths, path)
		}
	}
	return paths
}

func removePath(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil {
		t.Fatal(err)
	}
}

func assertNoBackup(t *testing.T, path string) {
	t.Helper()
	matches, err := filepath.Glob(path + ".*.bak")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("unexpected backups for %s: %v", path, matches)
	}
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
