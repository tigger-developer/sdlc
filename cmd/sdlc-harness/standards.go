// ABOUTME: Supplies installed SDLC standards to managed audit evidence.
// ABOUTME: Keeps canonical review rules separate from caller-selected project files.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var installedStandardsRoot = func() string {
	return filepath.Dir(filepath.Dir(auditPromptPath("")))
}

func rejectCallerStandards(project, root string, inputs []string) error {
	if len(inputs) == 0 {
		return nil
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("resolving installed audit standards %q: %w", root, err)
	}
	for _, path := range inputs {
		if !filepath.IsAbs(path) {
			path = filepath.Join(project, path)
		}
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("resolving project audit input %q: %w", path, err)
		}
		if withinPath(root, path) || withinPath(canonicalRoot, resolved) {
			return fmt.Errorf("caller-supplied audit standard %q is rejected; supply project evidence only", path)
		}
	}
	return nil
}

func managedAuditStandards(project, gate, standardsRoot string, inputs []string) ([]string, error) {
	required, err := managedRequiredStandards(project, gate, standardsRoot, inputs...)
	if err != nil {
		return nil, err
	}
	inventory, err := reviewReferenceDocuments(standardsRoot)
	if err != nil {
		return nil, err
	}
	paths := append([]string(nil), inputs...)
	seen := map[string]bool{}
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(project, path)
		}
		seen[filepath.Clean(path)] = true
	}
	profilePath := filepath.Join(project, ".sdlc", "project.yaml")
	if !seen[profilePath] {
		paths = append(paths, profilePath)
		seen[profilePath] = true
	}
	for _, path := range append(required, inventory...) {
		if !seen[path] {
			paths = append(paths, path)
			seen[path] = true
		}
	}
	return paths, nil
}

func managedRequiredStandards(project, gate, standardsRoot string, inputs ...string) ([]string, error) {
	technologies, err := projectReviewTechnologies(project)
	if err != nil {
		return nil, err
	}
	names := []string{"INDEPENDENT-AUDITOR.md", "READING.md", "REVIEW.md", "ARCHITECTURE.md", "TESTING.md"}
	switch gate {
	case "definition":
		names = append(names, "ISSUES.md", "CODING.md", "SECURITY.md", "DOCUMENTATION.md", "ORGMODE.md", "ORG-SCHEMA.md")
	case "test-code":
		names = append(names, "CODING.md", "GIT.md", "SECURITY.md", "ORG-SCHEMA.md")
	case "delivery-code":
		names = append(names, "CODING.md", "GIT.md", "DOCUMENTATION.md", "SECURITY.md", "ORG-SCHEMA.md")
	default:
		return nil, fmt.Errorf("unknown managed audit gate %q", gate)
	}
	return selectedReviewStandards(standardsRoot, names, technologies, inputs)
}

func projectReviewTechnologies(project string) ([]string, error) {
	profilePath := filepath.Join(project, ".sdlc", "project.yaml")
	contents, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("reading audit project profile: %w", err)
	}
	var profile struct {
		Standards struct {
			Technologies []string `yaml:"technologies"`
		} `yaml:"standards"`
	}
	if err := yaml.Unmarshal(contents, &profile); err != nil {
		return nil, fmt.Errorf("parsing audit project profile: %w", err)
	}
	return profile.Standards.Technologies, nil
}

