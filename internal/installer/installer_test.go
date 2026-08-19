package installer_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/tigger-developer/sdlc/internal/installer"
)

func TestDryRunReportsPlanWithoutWriting(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Input:     strings.NewReader(""),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertContains(t, output.String(), "DRY RUN")
	assertContains(t, output.String(), "would create symlink")
	_, err = os.Lstat(filepath.Join(agentHome, "sdlc"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry run created destination: Lstat() error = %v", err)
	}
}

func TestApplyCreatesSymlinkToCanonicalClone(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	destination := filepath.Join(agentHome, "sdlc")
	target, err := os.Readlink(destination)
	if err != nil {
		t.Fatalf("Readlink(%q) error = %v", destination, err)
	}
	if target != source {
		t.Fatalf("symlink target = %q, want %q", target, source)
	}
	assertSymlinkTarget(t, filepath.Join(agentHome, "commands", "build.md"), filepath.Join(source, "commands", "build.md"))
	assertSymlinkTarget(t, filepath.Join(agentHome, "skills", "audit-code"), filepath.Join(source, "skills", "audit-code"))
	if _, err := os.Lstat(filepath.Join(agentHome, "skills", "draft-issue")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("state-changing skill was installed: Lstat() error = %v", err)
	}
	assertContains(t, output.String(), "installed")
}

func TestCodexApplyLinksPromptLibraryAndAdvisorySkills(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertSymlinkTarget(t, filepath.Join(agentHome, "prompts-commands"), filepath.Join(source, "commands"))
	sharedSkills := filepath.Join(filepath.Dir(agentHome), ".agents", "skills")
	assertSymlinkTarget(t, filepath.Join(sharedSkills, "audit-code"), filepath.Join(source, "skills", "audit-code"))
	if _, err := os.Lstat(filepath.Join(sharedSkills, "draft-issue")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("state-changing skill was installed: Lstat() error = %v", err)
	}
}

func TestCopilotApplyLinksAdvisorySkills(t *testing.T) {
	source, agentHome := newFixture(t, ".copilot")

	err := installer.Run(installer.Options{
		Agent:     "copilot",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertSymlinkTarget(t, filepath.Join(agentHome, "prompts-commands"), filepath.Join(source, "commands"))
	assertSymlinkTarget(t, filepath.Join(agentHome, "skills", "audit-code"), filepath.Join(source, "skills", "audit-code"))
}

func TestApplyRefusesToReplaceExistingDestination(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	destination := filepath.Join(agentHome, "sdlc")
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	marker := filepath.Join(destination, "keep-me")
	if err := os.WriteFile(marker, []byte("authoritative\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want destination conflict")
	}
	assertContains(t, err.Error(), "already exists")
	data, readErr := os.ReadFile(marker)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if string(data) != "authoritative\n" {
		t.Fatalf("existing destination changed: marker = %q", data)
	}
}

func TestAgentIsInferredFromAgentHome(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "auto",
		AgentHome: agentHome,
		Source:    source,
		Input:     strings.NewReader(""),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertContains(t, output.String(), "Agent: codex")
}

func TestAgentIsInferredWhenCloneAlreadyLivesInAgentHome(t *testing.T) {
	root := t.TempDir()
	agentHome := filepath.Join(root, ".codex")
	source := filepath.Join(agentHome, "sdlc")
	writeFile(t, filepath.Join(source, "MAIN.md"), []byte("# SDLC\n"))
	writeFile(t, filepath.Join(source, "README.md"), []byte("# Quickstart\n"))
	writeFile(t, filepath.Join(source, "templates", "codex-sdlc.rules.example"), []byte("prefix_rule()\n"))
	writeFile(t, filepath.Join(source, "commands", "build.md"), []byte("# Build\n"))
	writeFile(t, filepath.Join(source, "skills", "audit-code", "SKILL.md"), []byte("# Audit\n"))
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:  "auto",
		Source: source,
		Input:  strings.NewReader(""),
		Output: &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertContains(t, output.String(), "Agent: codex")
	assertContains(t, output.String(), "already resolves")
}

func TestApplyIsIdempotentForManagedLinks(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	options := installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	}

	if err := installer.Run(options); err != nil {
		t.Fatalf("first Run() error = %v", err)
	}
	var secondOutput bytes.Buffer
	options.Output = &secondOutput
	if err := installer.Run(options); err != nil {
		t.Fatalf("second Run() error = %v", err)
	}
	assertContains(t, secondOutput.String(), "already resolves")
}

func TestClaudeAnalysisNamesMissingCommandRestrictions(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	settings := []byte(`{"permissions":{"allow":["Read(**/*)"],"deny":[]},"custom":"preserve"}`)
	writeFile(t, filepath.Join(agentHome, "settings.json"), settings)
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Input:     strings.NewReader(""),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	for _, rule := range []string{"Bash(rm:*)", "Bash(sed:*)", "Bash(awk:*)", "Bash(source:*)", "Bash(python:*)", "Bash(python3:*)"} {
		assertContains(t, output.String(), rule)
	}
}

func TestClaudeConfigurationRequiresConfirmationAndPreservesUnknownFields(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	settingsPath := filepath.Join(agentHome, "settings.json")
	original := []byte(`{"permissions":{"allow":["Read(**/*)","Bash(sed:*)","Bash(python:*)","Bash(python3:*)"],"deny":[]},"custom":"preserve"}`)
	writeFile(t, settingsPath, original)
	var declinedOutput bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("no\n"),
		Output:    &declinedOutput,
	})
	if err != nil {
		t.Fatalf("declined Run() error = %v", err)
	}
	assertFileEquals(t, settingsPath, original)
	assertContains(t, declinedOutput.String(), "Configuration unchanged")

	var acceptedOutput bytes.Buffer
	err = installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &acceptedOutput,
	})
	if err != nil {
		t.Fatalf("accepted Run() error = %v", err)
	}

	var settings map[string]any
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if settings["custom"] != "preserve" {
		t.Fatalf("unknown field was not preserved: custom = %#v", settings["custom"])
	}
	permissions := settings["permissions"].(map[string]any)
	deny := permissions["deny"].([]any)
	for _, rule := range []string{"Bash(rm:*)", "Bash(sed:*)", "Bash(awk:*)", "Bash(source:*)", "Bash(python:*)", "Bash(python3:*)"} {
		if !containsAnyString(deny, rule) {
			t.Errorf("deny list missing %q: %#v", rule, deny)
		}
	}
	allow := permissions["allow"].([]any)
	for _, rule := range []string{"Bash(sed:*)", "Bash(python:*)", "Bash(python3:*)"} {
		if containsAnyString(allow, rule) {
			t.Errorf("allow list retains conflicting %q: %#v", rule, allow)
		}
	}
	if !containsAnyString(allow, "Read(**/*)") {
		t.Errorf("unrelated allow entry was not preserved: %#v", allow)
	}
	assertContains(t, acceptedOutput.String(), "Proposed configuration change")
	assertContains(t, acceptedOutput.String(), "Configuration updated")
	if strings.Contains(acceptedOutput.String(), "DRY RUN") {
		t.Fatalf("configuration-writing run was labelled dry-run: %q", acceptedOutput.String())
	}
}

