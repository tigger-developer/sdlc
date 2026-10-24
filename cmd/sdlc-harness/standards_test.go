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
			for _, name := range []string{"INDEPENDENT-AUDITOR.md", "READING.md", "REVIEW.md", "ARCHITECTURE.md", "TESTING.md"} {
				if !present[filepath.Join(root, name)] {
					t.Errorf("%s audit omitted %s", gate, name)
				}
			}
			for _, name := range []string{"AUDIT.md", "STANDALONE-AUDIT.md", "SPIKE-AUDIT.md", "MAIN.md", "DELIVERY.md", "ADVICE.md", "AUDITS.md", "SELF-AUDIT.md", "EXTERNAL-AUDIT.md", "HARNESS.md", "PAIRING.md", "EMERGENCY.md"} {
				if present[filepath.Join(root, name)] {
					t.Errorf("%s audit exposed delivery or transport document %s", gate, name)
				}
			}
			if !present[profile] || !present[filepath.Join(root, "technologies", "GO.md")] {
				t.Errorf("%s audit omitted profile or selected Go standard", gate)
			}
		})
	}
}

func TestStandaloneAuditSelectsEvidenceStandardsWithoutDelivery(t *testing.T) {
	root := installedStandardsRoot()
	paths, err := standaloneAuditStandards(root, []string{"project.go"})
	if err != nil {
		t.Fatal(err)
	}
	if paths[0] != "project.go" {
		t.Fatal("project input was not preserved")
	}
	for _, name := range []string{"INDEPENDENT-AUDITOR.md", "READING.md", "REVIEW.md", "CODING.md", "TESTING.md", "technologies/GO.md"} {
		if !strings.Contains(strings.Join(paths, "\n"), filepath.Join(root, name)) {
			t.Fatalf("standalone inventory omitted %s", name)
		}
	}
	for _, name := range []string{"AUDIT.md", "STANDALONE-AUDIT.md", "SPIKE-AUDIT.md", "MAIN.md", "DELIVERY.md", "ADVICE.md", "AUDITS.md", "SELF-AUDIT.md", "EXTERNAL-AUDIT.md", "HARNESS.md", "technologies/RUST.md", "MARKDOWN.md"} {
		if strings.Contains(strings.Join(paths, "\n"), filepath.Join(root, name)) {
			t.Errorf("standalone review unnecessarily loaded %s", name)
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

func TestReviewSelectionCombinesDeclaredAndEvidenceStandards(t *testing.T) {
	root := installedStandardsRoot()
	paths, err := spikeAuditStandards(root, []string{"experiment.rs", "brief.org"}, "NODE", "HUGO")
	if err != nil {
		t.Fatal(err)
	}
	present := map[string]bool{}
	for _, path := range paths {
		if present[path] {
			t.Fatalf("duplicate review input %s", path)
		}
		present[path] = true
	}
	for _, name := range []string{"technologies/NODE.md", "technologies/JAVASCRIPT.md", "technologies/HUGO.md", "technologies/WEB.md", "technologies/RUST.md", "DOCUMENTATION.md", "ORGMODE.md"} {
		if !present[filepath.Join(root, name)] {
			t.Errorf("review selection omitted applicable standard %s", name)
		}
	}
	if present[filepath.Join(root, "MAIN.md")] || present[filepath.Join(root, "technologies/GO.md")] {
		t.Fatal("review selection loaded workflow or an unrelated technology")
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
