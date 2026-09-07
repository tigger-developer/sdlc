package projectinit

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestTechnologyAssessmentUsesSchemaHeuristics(t *testing.T) {
	schema, err := LoadConfigSchema(testSDLCRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	technologies, err := DiscoverTechnologies(filepath.Join(testSDLCRoot(t), "technologies"))
	if err != nil {
		t.Fatal(err)
	}
	assessment, err := detectProjectTechnologies(schema.TechnologyDetection, technologies, map[string]bool{
		"go.mod":                        true,
		"src/main.go":                   true,
		"hugo.toml":                     true,
		"docs/archive/old/component.ts": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if assessment.selection() != "GO,HUGO,WEB" {
		t.Fatalf("selection = %q", assessment.selection())
	}
	if strings.Contains(assessment.Technologies[0].Evidence, "docs/archive") {
		t.Fatalf("archived evidence was used: %#v", assessment)
	}
}

func TestTechnologyAssessmentRejectsRuleWithoutInstalledStandard(t *testing.T) {
	_, err := detectProjectTechnologies(TechnologyDetection{
		Rules: []TechnologyDetectionRule{{Technology: "RUST", Basenames: []string{"Cargo.toml"}}},
	}, []Technology{{Name: "GO"}}, map[string]bool{"Cargo.toml": true})
	if err == nil || !strings.Contains(err.Error(), "has no installed standard") {
		t.Fatalf("validation error = %v", err)
	}
}

func TestTechnologyAssessmentSkipsExplicitSelection(t *testing.T) {
	schema, err := LoadConfigSchema(testSDLCRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	options := defaultOptions(Options{Overrides: map[string]string{"SDLC_TECHNOLOGIES": "GO"}, Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}})
	assessment := assessProjectTechnologies(options, schema, []Technology{{Name: "GO"}}, nil, nil, t.TempDir())
	if assessment != nil {
		t.Fatalf("assessment = %#v", assessment)
	}
}

func TestTechnologyAssessmentUsesSchemaOrder(t *testing.T) {
	assessment, err := detectProjectTechnologies(TechnologyDetection{Rules: []TechnologyDetectionRule{
		{Technology: "WEB", Extensions: []string{".html"}},
		{Technology: "GO", Extensions: []string{".go"}},
	}}, []Technology{{Name: "GO"}, {Name: "WEB"}}, map[string]bool{"index.html": true, "main.go": true})
	if err != nil {
		t.Fatal(err)
	}
	if assessment.selection() != "GO,WEB" {
		t.Fatalf("selection = %q", assessment.selection())
	}
}

func TestLuaSourceSelectsLuaStandard(t *testing.T) {
	schema, err := LoadConfigSchema(testSDLCRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	technologies, err := DiscoverTechnologies(filepath.Join(testSDLCRoot(t), "technologies"))
	if err != nil {
		t.Fatal(err)
	}
	assessment, err := detectProjectTechnologies(schema.TechnologyDetection, technologies, map[string]bool{"filters/normalize.lua": true})
	if err != nil {
		t.Fatal(err)
	}
	if assessment.selection() != "LUA" {
		t.Fatalf("selection = %q", assessment.selection())
	}
}

func TestTechnologyRecommendationUsesExistingSchemaPrompt(t *testing.T) {
	var output bytes.Buffer
	value, explicit, err := promptField(
		bufio.NewReader(strings.NewReader("2\n\n")),
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
	if value != "GO" || !explicit {
		t.Fatalf("selection = %q, explicit = %v", value, explicit)
	}
	want := "\nSelect applicable technologies:\n[x] 1. GO\n[x] 2. WEB\nRecommended: GO,WEB.\nPress Enter to confirm, or enter numbers or names to toggle.\nSelection: \nCurrent selection:\n[x] 1. GO\n[ ] 2. WEB\nPress Enter to confirm, or enter numbers or names to toggle.\nSelection: "
	if output.String() != want {
		t.Fatalf("prompt output = %q, want %q", output.String(), want)
	}
}
