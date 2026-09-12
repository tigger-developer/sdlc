package projectinit

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// RT013.5 uses real local Git branches, never a provider or a client project.
func TestResumeAdvancedPrimary(t *testing.T) {
	for _, tc := range []struct {
		name, primary, answer string
		onMigration, refuse   bool
	}{
		{"migration", "master", "", true, false},
		{"master-yes", "master", "y\n", false, false},
		{"main-yes", "main", "yes\n", false, false},
		{"master-no", "master", "n\n", false, true},
		{"master-default-no", "master", "\n", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project, migration := resumeFixture(t, tc.primary)
			if tc.onMigration {
				runGitTest(t, project, "switch", migration)
			}
			var output bytes.Buffer
			options := defaultOptions(Options{Input: strings.NewReader(tc.answer), Output: &output, ErrorOutput: &output, RunCommand: localOnlyTestRunner})
			state, err := interruptedMigration(options, project, initializationWorkspacePath(project))
			if tc.refuse {
				if err == nil || state != nil {
					t.Fatalf("declined recovery = %#v, %v", state, err)
				}
				if branch := strings.TrimSpace(runGitTest(t, project, "branch", "--show-current")); branch != tc.primary {
					t.Fatalf("decline changed branch to %s", branch)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if state == nil || state.base != tc.primary || state.migration != migration {
					t.Fatalf("recovery = %#v", state)
				}
				if branch := strings.TrimSpace(runGitTest(t, project, "branch", "--show-current")); branch != migration {
					t.Fatalf("recovery branch = %s", branch)
				}
			}
			if strings.Contains(output.String(), "[y/N]") == tc.onMigration {
				t.Fatalf("confirmation output = %s", output.String())
			}
		})
	}
}

func TestResumeRejectsAmbiguityAndConflictingEdits(t *testing.T) {
	for _, scenario := range []string{"ambiguous", "conflicting-edit", "unrelated-archive"} {
		t.Run(scenario, func(t *testing.T) {
			project, migration := resumeFixture(t, "master")
			switch scenario {
			case "ambiguous":
				runGitTest(t, project, "branch", migration+"-2")
			case "conflicting-edit":
				writeProjectTestFile(t, filepath.Join(project, "README.md"), "primary advanced\n")
				runGitTest(t, project, "add", "README.md")
				runGitTest(t, project, "commit", "-m", "primary-only change")
				writeProjectTestFile(t, filepath.Join(project, "README.md"), "operator edit\n")
			case "unrelated-archive":
				runGitTest(t, project, "switch", "--orphan", "unrelated")
				writeProjectTestFile(t, filepath.Join(project, "OTHER.md"), "different history\n")
				runGitTest(t, project, "add", "OTHER.md")
				runGitTest(t, project, "commit", "-m", "unrelated baseline")
				runGitTest(t, project, "branch", "-f", "sdlc_v1_state_2026-09-12", "HEAD")
				runGitTest(t, project, "switch", "master")
			}
			var output bytes.Buffer
			state, err := interruptedMigration(defaultOptions(Options{Input: strings.NewReader("y\n"), Output: &output, ErrorOutput: &output, RunCommand: localOnlyTestRunner}), project, initializationWorkspacePath(project))
			if err == nil || state != nil {
				t.Fatalf("unsafe recovery = %#v, %v", state, err)
			}
			if branch := strings.TrimSpace(runGitTest(t, project, "branch", "--show-current")); branch != "master" {
				t.Fatalf("failed recovery changed branch to %s", branch)
			}
			if !exists(initializationWorkspacePath(project)) {
				t.Fatal("recovery workspace lost")
			}
			if scenario == "conflicting-edit" {
				data, readErr := readRegularFile(filepath.Join(project, "README.md"))
				if readErr != nil || string(data) != "operator edit\n" {
					t.Fatalf("operator edit changed: %q, %v", data, readErr)
				}
			}
		})
	}
}

func resumeFixture(t *testing.T, primary string) (string, string) {
	t.Helper()
	project := t.TempDir()
	runGitTest(t, project, "init", "-b", primary)
	runGitTest(t, project, "config", "user.name", "Test Operator")
	runGitTest(t, project, "config", "user.email", "operator@example.invalid")
	writeProjectTestFile(t, filepath.Join(project, "README.md"), "initial\n")
	runGitTest(t, project, "add", "README.md")
	runGitTest(t, project, "commit", "-m", "initial")
	runGitTest(t, project, "branch", "sdlc_v1_state_2026-09-12")
	migration := "sdlc-v3-migration-2026-09-12"
	runGitTest(t, project, "switch", "-c", migration)
	writeProjectTestFile(t, filepath.Join(project, "README.md"), "migration progress\n")
	runGitTest(t, project, "add", "README.md")
	runGitTest(t, project, "commit", "-m", "migration progress")
	runGitTest(t, project, "switch", primary)
	runGitTest(t, project, "merge", "--ff-only", migration)
	writeProjectTestFile(t, filepath.Join(initializationWorkspacePath(project), "checkpoint.yaml"), "retained\n")
	return project, migration
}
