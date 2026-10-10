package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runFixture injects an on-disk prompt registry at the filesystem boundary.
// Public audit invocations cannot override their canonical routing documents.
func runFixture(arguments []string, input io.Reader, output, diagnostics io.Writer) error {
	if err := prepareHermesSDKFixture(); err != nil {
		return err
	}
	var args []string
	registry, project := "", "."
	spike := false
	for i := 0; i < len(arguments); i++ {
		if arguments[i] == "--audit-prompts" && i+1 < len(arguments) {
			registry = arguments[i+1]
			i++
			continue
		}
		if arguments[i] == "--project" && i+1 < len(arguments) {
			project = arguments[i+1]
		}
		if arguments[i] == "--agent-spike" {
			spike = true
		}
		args = append(args, arguments[i])
	}
	if registry == "" {
		return run(args, input, output, diagnostics)
	}
	old := auditPromptPath
	auditPromptPath = func(string) string { return registry }
	defer func() { auditPromptPath = old }()
	initialized, err := profilePresent(project)
	if err != nil {
		return err
	}
	if initialized && !spike {
		profile := filepath.Join(project, ".sdlc", "project.yaml")
		if _, err := os.Stat(profile); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(profile), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(profile, []byte("version: 3\nstandards:\n  technologies: []\n"), 0o600); err != nil {
				return err
			}
		}
	}
	return run(args, input, output, diagnostics)
}

// The fake uv process models the installed-runtime boundary while existing
// synthetic Hermes providers continue to exercise model/provider/session args.
// No Python interpreter or external model is used by these Go CLI fixtures.
func prepareHermesSDKFixture() error {
	path, err := exec.LookPath("hermes")
	if err != nil {
		return nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(string(body), "#!/bin/sh\n") {
		return nil
	}
	root := filepath.Dir(filepath.Dir(path))
	if !strings.HasPrefix(root, os.TempDir()) {
		return nil
	}
	entry := filepath.Join(root, "libexec", "bin", "hermes")
	if err := os.MkdirAll(filepath.Dir(entry), 0700); err != nil {
		return err
	}
	interpreter := filepath.Join(root, "libexec", "bin", "python")
	if err := os.WriteFile(interpreter, []byte("fixture runtime"), 0600); err != nil {
		return err
	}
	if err := os.WriteFile(entry, []byte("#!"+interpreter+"\nfrom hermes_cli.main import main\n"), 0600); err != nil {
		return err
	}
	// #nosec G306 -- executable synthetic uv boundary, entirely fixture-owned.
	return os.WriteFile(filepath.Join(filepath.Dir(path), "uv"), []byte("#!/bin/sh\nexec \"$(dirname \"$0\")/hermes\" \"$@\"\n"), 0700)
}

func TestMain(m *testing.M) {
	protected, err := os.MkdirTemp("", "sdlc-protected-test-")
	if err != nil {
		os.Exit(1)
	}
	// Existing CLI fixtures occupy the real system temp directory. The policy
	// remains active against this separate synthetic root during those tests.
	temporaryAuditRoots = func() ([]string, error) { return []string{protected}, nil }
	profilePresent = func(string) (bool, error) { return true, nil }
	enforceAuditAnchor = func(project string) (string, error) { return filepath.Abs(project) }
	inventoryProject = func(string) (string, error) { return "\n\nInvocation project file inventory:\n- fixture\n", nil }
	installedStandardsRoot = func() string {
		root, err := filepath.Abs(filepath.Join("..", "..", "src"))
		if err != nil {
			return ""
		}
		return root
	}
	code := m.Run()
	_ = os.RemoveAll(protected) // TestMain removes only the temp directory it created.
	os.Exit(code)
}

func TestAuditAnchorUsesInvocationGitRoot(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := verifyAuditAnchor(".")
	if err != nil || got != want {
		t.Fatalf("invocation root = %q, %v; want %q", got, err, want)
	}
	if _, err := verifyAuditAnchor(t.TempDir()); err == nil {
		t.Fatal("accepted a caller-selected project outside the invocation Git tree")
	}
}

func TestAuditProjectPathsRejectEscapesAndScratchDirectories(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		filepath.Join(root, ".codex", "copy.go"),
		filepath.Join(root, "module", ".hermes", "copy.go"),
		filepath.Join(root, "tmp", "copy.go"),
		filepath.Join(root, "temp-copy", "copy.go"),
		filepath.Join(root, "tmp.backup", "copy.go"),
		filepath.Join(root, ".temporary", "copy.go"),
		filepath.Join(filepath.Dir(root), "outside.go"),
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := requireProjectFile(root, path); err == nil {
			t.Errorf("accepted excluded audit path %q", path)
		}
	}
	for _, path := range []string{
		filepath.Join(root, ".github", "workflows", "test.yaml"),
		filepath.Join(root, "templates", "page.go"),
		filepath.Join(root, "tests", "tmpfile.go"),
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := requireProjectFile(root, path); err != nil {
			t.Errorf("rejected legitimate project path %q: %v", path, err)
		}
	}
}

