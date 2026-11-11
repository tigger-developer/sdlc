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
	var changes []*configurationChange
	for _, item := range []struct {
		path  string
		value any
	}{
		{filepath.Join(root, "plugin.json"), manifest},
		{filepath.Join(root, "hooks", "hooks.json"), hooks},
	} {
		content, err := json.MarshalIndent(item.value, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("encoding Goose hook configuration: %w", err)
		}
		content = append(content, '\n')
		existing, err := os.ReadFile(item.path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading Goose hook configuration %q: %w", item.path, err)
		}
		linked, linkErr := pathIsSymlink(item.path)
		if linkErr != nil {
			return nil, linkErr
		}
		if bytes.Equal(existing, content) && !linked {
			continue
		}
		var previous map[string]any
		managed := false
		if json.Unmarshal(existing, &previous) == nil {
			managed = reflect.DeepEqual(previous, item.value)
			if filepath.Base(item.path) == "hooks.json" {
				managed = managedGooseHooks(previous)
			}
		}
		changes = append(changes, &configurationChange{
			path: item.path, beforeLabel: "existing Goose SDLC plugin file or absent",
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
