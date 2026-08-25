package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestHermesConfigurationOwnsOnlyCommandGuard(t *testing.T) {
	f := newFixture(t, "hermes")
	configPath := filepath.Join(f.root, ".hermes", "config.yaml")
	prompt := "Personal prompt.\n\n<!-- BEGIN AGENTS-DEPLOY OPERATIONS BOOTSTRAP -->\nRead ~/.hermes/OPERATIONS.md.\n<!-- END AGENTS-DEPLOY OPERATIONS BOOTSTRAP -->"
	writeFile(t, configPath, []byte("agent:\n  system_prompt: |\n"+indent(prompt, 4)+"\nhooks:\n  pre_tool_call:\n    - matcher: browser\n      command: /personal/hook\n      timeout: 9\n"))

	var output bytes.Buffer
	if err := Run(Options{
		Agent: "hermes", AgentHome: filepath.Join(f.root, ".hermes"), Source: f.source,
		Configure: true, Input: strings.NewReader("yes\n"), Output: &output,
	}); err != nil {
		t.Fatal(err)
	}
	parsed := readYAML(t, configPath)
	agent := parsed["agent"].(map[string]any)
	if got := agent["system_prompt"].(string); strings.TrimSpace(got) != strings.TrimSpace(prompt) {
		t.Fatalf("private prompt changed:\n%s", got)
	}
	serialized := string(mustReadFile(t, configPath))
	if strings.Contains(serialized, "SDLC-INSTALL OPERATIONS BOOTSTRAP") {
		t.Fatalf("SDLC inserted a private operations bootstrap:\n%s", serialized)
	}
	for _, expected := range []string{"/personal/hook", "agent-command-guard.sh"} {
		if !strings.Contains(serialized, expected) {
			t.Fatalf("configuration lacks %q:\n%s", expected, serialized)
		}
	}
	backup := assertOneBackup(t, configPath)
	if prior := string(mustReadFile(t, backup)); !strings.Contains(prior, "/personal/hook") || strings.Contains(prior, "agent-command-guard.sh") {
		t.Fatalf("Hermes backup did not preserve the prior configuration:\n%s", prior)
	}

	before := mustReadFile(t, configPath)
	output.Reset()
	if err := Run(Options{Agent: "hermes", AgentHome: filepath.Join(f.root, ".hermes"), Source: f.source, Configure: true, Input: strings.NewReader("yes\n"), Output: &output}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, mustReadFile(t, configPath)) || !strings.Contains(output.String(), "already contain the SDLC command guard") {
		t.Fatalf("Hermes configuration was not idempotent:\n%s", output.String())
	}
}

// RT-7.8
func TestHermesConfigurationMigratesProviderLocalCommandGuard(t *testing.T) {
	f := newFixture(t, "hermes")
	configPath := filepath.Join(f.root, ".hermes", "config.yaml")
	legacyCommand := "bash \"" + filepath.Join(f.root, ".hermes", "sdlc", "hooks", "agent-command-guard.sh") + "\""
	canonicalCommand := "bash \"" + filepath.Join(f.root, ".agents", "sdlc", "hooks", "agent-command-guard.sh") + "\""
	writeFile(t, configPath, []byte("hooks:\n  pre_tool_call:\n    - matcher: terminal\n      command: "+legacyCommand+"\n      timeout: 5\n    - matcher: terminal\n      command: "+canonicalCommand+"\n      timeout: 5\n"))

	if err := Run(Options{
		Agent: "hermes", AgentHome: filepath.Join(f.root, ".hermes"), Source: f.source,
		Configure: true, Input: strings.NewReader("yes\n"), Output: &bytes.Buffer{},
	}); err != nil {
		t.Fatal(err)
	}
	serialized := string(mustReadFile(t, configPath))
	if strings.Count(serialized, canonicalCommand) != 1 || strings.Contains(serialized, legacyCommand) {
		t.Fatalf("Hermes command guard was not migrated to the canonical root:\n%s", serialized)
	}
}

func TestHermesMalformedConfigurationStopsWithoutWriting(t *testing.T) {
	f := newFixture(t, "hermes")
	configPath := filepath.Join(f.root, ".hermes", "config.yaml")
	original := []byte("agent: [\n")
	writeFile(t, configPath, original)
	err := Run(Options{Agent: "hermes", AgentHome: filepath.Join(f.root, ".hermes"), Source: f.source, Configure: true, Input: strings.NewReader("yes\n"), Output: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "invalid YAML") {
		t.Fatalf("error = %v, want invalid YAML", err)
	}
	if !bytes.Equal(original, mustReadFile(t, configPath)) {
		t.Fatal("malformed Hermes configuration changed")
	}
}

// RT-7.3
func TestInteractiveHermesRequiresFirstRunConfiguration(t *testing.T) {
	f := newFixture(t, "hermes")
	var output bytes.Buffer
	err := RunInteractive(f.source, f.root, strings.NewReader("yes\n"), &output)
	if err == nil {
		t.Fatal("interactive install accepted Hermes without first-run configuration")
	}
	for _, expected := range []string{"⚠️", "Hermes first-run setup is incomplete", filepath.Join(f.root, ".hermes", "config.yaml")} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("error %q lacks %q", err, expected)
		}
	}
	assertAbsent(t, filepath.Join(f.root, ".agents", "sdlc"))
}

func indent(value string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	return prefix + strings.ReplaceAll(value, "\n", "\n"+prefix)
}

func readYAML(t *testing.T, path string) map[string]any {
	t.Helper()
	result := map[string]any{}
	if err := yaml.Unmarshal(mustReadFile(t, path), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
