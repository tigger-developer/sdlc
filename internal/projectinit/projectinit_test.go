package projectinit

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoverTechnologiesIsAlphabeticalAndAutomatic(t *testing.T) {
	directory := t.TempDir()
	writeTestFile(t, filepath.Join(directory, "WEB.md"), "# Web interfaces\n")
	writeTestFile(t, filepath.Join(directory, "GO.md"), "# Go\n")
	writeTestFile(t, filepath.Join(directory, "notes.txt"), "ignored\n")

	got, err := DiscoverTechnologies(directory)
	if err != nil {
		t.Fatal(err)
	}
	want := []Technology{
		{Name: "GO", Title: "Go", Path: filepath.Join(directory, "GO.md")},
		{Name: "WEB", Title: "Web interfaces", Path: filepath.Join(directory, "WEB.md")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DiscoverTechnologies() = %#v, want %#v", got, want)
	}

	writeTestFile(t, filepath.Join(directory, "RUST.md"), "# Rust\n")
	got, err = DiscoverTechnologies(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[1].Name != "RUST" {
		t.Fatalf("new technology was not discovered without code changes: %#v", got)
	}
}

func TestResolveConfigUsesCLIProjectUserPrecedence(t *testing.T) {
	user := map[string]string{keyAgentHarness: "hermes", keyAuditModel: "user-audit"}
	project := map[string]string{keyAgentHarness: "claude", keyAuditModel: "project-audit"}
	got := resolveConfig(Options{Harness: "codex"}, user, project)
	if got.Harness != "codex" || got.AuditModel != "project-audit" {
		t.Fatalf("resolved config = %#v", got)
	}
}

func TestProviderAndModelMustBeConfiguredTogether(t *testing.T) {
	if err := validateProviderModelPairs(resolvedConfig{AuditModel: "fast-model"}); err == nil {
		t.Fatal("audit model without provider was accepted")
	}
	if err := validateProviderModelPairs(resolvedConfig{DeliveryProvider: "provider", DeliveryModel: "model", AuditProvider: "reviewer", AuditModel: "fast-model"}); err != nil {
		t.Fatal(err)
	}
}

func TestRenderConstitutionIncludesUniversalAndSelectedStandardsOnce(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "standards")
	technologies := []Technology{{Name: "GO", Title: "Go engineering", Path: filepath.Join(root, "technologies", "GO.md")}}
	first := renderConstitution(root, technologies, resolvedConfig{})
	second := renderConstitution(root, technologies, resolvedConfig{})
	if !bytes.Equal(first, second) {
		t.Fatal("rendering is not deterministic")
	}
	text := string(first)
	for _, item := range universalStandards {
		if count := strings.Count(text, filepath.Join(root, item.Path)); count != 1 {
			t.Errorf("%s occurs %d times, want once", item.Path, count)
		}
	}
	if count := strings.Count(text, technologies[0].Path); count != 1 {
		t.Errorf("selected technology occurs %d times, want once", count)
	}
	if strings.Contains(text, "External Infrastructure Ownership") {
		t.Error("external integration was rendered while disabled")
	}
}

func TestRenderConstitutionIncludesExternalIntegration(t *testing.T) {
	enabled := true
	text := string(renderConstitution("/standards", nil, resolvedConfig{
		InfraEnabled: &enabled, InfraOwner: "Fleet project", InfraContract: "/fleet/INTEGRATION.md",
	}))
	for _, expected := range []string{"Fleet project", "/fleet/INTEGRATION.md", "Conflicting historical deployment requirements"} {
		if !strings.Contains(text, expected) {
			t.Errorf("rendered constitution omitted %q", expected)
		}
	}
}

func TestRenderConstitutionNamesAuditProviderWithModel(t *testing.T) {
	text := string(renderConstitution("/standards", nil, resolvedConfig{AuditProvider: "openai-codex", AuditModel: "gpt-5.6-luna"}))
	if !strings.Contains(text, "provider `openai-codex`, model `gpt-5.6-luna`") {
		t.Fatalf("rendered constitution omitted paired audit runtime:\n%s", text)
	}
}

func TestRunIsIdempotentAndDoesNotRelaunch(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	writeTestFile(t, filepath.Join(project, ".specify", "marker"), "present\n")
	writeTestFile(t, filepath.Join(project, ".specify", "presets", "sdlc-standards", "preset.yml"), "present\n")
	writeTestFile(t, filepath.Join(root, "technologies", "GO.md"), "# Go\n")
	writeTestFile(t, filepath.Join(root, "presets", "sdlc-standards", "preset.yml"), "present\n")
	disabled := false

	options := Options{
		ProjectRoot: project, SDLCRoot: root, UserConfigPath: filepath.Join(t.TempDir(), ".env"), NoLaunch: true,
		Harness: "codex", Technologies: []string{"GO"}, InfraEnabled: &disabled,
		Input: strings.NewReader(""), Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{},
	}
	if err := Run(options); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(project, ".specify", "templates", "overrides", "constitution-template.md")
	first, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	options.Output = &output
	options.Harness = ""
	options.Technologies = nil
	options.InfraEnabled = nil
	if err := Run(options); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) || output.Len() != 0 {
		t.Fatalf("idempotent rerun changed output: output=%q", output.String())
	}
	if _, err := os.Stat(filepath.Join(project, ".specify", "templates", "constitution-template.md")); !os.IsNotExist(err) {
		t.Fatalf("initializer wrote the Spec Kit core template: %v", err)
	}
}

