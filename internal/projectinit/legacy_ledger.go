package projectinit

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	legacyLedgerHeading = "* Legacy Acceptance Criteria (SDLC v1)"
	legacyLedgerID      = "legacy-acceptance-criteria"
)

// LegacyLedgerMerge describes one legacy-ledger consolidation.
type LegacyLedgerMerge struct {
	AcceptanceCriteria int
	LedgerChanged      bool
	ProfileChanged     bool
	SourceRemoved      bool
}

// MergeLegacyAcceptanceCriteria folds docs/ACs.org into docs/work.org and
// removes the now-redundant source ledger. It is safe to rerun after a
// successful merge.
func MergeLegacyAcceptanceCriteria(projectRoot string) (LegacyLedgerMerge, error) {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return LegacyLedgerMerge{}, fmt.Errorf("resolving project root: %w", err)
	}
	workPath := filepath.Join(root, "docs", "work.org")
	legacyPath := filepath.Join(root, "docs", "ACs.org")
	profilePath := filepath.Join(root, filepath.FromSlash(projectProfilePath))

	work, err := readRegularFile(workPath)
	if err != nil {
		return LegacyLedgerMerge{}, fmt.Errorf("reading work ledger: %w", err)
	}
	legacy, legacyErr := readRegularFile(legacyPath)
	if legacyErr != nil && !errors.Is(legacyErr, os.ErrNotExist) {
		return LegacyLedgerMerge{}, fmt.Errorf("reading legacy acceptance-criteria ledger: %w", legacyErr)
	}

	result := LegacyLedgerMerge{}
	merged := strings.Contains(string(work), legacyLedgerHeading+"\n")
	if errors.Is(legacyErr, os.ErrNotExist) {
		if !merged {
			return result, errors.New("docs/ACs.org is absent and docs/work.org contains no merged legacy acceptance-criteria subtree")
		}
		changed, updateErr := updateRequirementAuthorityProfile(profilePath)
		result.ProfileChanged = changed
		return result, updateErr
	}

	block, count, err := renderLegacyLedgerBlock(string(legacy))
	if err != nil {
		return result, err
	}
	result.AcceptanceCriteria = count
	updatedWork, changed, err := mergeLegacyBlock(string(work), block)
	if err != nil {
		return result, err
	}
	if changed {
		if err := writeAtomic(workPath, []byte(updatedWork)); err != nil {
			return result, fmt.Errorf("writing consolidated work ledger: %w", err)
		}
		result.LedgerChanged = true
	}

	profileChanged, err := updateRequirementAuthorityProfile(profilePath)
	if err != nil {
		return result, err
	}
	result.ProfileChanged = profileChanged
	if err := os.Remove(legacyPath); err != nil {
		return result, fmt.Errorf("removing merged legacy ledger %s: %w", legacyPath, err)
	}
	result.SourceRemoved = true
	return result, nil
}

func readRegularFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s must be a regular file, not a symbolic link", path)
	}
	return os.ReadFile(path)
}

func renderLegacyLedgerBlock(document string) (string, int, error) {
	document = strings.ReplaceAll(document, "\r\n", "\n")
	lines := strings.Split(document, "\n")
	firstHeading := -1
	sourceDate := ""
	for index, line := range lines {
		if strings.HasPrefix(line, "#+DATE:") {
			sourceDate = strings.TrimSpace(strings.TrimPrefix(line, "#+DATE:"))
		}
		if strings.HasPrefix(line, "* ") {
			firstHeading = index
			break
		}
	}
	if firstHeading < 0 {
		return "", 0, errors.New("docs/ACs.org contains no level-one Org headings")
	}

	sections := splitOrgTopLevelSections(lines[firstHeading:])
	acceptanceSection := -1
	for index, section := range sections {
		if section.title == "Acceptance criteria" {
			if acceptanceSection >= 0 {
				return "", 0, errors.New("docs/ACs.org contains more than one Acceptance criteria section")
			}
			acceptanceSection = index
		}
	}
	if acceptanceSection < 0 {
		return "", 0, errors.New("docs/ACs.org lacks the canonical Acceptance criteria section")
	}

	var output strings.Builder
	output.WriteString(legacyLedgerHeading)
	output.WriteString("\n:PROPERTIES:\n:CUSTOM_ID: ")
	output.WriteString(legacyLedgerID)
	output.WriteString("\n:VISIBILITY: folded\n")
	if sourceDate != "" {
		output.WriteString(":SOURCE_DATE: ")
		output.WriteString(sourceDate)
		output.WriteString("\n")
	}
	output.WriteString(":END:\n")

	criteria := 0
	for _, section := range sections {
		output.WriteString("\n")
		if section.title == "Acceptance criteria" {
			for _, line := range trimBlankEdges(section.lines[1:]) {
				output.WriteString(line)
				output.WriteString("\n")
				if strings.HasPrefix(line, "*** AC") {
					criteria++
				}
			}
			continue
		}
		for _, line := range section.lines {
			if strings.HasPrefix(line, "*") {
				line = "*" + line
			}
			output.WriteString(line)
			output.WriteString("\n")
		}
	}
	return strings.TrimRight(output.String(), "\n") + "\n", criteria, nil
}