func TestAuditProjectPathRejectsAliasIntoRuntimeDirectory(t *testing.T) {
	root := t.TempDir()
	runtime := filepath.Join(root, ".codex")
	if err := os.Mkdir(runtime, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtime, "copy.go"), []byte("package copy\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(runtime, filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	if err := requireProjectFile(root, filepath.Join(root, "alias", "copy.go")); err == nil {
		t.Fatal("accepted symlink alias into a provider runtime directory")
	}
}

func TestProjectInventoryExposesLegitimateHiddenFilesButNotRuntimeCopies(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{".github/workflows/test.yaml", "source.go", ".codex/copy.go", "tmp/copy.go"} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	inventory, err := auditProjectInventory(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(inventory, ".github/workflows/test.yaml") || !strings.Contains(inventory, "source.go") {
		t.Fatalf("inventory omits project files: %q", inventory)
	}
	if strings.Contains(inventory, "copy.go") {
		t.Fatalf("inventory includes runtime or scratch copies: %q", inventory)
	}
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
	err := runFixture([]string{"--project", project, "--gate", "definition", "--audit-record", "audits.yaml", "--work-item", "W001", "--input", input}, strings.NewReader(""), &output, &output)
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
	err := runFixture([]string{"--project", project, "--gate", "definition", "--audit-record", "audits.yaml", "--work-item", "W001", "--input", input}, strings.NewReader(""), &output, &output)
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
	err := runFixture([]string{"--project", project, "--input", "alias.go"}, strings.NewReader(""), &output, &output)
	if err == nil || !strings.Contains(err.Error(), "outside the invocation project") {
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
	// #nosec G306 -- private executable fixture never invokes a hosted provider.
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STANDALONE_PROBE", probe)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	oldProfile := profilePresent
	profilePresent = func(string) (bool, error) { return false, nil }
	t.Cleanup(func() { profilePresent = oldProfile })
	oldInventory := inventoryProject
	inventoryProject = auditProjectInventory
	t.Cleanup(func() { inventoryProject = oldInventory })
	args := []string{"--project", project, "--input", "go.mod", "--audit-prompts", filepath.Join(project, "src", "prompts", "audits.yaml"), "--global-config", filepath.Join(project, "absent-global.yaml"), "--harness", "codex", "--model", "fixture"}
	var output, diagnostics bytes.Buffer
	if err := runFixture(args, strings.NewReader("Review the supplied project files."), &output, &diagnostics); err != nil {
		t.Fatalf("standalone audit failed: %v\n%s", err, diagnostics.String())
	}
	if !strings.Contains(diagnostics.String(), "STANDALONE MODE") || !strings.Contains(output.String(), "AUDIT: standalone") {
		t.Fatalf("standalone output = %q; diagnostics = %q", output.String(), diagnostics.String())
	}
	prompt, err := os.ReadFile(probe)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"STANDALONE-AUDIT.md", "CODING.md", "DOCUMENTATION.md", "ORG-SCHEMA.md", "MARKDOWN.md", "go.mod"} {
		if !bytes.Contains(prompt, []byte(name)) {
			t.Fatalf("standalone prompt omitted %s", name)
		}
	}
	if bytes.Contains(prompt, []byte("Audit gate:")) {
		t.Fatal("standalone prompt contains an SDLC gate")
	}
	if !bytes.Contains(prompt, []byte("Invocation project file inventory")) || !bytes.Contains(prompt, []byte("README.md")) {
		t.Fatal("standalone prompt omitted the project inventory")
	}
}
