package installer

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var semanticReleasePattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$`)

func planGlobalReleaseConfiguration(commonHome, release string) (*configurationChange, error) {
	if release == "" {
		return nil, nil
	}
	if !semanticReleasePattern.MatchString(release) {
		return nil, fmt.Errorf("SDLC release %q is not a semantic-version tag", release)
	}

	path := filepath.Join(commonHome, "sdlc.yaml")
	current, mode, exists, err := readOptionalRegularFile(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		mode = 0o600
	}

	document, root, err := globalConfigurationDocument(path, current)
	if err != nil {
		return nil, err
	}
	changed, err := setGlobalConfigurationRelease(root, release)
	if err != nil {
		return nil, fmt.Errorf("validating global SDLC configuration %q: %w", path, err)
	}
	if !changed {
		return nil, nil
	}

	var candidate bytes.Buffer
	encoder := yaml.NewEncoder(&candidate)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		return nil, fmt.Errorf("rendering global SDLC configuration %q: %w", path, err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("closing global SDLC configuration encoder: %w", err)
	}
	return &configurationChange{
		path:        path,
		beforeLabel: "existing global defaults; unrelated values preserved",
		afterLabel:  fmt.Sprintf("deployed SDLC release set to %s", release),
		contents:    candidate.Bytes(),
		mode:        mode,
	}, nil
}

func globalConfigurationDocument(path string, current []byte) (*yaml.Node, *yaml.Node, error) {
	if len(bytes.TrimSpace(current)) == 0 {
		root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		return &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}, root, nil
	}
	var document yaml.Node
	if err := yaml.Unmarshal(current, &document); err != nil {
		return nil, nil, fmt.Errorf("parsing global SDLC configuration %q: %w", path, err)
	}
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("global SDLC configuration %q must contain one YAML mapping", path)
	}
	return &document, document.Content[0], nil
}

func setGlobalConfigurationRelease(root *yaml.Node, release string) (bool, error) {
	version, versionIndex := releaseMappingValue(root, "version")
	changed := false
	if version == nil {
		version = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: "3"}
		root.Content = append([]*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "version"},
			version,
		}, root.Content...)
		versionIndex = 0
		changed = true
	} else if version.Kind != yaml.ScalarNode || version.Tag != "!!int" || version.Value != "3" {
		return false, errorsForGlobalVersion(version)
	}

	currentRelease, _ := releaseMappingValue(root, "release")
	if currentRelease != nil {
		if currentRelease.Kind != yaml.ScalarNode || currentRelease.Tag == "!!null" {
			return false, fmt.Errorf("release must be a scalar semantic-version tag")
		}
		if currentRelease.Value == release && !changed {
			return false, nil
		}
		currentRelease.Tag = "!!str"
		currentRelease.Value = release
		return true, nil
	}

	releasePair := []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "release"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: release},
	}
	insertAt := versionIndex + 2
	root.Content = append(root.Content, nil, nil)
	copy(root.Content[insertAt+2:], root.Content[insertAt:])
	copy(root.Content[insertAt:], releasePair)
	return true, nil
}

func releaseMappingValue(mapping *yaml.Node, key string) (*yaml.Node, int) {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1], index
		}
	}
	return nil, -1
}

func errorsForGlobalVersion(version *yaml.Node) error {
	value := strings.TrimSpace(version.Value)
	if value == "" {
		value = "<empty>"
	}
	return fmt.Errorf("version must be the integer 3, got %s", value)
}
