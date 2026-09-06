package projectinit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

const authorityDiscoveryPromptPath = "prompts/discover-project-authorities.md"
const migratedWorkProposalFilename = "migrated-work-proposal.yaml"

type migratedWorkProposal struct {
	Version      int                     `yaml:"version"`
	MigratedWork []migratedWorkCandidate `yaml:"migrated_work"`
	Warnings     []string                `yaml:"warnings"`
}

type authorityChoice struct {
	Path     string
	Selected bool
}

type migratedWorkCandidate struct {
	Path        string `yaml:"path"`
	Descriptor  string `yaml:"descriptor"`
	Disposition string `yaml:"disposition"`
	Priority    string `yaml:"priority"`
	Created     string `yaml:"created"`
	Evidence    string `yaml:"evidence"`
}

func initializationWorkspacePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".sdlc", ".init")
}

func resolvePostMigrationConfiguration(
	options Options,
	schema ConfigSchema,
	technologies []Technology,
	global map[string]any,
	legacy map[string]string,
	sdlcRoot, projectRoot, workspace string,
	values map[string]string,
	explicit map[string]bool,
	generation *projectGeneration,
) error {
	fields := postMigrationFields(schema)
	available, err := availableProjectFiles(options, projectRoot)
	if err != nil {
		return err
	}
	for _, field := range fields {
		if field.DiscoveryCategory == "requirements" {
			continue
		}
		value, isExplicit, _ := initialValue(field, options.Overrides, legacy, global)
		if isExplicit {
			if err := field.ValidateValue(value, technologies); err != nil {
				return fmt.Errorf("%s: %w", field.Key, err)
			}
			value, err = validateAuthorityPaths(projectRoot, available, value)
			if err != nil {
				return fmt.Errorf("%s: %w", field.Key, err)
			}
		} else {
			value, err = promptDiscoveredAuthorityField(options, field, projectRoot)
			if err != nil {
				return err
			}
		}
		values[field.Key] = value
		explicit[field.Key] = true
	}

	available, err = availableProjectFiles(options, projectRoot)
	if err != nil {
		return err
	}
	requirements, err := deterministicRequirementAuthorities(available, generation.source)
	if err != nil {
		return err
	}
	values["SDLC_REQUIREMENT_AUTHORITIES"] = requirements
	explicit["SDLC_REQUIREMENT_AUTHORITIES"] = true

	if len(generation.legacyV2) == 0 {
		return nil
	}
	proposalPath, err := runMigratedWorkClassification(options, sdlcRoot, projectRoot, workspace, values["SDLC_AUDIT_MODEL"], generation.legacyV2)
	if err != nil {
		return err
	}
	proposal, err := readMigratedWorkProposal(proposalPath)
	if err != nil {
		return err
	}
	proposal, err = validateMigratedWorkProposal(proposal, projectRoot, available, generation.legacyV2)
	if err != nil {
		return fmt.Errorf("validating migrated-work proposal %s: %w", proposalPath, err)
	}
	for _, warning := range proposal.Warnings {
		if warning != "" {
			fmt.Fprintf(options.Output, "Migrated work classification warning: %s\n", warning)
		}
	}
	generation.migratedV2 = proposal.MigratedWork
	return nil
}

