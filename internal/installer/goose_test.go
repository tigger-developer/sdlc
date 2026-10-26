// ABOUTME: Verifies Goose installation paths, plugin decisions and idempotent selection.
// ABOUTME: Runs the command policy locally without invoking a model or external service.
package installer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGooseInteractiveInstallation(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	home := filepath.Join(root, ".config", "goose")
	installHarnessExecutables(t, root)
	writeFixtureFile(t, filepath.Join(home, "config.yaml"), "operator: preserved\n")
	writeFixtureFile(t, filepath.Join(source, "README.md"), "source\n")
	writeFixtureFile(t, filepath.Join(source, "src", "MAIN.md"), "public rules\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "define-change", "SKILL.md"), "definition\n")
	writeCommandFixtures(t, source)
	writeGuardFixture(t, filepath.Join(source, "hooks", "agent-command-guard.sh"))
	var output bytes.Buffer
	if err := RunInteractive(source, root, "v3.8.18", strings.NewReader("goose\n"), &output); err != nil {
		t.Fatalf("%v\n%s", err, output.String())
	}
	assertFixtureContent(t, filepath.Join(root, ".agents", "sdlc", "MAIN.md"), "public rules\n")
	assertFixtureContent(t, filepath.Join(home, "skills", "define-change", "SKILL.md"), "definition\n")
	assertFixtureContent(t, filepath.Join(home, "config.yaml"), "operator: preserved\n")
	if _, err := os.Stat(filepath.Join(root, ".config", ".agents")); !os.IsNotExist(err) {
		t.Fatalf("incorrect shared root: %v", err)
	}
	hooksPath := filepath.Join(root, ".agents", "plugins", "sdlc-command-guard-goose", "hooks", "hooks.json")
	var config struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command   string
				OnFailure string `json:"on_failure"`
			}
		}
	}
	if err := json.Unmarshal([]byte(readFixtureFile(t, hooksPath)), &config); err != nil {
		t.Fatal(err)
	}
	rules := config.Hooks["PreToolUse"]
	if len(rules) != 1 || len(rules[0].Hooks) != 1 || rules[0].Hooks[0].OnFailure != "block" || !strings.Contains(rules[0].Hooks[0].Command, "agent-command-guard.sh") {
		t.Fatalf("invalid blocking contract: %+v", config)
	}
	if !strings.Contains(output.String(), "external-audit-adapter=NOT-IN-SCOPE") {
		t.Fatalf("claims legacy adapter support: %s", output.String())
	}
	output.Reset()
	if err := RunInteractive(source, root, "v3.8.18", strings.NewReader("goose\n"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "current") {
		t.Fatalf("not idempotent: %s", output.String())
	}
}

func TestGooseNativeGuardPayload(t *testing.T) {
	for _, tc := range []struct {
		command string
		blocked bool
	}{
		{"printf 'hello'", false}, {"rm obsolete", true}, {"cat .env", true},
		{"python3 -V", true}, {"printf '%s' 'rm is data'", false},
	} {
		t.Run(tc.command, func(t *testing.T) {
			blocked, _, stderr := runGuard(t, map[string]any{"event": "PreToolUse", "tool_name": "shell", "tool_input": map[string]any{"command": tc.command}})
			if blocked != tc.blocked {
				t.Fatalf("blocked=%v want=%v: %s", blocked, tc.blocked, stderr)
			}
		})
	}
}

func TestGooseExplicitInstallationPreservesConflictingPlugin(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	home := filepath.Join(root, ".config", "goose")
	installHarnessExecutables(t, root)
	writeFixtureFile(t, filepath.Join(source, "README.md"), "source\n")
	writeFixtureFile(t, filepath.Join(source, "src", "MAIN.md"), "rules\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "define-change", "SKILL.md"), "skill\n")
	writeCommandFixtures(t, source)
	writeGuardFixture(t, filepath.Join(source, "hooks", "agent-command-guard.sh"))
	plugin := filepath.Join(root, ".agents", "plugins", "sdlc-command-guard-goose", "hooks", "hooks.json")
	writeFixtureFile(t, plugin, "operator-owned\n")
	unrelated := filepath.Join(root, ".agents", "plugins", "another-plugin", "plugin.json")
	writeFixtureFile(t, unrelated, "preserve\n")
	var output bytes.Buffer
	err := Run(Options{Agent: "goose", AgentHome: home, Source: source, Apply: true, Configure: true, Input: strings.NewReader("n\nn\n"), Output: &output})
	if err == nil || !strings.Contains(err.Error(), "declined replacement") {
		t.Fatalf("expected refusal to leave readiness unresolved, got %v", err)
	}
	assertFixtureContent(t, filepath.Join(root, ".agents", "sdlc", "MAIN.md"), "rules\n")
	assertFixtureContent(t, filepath.Join(home, "skills", "define-change", "SKILL.md"), "skill\n")
	assertFixtureContent(t, plugin, "operator-owned\n")
	assertFixtureContent(t, unrelated, "preserve\n")
}
