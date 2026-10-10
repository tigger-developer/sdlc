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
	all, err := installedAuditDocuments(root)
	if err != nil {
		t.Fatal(err)
	}
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
			for _, path := range all {
				if !present[path] {
					t.Errorf("%s audit omitted installed document %s", gate, path)
				}
			}
			if !present[profile] || !present[filepath.Join(root, "technologies", "GO.md")] {
				t.Errorf("%s audit omitted profile or selected Go standard", gate)
			}
		})
	}
}

func TestStandaloneAuditSuppliesFullInstalledInventory(t *testing.T) {
	root := installedStandardsRoot()
	all, err := installedAuditDocuments(root)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := standaloneAuditStandards(root, []string{"project.go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != len(all)+1 || paths[0] != "project.go" {
		t.Fatalf("standalone evidence count = %d; want %d", len(paths), len(all)+1)
	}
	for _, name := range []string{"ORG-SCHEMA.md", "MARKDOWN.md", "technologies/GO.md"} {
		if !strings.Contains(strings.Join(paths, "\n"), filepath.Join(root, name)) {
			t.Fatalf("standalone inventory omitted %s", name)
		}
	}
}

func TestInstalledAuditDocumentsDiscoverNewStandard(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"MAIN.md", "technologies/NEW.md", "prompts/reference.md"} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(name+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	paths, err := installedAuditDocuments(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 3 || paths[1] != filepath.Join(root, "prompts", "reference.md") || paths[2] != filepath.Join(root, "technologies", "NEW.md") {
		t.Fatalf("unexpected installed inventory: %v", paths)
	}
}

func TestStandaloneAuditRequiresItsCoreStandards(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "MAIN.md"), []byte("# Main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := standaloneAuditStandards(root, []string{"project.go"})
	if err == nil || !strings.Contains(err.Error(), "required standalone audit standard") {
		t.Fatalf("missing core standard error = %v", err)
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
