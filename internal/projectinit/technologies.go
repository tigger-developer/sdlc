package projectinit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const technologyAssessmentPromptPath = "prompts/discover-project-technologies.md"

type technologyAssessment struct {
	Version      int                             `yaml:"version"`
	Technologies []technologyAssessmentCandidate `yaml:"technologies"`
	Warnings     []string                        `yaml:"warnings"`
}

type technologyAssessmentCandidate struct {
	Name     string `yaml:"name"`
	Evidence string `yaml:"evidence"`
}

func assessProjectTechnologies(options Options, schema ConfigSchema, technologies []Technology, global map[string]any, legacy map[string]string, sdlcRoot, projectRoot string) *technologyAssessment {
	technologyField, ok := schemaField(schema, "SDLC_TECHNOLOGIES")
	if !ok {
		fmt.Fprintln(options.ErrorOutput, "Warning: technology assessment skipped because SDLC_TECHNOLOGIES is absent from the configuration schema.")
		return nil
	}
	if value, explicit, _ := initialValue(technologyField, options.Overrides, legacy, global); explicit && value != "" {
		return nil
	}

	model := ""
	if auditModelField, found := schemaField(schema, "SDLC_AUDIT_MODEL"); found {
		model, _, _ = initialValue(auditModelField, options.Overrides, legacy, global)
	}
	assessment, err := runTechnologyAssessment(options, sdlcRoot, projectRoot, model, technologies)
	if err != nil {
		fmt.Fprintf(options.ErrorOutput, "Warning: automatic technology assessment was unavailable: %v\nSelect the applicable technologies manually.\n", err)
		return nil
	}
	return &assessment
}

func schemaField(schema ConfigSchema, key string) (ConfigField, bool) {
	for _, field := range schema.Fields {
		if field.Key == key {
			return field, true
		}
	}
	return ConfigField{}, false
}

func runTechnologyAssessment(options Options, sdlcRoot, projectRoot, model string, technologies []Technology) (technologyAssessment, error) {
	promptPath := filepath.Join(sdlcRoot, filepath.FromSlash(technologyAssessmentPromptPath))
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		return technologyAssessment{}, fmt.Errorf("reading technology-assessment prompt %s: %w", promptPath, err)
	}
	available := make([]string, 0, len(technologies))
	for _, technology := range technologies {
		available = append(available, technology.Name)
	}
	choiceDocument, err := yaml.Marshal(map[string][]string{"available_technologies": available})
	if err != nil {
		return technologyAssessment{}, fmt.Errorf("rendering available technologies: %w", err)
	}
	prompt = append(prompt, []byte("\n\n## Available schema choices\n\nReturn only names from this list:\n\n```yaml\n")...)
	prompt = append(prompt, choiceDocument...)
	prompt = append(prompt, []byte("```\n")...)

	workspace, err := os.MkdirTemp("", "sdlc-init-technologies-")
	if err != nil {
		return technologyAssessment{}, fmt.Errorf("creating temporary technology-assessment workspace: %w", err)
	}
	defer func() { _ = os.RemoveAll(workspace) }()
	proposalPath := filepath.Join(workspace, "technology-assessment.yaml")
	arguments := []string{"exec", "--ephemeral", "--sandbox", "read-only", "--output-last-message", proposalPath}
	if strings.TrimSpace(model) != "" {
		arguments = append(arguments, "--model", model)
	}
	arguments = append(arguments, "-")
	if err := options.RunCommand("codex", arguments, projectRoot, bytes.NewReader(prompt), options.Output, options.ErrorOutput); err != nil {
		return technologyAssessment{}, fmt.Errorf("assessing the project technology stack with headless Codex: %w", err)
	}
	assessment, err := readTechnologyAssessment(proposalPath)
	if err != nil {
		return technologyAssessment{}, err
	}
	return validateTechnologyAssessment(assessment, technologies)
}

func readTechnologyAssessment(path string) (technologyAssessment, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return technologyAssessment{}, fmt.Errorf("reading technology assessment %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var assessment technologyAssessment
	if err := decoder.Decode(&assessment); err != nil {
		return technologyAssessment{}, fmt.Errorf("parsing technology assessment %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return technologyAssessment{}, fmt.Errorf("technology assessment %s contains multiple YAML documents", path)
		}
		return technologyAssessment{}, fmt.Errorf("parsing technology assessment %s: %w", path, err)
	}
	if assessment.Version != 1 {
		return technologyAssessment{}, fmt.Errorf("technology assessment %s must declare version: 1", path)
	}
	return assessment, nil
}

func validateTechnologyAssessment(assessment technologyAssessment, technologies []Technology) (technologyAssessment, error) {
	available := make(map[string]bool, len(technologies))
	for _, technology := range technologies {
		available[technology.Name] = true
	}
	seen := map[string]bool{}
	byName := make(map[string]technologyAssessmentCandidate, len(assessment.Technologies))
	for _, candidate := range assessment.Technologies {
		candidate.Name = strings.TrimSpace(candidate.Name)
		candidate.Evidence = singleLine(candidate.Evidence)
		if !available[candidate.Name] {
			return technologyAssessment{}, fmt.Errorf("technology assessment selected unknown schema choice %q", candidate.Name)
		}
		if seen[candidate.Name] {
			return technologyAssessment{}, fmt.Errorf("technology assessment selected %s more than once", candidate.Name)
		}
		if candidate.Evidence == "" {
			return technologyAssessment{}, fmt.Errorf("technology assessment selected %s without evidence", candidate.Name)
		}
		seen[candidate.Name] = true
		byName[candidate.Name] = candidate
	}
	assessment.Technologies = nil
	for _, technology := range technologies {
		if candidate, selected := byName[technology.Name]; selected {
			assessment.Technologies = append(assessment.Technologies, candidate)
		}
	}
	for index, warning := range assessment.Warnings {
		assessment.Warnings[index] = singleLine(warning)
	}
	return assessment, nil
}

func (assessment technologyAssessment) selection() string {
	selected := make([]string, 0, len(assessment.Technologies))
	for _, candidate := range assessment.Technologies {
		selected = append(selected, candidate.Name)
	}
	return strings.Join(selected, ",")
}

func renderTechnologyAssessment(output io.Writer, assessment technologyAssessment) {
	fmt.Fprintln(output, "\nProposed applicable technologies:")
	if len(assessment.Technologies) == 0 {
		fmt.Fprintln(output, "- No technology standard was identified with sufficient confidence.")
	}
	for _, candidate := range assessment.Technologies {
		fmt.Fprintf(output, "- %s: %s\n", candidate.Name, candidate.Evidence)
	}
	for _, warning := range assessment.Warnings {
		if warning != "" {
			fmt.Fprintf(output, "Warning: %s\n", warning)
		}
	}
}
