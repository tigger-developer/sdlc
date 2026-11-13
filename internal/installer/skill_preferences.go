// ABOUTME: Projects global phase model preferences into installed skill metadata.
// ABOUTME: Preferences are advisory and refresh only when the installer runs.
package installer

import (
	"bytes"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type skillModelPreference struct {
	Harness         string `yaml:"harness"`
	Provider        string `yaml:"provider"`
	Model           string `yaml:"model"`
	ReasoningEffort string `yaml:"reasoning_effort"`
}

func (plan *installationPlan) applySkillPreferences(config []byte) error {
	var selected struct {
		Delivery struct {
			Definition skillModelPreference `yaml:"definition"`
			TestCode   skillModelPreference `yaml:"test_code"`
			Build      skillModelPreference `yaml:"build"`
			Audit      skillModelPreference `yaml:"audit"`
		} `yaml:"delivery"`
	}
	if err := yaml.Unmarshal(config, &selected); err != nil {
		return fmt.Errorf("reading installed skill model preferences: %w", err)
	}
	for _, item := range []struct {
		name       string
		preference skillModelPreference
		testCode   *skillModelPreference
	}{
		{"define-change", selected.Delivery.Definition, nil},
		{"define-project-foundation", selected.Delivery.Definition, nil},
		{"deliver-change", selected.Delivery.Build, &selected.Delivery.TestCode},
		{"pair-change", selected.Delivery.Build, nil},
		{"emergency-change", selected.Delivery.Build, nil},
		{"hands-off-spike", selected.Delivery.Build, nil},
		{"definition-audit", selected.Delivery.Audit, nil},
		{"test-code-audit", selected.Delivery.Audit, nil},
		{"code-audit", selected.Delivery.Audit, nil},
	} {
		path := filepath.Join("skills", item.name, "SKILL.md")
		file, present := plan.tree[path]
		if !present {
			continue
		}
		updated, err := skillWithPreferences(file.Data, item.preference, item.testCode)
		if err != nil {
			return fmt.Errorf("projecting skill model preferences for %s: %w", item.name, err)
		}
		file.Data = updated
		plan.tree[path] = file
	}
	return nil
}

func skillWithPreferences(content []byte, primary skillModelPreference, testCode *skillModelPreference) ([]byte, error) {
	if primary == (skillModelPreference{}) && (testCode == nil || *testCode == (skillModelPreference{})) {
		return bytes.Clone(content), nil
	}
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, fmt.Errorf("skill has no YAML frontmatter")
	}
	end := bytes.Index(content[4:], []byte("\n---\n"))
	if end < 0 {
		return nil, fmt.Errorf("skill frontmatter is not closed")
	}
	end += 4
	var document yaml.Node
	if err := yaml.Unmarshal(content[4:end], &document); err != nil {
		return nil, err
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("skill frontmatter must be a mapping")
	}
	root := document.Content[0]
	metadata := mappingChild(root, "metadata")
	if metadata == nil {
		metadata = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "metadata"}, metadata)
	} else if metadata.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("skill metadata must be a mapping")
	}
	setPreference(metadata, "preferred_", primary)
	if testCode != nil {
		setPreference(metadata, "preferred_test_code_", *testCode)
	}
	frontmatter, err := yaml.Marshal(root)
	if err != nil {
		return nil, err
	}
	result := append([]byte("---\n"), frontmatter...)
	result = append(result, content[end+1:]...)
	return result, nil
}

func mappingChild(root *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == key {
			return root.Content[i+1]
		}
	}
	return nil
}

func setPreference(root *yaml.Node, prefix string, value skillModelPreference) {
	for _, field := range []struct{ name, value string }{
		{"harness", value.Harness},
		{"provider", value.Provider},
		{"model", value.Model},
		{"reasoning_effort", value.ReasoningEffort},
	} {
		key := prefix + field.name
		for i := 0; i+1 < len(root.Content); i += 2 {
			if root.Content[i].Value == key {
				root.Content = append(root.Content[:i], root.Content[i+2:]...)
				break
			}
		}
		if field.value != "" {
			root.Content = append(root.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
				&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: field.value})
		}
	}
}