func runMigratedWorkClassification(options Options, sdlcRoot, projectRoot, workspace, model string, legacyV2 []string) (string, error) {
	promptPath := filepath.Join(sdlcRoot, filepath.FromSlash(authorityDiscoveryPromptPath))
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("reading migrated-work classification prompt %s: %w", promptPath, err)
	}
	paths := make([]string, 0, len(legacyV2))
	for _, item := range legacyV2 {
		paths = append(paths, filepath.ToSlash(filepath.Join("docs", "archive", "sdlc-v2", "specs", item, "spec.md")))
	}
	pathDocument, err := yaml.Marshal(map[string][]string{"specifications": paths})
	if err != nil {
		return "", fmt.Errorf("rendering bounded migrated-work paths: %w", err)
	}
	prompt = append(prompt, []byte("\n\n## Exact specification scope\n\nInspect only these specification directories and their directly related files:\n\n```yaml\n")...)
	prompt = append(prompt, pathDocument...)
	prompt = append(prompt, []byte("```\n")...)
	proposalPath := filepath.Join(workspace, migratedWorkProposalFilename)
	arguments := []string{"exec", "--ephemeral", "--sandbox", "read-only", "--output-last-message", proposalPath}
	if strings.TrimSpace(model) != "" {
		arguments = append(arguments, "--model", model)
	}
	arguments = append(arguments, "-")
	if err := options.RunCommand("codex", arguments, projectRoot, bytes.NewReader(prompt), options.Output, options.ErrorOutput); err != nil {
		return "", fmt.Errorf("classifying archived Spec Kit work with headless Codex: %w", err)
	}
	return proposalPath, nil
}

func readMigratedWorkProposal(path string) (migratedWorkProposal, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return migratedWorkProposal{}, fmt.Errorf("reading migrated-work proposal %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var proposal migratedWorkProposal
	if err := decoder.Decode(&proposal); err != nil {
		return migratedWorkProposal{}, fmt.Errorf("parsing migrated-work proposal %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return migratedWorkProposal{}, fmt.Errorf("migrated-work proposal %s contains multiple YAML documents", path)
		}
		return migratedWorkProposal{}, fmt.Errorf("parsing migrated-work proposal %s: %w", path, err)
	}
	if proposal.Version != 1 {
		return migratedWorkProposal{}, fmt.Errorf("migrated-work proposal %s must declare version: 1", path)
	}
	return proposal, nil
}

func availableProjectFiles(options Options, projectRoot string) (map[string]bool, error) {
	output, err := commandOutput(options, projectRoot, "git", "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, fmt.Errorf("listing project files for authority validation: %w", err)
	}
	available := map[string]bool{}
	for _, path := range strings.Split(output, "\x00") {
		path = filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
		if path == "" || path == "." {
			continue
		}
		if _, statErr := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(path))); errors.Is(statErr, os.ErrNotExist) {
			continue
		} else if statErr != nil {
			return nil, fmt.Errorf("checking inventoried project path %s: %w", path, statErr)
		}
		available[path] = true
	}
	return available, nil
}

func validateAuthorityPaths(projectRoot string, available map[string]bool, value string) (string, error) {
	resolvedRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		return "", fmt.Errorf("resolving project root for authority validation: %w", err)
	}
	seen := map[string]bool{}
	var validated []string
	for _, path := range splitCSV(value) {
		if filepath.IsAbs(path) {
			return "", fmt.Errorf("authority path must be project-relative: %s", path)
		}
		clean := filepath.Clean(filepath.FromSlash(path))
		if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("authority path leaves the project: %s", path)
		}
		if filepath.Base(clean) == ".env" {
			return "", fmt.Errorf("authority path cannot name .env: %s", path)
		}
		relative := filepath.ToSlash(clean)
		if !available[relative] {
			return "", fmt.Errorf("authority path is not a tracked or migration-produced project file: %s", relative)
		}
		resolved, err := filepath.EvalSymlinks(filepath.Join(projectRoot, clean))
		if err != nil {
			return "", fmt.Errorf("resolving authority path %s: %w", relative, err)
		}
		within, err := filepath.Rel(resolvedRoot, resolved)
		if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("authority path resolves outside the project: %s", relative)
		}
		info, err := os.Stat(resolved)
		if err != nil || !info.Mode().IsRegular() {
			return "", fmt.Errorf("authority path is not a regular file: %s", relative)
		}
		if !seen[relative] {
			seen[relative] = true
			validated = append(validated, relative)
		}
	}
	return strings.Join(validated, ","), nil
}

