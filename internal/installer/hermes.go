package installer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func analyseHermesConfiguration(agentHome, _ string, output io.Writer) (*configurationChange, error) {
	path := filepath.Join(agentHome, "config.yaml")
	current, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("⚠️ Hermes first-run setup is incomplete. Launch Hermes, complete the startup TUI and model selection, then rerun make install. Expected configuration: %s", path)
	}
	if err != nil {
		return nil, fmt.Errorf("reading Hermes configuration %s: %w", path, err)
	}
	info, statErr := os.Lstat(path)
	if statErr != nil {
		return nil, fmt.Errorf("inspecting Hermes configuration %s: %w", path, statErr)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("Hermes configuration %s is not a regular file", path)
	}
	mode := info.Mode().Perm()
	hookCommand := fmt.Sprintf("bash %q", filepath.Join(agentHome, "sdlc", "hooks", "agent-command-guard.sh"))
	desired, changed, err := mergeHermesCommandGuardConfiguration(current, hookCommand)
	if err != nil {
		return nil, fmt.Errorf("analysing %s: %w", path, err)
	}
	if !changed {
		fmt.Fprintln(output, "Configuration: Hermes settings already contain the SDLC command guard.")
		return nil, nil
	}
	fmt.Fprintf(output, "Recommendation: update the SDLC command guard in %s.\n", path)
	return &configurationChange{
		path:        path,
		beforeLabel: "existing configuration; private instructions and unrelated values preserved; YAML formatting may be normalized",
		afterLabel:  fmt.Sprintf("managed SDLC command guard %s", hookCommand),
		contents:    desired,
		mode:        mode,
	}, nil
}

func mergeHermesCommandGuardConfiguration(current []byte, hookCommand string) ([]byte, bool, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace(current)) != 0 {
		if err := yaml.Unmarshal(current, &values); err != nil {
			return nil, false, fmt.Errorf("invalid YAML: %w", err)
		}
	}
	if err := mergeHermesCommandGuard(values, hookCommand); err != nil {
		return nil, false, err
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(values); err != nil {
		return nil, false, err
	}
	if err := encoder.Close(); err != nil {
		return nil, false, err
	}
	desired := output.Bytes()
	return desired, !bytes.Equal(current, desired), nil
}

func mergeHermesCommandGuard(values map[string]any, hookCommand string) error {
	hooksValue, exists := values["hooks"]
	var hooks map[string]any
	if !exists || hooksValue == nil {
		hooks = map[string]any{}
		values["hooks"] = hooks
	} else {
		var ok bool
		hooks, ok = hooksValue.(map[string]any)
		if !ok {
			return fmt.Errorf("hooks is %T, not a mapping", hooksValue)
		}
	}
	entriesValue, exists := hooks["pre_tool_call"]
	var entries []any
	if !exists || entriesValue == nil {
		entries = []any{}
	} else {
		var ok bool
		entries, ok = entriesValue.([]any)
		if !ok {
			return fmt.Errorf("hooks.pre_tool_call is %T, not a list", entriesValue)
		}
	}
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("hooks.pre_tool_call contains %T, not a mapping", raw)
		}
		if entry["command"] == hookCommand {
			entry["matcher"] = "terminal"
			entry["timeout"] = 5
			hooks["pre_tool_call"] = entries
			return nil
		}
	}
	hooks["pre_tool_call"] = append(entries, map[string]any{
		"matcher": "terminal",
		"command": hookCommand,
		"timeout": 5,
	})
	return nil
}
