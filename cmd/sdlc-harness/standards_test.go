package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestManagedAuditSuppliesCanonicalStandardsWithoutCallerInputs(t *testing.T) {
	project := t.TempDir()
	profile := filepath.Join(project, ".sdlc", "project.yaml")
	if err := os.Mkdir(filepath.Dir(profile), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profile, []byte("standards:\n  technologies: [GO]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	root := installedStandardsRoot()
	for _, gate := range []string{"definition", "test-code", "delivery-code"} {
		t.Run(gate, func(t *testing.T) {
			paths, err := managedAuditStandards(project, gate, root, nil)
			if err != nil {
				t.Fatal(err)
			}
			evidence, err := harness.CaptureEvidence(project, paths)
			if err != nil {
				t.Fatal(err)
			}
			present := map[string]bool{}
			for _, file := range evidence.Files {
				present[file.Path] = true
			}
			for _, name := range []string{"MAIN.md", "AUDITS.md", "ARCHITECTURE.md", "TESTING.md"} {
				if !present[filepath.Join(root, name)] {
					t.Errorf("%s audit omitted %s", gate, name)
				}
			}
			if !present[profile] || !present[filepath.Join(root, "technologies", "GO.md")] {
				t.Errorf("%s audit omitted profile or selected Go standard", gate)
			}
			if gate == "test-code" && present[filepath.Join(root, "DOCUMENTATION.md")] {
				t.Error("test-code audit included final-delivery documentation standard")
			}
		})
	}
}

func TestManagedAuditRejectsInvalidTechnologySelection(t *testing.T) {
	project := t.TempDir()
	profile := filepath.Join(project, ".sdlc", "project.yaml")
	if err := os.Mkdir(filepath.Dir(profile), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profile, []byte("standards:\n  technologies: [../escape]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := managedAuditStandards(project, "definition", installedStandardsRoot(), nil)
	if err == nil || !strings.Contains(err.Error(), "invalid selected audit technology") {
		t.Fatalf("invalid technology error = %v", err)
	}
}