func validateMigratedWorkProposal(proposal migratedWorkProposal, projectRoot string, available map[string]bool, legacyV2 []string) (migratedWorkProposal, error) {
	migratedWork, err := validateMigratedWork(proposal.MigratedWork, legacyV2, projectRoot, available)
	if err != nil {
		return migratedWorkProposal{}, err
	}
	proposal.MigratedWork = migratedWork
	for index, warning := range proposal.Warnings {
		proposal.Warnings[index] = singleLine(warning)
	}
	return proposal, nil
}

func validateMigratedWork(candidates []migratedWorkCandidate, legacyV2 []string, projectRoot string, available map[string]bool) ([]migratedWorkCandidate, error) {
	expected := make([]string, 0, len(legacyV2))
	expectedSet := make(map[string]bool, len(legacyV2))
	for _, item := range legacyV2 {
		path := filepath.ToSlash(filepath.Join("docs", "archive", "sdlc-v2", "specs", item, "spec.md"))
		expected = append(expected, path)
		expectedSet[path] = true
	}

	byPath := make(map[string]migratedWorkCandidate, len(candidates))
	for _, candidate := range candidates {
		path, err := validateAuthorityPaths(projectRoot, available, candidate.Path)
		if err != nil {
			return nil, fmt.Errorf("migrated work %q: %w", candidate.Path, err)
		}
		if path == "" || strings.Contains(path, ",") || !expectedSet[path] {
			return nil, fmt.Errorf("migrated work path is not an archived Spec Kit specification: %q", candidate.Path)
		}
		if _, duplicate := byPath[path]; duplicate {
			return nil, fmt.Errorf("migrated work specification appears more than once: %s", path)
		}
		candidate.Path = path
		candidate.Descriptor = singleLine(candidate.Descriptor)
		candidate.Disposition = strings.ToLower(singleLine(candidate.Disposition))
		candidate.Priority = singleLine(candidate.Priority)
		candidate.Created = strings.ToLower(singleLine(candidate.Created))
		candidate.Evidence = singleLine(candidate.Evidence)
		if candidate.Descriptor == "" || candidate.Priority == "" || candidate.Created == "" || candidate.Evidence == "" {
			return nil, fmt.Errorf("migrated work %s requires descriptor, priority, created, and evidence", path)
		}
		switch candidate.Disposition {
		case "delivered", "approved-undelivered", "abandoned", "unresolved":
		default:
			return nil, fmt.Errorf("migrated work %s has invalid disposition %q", path, candidate.Disposition)
		}
		if candidate.Created != "unknown" {
			if _, err := time.Parse("2006-01-02", candidate.Created); err != nil {
				return nil, fmt.Errorf("migrated work %s has invalid created date %q", path, candidate.Created)
			}
		}
		byPath[path] = candidate
	}

	validated := make([]migratedWorkCandidate, 0, len(expected))
	for _, path := range expected {
		candidate, ok := byPath[path]
		if !ok {
			return nil, fmt.Errorf("migrated-work proposal omits archived Spec Kit specification %s", path)
		}
		validated = append(validated, candidate)
	}
	return validated, nil
}

func singleLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func deterministicRequirementAuthorities(available map[string]bool, _ string) (string, error) {
	if !available["docs/work.org"] {
		return "", errors.New("docs/work.org is missing after project initialization")
	}
	hasOrg := available["docs/ACs.org"]
	hasMarkdown := available["docs/ACs.md"]
	if hasOrg && hasMarkdown {
		return "", errors.New("both docs/ACs.org and docs/ACs.md exist; resolve the conflicting legacy requirement authorities")
	}
	if hasOrg || hasMarkdown {
		return "", errors.New("legacy acceptance criteria remain outside docs/work.org after project initialization")
	}
	return "docs/work.org", nil
}