type orgSection struct {
	title string
	lines []string
}

func splitOrgTopLevelSections(lines []string) []orgSection {
	var sections []orgSection
	for _, line := range lines {
		if strings.HasPrefix(line, "* ") {
			sections = append(sections, orgSection{title: strings.TrimSpace(strings.TrimPrefix(line, "* "))})
		}
		if len(sections) != 0 {
			sections[len(sections)-1].lines = append(sections[len(sections)-1].lines, line)
		}
	}
	return sections
}

func trimBlankEdges(lines []string) []string {
	for len(lines) != 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) != 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func mergeLegacyBlock(work, block string) (string, bool, error) {
	if strings.Contains(work, legacyLedgerHeading+"\n") {
		existing := orgTopLevelSection(work, strings.TrimPrefix(legacyLedgerHeading, "* "))
		if strings.TrimSpace(existing) != strings.TrimSpace(block) {
			return "", false, errors.New("docs/work.org already contains a different Legacy Acceptance Criteria (SDLC v1) subtree")
		}
		updated := updateHistoricalRequirementsLink(work)
		return updated, updated != work, nil
	}
	marker := "* Migration record\n"
	if !strings.Contains(work, marker) {
		return "", false, errors.New("docs/work.org lacks the canonical Migration record section")
	}
	updated := strings.Replace(work, marker, block+"\n"+marker, 1)
	updated = updateHistoricalRequirementsLink(updated)
	return updated, true, nil
}

func orgTopLevelSection(document, title string) string {
	lines := strings.Split(document, "\n")
	start := -1
	for index, line := range lines {
		if line == "* "+title {
			start = index
			continue
		}
		if start >= 0 && index > start && strings.HasPrefix(line, "* ") {
			return strings.Join(lines[start:index], "\n") + "\n"
		}
	}
	if start >= 0 {
		return strings.Join(lines[start:], "\n")
	}
	return ""
}

func updateHistoricalRequirementsLink(document string) string {
	lines := strings.Split(document, "\n")
	for index, line := range lines {
		if strings.HasPrefix(line, "- *Historical requirements:*") {
			lines[index] = "- *Historical requirements:* [[#" + legacyLedgerID + "][Legacy Acceptance Criteria (SDLC v1)]]."
		}
	}
	return strings.Join(lines, "\n")
}

func updateRequirementAuthorityProfile(path string) (bool, error) {
	contents, err := readRegularFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading project profile: %w", err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(contents, &document); err != nil {
		return false, fmt.Errorf("parsing project profile: %w", err)
	}
	root, err := yamlDocumentMapping(&document)
	if err != nil {
		return false, err
	}
	authorities := yamlMappingValue(root, "authorities")
	if authorities == nil {
		authorities = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		appendYAMLMapping(root, "authorities", authorities)
	} else if authorities.Kind != yaml.MappingNode {
		return false, errors.New("project profile authorities must be a YAML mapping")
	}
	requirements := yamlMappingValue(authorities, "requirements")
	if requirements != nil && requirements.Kind == yaml.SequenceNode && len(requirements.Content) == 1 && requirements.Content[0].Value == "docs/work.org" {
		return false, nil
	}
	sequence := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: "docs/work.org"}}}
	if requirements == nil {
		appendYAMLMapping(authorities, "requirements", sequence)
	} else {
		*requirements = *sequence
	}
	var rendered bytes.Buffer
	encoder := yaml.NewEncoder(&rendered)
	encoder.SetIndent(4)
	if err := encoder.Encode(&document); err != nil {
		return false, fmt.Errorf("rendering project profile: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return false, fmt.Errorf("closing project profile encoder: %w", err)
	}
	if bytes.Equal(contents, rendered.Bytes()) {
		return false, nil
	}
	if err := writeAtomic(path, rendered.Bytes()); err != nil {
		return false, fmt.Errorf("writing project profile: %w", err)
	}
	return true, nil
}

func yamlDocumentMapping(document *yaml.Node) (*yaml.Node, error) {
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("project profile must be a YAML document mapping")
	}
	return document.Content[0], nil
}

func yamlMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	return nil
}

func appendYAMLMapping(mapping *yaml.Node, key string, value *yaml.Node) {
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		value,
	)
}

func writeAtomic(path string, contents []byte) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("%s must be a regular file, not a symbolic link", path)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".sdlc-legacy-merge-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