// Selection is shared by managed, standalone and spike reviews. Evidence adds
// technologies even when a project's declaration omitted them.
func selectedReviewStandards(standardsRoot string, names, selected, inputs []string) ([]string, error) {
	var paths []string
	seen := map[string]bool{}
	add := func(path string) error {
		if seen[path] {
			return nil
		}
		info, statErr := os.Stat(path)
		if statErr != nil {
			return fmt.Errorf("required audit standard %q: %w", path, statErr)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("required audit standard %q must be a regular file", path)
		}
		paths = append(paths, path)
		seen[path] = true
		return nil
	}
	for _, name := range names {
		if err := add(filepath.Join(standardsRoot, name)); err != nil {
			return nil, err
		}
	}
	for _, path := range inputs {
		var format string
		switch strings.ToLower(filepath.Ext(path)) {
		case ".md":
			format = "MARKDOWN.md"
		case ".org":
			format = "ORGMODE.md"
		}
		if format != "" {
			for _, name := range []string{"DOCUMENTATION.md", format} {
				if err := add(filepath.Join(standardsRoot, name)); err != nil {
					return nil, err
				}
			}
		}
	}
	technologies := append([]string(nil), selected...)
	for _, path := range inputs {
		technologies = append(technologies, evidenceTechnologies(path)...)
	}
	for _, technology := range technologies {
		switch technology {
		case "NODE":
			technologies = append(technologies, "JAVASCRIPT")
		case "HUGO":
			technologies = append(technologies, "WEB")
		}
	}
	for _, technology := range technologies {
		if technology == "" || strings.Trim(technology, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-") != "" {
			return nil, fmt.Errorf("invalid selected audit technology %q", technology)
		}
		if err := add(filepath.Join(standardsRoot, "technologies", technology+".md")); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

// Original evidence supplements the profile so an omitted language declaration
// cannot waive that language's audit standard.
func evidenceTechnologies(path string) []string {
	if filepath.Base(path) == "go.mod" {
		return []string{"GO"}
	}
	if filepath.Base(path) == "package.json" {
		return []string{"NODE", "JAVASCRIPT"}
	}
	names := map[string]string{".go": "GO", ".js": "JAVASCRIPT", ".ts": "JAVASCRIPT", ".jsx": "JAVASCRIPT", ".tsx": "JAVASCRIPT", ".html": "WEB", ".css": "WEB", ".sh": "SHELL", ".fish": "FISH", ".py": "PYTHON", ".lua": "LUA", ".pl": "PERL", ".rs": "RUST", ".swift": "SWIFT"}
	if name := names[strings.ToLower(filepath.Ext(path))]; name != "" {
		return []string{name}
	}
	return nil
}

// installedAuditDocuments inventories only the installed SDLC Markdown source.
// The gate prompt distinguishes required rules from available reference material.
func installedAuditDocuments(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(entry.Name()) != ".md" {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("installed audit document %q must be a regular file", path)
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("inventorying installed SDLC documents: %w", err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no installed SDLC Markdown documents under %q", root)
	}
	sort.Strings(paths)
	return paths, nil
}

func standaloneAuditStandards(standardsRoot string, inputs []string) ([]string, error) {
	names := []string{"INDEPENDENT-AUDITOR.md", "READING.md", "REVIEW.md", "CODING.md", "TESTING.md", "ARCHITECTURE.md", "SECURITY.md", "GIT.md", "DOCUMENTATION.md"}
	paths, err := selectedReviewStandards(standardsRoot, names, nil, inputs)
	if err != nil {
		return nil, fmt.Errorf("required standalone audit standard: %w", err)
	}
	return append(append([]string(nil), inputs...), paths...), nil
}

// spikeAuditStandards deliberately excludes workflow and documentation artefacts.
// Only declared and evidence-selected technologies are mandatory.
func spikeAuditStandards(root string, inputs []string, technologies ...string) ([]string, error) {
	names := []string{"INDEPENDENT-AUDITOR.md", "READING.md", "REVIEW.md", "CODING.md", "ARCHITECTURE.md", "SECURITY.md", "GIT.md", "TESTING.md"}
	paths, err := selectedReviewStandards(root, names, technologies, inputs)
	if err != nil {
		return nil, err
	}
	return append(append([]string(nil), inputs...), paths...), nil
}

// Delivery and transport documents are not even optional auditor references.
// New technology standards remain discoverable without loading every technology.
func isReviewReference(name string) bool {
	if strings.HasPrefix(name, "technologies/") && strings.HasSuffix(name, ".md") {
		return true
	}
	switch name {
	case "INDEPENDENT-AUDITOR.md", "READING.md", "REVIEW.md",
		"CODING.md", "TESTING.md", "ARCHITECTURE.md", "SECURITY.md", "GIT.md",
		"DOCUMENTATION.md", "ISSUES.md", "ORGMODE.md", "ORG-SCHEMA.md", "MARKDOWN.md":
		return true
	default:
		return false
	}
}

func reviewReferenceDocuments(root string) ([]string, error) {
	paths, err := installedAuditDocuments(root)
	if err != nil {
		return nil, err
	}
	var references []string
	for _, path := range paths {
		name, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		if isReviewReference(filepath.ToSlash(name)) {
			references = append(references, path)
		}
	}
	if len(references) == 0 {
		return nil, errors.New("installed audit review reference inventory is empty")
	}
	return references, nil
}

func appendUniqueAuditPaths(project string, paths, additions []string) []string {
	seen := map[string]bool{}
	key := func(path string) string {
		if !filepath.IsAbs(path) {
			path = filepath.Join(project, path)
		}
		return filepath.Clean(path)
	}
	for _, path := range paths {
		seen[key(path)] = true
	}
	for _, path := range additions {
		if !seen[key(path)] {
			paths = append(paths, path)
			seen[key(path)] = true
		}
	}
	return paths
}
