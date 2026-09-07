package projectinit

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

type technologyAssessment struct {
	Technologies []technologyAssessmentCandidate
	Warnings     []string
}

type technologyAssessmentCandidate struct {
	Name     string
	Evidence string
}

func assessProjectTechnologies(options Options, schema ConfigSchema, technologies []Technology, global map[string]any, legacy map[string]string, projectRoot string) *technologyAssessment {
	technologyField, ok := schemaField(schema, "SDLC_TECHNOLOGIES")
	if !ok {
		fmt.Fprintln(options.ErrorOutput, "Warning: technology assessment skipped because SDLC_TECHNOLOGIES is absent from the configuration schema.")
		return nil
	}
	if value, explicit, _ := initialValue(technologyField, options.Overrides, legacy, global); explicit && value != "" {
		return nil
	}

	available, err := availableProjectFiles(options, projectRoot)
	if err != nil {
		fmt.Fprintf(options.ErrorOutput, "Warning: deterministic technology detection was unavailable: %v\nSelect the applicable technologies manually.\n", err)
		return nil
	}
	assessment, err := detectProjectTechnologies(schema.TechnologyDetection, technologies, available)
	if err != nil {
		fmt.Fprintf(options.ErrorOutput, "Warning: deterministic technology detection failed: %v\nSelect the applicable technologies manually.\n", err)
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

func detectProjectTechnologies(detection TechnologyDetection, technologies []Technology, files map[string]bool) (technologyAssessment, error) {
	if err := validateTechnologyDetectionStandards(detection, technologies); err != nil {
		return technologyAssessment{}, err
	}

	installed := make(map[string]bool, len(technologies))
	for _, technology := range technologies {
		installed[technology.Name] = true
	}
	rules := make(map[string]TechnologyDetectionRule, len(detection.Rules))
	for _, rule := range detection.Rules {
		rules[rule.Technology] = rule
	}

	matches := map[string][]string{}
	for path := range files {
		if excludedTechnologyPath(path, detection) {
			continue
		}
		for _, rule := range detection.Rules {
			if technologyPathMatches(path, rule) {
				matches[rule.Technology] = append(matches[rule.Technology], filepath.ToSlash(path))
			}
		}
	}
	for technology := range matches {
		sort.Strings(matches[technology])
	}
	impliedBy := map[string][]string{}
	for technology := range matches {
		for _, implied := range rules[technology].Implies {
			impliedBy[implied] = append(impliedBy[implied], technology)
		}
	}

	assessment := technologyAssessment{}
	for _, technology := range technologies {
		paths := matches[technology.Name]
		if len(paths) != 0 {
			assessment.Technologies = append(assessment.Technologies, technologyAssessmentCandidate{
				Name: technology.Name, Evidence: technologyEvidence(paths),
			})
			continue
		}
		if sources := impliedBy[technology.Name]; len(sources) != 0 {
			sort.Strings(sources)
			assessment.Technologies = append(assessment.Technologies, technologyAssessmentCandidate{
				Name: technology.Name, Evidence: "Required by detected " + strings.Join(sources, ", ") + " standards.",
			})
		}
	}
	return assessment, nil
}

func validateTechnologyDetectionStandards(detection TechnologyDetection, technologies []Technology) error {
	installed := make(map[string]bool, len(technologies))
	for _, technology := range technologies {
		installed[technology.Name] = true
	}
	rules := make(map[string]bool, len(detection.Rules))
	for _, rule := range detection.Rules {
		if !installed[rule.Technology] {
			return fmt.Errorf("technology detection rule %s has no installed standard", rule.Technology)
		}
		for _, implied := range rule.Implies {
			if !installed[implied] {
				return fmt.Errorf("technology detection rule %s implies unavailable standard %s", rule.Technology, implied)
			}
		}
		rules[rule.Technology] = true
	}
	for technology := range installed {
		if !rules[technology] {
			return fmt.Errorf("installed technology standard %s has no detection rule", technology)
		}
	}
	return nil
}

func excludedTechnologyPath(path string, detection TechnologyDetection) bool {
	normalized := strings.ToLower(filepath.ToSlash(filepath.Clean(path)))
	for _, prefix := range detection.ExcludePrefixes {
		if strings.HasPrefix(normalized, strings.ToLower(filepath.ToSlash(prefix))) {
			return true
		}
	}
	excluded := map[string]bool{}
	for _, directory := range detection.ExcludeDirectories {
		excluded[strings.ToLower(directory)] = true
	}
	parts := strings.Split(normalized, "/")
	for _, part := range parts[:len(parts)-1] {
		if excluded[part] {
			return true
		}
	}
	return false
}

func technologyPathMatches(path string, rule TechnologyDetectionRule) bool {
	base := strings.ToLower(filepath.Base(path))
	extension := strings.ToLower(filepath.Ext(base))
	for _, candidate := range rule.Basenames {
		if base == strings.ToLower(candidate) {
			return true
		}
	}
	for _, candidate := range rule.Extensions {
		if extension == strings.ToLower(candidate) {
			return true
		}
	}
	return false
}

func technologyEvidence(paths []string) string {
	shown := paths
	if len(shown) > 3 {
		shown = shown[:3]
	}
	evidence := "Matched tracked file"
	if len(paths) != 1 {
		evidence += "s"
	}
	evidence += ": " + strings.Join(shown, ", ")
	if len(paths) > len(shown) {
		evidence += fmt.Sprintf(" and %d more", len(paths)-len(shown))
	}
	return evidence + "."
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