func TestClaudeConfigurationDoesNotReplaceSettingsSymlink(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	realSettings := filepath.Join(filepath.Dir(agentHome), "dotfiles", "claude-settings.json")
	original := []byte(`{"permissions":{"deny":[]},"custom":"preserve"}`)
	writeFile(t, realSettings, original)
	settingsPath := filepath.Join(agentHome, "settings.json")
	if err := os.Symlink(realSettings, settingsPath); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileEquals(t, realSettings, original)
	if info, err := os.Lstat(settingsPath); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("settings symlink was replaced: info = %#v, error = %v", info, err)
	}
	assertContains(t, output.String(), "symlink")
	assertContains(t, output.String(), "no automatic change")
}

func TestCodexConfigurationCreatesRulesOnlyAfterConfirmation(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	writeFile(t, filepath.Join(agentHome, "config.toml"), []byte("approval_policy = \"on-request\"\nsandbox_mode = \"workspace-write\"\n"))
	rulesPath := filepath.Join(agentHome, "rules", "sdlc.rules")
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	data, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", rulesPath, err)
	}
	assertContains(t, string(data), `pattern = ["sed"]`)
	assertContains(t, string(data), `pattern = ["python"]`)
	assertContains(t, string(data), `pattern = ["python3"]`)
	assertContains(t, output.String(), "Configuration updated")
}

func TestCodexConfigurationUpgradesRecognizedRulesAndIsIdempotent(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	rulesPath := filepath.Join(agentHome, "rules", "sdlc.rules")
	legacy := []byte("# Existing SDLC rules.\n" +
		"prefix_rule(\n    pattern = [\"rm\"],\n    decision = \"forbidden\",\n)\n\n" +
		"prefix_rule(\n    pattern = [\"sed\"],\n    decision = \"forbidden\",\n)\n\n" +
		"prefix_rule(\n    pattern = [\"awk\"],\n    decision = \"forbidden\",\n)\n\n" +
		"prefix_rule(\n    pattern = [\"source\"],\n    decision = \"forbidden\",\n)\n\n" +
		"# Unrelated local rule.\nprefix_rule(\n    pattern = [\"local-tool\"],\n    decision = \"prompt\",\n)\n")
	writeFile(t, rulesPath, legacy)

	var firstOutput bytes.Buffer
	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &firstOutput,
	})
	if err != nil {
		t.Fatalf("first Run() error = %v", err)
	}
	upgraded, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", rulesPath, err)
	}
	for _, expected := range []string{`pattern = ["python"]`, `pattern = ["python3"]`, `pattern = ["local-tool"]`, "BEGIN SDLC MANAGED PYTHON RULES"} {
		assertContains(t, string(upgraded), expected)
	}
	assertContains(t, firstOutput.String(), "Configuration updated")

	var secondOutput bytes.Buffer
	err = installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &secondOutput,
	})
	if err != nil {
		t.Fatalf("second Run() error = %v", err)
	}
	assertFileEquals(t, rulesPath, upgraded)
	assertContains(t, secondOutput.String(), "already contain the SDLC command restrictions")
}

