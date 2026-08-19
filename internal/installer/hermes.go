package installer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	hermesBootstrapStart       = "<!-- BEGIN SDLC-INSTALL OPERATIONS BOOTSTRAP -->"
	hermesBootstrapEnd         = "<!-- END SDLC-INSTALL OPERATIONS BOOTSTRAP -->"
	legacyHermesBootstrapStart = "<!-- BEGIN AGENTS-DEPLOY OPERATIONS BOOTSTRAP -->"
	legacyHermesBootstrapEnd   = "<!-- END AGENTS-DEPLOY OPERATIONS BOOTSTRAP -->"
)

func analyseHermesConfiguration(agentHome, source string, output io.Writer) (*configurationChange, error) {
	path := filepath.Join(agentHome, "config.yaml")
	current, err := os.ReadFile(path)
	mode := os.FileMode(0o600)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading Hermes configuration %s: %w", path, err)
	}
	if err == nil {
		info, statErr := os.Lstat(path)
		if statErr != nil {
			return nil, fmt.Errorf("inspecting Hermes configuration %s: %w", path, statErr)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("Hermes configuration %s is not a regular file", path)
		}
		mode = info.Mode().Perm()
	}
	bootstrapPath := filepath.Join(source, "templates", "hermes-bootstrap.md")
	bootstrap, err := os.ReadFile(bootstrapPath)
	if err != nil {
		return nil, fmt.Errorf("reading Hermes bootstrap template %s: %w", bootstrapPath, err)
	}
	hookCommand := fmt.Sprintf("bash %q", filepath.Join(agentHome, "sdlc", "hooks", "agent-command-guard.sh"))
	desired, changed, err := mergeHermesConfiguration(current, bootstrap, hookCommand)
	if err != nil {
		return nil, fmt.Errorf("analysing %s: %w", path, err)
	}
	if !changed {
		fmt.Fprintln(output, "Configuration: Hermes settings already contain the SDLC operations bootstrap and command guard.")
		return nil, nil
	}
	fmt.Fprintf(output, "Recommendation: update the SDLC operations bootstrap and command guard in %s.\n", path)
	promptState := "absent or empty"
	if len(bytes.TrimSpace(current)) != 0 {
		promptState = "existing configuration; custom prompt text will be preserved"
	}
	block := hermesBootstrapStart + "\n" + strings.TrimSpace(string(bootstrap)) + "\n" + hermesBootstrapEnd
	return &configurationChange{
		path:        path,
		beforeLabel: promptState + "; unrelated values preserved; YAML spacing and key order may be normalized",
		afterLabel:  fmt.Sprintf("managed operations bootstrap and command guard %s\n%s", hookCommand, block),
		contents:    desired,
		mode:        mode,
		backup:      true,
	}, nil
}

func mergeHermesConfiguration(current, bootstrap []byte, hookCommand string) ([]byte, bool, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace(current)) != 0 {
		if err := yaml.Unmarshal(current, &values); err != nil {
			return nil, false, fmt.Errorf("invalid YAML: %w", err)
		}
	}
	agentValue, exists := values["agent"]
	var agent map[string]any
	if !exists || agentValue == nil {
		agent = map[string]any{}
		values["agent"] = agent
	} else {
		var ok bool
		agent, ok = agentValue.(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf("agent is %T, not a mapping", agentValue)
		}
	}
	promptValue, exists := agent["system_prompt"]
	prompt := ""
	if exists && promptValue != nil {
		var ok bool
		prompt, ok = promptValue.(string)
		if !ok {
			return nil, false, fmt.Errorf("agent.system_prompt is %T, not a string", promptValue)
		}
	}
	block := hermesBootstrapStart + "\n" + strings.TrimSpace(string(bootstrap)) + "\n" + hermesBootstrapEnd
	startMarker, endMarker := hermesBootstrapStart, hermesBootstrapEnd
	startCount := strings.Count(prompt, startMarker)
	endCount := strings.Count(prompt, endMarker)
	legacyStartCount := strings.Count(prompt, legacyHermesBootstrapStart)
	legacyEndCount := strings.Count(prompt, legacyHermesBootstrapEnd)
	if startCount > 1 || endCount > 1 || startCount != endCount || legacyStartCount > 1 || legacyEndCount > 1 || legacyStartCount != legacyEndCount || (startCount == 1 && legacyStartCount == 1) {
		return nil, false, fmt.Errorf("ambiguous managed Hermes bootstrap markers")
	}
	if legacyStartCount == 1 {
		startMarker, endMarker = legacyHermesBootstrapStart, legacyHermesBootstrapEnd
		startCount = 1
	}
	if startCount == 1 {
		start := strings.Index(prompt, startMarker)
		endRelative := strings.Index(prompt[start:], endMarker)
		if endRelative < 0 {
			return nil, false, fmt.Errorf("ambiguous managed Hermes bootstrap markers")
		}
		end := start + endRelative + len(endMarker)
		agent["system_prompt"] = prompt[:start] + block + prompt[end:]
	} else if strings.TrimSpace(prompt) == "" {
		agent["system_prompt"] = block
	} else {
		agent["system_prompt"] = strings.TrimRight(prompt, "\n") + "\n\n" + block
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
