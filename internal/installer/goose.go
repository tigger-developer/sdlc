// ABOUTME: Projects the shared SDLC command guard into Goose's native plugin hooks.
// ABOUTME: Owns only a dedicated plugin and blocks hook failures before tool execution.
package installer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

func analyseGooseConfiguration(installRoot string) ([]*configurationChange, error) {
	root := filepath.Join(installRoot, "plugins", "sdlc-command-guard-goose")
	files, err := goosePluginContents()
	if err != nil {
		return nil, err
	}
	var changes []*configurationChange
	for name, content := range files {
		path := filepath.Join(root, name)
		existing, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading Goose hook configuration %q: %w", path, err)
		}
		linked, linkErr := pathIsSymlink(path)
		if linkErr != nil {
			return nil, linkErr
		}
		if bytes.Equal(existing, content) && !linked {
			continue
		}
		var previous map[string]any
		managed := false
		if json.Unmarshal(existing, &previous) == nil {
			var desired map[string]any
			if err := json.Unmarshal(content, &desired); err != nil {
				return nil, err
			}
			managed = reflect.DeepEqual(previous, desired)
			if filepath.Base(path) == "hooks.json" {
				managed = managedGooseHooks(previous)
			}
		}
		changes = append(changes, &configurationChange{
			path: path, beforeLabel: "existing Goose SDLC plugin file or absent",
			afterLabel: "Goose PreToolUse command guard with on_failure=block",
			contents:   content, mode: 0o644, conflict: (err == nil && !managed) || linked,
		})
	}
	return changes, nil
}

// Known plugin generations contain only SDLC policy handlers. Unknown additions
// remain conflicts rather than being discarded during an automatic migration.
func managedGooseHooks(root map[string]any) bool {
	if !onlyKeys(root, "hooks") {
		return false
	}
	hooks, ok := root["hooks"].(map[string]any)
	if !ok || !onlyKeys(hooks, "PreToolUse") {
		return false
	}
	groups, ok := hooks["PreToolUse"].([]any)
	if !ok || len(groups) == 0 {
		return false
	}
	for _, raw := range groups {
		group, ok := raw.(map[string]any)
		if !ok || !onlyKeys(group, "hooks") {
			return false
		}
		handlers, ok := group["hooks"].([]any)
		if !ok || len(handlers) == 0 {
			return false
		}
		for _, rawHandler := range handlers {
			handler, ok := rawHandler.(map[string]any)
			if !ok || !onlyKeys(handler, "type", "command", "timeout", "on_failure") || handler["type"] != "command" || handler["timeout"] != float64(10) || handler["on_failure"] != "block" {
				return false
			}
			command, _ := handler["command"].(string)
			if command != `"${PLUGIN_ROOT}/../../sdlc/hooks/agent-command-guard.sh"` && command != `"/opt/sdlc/hooks/agent-command-guard.sh"` {
				return false
			}
		}
	}
	return true
}

func goosePluginContents() (map[string][]byte, error) {
	manifest := map[string]any{
		"name": "sdlc-command-guard-goose", "version": "1.0.0",
		"description": "Apply the SDLC shell and protected-file policy before Goose tools run.",
	}
	hooks := map[string]any{"hooks": map[string]any{"PreToolUse": []any{
		map[string]any{"hooks": []any{map[string]any{
			"type": "command", "command": `"/opt/sdlc/hooks/agent-command-guard.sh"`,
			"timeout": 10, "on_failure": "block",
		}}},
	}}}

	result := map[string][]byte{}
	for name, value := range map[string]any{"plugin.json": manifest, filepath.Join("hooks", "hooks.json"): hooks} {
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return nil, err
		}
		result[name] = append(data, '\n')
	}
	return result, nil
}
