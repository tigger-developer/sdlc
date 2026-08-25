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
	hookCommand := fmt.Sprintf("bash %q", filepath.Join(filepath.Dir(agentHome), ".agents", "sdlc", "hooks", "agent-command-guard.sh"))
	legacyHookCommand := fmt.Sprintf("bash %q", filepath.Join(agentHome, "sdlc", "hooks", "agent-command-guard.sh"))
	desired, changed, err := mergeHermesCommandGuardConfiguration(current, hookCommand, legacyHookCommand)
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

func mergeHermesCommandGuardConfiguration(current []byte, hookCommand string, obsoleteHookCommands ...string) ([]byte, bool, error) {
	var document yaml.Node
	if len(bytes.TrimSpace(current)) != 0 {
		if err := yaml.Unmarshal(current, &document); err != nil {
			return nil, false, fmt.Errorf("invalid YAML: %w", err)
		}
	} else {
		document = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
	}
	root, err := hermesDocumentMapping(&document)
	if err != nil {
		return nil, false, err
	}
	changed, err := mergeHermesCommandGuard(root, hookCommand, obsoleteHookCommands...)
	if err != nil {
		return nil, false, err
	}
	if !changed {
		return current, false, nil
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(&document); err != nil {
		return nil, false, err
	}
	if err := encoder.Close(); err != nil {
		return nil, false, err
	}
	desired := output.Bytes()
	return desired, true, nil
}

func hermesDocumentMapping(document *yaml.Node) (*yaml.Node, error) {
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 {
		return nil, fmt.Errorf("configuration root is not a single YAML document")
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("configuration root is %s, not a mapping", hermesYAMLKind(root))
	}
	return root, nil
}

func mergeHermesCommandGuard(root *yaml.Node, hookCommand string, obsoleteHookCommands ...string) (bool, error) {
	hooks, exists := hermesMappingValue(root, "hooks")
	if !exists || hooks.Tag == "!!null" {
		hooks = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		hermesSetMappingValue(root, "hooks", hooks)
	} else if hooks.Kind != yaml.MappingNode {
		return false, fmt.Errorf("hooks is %s, not a mapping", hermesYAMLKind(hooks))
	}
	entries, exists := hermesMappingValue(hooks, "pre_tool_call")
	if !exists || entries.Tag == "!!null" {
		entries = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		hermesSetMappingValue(hooks, "pre_tool_call", entries)
	} else if entries.Kind != yaml.SequenceNode {
		return false, fmt.Errorf("hooks.pre_tool_call is %s, not a list", hermesYAMLKind(entries))
	}

	managedCount := 0
	compliant := false
	for _, entry := range entries.Content {
		if entry.Kind != yaml.MappingNode {
			return false, fmt.Errorf("hooks.pre_tool_call contains %s, not a mapping", hermesYAMLKind(entry))
		}
		command, _ := hermesScalarValue(entry, "command")
		if command != hookCommand && !containsString(obsoleteHookCommands, command) {
			continue
		}
		managedCount++
		matcher, matcherExists := hermesScalarValue(entry, "matcher")
		timeout, timeoutExists := hermesIntegerValue(entry, "timeout")
		compliant = managedCount == 1 && command == hookCommand && matcherExists && matcher == "terminal" && timeoutExists && timeout == 5
	}
	if managedCount == 1 && compliant {
		return false, nil
	}

	normalized := make([]*yaml.Node, 0, len(entries.Content)+1)
	managedEntryAdded := false
	for _, entry := range entries.Content {
		command, _ := hermesScalarValue(entry, "command")
		managed := command == hookCommand || containsString(obsoleteHookCommands, command)
		if !managed {
			normalized = append(normalized, entry)
			continue
		}
		if !managedEntryAdded {
			hermesSetScalar(entry, "command", "!!str", hookCommand)
			hermesSetScalar(entry, "matcher", "!!str", "terminal")
			hermesSetScalar(entry, "timeout", "!!int", "5")
			normalized = append(normalized, entry)
			managedEntryAdded = true
		}
	}
	if !managedEntryAdded {
		entry := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		hermesSetScalar(entry, "matcher", "!!str", "terminal")
		hermesSetScalar(entry, "command", "!!str", hookCommand)
		hermesSetScalar(entry, "timeout", "!!int", "5")
		normalized = append(normalized, entry)
	}
	entries.Content = normalized
	return true, nil
}

func hermesMappingValue(mapping *yaml.Node, key string) (*yaml.Node, bool) {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1], true
		}
	}
	return nil, false
}

func hermesSetMappingValue(mapping *yaml.Node, key string, value *yaml.Node) {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			mapping.Content[index+1] = value
			return
		}
	}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		value,
	)
}

func hermesScalarValue(mapping *yaml.Node, key string) (string, bool) {
	value, exists := hermesMappingValue(mapping, key)
	if !exists || value.Kind != yaml.ScalarNode || value.Tag == "!!null" {
		return "", false
	}
	return value.Value, true
}

func hermesIntegerValue(mapping *yaml.Node, key string) (int, bool) {
	value, exists := hermesMappingValue(mapping, key)
	if !exists || value.Kind != yaml.ScalarNode {
		return 0, false
	}
	var result int
	if err := value.Decode(&result); err != nil {
		return 0, false
	}
	return result, true
}

func hermesSetScalar(mapping *yaml.Node, key, tag, value string) {
	existing, exists := hermesMappingValue(mapping, key)
	if exists {
		existing.Kind = yaml.ScalarNode
		existing.Tag = tag
		existing.Value = value
		return
	}
	hermesSetMappingValue(mapping, key, &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value})
}

func hermesYAMLKind(node *yaml.Node) string {
	switch node.Kind {
	case yaml.MappingNode:
		return "a mapping"
	case yaml.SequenceNode:
		return "a list"
	case yaml.ScalarNode:
		return "a scalar"
	case yaml.AliasNode:
		return "an alias"
	default:
		return "an invalid value"
	}
}
