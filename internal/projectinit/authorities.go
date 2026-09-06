package projectinit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const authorityDiscoveryPromptPath = "prompts/discover-project-authorities.md"
const authorityProposalFilename = "authority-proposal.yaml"

type authorityProposal struct {
	Version      int                             `yaml:"version"`
	Authorities  map[string][]authorityCandidate `yaml:"authorities"`
	MigratedWork []migratedWorkCandidate         `yaml:"migrated_work"`
	Warnings     []string                        `yaml:"warnings"`
}

type authorityCandidate struct {
	Path       string `yaml:"path"`
	Descriptor string `yaml:"descriptor"`
	Rationale  string `yaml:"rationale"`
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
	var unresolved []ConfigField
	for _, field := range fields {
		value, isExplicit, _ := initialValue(field, options.Overrides, legacy, global)
		if !isExplicit {
			unresolved = append(unresolved, field)
			continue
		}
		if err := field.ValidateValue(value, technologies); err != nil {
			return fmt.Errorf("%s: %w", field.Key, err)
		}
		value, err = validateAuthorityPaths(projectRoot, available, value)
		if err != nil {
			return fmt.Errorf("%s: %w", field.Key, err)
		}
		if value != "" {
			values[field.Key] = value
		}
		explicit[field.Key] = true
	}
	if len(unresolved) == 0 && len(generation.legacyV2) == 0 {
		return nil
	}

	proposalPath, err := runAuthorityDiscovery(options, sdlcRoot, projectRoot, workspace, values["SDLC_AUDIT_MODEL"])
	if err != nil {
		return err
	}
	proposal, err := readAuthorityProposal(proposalPath)
	if err != nil {
		return err
	}
	proposal, err = validateAuthorityProposal(proposal, fields, projectRoot, available, generation.legacyV2)
	if err != nil {
		return fmt.Errorf("validating authority proposal %s: %w", proposalPath, err)
	}
	for _, warning := range proposal.Warnings {
		if warning != "" {
			fmt.Fprintf(options.Output, "Authority discovery warning: %s\n", warning)
		}
	}
	generation.migratedV2 = proposal.MigratedWork
	for _, field := range unresolved {
		value, err := promptAuthorityField(options, field, proposal.Authorities[field.DiscoveryCategory], projectRoot, available)
		if err != nil {
			return err
		}
		if value != "" {
			values[field.Key] = value
		}
		explicit[field.Key] = true
	}
	return nil
}

func runAuthorityDiscovery(options Options, sdlcRoot, projectRoot, workspace, model string) (string, error) {
	promptPath := filepath.Join(sdlcRoot, filepath.FromSlash(authorityDiscoveryPromptPath))
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("reading authority-discovery prompt %s: %w", promptPath, err)
	}
	proposalPath := filepath.Join(workspace, authorityProposalFilename)
	arguments := []string{"exec", "--ephemeral", "--sandbox", "read-only", "--output-last-message", proposalPath}
	if strings.TrimSpace(model) != "" {
		arguments = append(arguments, "--model", model)
	}
	arguments = append(arguments, "-")
	if err := options.RunCommand("codex", arguments, projectRoot, bytes.NewReader(prompt), options.Output, options.ErrorOutput); err != nil {
		return "", fmt.Errorf("discovering project authorities with headless Codex: %w", err)
	}
	return proposalPath, nil
}

func readAuthorityProposal(path string) (authorityProposal, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return authorityProposal{}, fmt.Errorf("reading authority proposal %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var proposal authorityProposal
	if err := decoder.Decode(&proposal); err != nil {
		return authorityProposal{}, fmt.Errorf("parsing authority proposal %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return authorityProposal{}, fmt.Errorf("authority proposal %s contains multiple YAML documents", path)
		}
		return authorityProposal{}, fmt.Errorf("parsing authority proposal %s: %w", path, err)
	}
	if proposal.Version != 1 {
		return authorityProposal{}, fmt.Errorf("authority proposal %s must declare version: 1", path)
	}
	if proposal.Authorities == nil {
		return authorityProposal{}, fmt.Errorf("authority proposal %s has no authorities mapping", path)
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
		if path != "" && path != "." {
			available[path] = true
		}
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

func validateAuthorityProposal(proposal authorityProposal, fields []ConfigField, projectRoot string, available map[string]bool, legacyV2 []string) (authorityProposal, error) {
	expected := map[string]bool{}
	for _, field := range fields {
		expected[field.DiscoveryCategory] = true
		candidates, ok := proposal.Authorities[field.DiscoveryCategory]
		if !ok {
			return authorityProposal{}, fmt.Errorf("authority proposal omits %s list", field.DiscoveryCategory)
		}
		seen := map[string]bool{}
		validated := make([]authorityCandidate, 0, len(candidates))
		for _, candidate := range candidates {
			if strings.TrimSpace(candidate.Descriptor) == "" || strings.TrimSpace(candidate.Rationale) == "" {
				return authorityProposal{}, fmt.Errorf("authority proposal entry %q requires a descriptor and rationale", candidate.Path)
			}
			path, err := validateAuthorityPaths(projectRoot, available, candidate.Path)
			if err != nil {
				return authorityProposal{}, err
			}
			if path == "" || strings.Contains(path, ",") {
				return authorityProposal{}, fmt.Errorf("authority proposal entry must contain exactly one path: %q", candidate.Path)
			}
			if seen[path] {
				continue
			}
			if field.DiscoveryCategory == "requirements" && (path == "docs/work.org" || strings.HasPrefix(path, "docs/archive/sdlc-v2/specs/")) {
				proposal.Warnings = append(proposal.Warnings, fmt.Sprintf("Excluded %s from requirement authorities because migrated work is indexed through docs/work.org.", path))
				continue
			}
			seen[path] = true
			candidate.Path = path
			candidate.Descriptor = strings.TrimSpace(candidate.Descriptor)
			candidate.Rationale = strings.TrimSpace(candidate.Rationale)
			validated = append(validated, candidate)
		}
		proposal.Authorities[field.DiscoveryCategory] = validated
	}
	for category := range proposal.Authorities {
		if !expected[category] {
			return authorityProposal{}, fmt.Errorf("authority proposal contains unknown category %q", category)
		}
	}
	migratedWork, err := validateMigratedWork(proposal.MigratedWork, legacyV2, projectRoot, available)
	if err != nil {
		return authorityProposal{}, err
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
			return nil, fmt.Errorf("authority proposal omits archived Spec Kit specification %s", path)
		}
		validated = append(validated, candidate)
	}
	return validated, nil
}

func singleLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func promptAuthorityField(options Options, field ConfigField, candidates []authorityCandidate, projectRoot string, available map[string]bool) (string, error) {
	fmt.Fprintf(options.Output, "\nProposed %s:\n", strings.ToLower(field.Prompt))
	paths := make([]string, 0, len(candidates))
	if len(candidates) == 0 {
		fmt.Fprintln(options.Output, "- None identified.")
	} else {
		for index, candidate := range candidates {
			fmt.Fprintf(options.Output, "%d. %s: %s\n   Evidence: %s\n", index+1, candidate.Path, candidate.Descriptor, candidate.Rationale)
			paths = append(paths, candidate.Path)
		}
	}
	fmt.Fprintf(options.Output, "Press Enter to accept, enter replacement paths separated by commas, or '-' for none: ")
	line, err := options.inputReader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("reading %s: %w", field.Key, err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		line = strings.Join(paths, ",")
	} else if line == "-" {
		line = ""
	}
	return validateAuthorityPaths(projectRoot, available, line)
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
