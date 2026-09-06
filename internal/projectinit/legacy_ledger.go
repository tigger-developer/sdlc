package projectinit

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	legacyLedgerHeading      = "* Legacy Acceptance Criteria (SDLC v1) :legacy:"
	legacyLedgerTitle        = "Legacy Acceptance Criteria (SDLC v1)"
	legacyLedgerID           = "legacy-acceptance-criteria"
	legacyLedgerPlaceholder  = "{{LEGACY_ACCEPTANCE_CRITERIA}}"
	obsoleteSpecKitAuthority = "Requirements established or changed through Spec Kit are governed by approved =specs/*/spec.md= artefacts."
)

var specKitReferencePattern = regexp.MustCompile(`(?i)spec.*kit`)

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
	existingBlock := legacyLedgerSection(string(work))
	if errors.Is(legacyErr, os.ErrNotExist) {
		if existingBlock == "" || strings.Contains(existingBlock, legacyLedgerPlaceholder) {
			return result, errors.New("docs/ACs.org is absent and docs/work.org contains no merged legacy acceptance-criteria subtree")
		}
		normalized, _, normalizeErr := normalizeLegacyLedgerBlock(existingBlock)
		if normalizeErr != nil {
			return result, normalizeErr
		}
		if strings.TrimSpace(normalized) != strings.TrimSpace(existingBlock) {
			updated := strings.Replace(string(work), existingBlock, normalized, 1)
			if err := writeAtomic(workPath, []byte(updated)); err != nil {
				return result, fmt.Errorf("normalizing consolidated work ledger: %w", err)
			}
			result.LedgerChanged = true
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
	lines := removeObsoleteSpecKitAuthority(strings.Split(document, "\n"))
	if err := rejectUnknownSpecKitReferences(lines); err != nil {
		return "", 0, err
	}
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
			converted, convertedCount, convertErr := normalizeLegacyAcceptanceCriteria(trimBlankEdges(section.lines[1:]))
			if convertErr != nil {
				return "", 0, convertErr
			}
			criteria += convertedCount
			for _, line := range converted {
				output.WriteString(line)
				output.WriteString("\n")
			}
			continue
		}
		for _, line := range section.lines {
			if strings.HasPrefix(line, "*") {
				line = "*" + line
				if line == "** Ledger authority" {
					line = "** Legacy ledger authority at migration"
				}
				if line == "** Status vocabulary" {
					line = "** Legacy status vocabulary at migration"
				}
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

func removeObsoleteSpecKitAuthority(lines []string) []string {
	var output []string
	for index := 0; index < len(lines); {
		if strings.TrimSpace(lines[index]) == "" {
			output = append(output, lines[index])
			index++
			continue
		}

		end := index + 1
		for end < len(lines) && strings.TrimSpace(lines[end]) != "" {
			end++
		}
		paragraph := strings.Join(strings.Fields(strings.Join(lines[index:end], " ")), " ")
		if paragraph != obsoleteSpecKitAuthority {
			output = append(output, lines[index:end]...)
		}
		index = end
	}
	return output
}

func rejectUnknownSpecKitReferences(lines []string) error {
	var references []string
	for index, line := range lines {
		if specKitReferencePattern.MatchString(line) {
			references = append(references, fmt.Sprintf("line %d: %s", index+1, strings.TrimSpace(line)))
		}
	}
	if len(references) == 0 {
		return nil
	}
	return fmt.Errorf("legacy acceptance-criteria ledger contains unrecognized Spec Kit references; review them before migration:\n%s", strings.Join(references, "\n"))
}

func mergeLegacyBlock(work, block string) (string, bool, error) {
	existing := legacyLedgerSection(work)
	if existing != "" {
		if strings.Contains(existing, legacyLedgerPlaceholder) {
			updated := strings.Replace(work, existing, block, 1)
			updated = updateHistoricalRequirementsLink(updated)
			return updated, true, nil
		}
		normalized, _, err := normalizeLegacyLedgerBlock(existing)
		if err != nil {
			return "", false, err
		}
		if strings.TrimSpace(normalized) != strings.TrimSpace(block) {
			return "", false, errors.New("docs/work.org already contains a different Legacy Acceptance Criteria (SDLC v1) subtree")
		}
		updated := work
		if normalized != existing {
			updated = strings.Replace(updated, existing, normalized, 1)
		}
		updated = updateHistoricalRequirementsLink(updated)
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

func legacyLedgerSection(document string) string {
	lines := strings.Split(document, "\n")
	start := -1
	for index, line := range lines {
		if strings.HasPrefix(line, "* "+legacyLedgerTitle) {
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

func normalizeLegacyLedgerBlock(block string) (string, int, error) {
	lines := removeObsoleteSpecKitAuthority(strings.Split(strings.ReplaceAll(block, "\r\n", "\n"), "\n"))
	if err := rejectUnknownSpecKitReferences(lines); err != nil {
		return "", 0, err
	}
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "* "+legacyLedgerTitle) {
		return "", 0, errors.New("legacy acceptance-criteria subtree has an invalid heading")
	}
	lines[0] = legacyLedgerHeading
	for index, line := range lines {
		switch line {
		case "** Ledger authority":
			lines[index] = "** Legacy ledger authority at migration"
		case "** Status vocabulary":
			lines[index] = "** Legacy status vocabulary at migration"
		}
	}
	converted, count, err := normalizeLegacyAcceptanceCriteria(lines)
	if err != nil {
		return "", 0, err
	}
	return strings.TrimRight(strings.Join(converted, "\n"), "\n") + "\n", count, nil
}

func normalizeLegacyAcceptanceCriteria(lines []string) ([]string, int, error) {
	var output []string
	criteria := 0
	for index := 0; index < len(lines); {
		line := lines[index]
		if !isLegacyACHeading(line) {
			output = append(output, line)
			index++
			continue
		}

		end := index + 1
		for end < len(lines) && !isLegacyACHeading(lines[end]) && !strings.HasPrefix(lines[end], "** ") {
			end++
		}
		normalized, err := normalizeLegacyAcceptanceCriterion(lines[index:end])
		if err != nil {
			return nil, 0, err
		}
		output = append(output, normalized...)
		criteria++
		index = end
	}
	return output, criteria, nil
}

func isLegacyACHeading(line string) bool {
	if !strings.HasPrefix(line, "*** ") {
		return false
	}
	text := strings.TrimSpace(strings.TrimPrefix(line, "*** "))
	for _, state := range []string{"PENDING", "FAILING", "SUPERSEDED", "HOLD", "HOLDING", "ASSUMED_PASS"} {
		text = strings.TrimSpace(strings.TrimPrefix(text, state+" "))
	}
	return strings.HasPrefix(text, "AC")
}

func normalizeLegacyAcceptanceCriterion(lines []string) ([]string, error) {
	heading := strings.TrimSpace(strings.TrimPrefix(lines[0], "*** "))
	existingState := ""
	for _, state := range []string{"ASSUMED_PASS", "SUPERSEDED", "HOLDING", "PENDING", "FAILING", "HOLD"} {
		if strings.HasPrefix(heading, state+" ") {
			existingState = normalizeLegacyState(state)
			heading = strings.TrimSpace(strings.TrimPrefix(heading, state+" "))
			break
		}
	}

	statusStart := -1
	for index := 1; index < len(lines); index++ {
		if lines[index] == "**** Status" {
			if statusStart >= 0 {
				return nil, fmt.Errorf("%s has more than one Status field", heading)
			}
			statusStart = index
		}
	}
	statusEnd := -1
	if statusStart >= 0 {
		statusEnd = statusStart + 1
		for statusEnd < len(lines) && !strings.HasPrefix(lines[statusEnd], "**** ") {
			statusEnd++
		}
	}

	state := existingState
	var qualification []string
	if statusStart >= 0 {
		parsedState, parsedQualification, err := parseLegacyStatus(lines[statusStart+1 : statusEnd])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", heading, err)
		}
		if state != "" && state != parsedState {
			return nil, fmt.Errorf("%s has conflicting headline and Status states", heading)
		}
		state = parsedState
		qualification = parsedQualification
	}
	if state == "" {
		return nil, fmt.Errorf("%s has no recognized legacy requirement state", heading)
	}

	output := []string{"*** " + state + " " + heading}
	if statusStart < 0 {
		return append(output, lines[1:]...), nil
	}
	output = append(output, lines[1:statusStart]...)
	if len(trimBlankEdges(qualification)) != 0 {
		output = append(output, "**** Status qualification", "")
		output = append(output, trimBlankEdges(qualification)...)
	}
	output = append(output, lines[statusEnd:]...)
	return output, nil
}

func parseLegacyStatus(lines []string) (string, []string, error) {
	trimmed := trimBlankEdges(lines)
	if len(trimmed) == 0 {
		return "", nil, errors.New("Status section is empty")
	}
	value := strings.TrimSpace(trimmed[0])
	value = strings.TrimSpace(strings.TrimPrefix(value, "- "))
	value = strings.TrimLeft(value, "*=")
	upper := strings.ToUpper(value)
	for _, candidate := range []string{"ASSUMED PASS", "ASSUMED_PASS", "SUPERSEDED", "HOLDING", "PENDING", "FAILING", "HOLD"} {
		if !strings.HasPrefix(upper, candidate) {
			continue
		}
		if len(value) > len(candidate) && !strings.ContainsRune(" \t,:*-_=", rune(value[len(candidate)])) {
			continue
		}
		remainder := strings.TrimLeft(value[len(candidate):], " \t,:*-=")
		qualification := append([]string(nil), trimmed[1:]...)
		if remainder != "" {
			qualification = append([]string{remainder}, qualification...)
		}
		return normalizeLegacyState(candidate), qualification, nil
	}
	return "", nil, fmt.Errorf("unrecognized Status %q", value)
}

func normalizeLegacyState(state string) string {
	switch strings.ReplaceAll(strings.ToUpper(state), " ", "_") {
	case "HOLDING":
		return "HOLD"
	case "ASSUMED_PASS":
		return "ASSUMED_PASS"
	default:
		return strings.ToUpper(state)
	}
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

func removeEmptyLegacyLedgerSection(document string) string {
	section := legacyLedgerSection(document)
	if section == "" || !strings.Contains(section, legacyLedgerPlaceholder) {
		return document
	}
	return strings.Replace(document, section, "", 1)
}

func mergeWorkLedgerPreamble(existing, rendered string) string {
	canonicalPreamble, _ := splitOrgPreamble(rendered)
	existingPreamble, existingBody := splitOrgPreamble(existing)
	existingTitle := orgDirective(existingPreamble, "TITLE")
	if existingTitle != "" {
		canonicalPreamble = replaceOrgDirective(canonicalPreamble, "TITLE", existingTitle)
	}
	if strings.TrimSpace(existingPreamble) == strings.TrimSpace(canonicalPreamble) {
		return canonicalPreamble + existingBody
	}

	extra := stripKnownWorkPreamble(existingPreamble)
	if strings.Contains(extra, "This is the project's canonical work and requirements index.") {
		extra = ""
	}
	result := strings.TrimRight(canonicalPreamble, "\n") + "\n"
	if strings.TrimSpace(extra) != "" {
		result += "\n" + strings.TrimSpace(extra) + "\n"
	}
	return result + existingBody
}

func splitOrgPreamble(document string) (string, string) {
	lines := strings.Split(document, "\n")
	for index, line := range lines {
		if strings.HasPrefix(line, "* ") {
			return strings.Join(lines[:index], "\n") + "\n", strings.Join(lines[index:], "\n")
		}
	}
	return document, ""
}

func orgDirective(document, name string) string {
	prefix := "#+" + name + ":"
	for _, line := range strings.Split(document, "\n") {
		if strings.HasPrefix(strings.ToUpper(line), prefix) {
			return line
		}
	}
	return ""
}

func replaceOrgDirective(document, name, replacement string) string {
	prefix := "#+" + name + ":"
	lines := strings.Split(document, "\n")
	for index, line := range lines {
		if strings.HasPrefix(strings.ToUpper(line), prefix) {
			lines[index] = replacement
			return strings.Join(lines, "\n")
		}
	}
	return replacement + "\n" + document
}

func stripKnownWorkPreamble(document string) string {
	lines := strings.Split(document, "\n")
	var output []string
	for index := 0; index < len(lines); index++ {
		upper := strings.ToUpper(lines[index])
		knownDirective := false
		for _, name := range []string{"TITLE", "STARTUP", "TODO", "TYP_TODO", "TAGS", "OPTIONS"} {
			if strings.HasPrefix(upper, "#+"+name+":") {
				knownDirective = true
				break
			}
		}
		if knownDirective {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(lines[index]), "#+begin_comment") {
			end := index + 1
			for end < len(lines) && !strings.EqualFold(strings.TrimSpace(lines[end]), "#+end_comment") {
				end++
			}
			if end < len(lines) {
				block := strings.Join(lines[index:end+1], "\n")
				if strings.Contains(block, "canonical SDLC v3 work-ledger template") {
					index = end
					continue
				}
			}
		}
		output = append(output, lines[index])
	}
	return strings.Join(output, "\n")
}

func consolidateStateSections(document string) (string, error) {
	var entries []string
	result := document
	for _, title := range []string{"Open defects", "Undelivered features", "Active delivery", "Human review", "Closed work"} {
		section := orgTopLevelSection(result, title)
		if section == "" {
			continue
		}
		lines := strings.Split(strings.TrimRight(section, "\n"), "\n")
		body := trimBlankEdges(lines[1:])
		if len(body) != 0 {
			entries = append(entries, strings.Join(body, "\n"))
		}
		result = strings.Replace(result, section, "", 1)
	}
	if len(entries) == 0 {
		return result, nil
	}
	return appendToOrgTopLevelSection(result, "Work items", strings.Join(entries, "\n\n"))
}

func appendToOrgTopLevelSection(document, title, addition string) (string, error) {
	section := orgTopLevelSection(document, title)
	if section == "" {
		return "", fmt.Errorf("work ledger lacks section %q", title)
	}
	updatedSection := strings.TrimRight(section, "\n") + "\n\n" + strings.TrimSpace(addition) + "\n\n"
	return strings.Replace(document, section, updatedSection, 1), nil
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
