package installer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func analyseHermesRetirement(agentHome string, output io.Writer) (*configurationChange, error) {
	path := filepath.Join(agentHome, "config.yaml")
	current, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading Hermes configuration %s: %w", path, err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("Hermes configuration %s is not a regular file", path)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(current, &document); err != nil {
		return nil, fmt.Errorf("parsing Hermes configuration %s: %w", path, err)
	}
	root, err := hermesDocumentMapping(&document)
	if err != nil {
		return nil, err
	}
	changed, err := removeHermesCommandGuard(root)
	if err != nil {
		return nil, err
	}
	if !changed {
		return nil, nil
	}
	var candidate bytes.Buffer
	encoder := yaml.NewEncoder(&candidate)
	encoder.SetIndent(2)
	if err := encoder.Encode(&document); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	fmt.Fprintf(output, "Recommendation: retire the SDLC v1/v2 tool guard in %s.\n", path)
	return &configurationChange{
		path: path, beforeLabel: "existing configuration; unrelated values preserved",
		afterLabel: "SDLC v1/v2 tool guard removed", contents: candidate.Bytes(), mode: info.Mode().Perm(),
	}, nil
}

func removeHermesCommandGuard(root *yaml.Node) (bool, error) {
	hooks, exists := hermesMappingValue(root, "hooks")
	if !exists {
		return false, nil
	}
	if hooks.Kind != yaml.MappingNode {
		return false, fmt.Errorf("hooks is %s, not a mapping", hermesYAMLKind(hooks))
	}
	entries, exists := hermesMappingValue(hooks, "pre_tool_call")
	if !exists {
		return false, nil
	}
	if entries.Kind != yaml.SequenceNode {
		return false, fmt.Errorf("hooks.pre_tool_call is %s, not a list", hermesYAMLKind(entries))
	}
	filtered := make([]*yaml.Node, 0, len(entries.Content))
	changed := false
	for _, entry := range entries.Content {
		command, _ := hermesScalarValue(entry, "command")
		if isManagedGuardCommand(command) {
			changed = true
			continue
		}
		filtered = append(filtered, entry)
	}
	if !changed {
		return false, nil
	}
	if len(filtered) == 0 {
		hermesDeleteMappingValue(hooks, "pre_tool_call")
	} else {
		entries.Content = filtered
	}
	if len(hooks.Content) == 0 {
		hermesDeleteMappingValue(root, "hooks")
	}
	return true, nil
}

func hermesDeleteMappingValue(mapping *yaml.Node, key string) {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			mapping.Content = append(mapping.Content[:index], mapping.Content[index+2:]...)
			return
		}
	}
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

func hermesMappingValue(mapping *yaml.Node, key string) (*yaml.Node, bool) {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1], true
		}
	}
	return nil, false
}

func hermesScalarValue(mapping *yaml.Node, key string) (string, bool) {
	value, exists := hermesMappingValue(mapping, key)
	if !exists || value.Kind != yaml.ScalarNode || value.Tag == "!!null" {
		return "", false
	}
	return value.Value, true
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
