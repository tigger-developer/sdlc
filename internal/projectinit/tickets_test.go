// ABOUTME: Exercises v3 ticket-directory migration through the initializer.
// ABOUTME: Protects existing evidence, links and refusal-before-mutation behaviour.
package projectinit

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ticketProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGitTest(t, root, "init")
	runGitTest(t, root, "config", "user.name", "Test")
	runGitTest(t, root, "config", "user.email", "test@example.invalid")
	writeProjectTestFile(t, filepath.Join(root, projectProfilePath), "version: 3\nproject:\n  primary_branch: master\n  custom: preserved\n")
	writeProjectTestFile(t, filepath.Join(root, "docs/work.org"), "* Work items\n** TODO W123 - Existing\n[[file:../specs/W123-existing/spec.org][Existing]]\n** TODO D124 - Defect\n** TODO S125 - Spike\n")
	writeProjectTestFile(t, filepath.Join(root, "specs/W123-existing/spec.org"), "#+TITLE: Existing\n[[file:../../README.md][Project]]\n[[file:../../specs/W123-existing/spec.org][Self]]\n")
	writeProjectTestFile(t, filepath.Join(root, "specs/W123-existing/audits.yaml"), "historical: specs/W123-existing/spec.org\n")
	writeProjectTestFile(t, filepath.Join(root, "README.md"), "See [existing](specs/W123-existing/spec.org).\nHistorical: docs/archive/specs/old.md\n")
	runGitTest(t, root, "add", ".")
	runGitTest(t, root, "commit", "-m", "baseline")
	return root
}

func TestRunMigratesV3SpecsToTickets(t *testing.T) {
	root := ticketProject(t)
	profile, _ := os.ReadFile(filepath.Join(root, projectProfilePath))
	var output bytes.Buffer
	usedGitMove := false
	runner := func(name string, args []string, dir string, input io.Reader, out, errOut io.Writer) error {
		if name != "git" {
			t.Fatalf("v3 migration invoked non-Git command %s", name)
		}
		if len(args) > 0 && args[0] == "mv" {
			usedGitMove = true
		}
		return localOnlyTestRunner(name, args, dir, input, out, errOut)
	}
	opts := Options{ProjectRoot: root, SDLCRoot: filepath.Join(root, "deliberately-unavailable-framework"), Input: strings.NewReader("n\n"), Output: &output, ErrorOutput: &output, RunCommand: runner}
	if err := Run(opts); err != nil {
		t.Fatal(err)
	}
	if !usedGitMove {
		t.Fatal("migration did not use git mv")
	}
	if exists(filepath.Join(root, "specs")) || !exists(filepath.Join(root, "tickets/W123-existing/spec.org")) {
		t.Fatal("ticket root not migrated")
	}
	for _, p := range []string{"docs/work.org", "README.md", "tickets/W123-existing/spec.org"} {
		b, err := os.ReadFile(filepath.Join(root, p))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), "tickets/W123-existing/spec.org") {
			t.Fatalf("unrepaired link in %s: %s", p, b)
		}
	}
	b, _ := os.ReadFile(filepath.Join(root, projectProfilePath))
	if !bytes.Equal(profile, b) {
		t.Fatal("profile replaced")
	}
	b, _ = os.ReadFile(filepath.Join(root, "tickets/W123-existing/audits.yaml"))
	if string(b) != "historical: tickets/W123-existing/spec.org\n" {
		t.Fatal("audit reference not repaired")
	}
	b, _ = os.ReadFile(filepath.Join(root, "README.md"))
	if !strings.Contains(string(b), "docs/archive/tickets/old.md") {
		t.Fatal("reference not replaced")
	}
	if got := highestHistoricalWorkNumber(root); got != 125 {
		t.Fatalf("shared namespace highest=%d", got)
	}
	if got := strings.TrimSpace(runGitTest(t, root, "status", "--porcelain")); got != "" {
		t.Fatalf("uncommitted migration: %s", got)
	}
	if err := Run(opts); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
}

func TestTicketMigrationRefusesConflictingRoot(t *testing.T) {
	root := ticketProject(t)
	writeProjectTestFile(t, filepath.Join(root, "tickets/D126-conflict/defect.org"), "existing\n")
	runGitTest(t, root, "add", ".")
	runGitTest(t, root, "commit", "-m", "existing tickets")
	before := runGitTest(t, root, "rev-parse", "HEAD")
	err := Run(Options{ProjectRoot: root, Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}})
	if err == nil {
		t.Fatal("accepted conflicting roots")
	}
	if !exists(filepath.Join(root, "specs/W123-existing/spec.org")) || runGitTest(t, root, "rev-parse", "HEAD") != before {
		t.Fatal("mutated refused migration")
	}
}

func TestTicketMigrationRejectsDirtyTreeBeforeMoving(t *testing.T) {
	root := ticketProject(t)
	writeProjectTestFile(t, filepath.Join(root, "README.md"), "operator change\n")
	err := Run(Options{ProjectRoot: root, Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}})
	if err == nil || !exists(filepath.Join(root, "specs")) || exists(filepath.Join(root, "tickets")) {
		t.Fatalf("dirty migration mutated state: %v", err)
	}
}

func TestTicketMigrationIncludesYAMLAndExcludesOtherFiles(t *testing.T) {
	root := ticketProject(t)
	writeProjectTestFile(t, filepath.Join(root, "links.yaml"), "path: specs/W123-existing/spec.org\n")
	writeProjectTestFile(t, filepath.Join(root, "script.sh"), "printf 'specs/ is untouched here'\n")
	runGitTest(t, root, "add", ".")
	runGitTest(t, root, "commit", "-m", "other formats")
	if err := Run(Options{ProjectRoot: root, Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(root, "links.yaml"))
	if string(b) != "path: tickets/W123-existing/spec.org\n" {
		t.Fatal(string(b))
	}
	b, _ = os.ReadFile(filepath.Join(root, "script.sh"))
	if !strings.Contains(string(b), "specs/") {
		t.Fatal("rewrote out-of-scope format")
	}
}
