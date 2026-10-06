package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	protected, err := os.MkdirTemp("", "sdlc-protected-test-")
	if err != nil {
		os.Exit(1)
	}
	// Existing CLI fixtures occupy the real system temp directory. The policy
	// remains active against this separate synthetic root during those tests.
	temporaryAuditRoots = func() ([]string, error) { return []string{protected}, nil }
	profilePresent = func(string) (bool, error) { return true, nil }
	code := m.Run()
	_ = os.RemoveAll(protected) // TestMain removes only the temp directory it created.
	os.Exit(code)
}

func TestTemporaryAuditPathRefusedBeforeProvider(t *testing.T) {
	root := t.TempDir()
	protected := filepath.Join(root, "protected")
	if err := os.Mkdir(protected, 0o700); err != nil {
		t.Fatal(err)
	}
	oldRoots := temporaryAuditRoots
	temporaryAuditRoots = func() ([]string, error) { return []string{protected}, nil }
	t.Cleanup(func() { temporaryAuditRoots = oldRoots })
	project := filepath.Join(protected, "project")
	if err := os.Mkdir(project, 0o700); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(project, "code.go")
	if err := os.WriteFile(input, []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err := run([]string{"--project", project, "--gate", "definition", "--audit-record", "audits.yaml", "--work-item", "W001", "--input", input}, strings.NewReader(""), &output, &output)
	if err == nil || !strings.Contains(err.Error(), "temporary project or evidence") {
		t.Fatalf("temporary project audit error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, "audits.yaml")); !os.IsNotExist(err) {
		t.Fatalf("refusal created an audit record: %v", err)
	}
}

func TestTemporaryEvidenceRefusedBeforeProvider(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	protected := filepath.Join(root, "protected")
	for _, path := range []string{project, protected} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	input := filepath.Join(protected, "copy.go")
	if err := os.WriteFile(input, []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldRoots := temporaryAuditRoots
	temporaryAuditRoots = func() ([]string, error) { return []string{protected}, nil }
	t.Cleanup(func() { temporaryAuditRoots = oldRoots })
	var output bytes.Buffer
	err := run([]string{"--project", project, "--gate", "definition", "--audit-record", "audits.yaml", "--work-item", "W001", "--input", input}, strings.NewReader(""), &output, &output)
	if err == nil || !strings.Contains(err.Error(), "temporary project or evidence") {
		t.Fatalf("temporary evidence audit error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, "audits.yaml")); !os.IsNotExist(err) {
		t.Fatalf("refusal created an audit record: %v", err)
	}
}

func TestTemporaryAuditPathRejectsBothSymlinkDirections(t *testing.T) {
	root := t.TempDir()
	protected := filepath.Join(root, "protected")
	original := filepath.Join(root, "original")
	for _, path := range []string{protected, original} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	aliasInTemp := filepath.Join(protected, "alias")
	aliasToTemp := filepath.Join(original, "alias")
	if err := os.Symlink(original, aliasInTemp); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(protected, aliasToTemp); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{protected, aliasInTemp, aliasToTemp} {
		if err := rejectTemporaryAuditPath(path, []string{protected}); err == nil {
			t.Errorf("accepted temporary path %s", path)
		}
	}
	if err := rejectTemporaryAuditPath(original, []string{protected}); err != nil {
		t.Fatalf("rejected original path: %v", err)
	}
}

func TestRealSystemTemporaryPathIsRejected(t *testing.T) {
	if err := rejectTemporaryAuditPath(t.TempDir(), []string{os.TempDir()}); err == nil {
		t.Fatal("accepted a project under the operating system temporary directory")
	}
}

func TestStandaloneRejectsInputEscapingProject(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	outside := filepath.Join(root, "outside.go")
	if err := os.Mkdir(project, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("package outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(project, "alias.go")); err != nil {
		t.Fatal(err)
	}
	oldProfile := profilePresent
	profilePresent = func(string) (bool, error) { return false, nil }
	t.Cleanup(func() { profilePresent = oldProfile })
	var output bytes.Buffer
	err := run([]string{"--project", project, "--input", "alias.go"}, strings.NewReader(""), &output, &output)
	if err == nil || !strings.Contains(err.Error(), "outside the selected project") {
		t.Fatalf("escaped input error = %v", err)
	}
}

func TestStandaloneVerdictFormat(t *testing.T) {
	for _, body := range []string{
		"AUDIT: standalone\nREVISION: candidate\nVERDICT: PASS\n",
		"AUDIT: standalone\nREVISION: candidate\nVERDICT: FAIL\nFinding\n",
	} {
		if err := validateStandaloneVerdict(body); err != nil {
			t.Fatal(err)
		}
	}
	for _, body := range []string{
		"GATE: delivery-code\nREVISION: candidate\nVERDICT: PASS\n",
		"AUDIT: standalone\nREVISION: candidate\nVERDICT: PASS\nVERDICT: FAIL\n",
	} {
		if err := validateStandaloneVerdict(body); err == nil {
			t.Fatalf("accepted invalid standalone result %q", body)
		}
	}
}

func TestStandaloneAuditReviewsProjectWithInstalledStandards(t *testing.T) {
	project, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	probe := filepath.Join(bin, "prompt")
	script := `#!/bin/sh
output=
previous=
for argument in "$@"; do
  if [ "$previous" = "-o" ]; then output="$argument"; fi
  previous="$argument"
done
/bin/cat > "$STANDALONE_PROBE"
printf 'AUDIT: standalone\nREVISION: candidate\nVERDICT: PASS\n' > "$output"
printf '{"type":"thread.started","thread_id":"standalone-fixture"}\n'
`
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STANDALONE_PROBE", probe)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	oldProfile := profilePresent
	profilePresent = func(string) (bool, error) { return false, nil }
	t.Cleanup(func() { profilePresent = oldProfile })
	args := []string{"--project", project, "--input", "go.mod", "--audit-prompts", filepath.Join(project, "src", "prompts", "audits.yaml"), "--global-config", filepath.Join(project, "absent-global.yaml"), "--harness", "codex", "--model", "fixture"}
	var output, diagnostics bytes.Buffer
	if err := run(args, strings.NewReader("Review the supplied project files."), &output, &diagnostics); err != nil {
		t.Fatalf("standalone audit failed: %v\n%s", err, diagnostics.String())
	}
	if !strings.Contains(diagnostics.String(), "STANDALONE MODE") || !strings.Contains(output.String(), "AUDIT: standalone") {
		t.Fatalf("standalone output = %q; diagnostics = %q", output.String(), diagnostics.String())
	}
	prompt, err := os.ReadFile(probe)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"STANDALONE-AUDIT.md", "CODING.md", "DOCUMENTATION.md", "go.mod"} {
		if !bytes.Contains(prompt, []byte(name)) {
			t.Fatalf("standalone prompt omitted %s", name)
		}
	}
	if bytes.Contains(prompt, []byte("Audit gate:")) {
		t.Fatal("standalone prompt contains an SDLC gate")
	}
}
