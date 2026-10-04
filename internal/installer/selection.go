package installer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const activeHarnessKey = "active-harness"

func selectInteractiveAgents(userHome string, reader *bufio.Reader, output io.Writer) ([]string, error) {
	detected, err := detectedAgents(userHome)
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(userHome, ".agents", "sdlc.yaml")
	configured, present, err := configuredActiveHarnesses(configPath)
	if err != nil {
		return nil, err
	}
	if present {
		selected := availableAgents(configured, detected)
		fmt.Fprintf(output, "Active harnesses from %s with existing agent homes: %s\n", configPath, displayAgentNames(selected))
		return selected, nil
	}
	fmt.Fprintf(output, "Detected agent homes: %s\n", displayAgentNames(detected))
	if len(detected) == 0 {
		return nil, nil
	}
	fmt.Fprint(output, "Install the SDLC for which agents? Enter comma-separated names, or press Enter for all detected agents: ")
	answer, readErr := reader.ReadString('\n')
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return nil, fmt.Errorf("reading agent selection: %w", readErr)
	}
	return parseAgentSelection(answer, detected)
}

func configuredActiveHarnesses(path string) ([]string, bool, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("reading global SDLC configuration %q: %w", path, err)
	}
	_, root, err := globalConfigurationDocument(path, contents)
	if err != nil {
		return nil, false, err
	}
	value, _ := releaseMappingValue(root, activeHarnessKey)
	if value == nil {
		return nil, false, nil
	}
	if value.Kind != yaml.SequenceNode {
		return nil, true, fmt.Errorf("%s in %q must be a YAML list", activeHarnessKey, path)
	}
	var configured []string
	for _, item := range value.Content {
		if item.Kind != yaml.ScalarNode {
			return nil, true, fmt.Errorf("%s entries in %q must be agent names", activeHarnessKey, path)
		}
		configured = append(configured, item.Value)
	}
	validated, err := validateAgentNames(configured, supportedProviderNames())
	if err != nil {
		return nil, true, fmt.Errorf("validating %s in %q: %w", activeHarnessKey, path, err)
	}
	return validated, true, nil
}

func supportedProviderNames() []string {
	names := make([]string, 0, len(providerDefinitions))
	for _, provider := range providerDefinitions {
		names = append(names, provider.name)
	}
	return names
}

func availableAgents(configured, detected []string) []string {
	detectedSet := map[string]bool{}
	for _, agent := range detected {
		detectedSet[agent] = true
	}
	var selected []string
	for _, agent := range configured {
		if detectedSet[agent] {
			selected = append(selected, agent)
		}
	}
	return selected
}

func parseAgentSelection(answer string, detected []string) ([]string, error) {
	answer = strings.TrimSpace(answer)
	if answer == "" || strings.EqualFold(answer, "all") {
		return append([]string(nil), detected...), nil
	}
	return validateAgentNames(strings.Split(answer, ","), detected)
}

func validateAgentNames(requested, allowed []string) ([]string, error) {
	allowedSet := map[string]bool{}
	for _, agent := range allowed {
		allowedSet[agent] = true
	}
	seen := map[string]bool{}
	var selected []string
	for _, raw := range requested {
		agent := strings.ToLower(strings.TrimSpace(raw))
		if !allowedSet[agent] {
			return nil, fmt.Errorf("agent %q is not available; choose from %s", raw, displayAgentNames(allowed))
		}
		if !seen[agent] {
			selected = append(selected, agent)
			seen[agent] = true
		}
	}
	return selected, nil
}

func displayAgentNames(agents []string) string {
	if len(agents) == 0 {
		return "none"
	}
	return strings.Join(agents, ", ")
}
