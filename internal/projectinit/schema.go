package projectinit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const configSchemaPath = "config/project-init.schema.yaml"

var environmentKeyPattern = regexp.MustCompile(`^SDLC_[A-Z0-9_]+$`)

// ConfigSchema is the single source of truth for initializer questions and
// project/global configuration resolution.
type ConfigSchema struct {
	Version                int           `yaml:"version"`
	Precedence             []string      `yaml:"precedence"`
	RetiredEnvironmentKeys []string      `yaml:"retired_environment_keys"`
	Fields                 []ConfigField `yaml:"fields"`
}

// ConfigField describes one configurable value and its project YAML path.
type ConfigField struct {
	Key              string          `yaml:"key"`
	Path             string          `yaml:"path"`
	Flag             string          `yaml:"flag"`
	Help             string          `yaml:"help"`
	Type             string          `yaml:"type"`
	Choices          []string        `yaml:"choices"`
	ChoicesFrom      string          `yaml:"choices_from"`
	Prompt           string          `yaml:"prompt"`
	Default          string          `yaml:"default"`
	AllowGlobal      bool            `yaml:"allow_global"`
	Required         bool            `yaml:"required"`
	MigrationAliases []string        `yaml:"migration_aliases"`
	When             *FieldCondition `yaml:"when"`
}

// FieldCondition limits a question to selected values of an earlier field.
type FieldCondition struct {
	Field  string   `yaml:"field"`
	Values []string `yaml:"values"`
}

// LoadConfigSchema loads and validates the deployed schema.
func LoadConfigSchema(sdlcRoot string) (ConfigSchema, error) {
	path := filepath.Join(sdlcRoot, filepath.FromSlash(configSchemaPath))
	contents, err := os.ReadFile(path)
	if err != nil {
		return ConfigSchema{}, fmt.Errorf("reading configuration schema %q: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var schema ConfigSchema
	if err := decoder.Decode(&schema); err != nil {
		return ConfigSchema{}, fmt.Errorf("parsing configuration schema %q: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return ConfigSchema{}, errors.New("configuration schema contains multiple YAML documents")
		}
		return ConfigSchema{}, fmt.Errorf("parsing configuration schema %q: %w", path, err)
	}
	if err := schema.Validate(); err != nil {
		return ConfigSchema{}, fmt.Errorf("validating configuration schema %q: %w", path, err)
	}
	return schema, nil
}

// Validate rejects ambiguous schemas before any project mutation.
func (schema ConfigSchema) Validate() error {
	if schema.Version != 3 {
		return fmt.Errorf("unsupported schema version %d", schema.Version)
	}
	want := []string{"cli", "environment", "project", "global", "default"}
	if len(schema.Precedence) != len(want) {
		return errors.New("precedence must be cli, environment, project, global, default")
	}
	for index := range want {
		if schema.Precedence[index] != want[index] {
			return errors.New("precedence must be cli, environment, project, global, default")
		}
	}
	keys := map[string]bool{}
	paths := map[string]bool{}
	flags := map[string]bool{}
	for _, field := range schema.Fields {
		if !environmentKeyPattern.MatchString(field.Key) || keys[field.Key] {
			return fmt.Errorf("invalid or duplicate key %q", field.Key)
		}
		if !validYAMLPath(field.Path) || paths[field.Path] {
			return fmt.Errorf("invalid or duplicate path %q", field.Path)
		}
		if field.Flag == "" || flags[field.Flag] {
			return fmt.Errorf("missing or duplicate flag %q", field.Flag)
		}
		keys[field.Key], paths[field.Path], flags[field.Flag] = true, true, true
		switch field.Type {
		case "string", "string-list", "boolean", "duration":
			if len(field.Choices) != 0 || field.ChoicesFrom != "" {
				return fmt.Errorf("field %s cannot define choices", field.Key)
			}
		case "choice":
			if len(field.Choices) == 0 || field.ChoicesFrom != "" {
				return fmt.Errorf("field %s requires fixed choices", field.Key)
			}
		case "multi-choice":
			if field.ChoicesFrom != "technologies" || len(field.Choices) != 0 {
				return fmt.Errorf("field %s requires technologies choice discovery", field.Key)
			}
		default:
			return fmt.Errorf("field %s has unsupported type %q", field.Key, field.Type)
		}
		for _, alias := range field.MigrationAliases {
			if !environmentKeyPattern.MatchString(alias) {
				return fmt.Errorf("invalid migration alias %q", alias)
			}
		}
		if field.Default != "" {
			if err := field.ValidateValue(field.Default, nil); err != nil {
				return fmt.Errorf("invalid default for %s: %w", field.Key, err)
			}
		}
		if field.When != nil {
			if !keys[field.When.Field] || len(field.When.Values) == 0 {
				return fmt.Errorf("field %s has an invalid condition", field.Key)
			}
		}
	}
	for _, key := range schema.RetiredEnvironmentKeys {
		if !environmentKeyPattern.MatchString(key) || keys[key] {
			return fmt.Errorf("invalid, duplicate, or still-active retired environment key %q", key)
		}
		keys[key] = true
	}
	return nil
}

func validYAMLPath(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" || strings.ContainsAny(part, " \t\r\n") {
			return false
		}
	}
	return true
}

// ValidateValue validates and canonicalizes a field value.
func (field ConfigField) ValidateValue(value string, technologies []Technology) error {
	value = strings.TrimSpace(value)
	if value == "" {
		if field.Required {
			return errors.New("value is required")
		}
		return nil
	}
	switch field.Type {
	case "boolean":
		if value != "true" && value != "false" {
			return errors.New("value must be true or false")
		}
	case "choice":
		for _, choice := range field.Choices {
			if value == choice {
				return nil
			}
		}
		return fmt.Errorf("value must be one of %s", strings.Join(field.Choices, ", "))
	case "multi-choice":
		available := map[string]bool{}
		for _, technology := range technologies {
			available[technology.Name] = true
		}
		for _, selected := range splitCSV(value) {
			if !available[selected] {
				return fmt.Errorf("unknown technology %q", selected)
			}
		}
	case "string-list":
		if len(splitCSV(value)) == 0 {
			return errors.New("value must contain at least one item")
		}
	case "duration":
		duration, err := time.ParseDuration(value)
		if err != nil || duration <= 0 {
			return errors.New("value must be a positive duration such as 4m or 300s")
		}
	}
	return nil
}

// ManagedEnvironmentKeys returns the allowlist used for one-time .env import.
func (schema ConfigSchema) ManagedEnvironmentKeys() []string {
	seen := map[string]bool{}
	var keys []string
	for _, field := range schema.Fields {
		for _, key := range append([]string{field.Key}, field.MigrationAliases...) {
			if !seen[key] {
				keys = append(keys, key)
				seen[key] = true
			}
		}
	}
	for _, key := range schema.RetiredEnvironmentKeys {
		if !seen[key] {
			keys = append(keys, key)
			seen[key] = true
		}
	}
	return keys
}

func fieldApplies(field ConfigField, values map[string]string) bool {
	if field.When == nil {
		return true
	}
	actual := values[field.When.Field]
	for _, value := range field.When.Values {
		if actual == value {
			return true
		}
	}
	return false
}

func splitCSV(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			values = append(values, item)
		}
	}
	return values
}
