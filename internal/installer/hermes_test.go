package installer_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/installer"
	"gopkg.in/yaml.v3"
)

func TestHermesConfigurationPreservesUnrelatedValuesAndIsIdempotent(t *testing.T) {
	source, agentHome := newFixture(t, ".hermes")
	configPath := filepath.Join(agentHome, "config.yaml")
	writeFile(t, configPath, []byte("model:\n  provider: example\nagent:\n  system_prompt: |\n    Existing custom prompt.\nterminal:\n  backend: local\nhooks_auto_accept: false\nhooks:\n  pre_tool_call:\n    - matcher: browser\n      command: /custom/existing-hook\n      timeout: 9\n"))
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent: "hermes", AgentHome: agentHome, Source: source,
		Configure: true, Input: strings.NewReader("yes\n"), Output: &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(output.String(), "Backup:") {
		t.Fatalf("Hermes configuration backup was not reported:\n%s", output.String())
	}
	first, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	parsed := parseHermesYAML(t, first)
	if nestedHermesString(t, parsed, "model", "provider") != "example" || nestedHermesString(t, parsed, "terminal", "backend") != "local" {
		t.Fatal("unrelated Hermes configuration values changed")
	}
	prompt := nestedHermesString(t, parsed, "agent", "system_prompt")
	assertContains(t, prompt, "Existing custom prompt.")
	assertContains(t, prompt, "BEGIN SDLC-INSTALL OPERATIONS BOOTSTRAP")
	if parsed["hooks_auto_accept"] != false {
		t.Fatalf("hooks_auto_accept changed: %#v", parsed["hooks_auto_accept"])
	}
	assertHermesHook(t, parsed, "browser", "/custom/existing-hook", 9)
	assertHermesHook(t, parsed, "terminal", fmt.Sprintf("bash %q", filepath.Join(agentHome, "sdlc", "hooks", "agent-command-guard.sh")), 5)

	output.Reset()
	err = installer.Run(installer.Options{
		Agent: "hermes", AgentHome: agentHome, Source: source,
		Configure: true, Input: strings.NewReader("yes\n"), Output: &output,
	})
	if err != nil {
		t.Fatalf("second Run() error = %v", err)
	}
	assertFileEquals(t, configPath, first)
	assertContains(t, output.String(), "already contain the SDLC operations bootstrap and command guard")
}

func TestHermesDryRunRecommendsConfigurationWithoutWriting(t *testing.T) {
	source, agentHome := newFixture(t, ".hermes")
	configPath := filepath.Join(agentHome, "config.yaml")
	original := []byte("agent:\n  system_prompt: |\n    Existing prompt.\n")
	writeFile(t, configPath, original)
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent: "hermes", AgentHome: agentHome, Source: source,
		Input: strings.NewReader(""), Output: &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileEquals(t, configPath, original)
	assertContains(t, output.String(), "re-run with --configure")
}

func TestHermesConfigurationMigratesLegacyAgentsDeployMarkers(t *testing.T) {
	source, agentHome := newFixture(t, ".hermes")
	configPath := filepath.Join(agentHome, "config.yaml")
	writeFile(t, configPath, []byte("agent:\n  system_prompt: |\n    Existing prompt.\n\n    <!-- BEGIN AGENTS-DEPLOY OPERATIONS BOOTSTRAP -->\n    Legacy bootstrap.\n    <!-- END AGENTS-DEPLOY OPERATIONS BOOTSTRAP -->\n"))
	err := installer.Run(installer.Options{
		Agent: "hermes", AgentHome: agentHome, Source: source,
		Configure: true, Input: strings.NewReader("yes\n"), Output: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(updated), "BEGIN AGENTS-DEPLOY") {
		t.Fatalf("legacy management marker remains:\n%s", updated)
	}
	assertContains(t, string(updated), "BEGIN SDLC-INSTALL OPERATIONS BOOTSTRAP")
}

func TestHermesIsInferredFromAgentHome(t *testing.T) {
	source, agentHome := newFixture(t, ".hermes")
	var output bytes.Buffer
	err := installer.Run(installer.Options{
		Agent: "auto", AgentHome: agentHome, Source: source,
		Input: strings.NewReader(""), Output: &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertContains(t, output.String(), "Agent: hermes")
}

func TestHermesDuplicateManagedMarkersStopWithoutWriting(t *testing.T) {
	source, agentHome := newFixture(t, ".hermes")
	configPath := filepath.Join(agentHome, "config.yaml")
	original := []byte("agent:\n  system_prompt: |\n    <!-- BEGIN AGENTS-DEPLOY OPERATIONS BOOTSTRAP -->\n    one\n    <!-- BEGIN AGENTS-DEPLOY OPERATIONS BOOTSTRAP -->\n")
	writeFile(t, configPath, original)
	err := installer.Run(installer.Options{
		Agent: "hermes", AgentHome: agentHome, Source: source,
		Configure: true, Input: strings.NewReader("yes\n"), Output: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want marker error")
	}
	assertContains(t, err.Error(), "ambiguous managed Hermes bootstrap markers")
	assertFileEquals(t, configPath, original)
}

func parseHermesYAML(t *testing.T, content []byte) map[string]any {
	t.Helper()
	var parsed map[string]any
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		t.Fatal(err)
	}
	return parsed
}

func nestedHermesString(t *testing.T, values map[string]any, keys ...string) string {
	t.Helper()
	var current any = values
	for _, key := range keys {
		mapping, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("%s is not a mapping in %#v", key, current)
		}
		current, ok = mapping[key]
		if !ok {
			t.Fatalf("missing key %s", key)
		}
	}
	value, ok := current.(string)
	if !ok {
		t.Fatalf("%v is not a string: %T", keys, current)
	}
	return value
}

func assertHermesHook(t *testing.T, values map[string]any, matcher, command string, timeout int) {
	t.Helper()
	hooks, ok := values["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("hooks is not a mapping: %#v", values["hooks"])
	}
	entries, ok := hooks["pre_tool_call"].([]any)
	if !ok {
		t.Fatalf("hooks.pre_tool_call is not a list: %#v", hooks["pre_tool_call"])
	}
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if ok && entry["matcher"] == matcher && entry["command"] == command && entry["timeout"] == timeout {
			return
		}
	}
	t.Fatalf("hook matcher=%q command=%q timeout=%d missing from %#v", matcher, command, timeout, entries)
}
