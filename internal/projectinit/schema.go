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

var environmentKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// ConfigSchema defines project configuration fields and their resolution.
type ConfigSchema struct {
	Version    int           `yaml:"version"`
	Precedence []string      `yaml:"precedence"`
	Fields     []ConfigField `yaml:"fields"`
	Phases     []PhaseConfig `yaml:"phases"`
}

// ConfigField defines one project configuration value.
type ConfigField struct {
	Key              string        `yaml:"key"`
	Flag             string        `yaml:"flag"`
	Help             string        `yaml:"help"`
	Type             string        `yaml:"type"`
	Choices          []string      `yaml:"choices"`
	ChoicesFrom      string        `yaml:"choices_from"`
	Prompt           string        `yaml:"prompt"`
	Fallback         string        `yaml:"fallback"`
	Default          string        `yaml:"default"`
	RequiredWhen     *RequiredWhen `yaml:"required_when"`
	AllowUserDefault bool          `yaml:"allow_user_default"`
	Persist          bool          `yaml:"persist"`
}

// RequiredWhen makes a field mandatory for selected values of another field.
type RequiredWhen struct {
	Field  string   `yaml:"field"`
	Values []string `yaml:"values"`
}

// PhaseConfig relates one phase's harness, provider, and model fields.
type PhaseConfig struct {
	Name              string   `yaml:"name"`
	Harness           string   `yaml:"harness"`
	Provider          string   `yaml:"provider"`
	Model             string   `yaml:"model"`
	ProviderHarnesses []string `yaml:"provider_harnesses"`
}