func TestWriteManagedEnvPreservesUnmanagedValuesAndGitignore(t *testing.T) {
	directory := t.TempDir()
	envPath := filepath.Join(directory, ".env")
	writeTestFile(t, envPath, "PRIVATE_TOKEN=keep\nSDLC_AGENT_HARNESS=claude\n")
	if err := writeManagedEnv(envPath, map[string]string{keyAgentHarness: "codex"}); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "PRIVATE_TOKEN=keep") || !strings.Contains(string(contents), "SDLC_AGENT_HARNESS=\"codex\"") || strings.Contains(string(contents), "SDLC_AGENT_HARNESS=claude") {
		t.Fatalf("managed environment update was incorrect:\n%s", contents)
	}
	ignorePath := filepath.Join(directory, ".gitignore")
	if changed, err := ensureEnvIgnored(ignorePath); err != nil || !changed {
		t.Fatal(err)
	}
	if changed, err := ensureEnvIgnored(ignorePath); err != nil || changed {
		t.Fatal(err)
	}
	ignored, err := os.ReadFile(ignorePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(ignored) != ".env\n" {
		t.Fatalf(".gitignore = %q", ignored)
	}
}

func TestEnsurePresetBacksUpAndUpdatesChangedCopy(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	source := filepath.Join(root, "presets", "sdlc-standards")
	destination := filepath.Join(project, ".specify", "presets", "sdlc-standards")
	writeTestFile(t, filepath.Join(source, "preset.yml"), "version: 2\n")
	writeTestFile(t, filepath.Join(destination, "preset.yml"), "version: 1\n")
	writeTestFile(t, filepath.Join(destination, "project-edit.md"), "preserve in backup\n")

	runner := func(_ string, arguments []string, _ string, _ io.Reader, _, _ io.Writer) error {
		switch arguments[1] {
		case "remove":
			return os.RemoveAll(destination)
		case "add":
			return copyDirectory(source, destination)
		default:
			t.Fatalf("unexpected preset command: %v", arguments)
			return nil
		}
	}
	if err := ensurePreset(project, root, Options{Output: &bytes.Buffer{}, RunCommand: runner}); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(filepath.Join(destination, "preset.yml"))
	if err != nil || string(updated) != "version: 2\n" {
		t.Fatalf("updated preset = %q, %v", updated, err)
	}
	backups, err := filepath.Glob(destination + ".*.bak")
	if err != nil || len(backups) != 1 {
		t.Fatalf("preset backups = %v, %v", backups, err)
	}
	preserved, err := os.ReadFile(filepath.Join(backups[0], "project-edit.md"))
	if err != nil || string(preserved) != "preserve in backup\n" {
		t.Fatalf("backed-up project edit = %q, %v", preserved, err)
	}
}

func TestDirectoriesEqualIgnoresSpecKitComposedOutput(t *testing.T) {
	source := t.TempDir()
	destination := t.TempDir()
	writeTestFile(t, filepath.Join(source, "preset.yml"), "version: 2\n")
	writeTestFile(t, filepath.Join(destination, "preset.yml"), "version: 2\n")
	writeTestFile(t, filepath.Join(destination, ".composed", "generated.md"), "generated\n")
	equal, err := directoriesEqual(source, destination)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Spec Kit generated .composed output created a preset variance")
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
