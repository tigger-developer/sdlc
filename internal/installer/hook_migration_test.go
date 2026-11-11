// ABOUTME: Verifies legacy guard migration and duplicate removal across harnesses.
// ABOUTME: Preserves unrelated hooks and proves repeated configuration is stable.
package installer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestJSONHarnessGuardMigration(t *testing.T) {
	for name, normalize := range map[string]func(map[string]any) (bool, error){"codex": normalizeCodexToolGuard, "claude": normalizeClaudeToolGuard} {
		t.Run(name, func(t *testing.T) {
			var root map[string]any
			input := `{"personal":true,"hooks":{"PostToolUse":[{"personal":true}],"PreToolUse":[{"matcher":"Bash","hooks":[{"command":"bash ~/.agents/sdlc/hooks/agent-command-guard.sh","type":"command","timeout":5},{"command":"personal-hook","type":"command"}]},{"matcher":"*","hooks":[{"command":"bash /opt/sdlc/hooks/agent-command-guard.sh","type":"command","timeout":5}]}]}}`
			if err := json.Unmarshal([]byte(input), &root); err != nil {
				t.Fatal(err)
			}
			if changed, err := normalize(root); err != nil || !changed {
				t.Fatalf("migration: %v %v", changed, err)
			}
			hooks := root["hooks"].(map[string]any)
			groups := hooks["PreToolUse"].([]any)
			if len(groups) != 2 {
				t.Fatalf("duplicate groups: %+v", groups)
			}
			personal := groups[0].(map[string]any)
			if personal["matcher"] != "Bash" || personal["hooks"].([]any)[0].(map[string]any)["command"] != "personal-hook" || root["personal"] != true || hooks["PostToolUse"] == nil {
				t.Fatal("unrelated configuration changed")
			}
			guard := groups[1].(map[string]any)
			if guard["matcher"] != "*" || len(guard["hooks"].([]any)) != 1 || guard["hooks"].([]any)[0].(map[string]any)["command"] != toolGuardCommand {
				t.Fatalf("guard: %+v", guard)
			}
			encoded, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), "~/.agents") {
				t.Fatal("stale guard remains")
			}
			if changed, err := normalize(root); err != nil || changed {
				t.Fatalf("repeat: %v %v", changed, err)
			}
		})
	}
}

func TestHermesGuardMigrationDuplicates(t *testing.T) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte("personal: true\nhooks:\n  pre_tool_call:\n    - command: bash ~/.agents/sdlc/hooks/agent-command-guard.sh\n      matcher: terminal\n      timeout: 5\n    - command: bash /opt/sdlc/hooks/agent-command-guard.sh\n      matcher: .*\n      timeout: 5\n    - command: personal-hook\n      matcher: file\n"), &doc); err != nil {
		t.Fatal(err)
	}
	root, err := hermesDocumentMapping(&doc)
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := ensureHermesCommandGuard(root); err != nil || !changed {
		t.Fatalf("migration %v %v", changed, err)
	}
	hooks, _ := hermesMappingValue(root, "hooks")
	entries, _ := hermesMappingValue(hooks, "pre_tool_call")
	if len(entries.Content) != 2 {
		t.Fatal("duplicates remain")
	}
	if command, _ := hermesScalarValue(entries.Content[0], "command"); command != toolGuardCommand {
		t.Fatal(command)
	}
	if command, _ := hermesScalarValue(entries.Content[1], "command"); command != "personal-hook" {
		t.Fatal(command)
	}
	if changed, err := ensureHermesCommandGuard(root); err != nil || changed {
		t.Fatalf("repeat %v %v", changed, err)
	}
}

func TestCopilotMigratesMixedHooks(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "hooks", "sdlc-tool-guard.json")
	writeFixtureFile(t, path, `{"version":1,"personal":true,"hooks":{"postToolUse":[],"preToolUse":[{"type":"command","command":"bash ~/.agents/sdlc/hooks/agent-command-guard.sh","timeoutSec":5},{"type":"command","command":"personal-hook"},{"type":"command","command":"bash /opt/sdlc/hooks/agent-command-guard.sh","timeoutSec":5}]}}`)
	change, err := analyseCopilotConfiguration(home, &bytes.Buffer{})
	if err != nil || change == nil || change.conflict {
		t.Fatalf("migration %+v %v", change, err)
	}
	var root map[string]any
	if err := json.Unmarshal(change.contents, &root); err != nil {
		t.Fatal(err)
	}
	hooks := root["hooks"].(map[string]any)
	entries := hooks["preToolUse"].([]any)
	if len(entries) != 2 || entries[0].(map[string]any)["command"] != "personal-hook" || entries[1].(map[string]any)["command"] != toolGuardCommand || root["personal"] != true || hooks["postToolUse"] == nil {
		t.Fatalf("unexpected configuration %+v", root)
	}
	if err := os.WriteFile(path, change.contents, 0o600); err != nil {
		t.Fatal(err)
	}
	if next, err := analyseCopilotConfiguration(home, &bytes.Buffer{}); err != nil || next != nil {
		t.Fatalf("repeat %+v %v", next, err)
	}
}

func TestGooseMigratesKnownPluginWithoutConflict(t *testing.T) {
	root := t.TempDir()
	changes, err := analyseGooseConfiguration(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range changes {
		writeFixtureFile(t, change.path, string(change.contents))
	}
	path := filepath.Join(root, "plugins", "sdlc-command-guard-goose", "hooks", "hooks.json")
	writeFixtureFile(t, path, `{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"\"${PLUGIN_ROOT}/../../sdlc/hooks/agent-command-guard.sh\"","timeout":10,"on_failure":"block"},{"type":"command","command":"\"/opt/sdlc/hooks/agent-command-guard.sh\"","timeout":10,"on_failure":"block"}]}]}}`)
	changes, err = analyseGooseConfiguration(root)
	if err != nil || len(changes) != 1 || changes[0].conflict {
		t.Fatalf("migration %+v %v", changes, err)
	}
	writeFixtureFile(t, path, string(changes[0].contents))
	if next, err := analyseGooseConfiguration(root); err != nil || len(next) != 0 {
		t.Fatalf("repeat %+v %v", next, err)
	}
}