func TestCodexConfigurationLeavesAmbiguousRulesUnchanged(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	rulesPath := filepath.Join(agentHome, "rules", "sdlc.rules")
	original := []byte("prefix_rule(\n    pattern = [\"local-tool\"],\n    decision = \"prompt\",\n)\n")
	writeFile(t, rulesPath, original)
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileEquals(t, rulesPath, original)
	assertContains(t, output.String(), "cannot safely migrate")
}

func TestMalformedProviderConfigurationStopsWithContext(t *testing.T) {
	tests := []struct {
		name       string
		agent      string
		homeName   string
		configName string
		contents   string
	}{
		{name: "claude json", agent: "claude", homeName: ".claude", configName: "settings.json", contents: "{"},
		{name: "claude null", agent: "claude", homeName: ".claude", configName: "settings.json", contents: "null"},
		{name: "codex toml", agent: "codex", homeName: ".codex", configName: "config.toml", contents: "approval_policy = ["},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source, agentHome := newFixture(t, test.homeName)
			writeFile(t, filepath.Join(agentHome, test.configName), []byte(test.contents))
			err := installer.Run(installer.Options{
				Agent:     test.agent,
				AgentHome: agentHome,
				Source:    source,
				Input:     strings.NewReader(""),
				Output:    &bytes.Buffer{},
			})
			if err == nil {
				t.Fatal("Run() error = nil, want parse error")
			}
			assertContains(t, err.Error(), test.configName)
		})
	}
}

func TestDistributedProviderExamplesParse(t *testing.T) {
	claudePath := filepath.Join("..", "..", "templates", "claude-settings.example.json")
	claudeData, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", claudePath, err)
	}
	var claudeSettings map[string]any
	if err := json.Unmarshal(claudeData, &claudeSettings); err != nil {
		t.Fatalf("Claude settings example is invalid JSON: %v", err)
	}

	codexPath := filepath.Join("..", "..", "templates", "codex-config.example.toml")
	var codexSettings map[string]any
	if _, err := toml.DecodeFile(codexPath, &codexSettings); err != nil {
		t.Fatalf("Codex config example is invalid TOML: %v", err)
	}
}

func newFixture(t *testing.T, homeName string) (string, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "canonical", "sdlc")
	agentHome := filepath.Join(root, homeName)
	if err := os.MkdirAll(filepath.Join(source, "templates"), 0o700); err != nil {
		t.Fatalf("MkdirAll(source) error = %v", err)
	}
	if err := os.MkdirAll(agentHome, 0o700); err != nil {
		t.Fatalf("MkdirAll(agentHome) error = %v", err)
	}
	writeFile(t, filepath.Join(source, "MAIN.md"), []byte("# SDLC\n"))
	writeFile(t, filepath.Join(source, "README.md"), []byte("# Quickstart\n"))
	writeFile(t, filepath.Join(source, "templates", "codex-sdlc.rules.example"), []byte("prefix_rule(\n    pattern = [\"sed\"],\n    decision = \"forbidden\",\n)\n\n# BEGIN SDLC MANAGED PYTHON RULES\nprefix_rule(\n    pattern = [\"python\"],\n    decision = \"forbidden\",\n)\n\nprefix_rule(\n    pattern = [\"python3\"],\n    decision = \"forbidden\",\n)\n# END SDLC MANAGED PYTHON RULES\n"))
	writeFile(t, filepath.Join(source, "commands", "build.md"), []byte("# Build\n"))
	writeFile(t, filepath.Join(source, "skills", "audit-code", "SKILL.md"), []byte("# Audit code\n"))
	writeFile(t, filepath.Join(source, "skills", "draft-issue", "SKILL.md"), []byte("# Draft issue\n"))
	return source, agentHome
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("output %q does not contain %q", got, want)
	}
}

func assertFileEquals(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("file %q = %q, want %q", path, got, want)
	}
}

func containsAnyString(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func assertSymlinkTarget(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("Readlink(%q) error = %v", path, err)
	}
	if got != want {
		t.Fatalf("symlink %q target = %q, want %q", path, got, want)
	}
}
