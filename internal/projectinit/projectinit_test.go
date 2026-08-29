package projectinit

import (
	"bytes"
	"fmt"
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
	user := map[string]string{keyAgentHarness: "hermes", keySpecModel: "user-spec", keyAuditModel: "user-audit"}
	project := map[string]string{keyAgentHarness: "claude", keySpecModel: "project-spec", keyAuditModel: "project-audit"}
	got := resolveConfig(Options{Harness: "codex", SpecModel: "cli-spec"}, user, project)
	if got.Harness != "codex" || got.SpecModel != "cli-spec" || got.AuditModel != "project-audit" {
		t.Fatalf("resolved config = %#v", got)
	}
}

func TestUserDefaultsLiveUnderCommonAgentsRoot(t *testing.T) {
	home := filepath.Join(string(filepath.Separator), "home", "operator")
	want := filepath.Join(home, ".agents", ".env")
	if got := userDefaultsPath(home); got != want {
		t.Fatalf("userDefaultsPath() = %q, want %q", got, want)
	}
}

func TestProviderAndModelMustBeConfiguredTogether(t *testing.T) {
	if err := validateProviderModelPairs(resolvedConfig{AuditModel: "fast-model"}); err == nil {
		t.Fatal("audit model without provider was accepted")
	}
	if err := validateProviderModelPairs(resolvedConfig{
		SpecProvider: "provider", SpecModel: "spec-model",
		BuildProvider: "provider", BuildModel: "build-model",
		AuditProvider: "reviewer", AuditModel: "fast-model",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyDeliveryDefaultsBecomeSpecificationDefaults(t *testing.T) {
	values := map[string]string{
		legacyDeliveryProvider: "openai-codex",
		legacyDeliveryModel:    "gpt-5.6-sol",
	}
	if !normalizeLegacyDelivery(values) {
		t.Fatal("legacy delivery configuration was not identified")
	}
	if values[keySpecProvider] != "openai-codex" || values[keySpecModel] != "gpt-5.6-sol" {
		t.Fatalf("normalized legacy configuration = %#v", values)
	}
	if values[legacyDeliveryProvider] != "" || values[legacyDeliveryModel] != "" {
		t.Fatalf("legacy keys remained after normalization: %#v", values)
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
	if !strings.Contains(text, "**Last Revised**: [LAST_AMENDED_DATE]") || strings.Contains(text, "**Last Amended**") {
		t.Error("unratified constitution template used amendment terminology")
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

func TestRenderConstitutionPinsSDLCRevision(t *testing.T) {
	text := string(renderConstitution("/standards", nil, resolvedConfig{SDLCRevision: "0123456789abcdef"}))
	if !strings.Contains(text, "The adopted SDLC revision is `0123456789abcdef`.") || strings.Contains(text, "TODO(SDLC_REVISION)") {
		t.Fatalf("rendered constitution did not pin the SDLC revision:\n%s", text)
	}

	unresolved := string(renderConstitution("/standards", nil, resolvedConfig{}))
	if !strings.Contains(unresolved, "TODO(SDLC_REVISION)") {
		t.Fatalf("unversioned constitution omitted the revision TODO:\n%s", unresolved)
	}
}

func TestRenderConstitutionNamesAuditProviderWithModel(t *testing.T) {
	text := string(renderConstitution("/standards", nil, resolvedConfig{AuditProvider: "openai-codex", AuditModel: "gpt-5.6-luna"}))
	if !strings.Contains(text, "provider `openai-codex`, model `gpt-5.6-luna`") {
		t.Fatalf("rendered constitution omitted paired audit runtime:\n%s", text)
	}
}

func TestRunInitializesGreenfieldSpecKitProject(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "technologies", "GO.md"), "# Go\n")
	writeTestFile(t, filepath.Join(root, "presets", "sdlc-standards", "preset.yml"), "present\n")
	disabled := false
	var commands []string
	runner := func(name string, arguments []string, directory string, _ io.Reader, _, _ io.Writer) error {
		commands = append(commands, name+" "+strings.Join(arguments, " "))
		switch {
		case name == "specify" && len(arguments) != 0 && arguments[0] == "init":
			return os.MkdirAll(filepath.Join(directory, ".specify"), 0o700)
		case name == "specify" && len(arguments) >= 2 && arguments[0] == "preset" && arguments[1] == "add":
			return copyDirectory(filepath.Join(root, "presets", "sdlc-standards"), filepath.Join(directory, ".specify", "presets", "sdlc-standards"))
		default:
			return fmt.Errorf("unexpected command: %s %v", name, arguments)
		}
	}
	options := Options{
		ProjectRoot: project, SDLCRoot: root, UserConfigPath: filepath.Join(t.TempDir(), ".env"), NoLaunch: true,
		Harness: "codex", Technologies: []string{"GO"}, InfraEnabled: &disabled, SDLCRevision: "0123456789abcdef",
		Input: strings.NewReader("yes\n"), Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}, RunCommand: runner,
	}
	if err := Run(options); err != nil {
		t.Fatal(err)
	}
	if len(commands) != 2 || !strings.Contains(commands[0], "specify init") || !strings.Contains(commands[1], "specify preset add") {
		t.Fatalf("greenfield commands = %#v", commands)
	}
	constitution, err := os.ReadFile(filepath.Join(project, ".specify", "templates", "overrides", "constitution-template.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(constitution), "The adopted SDLC revision is `0123456789abcdef`.") {
		t.Fatalf("greenfield constitution omitted source revision:\n%s", constitution)
	}
}

func TestRunPreservesBrownfieldProjectAndIsIdempotent(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	writeTestFile(t, filepath.Join(project, "README.md"), "# Existing project\n")
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
	readme, err := os.ReadFile(filepath.Join(project, "README.md"))
	if err != nil || string(readme) != "# Existing project\n" {
		t.Fatalf("brownfield README changed: %q, %v", readme, err)
	}
}

func TestRunSnapshotsEveryResolvedGlobalDefaultIntoProject(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	userConfig := filepath.Join(t.TempDir(), ".env")
	writeTestFile(t, filepath.Join(project, ".specify", "marker"), "present\n")
	writeTestFile(t, filepath.Join(project, ".specify", "presets", "sdlc-standards", "preset.yml"), "present\n")
	writeTestFile(t, filepath.Join(root, "technologies", "GO.md"), "# Go\n")
	writeTestFile(t, filepath.Join(root, "presets", "sdlc-standards", "preset.yml"), "present\n")
	writeTestFile(t, userConfig, strings.Join([]string{
		`SDLC_AGENT_HARNESS="codex"`,
		`SDLC_SPEC_PROVIDER="openai-codex"`,
		`SDLC_SPEC_MODEL="gpt-5.6-sol"`,
		`SDLC_BUILD_PROVIDER="openai-codex"`,
		`SDLC_BUILD_MODEL="gpt-5.6-terra"`,
		`SDLC_AUDIT_PROVIDER="openai-codex"`,
		`SDLC_AUDIT_MODEL="gpt-5.6-luna"`,
		`SDLC_TECHNOLOGIES="GO"`,
		`SDLC_INFRA_ENABLED="true"`,
		`SDLC_INFRA_OWNER="Exodan"`,
		`SDLC_INFRA_CONTRACT="~/code/exodan/deploy/docs/PROJECT-INTEGRATION.md"`,
		"",
	}, "\n"))

	if err := Run(Options{
		ProjectRoot: project, SDLCRoot: root, UserConfigPath: userConfig, NoLaunch: true,
		Input: strings.NewReader(""), Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{},
	}); err != nil {
		t.Fatal(err)
	}
	got, err := readManagedEnv(filepath.Join(project, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := readManagedEnv(userConfig)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("project snapshot = %#v, want global defaults %#v", got, want)
	}

	writeTestFile(t, userConfig, strings.ReplaceAll(string(mustReadFile(t, userConfig)), "gpt-5.6-sol", "new-global-spec"))
	var rerunOutput bytes.Buffer
	if err := Run(Options{
		ProjectRoot: project, SDLCRoot: root, UserConfigPath: userConfig, NoLaunch: true,
		Input: strings.NewReader(""), Output: &rerunOutput, ErrorOutput: &bytes.Buffer{},
	}); err != nil {
		t.Fatal(err)
	}
	afterGlobalChange, err := readManagedEnv(filepath.Join(project, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(afterGlobalChange, got) || rerunOutput.Len() != 0 {
		t.Fatalf("global default change altered existing snapshot: values=%#v output=%q", afterGlobalChange, rerunOutput.String())
	}
}

func TestLaunchConstitutionRequiresProjectWideFiltering(t *testing.T) {
	var command string
	var arguments []string
	var directory string
	runner := func(name string, gotArguments []string, gotDirectory string, _ io.Reader, _, _ io.Writer) error {
		command = name
		arguments = append([]string(nil), gotArguments...)
		directory = gotDirectory
		return nil
	}
	templatePath := filepath.Join("project", ".specify", "templates", "overrides", "constitution-template.md")
	projectRoot := filepath.Join("project")
	options := Options{RunCommand: runner, Input: strings.NewReader(""), Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}}
	if err := launchConstitution(resolvedConfig{Harness: "codex", SpecProvider: "openai-codex", SpecModel: "gpt-5.6-sol"}, projectRoot, templatePath, options); err != nil {
		t.Fatal(err)
	}
	if command != "codex" || directory != projectRoot || len(arguments) != 3 || arguments[0] != "--model" || arguments[1] != "gpt-5.6-sol" {
		t.Fatalf("constitution launch = command %q, arguments %#v, directory %q", command, arguments, directory)
	}
	prompt := arguments[2]
	for _, required := range []string{
		"`" + templatePath + "` as an immutable baseline",
		"This is a filtering exercise, not a summary",
		"It applies across unrelated future features",
		"Changing it would require a constitutional decision",
		"supported by authoritative project documentation",
		"feature requirements and acceptance criteria",
		"migration algorithms, schema procedures, commands, test gates",
		"The constitution may elevate a concise project-wide invariant",
		"Importance alone does not make something constitutional",
		"Add up to four concise project-specific principles",
		"Zero is valid only when no evidenced project-wide invariant",
		"not, by itself, a reason to omit it",
		"Make the authority hierarchy concern-specific",
		"Approved feature specifications govern observable behaviour",
		"Code and tests record implemented state and evidence",
		"Do not place undifferentiated project documentation above approved specifications",
		"The Governance section must name the human ratification and amendment authority",
		"Use a pre-1.0 version for an unratified draft",
		"List every unresolved ratification blocker",
		"Would changing this require amending the constitution?",
		"Produce only `.specify/memory/constitution.md`",
	} {
		if !strings.Contains(prompt, required) {
			t.Errorf("constitution launch prompt omitted %q:\n%s", required, prompt)
		}
	}
}

func TestWriteManagedEnvPreservesUnmanagedValuesAndGitignore(t *testing.T) {
	directory := t.TempDir()
	envPath := filepath.Join(directory, ".env")
	writeTestFile(t, envPath, "PRIVATE_TOKEN=keep\nSDLC_AGENT_HARNESS=claude\nSDLC_DELIVERY_MODEL=legacy-model\n")
	if err := writeManagedEnv(envPath, map[string]string{keyAgentHarness: "codex", keySpecModel: "spec-model"}); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "PRIVATE_TOKEN=keep") || !strings.Contains(string(contents), "SDLC_AGENT_HARNESS=\"codex\"") || !strings.Contains(string(contents), "SDLC_SPEC_MODEL=\"spec-model\"") || strings.Contains(string(contents), "SDLC_AGENT_HARNESS=claude") || strings.Contains(string(contents), "SDLC_DELIVERY_MODEL") {
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

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}
