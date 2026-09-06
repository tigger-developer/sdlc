package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV3InteractiveInstallUsesOneGlobalTreeAndRetiresUnsupportedCopies(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	writeFixtureFile(t, filepath.Join(source, "README.md"), "# SDLC\n")
	writeFixtureFile(t, filepath.Join(source, "src", "MAIN.md"), "# Lean SDLC\n")
	writeFixtureFile(t, filepath.Join(source, "src", "prompts", "discover-project-authorities.md"), "Discover project authorities.\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "audit-code", "SKILL.md"), "---\nname: audit-code\ndescription: Review code.\n---\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "audit-test-code", "SKILL.md"), "---\nname: audit-test-code\ndescription: Review implemented tests.\n---\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "audit-test-definitions", "SKILL.md"), "---\nname: audit-test-definitions\ndescription: Review test definitions.\n---\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "define-change", "SKILL.md"), "---\nname: define-change\ndescription: Define a change.\n---\n")
	writeFixtureFile(t, filepath.Join(source, "hooks", "agent-command-guard.sh"), "#!/usr/bin/env bash\n")
	writeFixtureFile(t, filepath.Join(source, "templates", "codex-sdlc.rules.example"), codexPythonRulesStart+"\nprefix_rule(pattern=[\"python3\"], decision=\"forbidden\")\n"+codexPythonRulesEnd+"\n")

	for _, provider := range []string{"claude", "codex", "copilot", "hermes"} {
		if err := os.MkdirAll(filepath.Join(root, "."+provider), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, provider := range []string{"claude", "copilot", "hermes"} {
		writeFixtureFile(t, filepath.Join(root, "."+provider, "skills", "audit-code", "SKILL.md"), "old\n")
		writeFixtureFile(t, filepath.Join(root, "."+provider, "skills", "define-change", "SKILL.md"), "old\n")
	}
	writeFixtureFile(t, filepath.Join(root, ".copilot", "hooks", "sdlc-tool-guard.json"), "{}\n")
	writeFixtureFile(t, filepath.Join(root, ".agents", "sdlc", "skills", "audit-code", "SKILL.md"), "duplicate\n")
	writeFixtureFile(t, filepath.Join(root, ".agents", "skills", "audit-tests", "SKILL.md"), "obsolete\n")
	writeFixtureFile(t, filepath.Join(root, ".claude", "settings.json"), `{"personal":true,"hooks":{"PreToolUse":[{"matcher":"*","hooks":[{"type":"command","command":"bash ~/.agents/sdlc/hooks/agent-command-guard.sh","timeout":5}]}]}}`)
	writeFixtureFile(t, filepath.Join(root, ".hermes", "config.yaml"), "personal: true\nhooks:\n  pre_tool_call:\n    - command: bash ~/.agents/sdlc/hooks/agent-command-guard.sh\n      matcher: .*\n      timeout: 5\n")

	var output bytes.Buffer
	if err := RunInteractive(source, root, "v3.0.0", strings.NewReader("yes\n"), &output); err != nil {
		t.Fatalf("install: %v\n%s", err, output.String())
	}
	assertFixtureContent(t, filepath.Join(root, ".agents", "sdlc", "MAIN.md"), "# Lean SDLC\n")
	assertFixtureContent(t, filepath.Join(root, ".agents", "skills", "audit-code", "SKILL.md"), "---\nname: audit-code\ndescription: Review code.\n---\n")
	assertFixtureContent(t, filepath.Join(root, ".agents", "skills", "audit-test-code", "SKILL.md"), "---\nname: audit-test-code\ndescription: Review implemented tests.\n---\n")
	assertFixtureContent(t, filepath.Join(root, ".agents", "skills", "audit-test-definitions", "SKILL.md"), "---\nname: audit-test-definitions\ndescription: Review test definitions.\n---\n")
	assertFixtureContent(t, filepath.Join(root, ".agents", "skills", "define-change", "SKILL.md"), "---\nname: define-change\ndescription: Define a change.\n---\n")
	if _, err := os.Lstat(filepath.Join(root, ".agents", "skills", "audit-tests")); !os.IsNotExist(err) {
		t.Fatalf("retired global audit-tests skill remains: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(root, ".agents", "sdlc", "skills")); !os.IsNotExist(err) {
		t.Fatalf("duplicate canonical-root skills directory remains: %v", err)
	}
	for _, provider := range []string{"claude", "copilot", "hermes"} {
		for _, skill := range []string{"audit-code", "define-change"} {
			if _, err := os.Lstat(filepath.Join(root, "."+provider, "skills", skill)); !os.IsNotExist(err) {
				t.Fatalf("unsupported provider skill %s remains for %s: %v", skill, provider, err)
			}
		}
	}
	claude := readFixtureFile(t, filepath.Join(root, ".claude", "settings.json"))
	if !strings.Contains(claude, `"personal": true`) || strings.Contains(claude, "agent-command-guard") {
		t.Fatalf("Claude cleanup did not preserve personal data and remove guard: %s", claude)
	}
	hermes := readFixtureFile(t, filepath.Join(root, ".hermes", "config.yaml"))
	if !strings.Contains(hermes, "personal: true") || strings.Contains(hermes, "agent-command-guard") {
		t.Fatalf("Hermes cleanup did not preserve personal data and remove guard: %s", hermes)
	}
	global := readFixtureFile(t, filepath.Join(root, ".agents", "sdlc.yaml"))
	if !strings.Contains(global, "version: 3") || !strings.Contains(global, "release: v3.0.0") {
		t.Fatalf("global configuration does not identify the schema and deployed release:\n%s", global)
	}
}

func TestV3InteractiveInstallIsSilentNoOpAfterSynchronization(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	writeFixtureFile(t, filepath.Join(source, "README.md"), "# SDLC\n")
	writeFixtureFile(t, filepath.Join(source, "src", "MAIN.md"), "# Lean SDLC\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "audit-code", "SKILL.md"), "---\nname: audit-code\ndescription: Review code.\n---\n")
	writeFixtureFile(t, filepath.Join(source, "hooks", "agent-command-guard.sh"), "#!/usr/bin/env bash\n")
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, filepath.Join(source, "templates", "codex-sdlc.rules.example"), codexPythonRulesStart+"\n"+codexPythonRulesEnd+"\n")
	var first bytes.Buffer
	writeFixtureFile(t, filepath.Join(root, ".agents", "sdlc.yaml"), "# operator settings\nversion: 3\nrelease: v2.1.0\ndelivery:\n  branch_strategy: feature\n")
	if err := RunInteractive(source, root, "v3.0.0", strings.NewReader("yes\n"), &first); err != nil {
		t.Fatal(err)
	}
	global := readFixtureFile(t, filepath.Join(root, ".agents", "sdlc.yaml"))
	for _, expected := range []string{"# operator settings", "version: 3", "release: v3.0.0", "branch_strategy: feature"} {
		if !strings.Contains(global, expected) {
			t.Fatalf("global configuration lost %q:\n%s", expected, global)
		}
	}
	var second bytes.Buffer
	if err := RunInteractive(source, root, "v3.0.0", strings.NewReader("unexpected\n"), &second); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(second.String(), "All detected SDLC copies are current.") || strings.Contains(second.String(), "Deploy all") {
		t.Fatalf("second run was noisy or prompted:\n%s", second.String())
	}
}

func writeFixtureFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	// #nosec G306 -- fixtures model tracked public project files.
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFixtureFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func assertFixtureContent(t *testing.T, path, expected string) {
	t.Helper()
	if actual := readFixtureFile(t, path); actual != expected {
		t.Fatalf("%s = %q, want %q", path, actual, expected)
	}
}
