package projectinit

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigSchemaIsStrictAndStable(t *testing.T) {
	schema := loadTestSchema(t)
	if len(schema.Fields) == 0 || schema.Fields[0].Key != keyAgentHarness {
		t.Fatalf("unexpected schema order: %#v", schema.Fields)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "config", "project-init.schema.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("version: 1\nunknown: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfigSchema(directory); err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("unknown schema field was not rejected: %v", err)
	}
}

func TestSchemaRejectsFallbackCycles(t *testing.T) {
	schema := ConfigSchema{
		Version:    1,
		Precedence: []string{"cli", "environment", "project", "user", "fallback"},
		Fields: []ConfigField{
			{Key: "A", Flag: "a", Type: "string", Fallback: "B"},
			{Key: "B", Flag: "b", Type: "string", Fallback: "A"},
		},
	}
	if err := schema.Validate(); err == nil || !strings.Contains(err.Error(), "fallback cycle") {
		t.Fatalf("fallback cycle was not rejected: %v", err)
	}
}

func TestResolveConfigUsesDeclaredPrecedenceAndFallback(t *testing.T) {
	schema := loadTestSchema(t)
	values := resolveConfigValues(schema,
		map[string]string{keySpecModel: "cli-spec"},
		map[string]string{keyAgentHarness: "hermes", keySpecModel: "user-spec", keyProjectType: "brownfield"},
		map[string]string{keyAgentHarness: "claude", keySpecModel: "project-spec"},
		map[string]string{keyAgentHarness: "codex", keySpecModel: "environment-spec"},
	)
	if values[keySpecModel] != "cli-spec" || values[keyAgentHarness] != "codex" {
		t.Fatalf("precedence result = %#v", values)
	}
	if values[keySpecHarness] != "codex" || values[keyBuildHarness] != "codex" || values[keyAuditHarness] != "codex" {
		t.Fatalf("harness fallbacks = %#v", values)
	}
	if values[keyProjectType] != "" {
		t.Fatalf("user-level project type became a default: %#v", values)
	}
}

func TestCompleteConfigUsesNumberedChoicePrompts(t *testing.T) {
	schema := ConfigSchema{
		Version:    1,
		Precedence: []string{"cli", "environment", "project", "user", "fallback"},
		Fields:     []ConfigField{{Key: "ROLE", Flag: "role", Type: "choice", Choices: []string{"none", "consumer", "provider"}, Prompt: "Select role:"}},
	}
	values := map[string]string{}
	var output strings.Builder
	changed, err := completeConfig(schema, values, bufioReader("2\n"), &output, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || values["ROLE"] != "consumer" {
		t.Fatalf("choice result = %#v", values)
	}
	for _, expected := range []string{"Select role:", "1. none", "2. consumer", "3. provider"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("prompt omitted %q: %q", expected, output.String())
		}
	}
}

func TestCompleteConfigSkipsResolvedFields(t *testing.T) {
	schema := ConfigSchema{
		Version:    1,
		Precedence: []string{"cli", "environment", "project", "user", "fallback"},
		Fields:     []ConfigField{{Key: "ROLE", Flag: "role", Type: "choice", Choices: []string{"none", "consumer", "provider"}, Prompt: "Select role:"}},
	}
	values := map[string]string{"ROLE": "provider"}
	var output strings.Builder
	changed, err := completeConfig(schema, values, bufioReader(""), &output, nil)
	if err != nil {
		t.Fatal(err)
	}
	if changed || output.Len() != 0 {
		t.Fatalf("resolved field prompted or changed: changed=%t output=%q", changed, output.String())
	}
}

func bufioReader(input string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(input))
}