// LoadConfigSchema loads and strictly validates the deployed YAML schema.
func LoadConfigSchema(sdlcRoot string) (ConfigSchema, error) {
	path := filepath.Join(sdlcRoot, filepath.FromSlash(configSchemaPath))
	contents, err := os.ReadFile(path)
	if err != nil {
		return ConfigSchema{}, fmt.Errorf("reading project configuration schema %q: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var schema ConfigSchema
	if err := decoder.Decode(&schema); err != nil {
		return ConfigSchema{}, fmt.Errorf("parsing project configuration schema %q: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return ConfigSchema{}, fmt.Errorf("parsing project configuration schema %q: multiple YAML documents are not supported", path)
		}
		return ConfigSchema{}, fmt.Errorf("parsing project configuration schema %q: %w", path, err)
	}
	if err := schema.Validate(); err != nil {
		return ConfigSchema{}, fmt.Errorf("validating project configuration schema %q: %w", path, err)
	}
	return schema, nil
}

// Validate rejects ambiguous or internally inconsistent schemas.
func (schema ConfigSchema) Validate() error {
	if schema.Version != 1 {
		return fmt.Errorf("unsupported version %d", schema.Version)
	}
	wantOrder := []string{"cli", "environment", "project", "user", "fallback"}
	wantPrecedence := map[string]bool{"cli": true, "environment": true, "project": true, "user": true, "fallback": true}
	seenPrecedence := map[string]bool{}
	for _, source := range schema.Precedence {
		if !wantPrecedence[source] {
			return fmt.Errorf("unknown precedence source %q", source)
		}
		if seenPrecedence[source] {
			return fmt.Errorf("duplicate precedence source %q", source)
		}
		seenPrecedence[source] = true
	}
	if len(seenPrecedence) != len(wantPrecedence) {
		return errors.New("precedence must contain cli, environment, project, user, and fallback exactly once")
	}
	for index, source := range wantOrder {
		if schema.Precedence[index] != source {
			return errors.New("precedence must be cli, environment, project, user, then fallback")
		}
	}

	byKey := map[string]ConfigField{}
	seenFlags := map[string]bool{}
	for _, field := range schema.Fields {
		if !environmentKeyPattern.MatchString(field.Key) || byKey[field.Key].Key != "" {
			return fmt.Errorf("missing or duplicate field key %q", field.Key)
		}
		byKey[field.Key] = field
		if field.Flag == "" || seenFlags[field.Flag] {
			return fmt.Errorf("missing or duplicate flag %q", field.Flag)
		}
		seenFlags[field.Flag] = true
		switch field.Type {
		case "string", "duration":
			if len(field.Choices) != 0 || field.ChoicesFrom != "" {
				return fmt.Errorf("%s field %s cannot define choices", field.Type, field.Key)
			}
		case "choice":
			if len(field.Choices) == 0 || field.ChoicesFrom != "" {
				return fmt.Errorf("choice field %s requires fixed choices", field.Key)
			}
		case "multi-choice":
			if (len(field.Choices) == 0) == (field.ChoicesFrom == "") {
				return fmt.Errorf("multi-choice field %s requires exactly one choice source", field.Key)
			}
		default:
			return fmt.Errorf("field %s has unsupported type %q", field.Key, field.Type)
		}
		if field.ChoicesFrom != "" && field.ChoicesFrom != "technologies" {
			return fmt.Errorf("field %s has unsupported choice source %q", field.Key, field.ChoicesFrom)
		}
		seenChoices := map[string]bool{}
		for _, choice := range field.Choices {
			key := strings.ToLower(strings.TrimSpace(choice))
			if key == "" || seenChoices[key] {
				return fmt.Errorf("field %s has an empty or duplicate choice %q", field.Key, choice)
			}
			seenChoices[key] = true
		}
	}
	for _, field := range schema.Fields {
		if field.Fallback != "" && field.Default != "" {
			return fmt.Errorf("field %s cannot define both fallback and default", field.Key)
		}
		if field.Fallback != "" {
			if _, ok := byKey[field.Fallback]; !ok {
				return fmt.Errorf("field %s has unknown fallback %s", field.Key, field.Fallback)
			}
			seen := map[string]bool{field.Key: true}
			for key := field.Fallback; key != ""; key = byKey[key].Fallback {
				if seen[key] {
					return fmt.Errorf("fallback cycle includes %s", key)
				}
				seen[key] = true
			}
		}
		if field.Default != "" && field.Type == "choice" {
			if _, ok := canonicalChoice(field, field.Default); !ok {
				return fmt.Errorf("field %s has unknown default %q", field.Key, field.Default)
			}
		}
		if field.Default != "" && field.Type == "duration" {
			if err := validateWholeSecondDuration(field.Key, field.Default); err != nil {
				return err
			}
		}
		if field.RequiredWhen != nil {
			conditionField, ok := byKey[field.RequiredWhen.Field]
			if !ok {
				return fmt.Errorf("field %s has unknown condition field %s", field.Key, field.RequiredWhen.Field)
			}
			if len(field.RequiredWhen.Values) == 0 {
				return fmt.Errorf("field %s has an empty required_when value list", field.Key)
			}
			for _, value := range field.RequiredWhen.Values {
				if _, ok := canonicalChoice(conditionField, value); !ok {
					return fmt.Errorf("field %s condition has unknown %s value %q", field.Key, field.RequiredWhen.Field, value)
				}
			}
		}
	}
	seenPhases := map[string]bool{}
	seenProviderFields := map[string]bool{}
	for _, phase := range schema.Phases {
		if phase.Name == "" || seenPhases[phase.Name] {
			return fmt.Errorf("missing or duplicate phase name %q", phase.Name)
		}
		seenPhases[phase.Name] = true
		harness, harnessOK := byKey[phase.Harness]
		provider, providerOK := byKey[phase.Provider]
		model, modelOK := byKey[phase.Model]
		if !harnessOK || harness.Type != "choice" {
			return fmt.Errorf("phase %s has unknown or non-choice harness field %s", phase.Name, phase.Harness)
		}
		if !providerOK || provider.Type != "string" {
			return fmt.Errorf("phase %s has unknown or non-string provider field %s", phase.Name, phase.Provider)
		}
		if !modelOK || model.Type != "string" {
			return fmt.Errorf("phase %s has unknown or non-string model field %s", phase.Name, phase.Model)
		}
		if seenProviderFields[phase.Provider] {
			return fmt.Errorf("provider field %s belongs to more than one phase", phase.Provider)
		}
		seenProviderFields[phase.Provider] = true
		if len(phase.ProviderHarnesses) == 0 {
			return fmt.Errorf("phase %s must declare provider_harnesses", phase.Name)
		}
		seenHarnesses := map[string]bool{}
		for _, value := range phase.ProviderHarnesses {
			canonical, ok := canonicalChoice(harness, value)
			if !ok || seenHarnesses[canonical] {
				return fmt.Errorf("phase %s has unknown or duplicate provider harness %q", phase.Name, value)
			}
			seenHarnesses[canonical] = true
		}
	}
	for key := range byKey {
		if strings.HasSuffix(key, "_PROVIDER") && !seenProviderFields[key] {
			return fmt.Errorf("provider field %s has no phase harness rule", key)
		}
	}
	return nil
}

// ManagedKeys returns schema fields in stable persistence order.
func (schema ConfigSchema) ManagedKeys() []string {
	keys := make([]string, 0, len(schema.Fields))
	for _, field := range schema.Fields {
		keys = append(keys, field.Key)
	}
	return keys
}

func canonicalChoice(field ConfigField, value string) (string, bool) {
	for _, choice := range field.Choices {
		if strings.EqualFold(strings.TrimSpace(value), choice) {
			return choice, true
		}
	}
	return "", false
}

func validateWholeSecondDuration(key, value string) error {
	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || duration < time.Second || duration%time.Second != 0 {
		return fmt.Errorf("%s must be a whole-second duration of at least one second, got %q", key, value)
	}
	return nil
}
