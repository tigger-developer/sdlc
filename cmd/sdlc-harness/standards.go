// ABOUTME: Supplies installed SDLC standards to managed audit evidence.
// ABOUTME: Keeps canonical review rules separate from caller-selected project files.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var installedStandardsRoot = func() string {
	return filepath.Dir(filepath.Dir(auditPromptPath("")))
}

func managedAuditStandards(project, gate, standardsRoot string, inputs []string) ([]string, error) {
	profilePath := filepath.Join(project, ".sdlc", "project.yaml")
	contents, err := os.ReadFile(profilePath)
	if errors.Is(err, os.ErrNotExist) {
		// Older CLI fixtures select a managed gate without writing a profile.
		return inputs, nil
	}
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
	names := []string{"MAIN.md", "AUDITS.md", "ARCHITECTURE.md", "TESTING.md"}
	switch gate {
	case "definition":
		names = append(names, "ISSUES.md", "DOCUMENTATION.md", "ORGMODE.md")
	case "test-code":
		names = append(names, "CODING.md")
	case "delivery-code":
		names = append(names, "CODING.md", "DOCUMENTATION.md", "SECURITY.md")
	default:
		return nil, fmt.Errorf("unknown managed audit gate %q", gate)
	}
	paths := append([]string(nil), inputs...)
	seen := make(map[string]bool, len(paths)+len(names)+len(profile.Standards.Technologies)+1)
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(project, path)
		}
		seen[filepath.Clean(path)] = true
	}
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
	if err := add(profilePath); err != nil {
		return nil, err
	}
	for _, name := range names {
		if err := add(filepath.Join(standardsRoot, name)); err != nil {
			return nil, err
		}
	}
	for _, technology := range profile.Standards.Technologies {
		if technology == "" || strings.Trim(technology, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-") != "" {
			return nil, fmt.Errorf("invalid selected audit technology %q", technology)
		}
		if err := add(filepath.Join(standardsRoot, "technologies", technology+".md")); err != nil {
			return nil, err
		}
	}
	return paths, nil
}
