package projectinit

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"reflect"
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

func TestSchemaRequiresEveryProviderFieldToDeclareHarnessUse(t *testing.T) {
	schema := ConfigSchema{
		Version:    1,
		Precedence: []string{"cli", "environment", "project", "user", "fallback"},
		Fields: []ConfigField{
			{Key: "PHASE_HARNESS", Flag: "harness", Type: "choice", Choices: []string{"hermes"}},
			{Key: "PHASE_PROVIDER", Flag: "provider", Type: "string"},
		},
	}
	if err := schema.Validate(); err == nil || !strings.Contains(err.Error(), "has no phase harness rule") {
		t.Fatalf("unscoped provider field was accepted: %v", err)
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

func TestResolveConfigUsesDeclaredLiteralDefault(t *testing.T) {
	schema := ConfigSchema{
		Version:    1,
		Precedence: []string{"cli", "environment", "project", "user", "fallback"},
		Fields: []ConfigField{{
			Key: "STRATEGY", Flag: "strategy", Type: "choice",
			Choices: []string{"current", "feature"}, Default: "current",
		}},
	}
	if err := schema.Validate(); err != nil {
		t.Fatal(err)
	}
	values := resolveConfigValues(schema, nil, nil, nil, nil)
	if values["STRATEGY"] != "current" {
		t.Fatalf("literal default = %q, want current", values["STRATEGY"])
	}
}

func TestSchemaValidatesDurationDefaults(t *testing.T) {
	for _, value := range []string{"eventually", "500ms", "1500ms"} {
		t.Run(value, func(t *testing.T) {
			schema := ConfigSchema{
				Version:    1,
				Precedence: []string{"cli", "environment", "project", "user", "fallback"},
				Fields:     []ConfigField{{Key: "TIMEOUT", Flag: "timeout", Type: "duration", Default: value}},
			}
			if err := schema.Validate(); err == nil || !strings.Contains(err.Error(), "whole-second duration") {
				t.Fatalf("duration default %q error = %v", value, err)
			}
		})
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
	prompted, err := completeConfig(schema, values, bufioReader("2\n"), &output, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !prompted["ROLE"] || values["ROLE"] != "consumer" {
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
	prompted, err := completeConfig(schema, values, bufioReader(""), &output, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(prompted) != 0 || output.Len() != 0 {
		t.Fatalf("resolved field prompted: prompted=%v output=%q", prompted, output.String())
	}
}

func TestCompleteConfigDistinguishesPromptAnswersFromFallbacks(t *testing.T) {
	schema := ConfigSchema{
		Version:    1,
		Precedence: []string{"cli", "environment", "project", "user", "fallback"},
		Fields: []ConfigField{
			{Key: "HARNESS", Flag: "harness", Type: "choice", Choices: []string{"codex", "hermes"}, Prompt: "Select harness:"},
			{Key: "AUDIT_HARNESS", Flag: "audit-harness", Type: "choice", Choices: []string{"codex", "hermes"}, Fallback: "HARNESS"},
		},
	}
	values := map[string]string{}
	prompted, err := completeConfig(schema, values, bufioReader("1\n"), io.Discard, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(prompted, map[string]bool{"HARNESS": true}) {
		t.Fatalf("prompted fields = %#v", prompted)
	}
	if values["AUDIT_HARNESS"] != "codex" {
		t.Fatalf("fallback harness = %q", values["AUDIT_HARNESS"])
	}
}

func bufioReader(input string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(input))
}