func discoverAuthorityChoices(available map[string]bool, category string) []authorityChoice {
	var tokens []string
	switch category {
	case "product":
		tokens = []string{"readme", "vision"}
	case "architecture":
		tokens = []string{"architecture"}
	default:
		return nil
	}
	var choices []authorityChoice
	for path := range available {
		extension := strings.ToLower(filepath.Ext(path))
		if extension != ".md" && extension != ".org" {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if !stemContainsAnyToken(stem, tokens) {
			continue
		}
		choices = append(choices, authorityChoice{Path: path, Selected: canonicalAuthorityPath(path, category)})
	}
	sort.Slice(choices, func(left, right int) bool {
		return choices[left].Path < choices[right].Path
	})
	return choices
}

func stemContainsAnyToken(stem string, wanted []string) bool {
	tokens := strings.FieldsFunc(strings.ToLower(stem), func(character rune) bool {
		return !unicode.IsLetter(character) && !unicode.IsNumber(character)
	})
	for _, token := range tokens {
		for _, want := range wanted {
			if token == want {
				return true
			}
		}
	}
	return false
}

func canonicalAuthorityPath(path, category string) bool {
	normalized := strings.ToLower(filepath.ToSlash(path))
	switch category {
	case "product":
		return normalized == "readme.md" || normalized == "readme.org" || normalized == "docs/vision.md" || normalized == "docs/vision.org"
	case "architecture":
		return normalized == "docs/architecture.md" || normalized == "docs/architecture.org"
	default:
		return false
	}
}

func promptDiscoveredAuthorityField(options Options, field ConfigField, projectRoot string) (string, error) {
	selected := map[string]bool{}
	seen := map[string]bool{}
	for {
		available, err := availableProjectFiles(options, projectRoot)
		if err != nil {
			return "", err
		}
		choices := discoverAuthorityChoices(available, field.DiscoveryCategory)
		present := map[string]bool{}
		for index := range choices {
			path := choices[index].Path
			present[path] = true
			if !seen[path] {
				selected[path] = choices[index].Selected
				seen[path] = true
			}
			choices[index].Selected = selected[path]
		}
		for path := range selected {
			if !present[path] {
				delete(selected, path)
			}
		}

		displayChoices := make([]promptChoice, 0, len(choices))
		for _, choice := range choices {
			displayChoices = append(displayChoices, promptChoice{Label: choice.Path, Selected: choice.Selected})
		}
		renderPromptChoices(options.Output, "Select "+strings.ToLower(field.Prompt)+":", displayChoices)
		if len(choices) == 0 {
			fmt.Fprintln(options.Output, "- No matching Markdown or Org documents found.")
		}
		fmt.Fprint(options.Output, "Enter numbers to toggle, paths to replace, 'r' to rescan, '-' for none, or press Enter to accept: ")
		line, err := options.inputReader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("reading %s: %w", field.Key, err)
		}
		line = strings.TrimSpace(line)
		switch line {
		case "":
			var paths []string
			for _, choice := range choices {
				if choice.Selected {
					paths = append(paths, choice.Path)
				}
			}
			return validateAuthorityPaths(projectRoot, available, strings.Join(paths, ","))
		case "r", "R":
			continue
		case "-":
			return "", nil
		}

		items := splitCSV(line)
		numbers := make([]int, 0, len(items))
		allNumbers := len(items) != 0
		for _, item := range items {
			number, numberErr := strconv.Atoi(item)
			if numberErr != nil {
				allNumbers = false
				break
			}
			numbers = append(numbers, number)
		}
		if !allNumbers {
			return validateAuthorityPaths(projectRoot, available, line)
		}
		for _, number := range numbers {
			if number < 1 || number > len(choices) {
				return "", fmt.Errorf("selection %d is outside the available range", number)
			}
			path := choices[number-1].Path
			selected[path] = !selected[path]
		}
	}
}

func postMigrationFields(schema ConfigSchema) []ConfigField {
	var fields []ConfigField
	for _, field := range schema.Fields {
		if field.Phase == "post-migration" {
			fields = append(fields, field)
		}
	}
	return fields
}
