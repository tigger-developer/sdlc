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

const authorityDiscoveryPromptPath = "prompts/discover-project-authorities.md"
const authorityProposalFilename = "authority-proposal.yaml"

type authorityProposal struct {
	Version     int                             `yaml:"version"`
	Authorities map[string][]authorityCandidate `yaml:"authorities"`
	Warnings    []string                        `yaml:"warnings"`
}

type authorityCandidate struct {
	Path       string `yaml:"path"`
	Descriptor string `yaml:"descriptor"`
	Rationale  string `yaml:"rationale"`
}

func initializationWorkspacePath(options Options, projectRoot string) (string, error) {
	path, err := commandOutput(options, projectRoot, "git", "rev-parse", "--git-path", initializationWorkspaceName)
	if err != nil {
		return "", fmt.Errorf("resolving temporary initialization workspace: %w", err)
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("Git returned an empty temporary initialization workspace path")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectRoot, path)
	}
	return filepath.Clean(path), nil
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
	if len(unresolved) == 0 {
		return nil
	}

	proposalPath, err := runAuthorityDiscovery(options, sdlcRoot, projectRoot, workspace, values["SDLC_SPEC_MODEL"])
	if err != nil {
		return err
	}
	proposal, err := readAuthorityProposal(proposalPath)
	if err != nil {
		return err
	}
	proposal, err = validateAuthorityProposal(proposal, fields, projectRoot, available)
	if err != nil {
		return fmt.Errorf("validating authority proposal %s: %w", proposalPath, err)
	}
	for _, warning := range proposal.Warnings {
		if warning != "" {
			fmt.Fprintf(options.Output, "Authority discovery warning: %s\n", warning)
		}
	}
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

func validateAuthorityProposal(proposal authorityProposal, fields []ConfigField, projectRoot string, available map[string]bool) (authorityProposal, error) {
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
	for index, warning := range proposal.Warnings {
		proposal.Warnings[index] = strings.TrimSpace(warning)
	}
	return proposal, nil
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
