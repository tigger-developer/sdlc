// ABOUTME: Projects the shared SDLC command guard into Goose's native plugin hooks.
// ABOUTME: Owns only a dedicated plugin and blocks hook failures before tool execution.
package installer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func analyseGooseConfiguration(agentHome string) ([]*configurationChange, error) {
	root := filepath.Join(commonAgentHome(agentGoose, agentHome), "plugins", "sdlc-command-guard-goose")
	manifest := map[string]any{
		"name": "sdlc-command-guard-goose", "version": "1.0.0",
		"description": "Apply the SDLC shell and protected-file policy before Goose tools run.",
	}
	hooks := map[string]any{"hooks": map[string]any{"PreToolUse": []any{
		map[string]any{"hooks": []any{map[string]any{
			"type": "command", "command": `"${PLUGIN_ROOT}/../../sdlc/hooks/agent-command-guard.sh"`,
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
		changes = append(changes, &configurationChange{
			path: item.path, beforeLabel: "existing Goose SDLC plugin file or absent",
			afterLabel: "Goose PreToolUse command guard with on_failure=block",
			contents:   content, mode: 0o600, conflict: err == nil || linked,
		})
	}
	return changes, nil
}
