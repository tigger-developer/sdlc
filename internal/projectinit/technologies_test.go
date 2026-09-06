package projectinit

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTechnologyAssessmentUsesAuditModelAndSchemaChoices(t *testing.T) {
	project := t.TempDir()
	writeProjectTestFile(t, filepath.Join(project, "go.mod"), "module example.test/project\n\ngo 1.25\n")
	technologies, err := DiscoverTechnologies(filepath.Join(testSDLCRoot(t), "technologies"))
	if err != nil {
		t.Fatal(err)
	}
	var capturedPrompt string
	options := defaultOptions(Options{
		Output:      &bytes.Buffer{},
		ErrorOutput: &bytes.Buffer{},
		RunCommand: func(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
			if name != "codex" {
				return fmt.Errorf("unexpected command %s", name)
			}
			if argumentValue(arguments, "--model") != "gpt-5.6-luna" || argumentValue(arguments, "--sandbox") != "read-only" || !containsArgument(arguments, "--ephemeral") {
				return fmt.Errorf("unbounded or wrongly configured Codex invocation: %v", arguments)
			}
			contents, readErr := io.ReadAll(input)
			if readErr != nil {
				return readErr
			}
			capturedPrompt = string(contents)
			proposalPath := argumentValue(arguments, "--output-last-message")
			return os.WriteFile(proposalPath, []byte("version: 1\ntechnologies:\n  - name: GO\n    evidence: go.mod declares the maintained Go product.\nwarnings: []\n"), 0o600)
		},
	})
	assessment, err := runTechnologyAssessment(options, testSDLCRoot(t), project, "gpt-5.6-luna", technologies)
	if err != nil {
		t.Fatal(err)
	}
	if assessment.selection() != "GO" {
		t.Fatalf("selection = %q", assessment.selection())
	}
	if !strings.Contains(capturedPrompt, "available_technologies:") || !strings.Contains(capturedPrompt, "- GO") || !strings.Contains(capturedPrompt, "docs/archive") {
		t.Fatalf("bounded prompt is missing choices or exclusions:\n%s", capturedPrompt)
	}
}

func TestTechnologyAssessmentRejectsUnknownOrUnsupportedOutput(t *testing.T) {
	_, err := validateTechnologyAssessment(technologyAssessment{
		Version: 1,
		Technologies: []technologyAssessmentCandidate{
			{Name: "RUST", Evidence: "Cargo.toml exists."},
		},
	}, []Technology{{Name: "GO"}})
	if err == nil || !strings.Contains(err.Error(), "unknown schema choice") {
		t.Fatalf("validation error = %v", err)
	}
}

func TestTechnologyAssessmentSkipsExplicitSelection(t *testing.T) {
	schema, err := LoadConfigSchema(testSDLCRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	options := defaultOptions(Options{
		Overrides:   map[string]string{"SDLC_TECHNOLOGIES": "GO"},
		Output:      &bytes.Buffer{},
		ErrorOutput: &bytes.Buffer{},
		RunCommand: func(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
			t.Fatalf("explicit technology selection invoked %s", name)
			return nil
		},
	})
	assessment := assessProjectTechnologies(options, schema, []Technology{{Name: "GO"}}, nil, nil, testSDLCRoot(t), t.TempDir())
	if assessment != nil {
		t.Fatalf("assessment = %#v", assessment)
	}
}

func TestTechnologyAssessmentUsesConfiguredGlobalAuditModel(t *testing.T) {
	t.Setenv("SDLC_TECHNOLOGIES", "")
	t.Setenv("SDLC_AUDIT_MODEL", "")
	schema, err := LoadConfigSchema(testSDLCRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	global := map[string]any{
		"delivery": map[string]any{
			"audit": map[string]any{"model": "gpt-5.6-luna"},
		},
	}
	options := defaultOptions(Options{
		Output:      &bytes.Buffer{},
		ErrorOutput: &bytes.Buffer{},
		RunCommand: func(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
			if name != "codex" || argumentValue(arguments, "--model") != "gpt-5.6-luna" {
				return fmt.Errorf("audit model was not used: %s %v", name, arguments)
			}
			return os.WriteFile(argumentValue(arguments, "--output-last-message"), []byte("version: 1\ntechnologies: []\nwarnings: []\n"), 0o600)
		},
	})
	assessment := assessProjectTechnologies(options, schema, []Technology{{Name: "GO"}}, global, nil, testSDLCRoot(t), t.TempDir())
	if assessment == nil {
		t.Fatal("configured assessment was not returned")
	}
}

func TestTechnologyAssessmentUsesSchemaOrder(t *testing.T) {
	assessment, err := validateTechnologyAssessment(technologyAssessment{
		Version: 1,
		Technologies: []technologyAssessmentCandidate{
			{Name: "WEB", Evidence: "Maintained browser UI."},
			{Name: "GO", Evidence: "Maintained Go runtime."},
		},
	}, []Technology{{Name: "GO"}, {Name: "WEB"}})
	if err != nil {
		t.Fatal(err)
	}
	if assessment.selection() != "GO,WEB" {
		t.Fatalf("selection = %q", assessment.selection())
	}
}

func TestTechnologyRecommendationUsesExistingSchemaPrompt(t *testing.T) {
	var output bytes.Buffer
	value, explicit, err := promptField(
		bufio.NewReader(strings.NewReader("\n")),
		&output,
		ConfigField{
			Key:         "SDLC_TECHNOLOGIES",
			Type:        "multi-choice",
			ChoicesFrom: "technologies",
			Prompt:      "Select applicable technologies:",
		},
		"GO,WEB",
		"assessment",
		[]Technology{{Name: "GO"}, {Name: "WEB"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if value != "GO,WEB" || explicit {
		t.Fatalf("selection = %q, explicit = %v", value, explicit)
	}
	want := "\nSelect applicable technologies:\n[x] 1. GO\n[x] 2. WEB\nRecommended: GO,WEB. Press Enter to accept, or enter a project value.\nSelection: "
	if output.String() != want {
		t.Fatalf("prompt output = %q, want %q", output.String(), want)
	}
}
