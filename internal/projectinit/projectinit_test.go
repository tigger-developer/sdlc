package projectinit

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	user := map[string]string{keyAgentHarness: "hermes", keySpecModel: "user-spec", keyAuditModel: "user-audit", keyProjectType: "brownfield"}
	project := map[string]string{keyAgentHarness: "claude", keySpecModel: "project-spec", keyAuditModel: "project-audit", keyProjectType: "greenfield"}
	got := resolveConfig(Options{Harness: "codex", SpecModel: "cli-spec", ProjectType: "brownfield"}, user, project)
	if got.Harness != "codex" || got.SpecModel != "cli-spec" || got.AuditModel != "project-audit" || got.ProjectType != "brownfield" {
		t.Fatalf("resolved config = %#v", got)
	}
	withoutProjectValue := resolveConfig(Options{}, user, map[string]string{})
	if withoutProjectValue.ProjectType != "" {
		t.Fatalf("user project type became a global default: %#v", withoutProjectValue)
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

func TestProjectTypeMustBeGreenfieldOrBrownfield(t *testing.T) {
	for _, accepted := range []string{"greenfield", "BROWNFIELD"} {
		if err := validateProjectType(accepted); err != nil {
			t.Errorf("validateProjectType(%q): %v", accepted, err)
		}
	}
	if err := validateProjectType("existing"); err == nil {
		t.Fatal("unsupported project type was accepted")
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
	first := renderConstitutionForTest(t, root, technologies, resolvedConfig{})
	second := renderConstitutionForTest(t, root, technologies, resolvedConfig{})
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
	text := string(renderConstitutionForTest(t, "/standards", nil, resolvedConfig{
		InfraEnabled: &enabled, InfraOwner: "Fleet project", InfraContract: "/fleet/INTEGRATION.md",
	}))
	for _, expected := range []string{"Fleet project", "/fleet/INTEGRATION.md", "Conflicting historical deployment requirements"} {
		if !strings.Contains(text, expected) {
			t.Errorf("rendered constitution omitted %q", expected)
		}
	}
}

func TestRenderConstitutionPinsSDLCRevision(t *testing.T) {
	text := string(renderConstitutionForTest(t, "/standards", nil, resolvedConfig{SDLCRevision: "0123456789abcdef"}))
	if !strings.Contains(text, "The adopted SDLC revision is `0123456789abcdef`.") || strings.Contains(text, "TODO(SDLC_REVISION)") {
		t.Fatalf("rendered constitution did not pin the SDLC revision:\n%s", text)
	}

	unresolved := string(renderConstitutionForTest(t, "/standards", nil, resolvedConfig{}))
	if !strings.Contains(unresolved, "TODO(SDLC_REVISION)") {
		t.Fatalf("unversioned constitution omitted the revision TODO:\n%s", unresolved)
	}
}

func TestRenderConstitutionOmitsAuditRuntimeAndIncludesGovernance(t *testing.T) {
	text := string(renderConstitutionForTest(t, "/standards", nil, resolvedConfig{AuditProvider: "openai-codex", AuditModel: "gpt-5.6-luna"}))
	for _, runtimeValue := range []string{"openai-codex", "gpt-5.6-luna", "configured audit runtime"} {
		if strings.Contains(text, runtimeValue) {
			t.Fatalf("rendered constitution contains audit runtime %q:\n%s", runtimeValue, text)
		}
	}
	if !strings.Contains(text, "For staged Spec Kit delivery, each audit MUST run in a fresh agent context") {
		t.Fatalf("rendered constitution omitted durable audit governance:\n%s", text)
	}
	if !strings.Contains(text, "explicitly selects paired development under `~/.agents/sdlc/PAIRING.md`") {
		t.Fatalf("rendered constitution omitted paired-development governance:\n%s", text)
	}
}

func TestRenderConstitutionUsesFixedSpecificationBaseline(t *testing.T) {
	greenfield := string(renderConstitutionForTest(t, "/standards", nil, resolvedConfig{ProjectType: "greenfield"}))
	for _, expected := range []string{
		"## Specification Baseline",
		"**Project classification:** Greenfield",
		"No pre-existing requirement baseline existed at ratification",
		"Approved feature specifications establish requirements prospectively",
	} {
		if !strings.Contains(greenfield, expected) {
			t.Errorf("greenfield scaffold omitted %q:\n%s", expected, greenfield)
		}
	}
	if strings.Contains(greenfield, "### Historical Requirement Authority") {
		t.Fatal("greenfield scaffold contains brownfield authority placeholders")
	}

	brownfield := string(renderConstitutionForTest(t, "/standards", nil, resolvedConfig{ProjectType: "brownfield"}))
	for _, expected := range []string{
		"<!-- SYNC IMPACT: [OLD_VERSION] -> [CONSTITUTION_VERSION] | Principles: [CHANGES OR NONE] | Added: [SECTIONS OR NONE] | Removed: [SECTIONS OR NONE] | TODOs: [ITEMS OR NONE] -->",
		"SDLC-GENERATED-SCAFFOLD: editable until ratification",
		"**Project classification:** Brownfield",
		"### Requirement Authority",
		"DISTINGUISH LEGACY-PROCESS RECORDS FROM APPROVED SPEC KIT FEATURE SPECIFICATIONS",
		"### Historical Requirement Authority",
		"### Design Authority",
		"EXCLUDE ARCHIVED IMPLEMENTATION PLANS",
		"### Regression Evidence and Traceability",
		"Tests and code provide evidence of implemented behaviour. They do not approve requirements.",
		"### Precedence and Supersession",
		"Before ratification, this scaffold has no authority",
	} {
		if !strings.Contains(brownfield, expected) {
			t.Errorf("brownfield scaffold omitted %q:\n%s", expected, brownfield)
		}
	}
	if strings.Contains(brownfield, "immutable") || strings.Contains(brownfield, "preserve the generated baseline") {
		t.Fatalf("brownfield scaffold claimed pre-ratification authority:\n%s", brownfield)
	}
}

func TestRunInitializesGreenfieldSpecKitProject(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	installProjectInitResourcesForTest(t, root)
	writeTestFile(t, filepath.Join(root, "technologies", "GO.md"), "# Go\n")
	writeTestFile(t, filepath.Join(root, "presets", "sdlc-standards", "preset.yml"), "present\n")
	disabled := false
	var commands []string
	runner := func(name string, arguments []string, directory string, _ io.Reader, output, _ io.Writer) error {
		commands = append(commands, name+" "+strings.Join(arguments, " "))
		switch {
		case name == "specify" && len(arguments) != 0 && arguments[0] == "init":
			return os.MkdirAll(filepath.Join(directory, ".specify"), 0o700)
		case name == "specify" && len(arguments) >= 2 && arguments[0] == "preset" && arguments[1] == "add":
			return copyDirectory(filepath.Join(root, "presets", "sdlc-standards"), filepath.Join(directory, ".specify", "presets", "sdlc-standards"))
		case name == "git" && len(arguments) != 0 && arguments[0] == "status":
			fmt.Fprintln(output, "?? .specify/templates/overrides/constitution-template.md")
			return nil
		case name == "git" && len(arguments) != 0 && (arguments[0] == "add" || arguments[0] == "commit"):
			return nil
		default:
			return fmt.Errorf("unexpected command: %s %v", name, arguments)
		}
	}
	options := Options{
		ProjectRoot: project, SDLCRoot: root, UserConfigPath: filepath.Join(t.TempDir(), ".env"), NoLaunch: true,
		Harness: "codex", ProjectType: "greenfield", Technologies: []string{"GO"}, InfraEnabled: &disabled, SDLCRevision: "0123456789abcdef",
		Input: strings.NewReader("yes\n"), Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}, RunCommand: runner,
	}
	if err := Run(options); err != nil {
		t.Fatal(err)
	}
	if len(commands) != 5 || !strings.Contains(commands[0], "specify init") || !strings.Contains(commands[1], "specify preset add") ||
		!strings.Contains(commands[2], "git status") || !strings.Contains(commands[3], "git add") || !strings.Contains(commands[4], "git commit --only --quiet") {
		t.Fatalf("greenfield commands = %#v", commands)
	}
	constitution, err := os.ReadFile(filepath.Join(project, ".specify", "templates", "overrides", "constitution-template.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(constitution), "The adopted SDLC revision is `0123456789abcdef`.") {
		t.Fatalf("greenfield constitution omitted source revision:\n%s", constitution)
	}
	if !strings.Contains(string(constitution), "**Project classification:** Greenfield") {
		t.Fatalf("greenfield constitution omitted project classification:\n%s", constitution)
	}
}

func TestRunPreservesBrownfieldProjectAndIsIdempotent(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	installProjectInitResourcesForTest(t, root)
	writeTestFile(t, filepath.Join(project, "README.md"), "# Existing project\n")
	writeTestFile(t, filepath.Join(project, ".specify", "marker"), "present\n")
	writeTestFile(t, filepath.Join(project, ".specify", "presets", "sdlc-standards", "preset.yml"), "present\n")
	writeTestFile(t, filepath.Join(root, "technologies", "GO.md"), "# Go\n")
	writeTestFile(t, filepath.Join(root, "presets", "sdlc-standards", "preset.yml"), "present\n")
	disabled := false

	options := Options{
		ProjectRoot: project, SDLCRoot: root, UserConfigPath: filepath.Join(t.TempDir(), ".env"), NoLaunch: true,
		Harness: "codex", ProjectType: "brownfield", Technologies: []string{"GO"}, InfraEnabled: &disabled,
		Input: strings.NewReader(""), Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}, RunCommand: cleanGitRunner,
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

func TestEnsureManagedPrefixPrependsOpaqueDocumentOnce(t *testing.T) {
	project := t.TempDir()
	block := managedBlockForTest(t)
	target := filepath.Join(project, filepath.FromSlash(legacyACDocumentPath))
	original := []byte("\n## Arbitrary existing content\n\nDo not interpret this document.\n")
	writeTestFile(t, target, string(original))

	changed, err := ensureManagedPrefix(project, legacyACDocumentPath, block)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("legacy block was not prepended")
	}
	first := mustReadFile(t, target)
	want := append(append([]byte(nil), block.content...), '\n')
	want = append(want, original...)
	if !bytes.Equal(first, want) {
		t.Fatalf("legacy block did not preserve opaque content:\n%s", first)
	}

	changed, err = ensureManagedPrefix(project, legacyACDocumentPath, block)
	if err != nil {
		t.Fatal(err)
	}
	if changed || !bytes.Equal(mustReadFile(t, target), first) {
		t.Fatal("marked authority introduction was not idempotent")
	}

	embedded := "Existing text\n\n" + string(block.content) + "\n"
	writeTestFile(t, target, embedded)
	if _, err = ensureManagedPrefix(project, legacyACDocumentPath, block); err == nil {
		t.Fatal("managed marker outside the controlled prefix was accepted")
	}
	if string(mustReadFile(t, target)) != embedded {
		t.Fatal("invalid managed marker placement changed the document")
	}
}

func TestEnsureManagedPrefixReplacesOnlyStalePrefix(t *testing.T) {
	project := t.TempDir()
	block := managedBlockForTest(t)
	target := filepath.Join(project, filepath.FromSlash(legacyACDocumentPath))
	body := []byte("\n# Existing ledger\r\n\r\nOpaque body bytes.\r\n")
	stale := bytes.Replace(block.content, []byte("authoritative"), []byte("previous"), 1)
	contents := append(append(stale, '\n'), body...)
	writeTestFile(t, target, string(contents))

	changed, err := ensureManagedPrefix(project, legacyACDocumentPath, block)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("stale managed prefix was not replaced")
	}
	want := append(append([]byte(nil), block.content...), '\n')
	want = append(want, body...)
	if got := mustReadFile(t, target); !bytes.Equal(got, want) {
		t.Fatalf("managed-prefix replacement changed the opaque body:\n%s", got)
	}
}

func TestMigrateBrownfieldDocumentsIsMechanicalReviewedAndIdempotent(t *testing.T) {
	project := initializeLegacyBrownfieldProject(t)
	block := managedBlockForTest(t)
	writeTestFile(t, filepath.Join(project, "operator-notes.md"), "operator change\n")
	runGitForTest(t, project, "add", "operator-notes.md")
	planBefore := mustReadFile(t, filepath.Join(project, "docs", "implementation_plan.md"))
	visionPath := filepath.Join(project, "docs", "VISION.md")
	writeTestFile(t, visionPath, string(mustReadFile(t, visionPath))+"\nCurrent product clarification.\n")
	visionBefore := mustReadFile(t, visionPath)
	architectureBefore := mustReadFile(t, filepath.Join(project, "docs", "architecture.md"))

	var output bytes.Buffer
	options := defaultOptions(Options{
		Input: strings.NewReader("yes\n"), Output: &output, ErrorOutput: &bytes.Buffer{},
	})
	proceed, err := migrateBrownfieldDocuments(project, block, bufio.NewReader(options.Input), options)
	if err != nil {
		t.Fatal(err)
	}
	if !proceed {
		t.Fatal("accepted migration stopped initialization")
	}
	if _, err := os.Stat(filepath.Join(project, "docs", "implementation_plan.md")); !os.IsNotExist(err) {
		t.Fatalf("legacy implementation plan remains at its active path: %v", err)
	}
	planAfter := mustReadFile(t, filepath.Join(project, "docs", "archive", "implementation_plan.md"))
	if !bytes.Equal(planBefore, planAfter) {
		t.Fatal("archived implementation plan content changed")
	}
	requirements := string(mustReadFile(t, filepath.Join(project, filepath.FromSlash(legacyACDocumentPath))))
	if !strings.HasPrefix(requirements, string(block.content)) || strings.Count(requirements, string(block.marker)) != 1 {
		t.Errorf("legacy requirements block is missing or duplicated:\n%s", requirements)
	}
	if !bytes.Equal(visionBefore, mustReadFile(t, filepath.Join(project, "docs", "VISION.md"))) {
		t.Fatal("product vision changed during migration")
	}
	if !bytes.Equal(architectureBefore, mustReadFile(t, filepath.Join(project, "docs", "architecture.md"))) {
		t.Fatal("architecture changed during migration")
	}
	readme := string(mustReadFile(t, filepath.Join(project, "README.md")))
	if !strings.Contains(readme, "docs/implementation_plan.md") || strings.Contains(readme, "docs/archive/implementation_plan.md") {
		t.Fatalf("README was changed by the mechanical migration:\n%s", readme)
	}
	if !strings.Contains(output.String(), "Brownfield documentation migration variances:") || !strings.Contains(output.String(), "Committed brownfield documentation migration.") {
		t.Fatalf("migration review output = %q", output.String())
	}
	staged := strings.TrimSpace(runGitForTest(t, project, "diff", "--cached", "--name-only"))
	if staged != "operator-notes.md" {
		t.Fatalf("unrelated staged change was disturbed: %q", staged)
	}
	committed := runGitForTest(t, project, "show", "--pretty=format:", "--name-only", "HEAD")
	if strings.Contains(committed, "operator-notes.md") || strings.Contains(committed, "docs/VISION.md") {
		t.Fatalf("migration committed unrelated path:\n%s", committed)
	}

	var rerunOutput bytes.Buffer
	rerunOptions := defaultOptions(Options{
		Input: strings.NewReader(""), Output: &rerunOutput, ErrorOutput: &bytes.Buffer{},
	})
	proceed, err = migrateBrownfieldDocuments(project, block, bufio.NewReader(rerunOptions.Input), rerunOptions)
	if err != nil {
		t.Fatal(err)
	}
	if !proceed || rerunOutput.Len() != 0 {
		t.Fatalf("current migration was not a silent no-op: proceed=%t output=%q", proceed, rerunOutput.String())
	}

	architecturePath := filepath.Join(project, "docs", "architecture.md")
	architecture := mustReadFile(t, architecturePath)
	writeTestFile(t, architecturePath, string(architecture)+"\nCurrent design clarification.\n")
	rerunOutput.Reset()
	proceed, err = migrateBrownfieldDocuments(project, block, bufio.NewReader(rerunOptions.Input), rerunOptions)
	if err != nil {
		t.Fatal(err)
	}
	if !proceed || rerunOutput.Len() != 0 {
		t.Fatalf("current migration reacted to later documentation work: proceed=%t output=%q", proceed, rerunOutput.String())
	}
	if diff := runGitForTest(t, project, "diff", "--name-only", "--", "docs/architecture.md"); strings.TrimSpace(diff) != "docs/architecture.md" {
		t.Fatalf("later documentation work was disturbed: %q", diff)
	}
	runGitForTest(t, project, "add", "docs/architecture.md")
	rerunOutput.Reset()
	proceed, err = migrateBrownfieldDocuments(project, block, bufio.NewReader(rerunOptions.Input), rerunOptions)
	if err != nil {
		t.Fatal(err)
	}
	if !proceed || rerunOutput.Len() != 0 {
		t.Fatalf("current migration reacted to staged architecture work: proceed=%t output=%q", proceed, rerunOutput.String())
	}
	if staged := strings.TrimSpace(runGitForTest(t, project, "diff", "--cached", "--name-only")); staged != "docs/architecture.md\noperator-notes.md" && staged != "operator-notes.md\ndocs/architecture.md" {
		t.Fatalf("staged project work was disturbed: %q", staged)
	}
}

func TestMigrateBrownfieldDocumentsDeclineLeavesReviewedChanges(t *testing.T) {
	project := initializeLegacyBrownfieldProject(t)
	block := managedBlockForTest(t)
	before := strings.TrimSpace(runGitForTest(t, project, "rev-parse", "HEAD"))
	var output bytes.Buffer
	options := defaultOptions(Options{
		Input: strings.NewReader("no\n"), Output: &output, ErrorOutput: &bytes.Buffer{},
	})
	proceed, err := migrateBrownfieldDocuments(project, block, bufio.NewReader(options.Input), options)
	if err != nil {
		t.Fatal(err)
	}
	if proceed {
		t.Fatal("declined migration continued initialization")
	}
	after := strings.TrimSpace(runGitForTest(t, project, "rev-parse", "HEAD"))
	if after != before {
		t.Fatalf("declined migration created a commit: before=%s after=%s", before, after)
	}
	if !strings.Contains(output.String(), "remains staged and constitution generation has stopped") {
		t.Fatalf("declined migration output = %q", output.String())
	}

	var acceptedOutput bytes.Buffer
	acceptedOptions := defaultOptions(Options{
		Input: strings.NewReader("yes\n"), Output: &acceptedOutput, ErrorOutput: &bytes.Buffer{},
	})
	proceed, err = migrateBrownfieldDocuments(project, block, bufio.NewReader(acceptedOptions.Input), acceptedOptions)
	if err != nil {
		t.Fatal(err)
	}
	if !proceed || strings.TrimSpace(runGitForTest(t, project, "rev-parse", "HEAD")) == before {
		t.Fatal("reviewed migration could not be accepted on rerun")
	}
}

func TestMigrateBrownfieldDocumentsIgnoresOtherBrownfieldLayouts(t *testing.T) {
	project := t.TempDir()
	block := managedBlockForTest(t)
	var output bytes.Buffer
	options := defaultOptions(Options{
		Input: strings.NewReader(""), Output: &output, ErrorOutput: &bytes.Buffer{},
	})
	proceed, err := migrateBrownfieldDocuments(project, block, bufio.NewReader(options.Input), options)
	if err != nil {
		t.Fatal(err)
	}
	if !proceed || output.Len() != 0 {
		t.Fatalf("unrecognized brownfield layout was not ignored: proceed=%t output=%q", proceed, output.String())
	}
}

func TestMigrateBrownfieldDocumentsDoesNotRequireREADME(t *testing.T) {
	project := initializeLegacyBrownfieldProject(t)
	block := managedBlockForTest(t)
	if err := os.Remove(filepath.Join(project, "README.md")); err != nil {
		t.Fatal(err)
	}
	options := defaultOptions(Options{
		Input: strings.NewReader("yes\n"), Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{},
	})
	proceed, err := migrateBrownfieldDocuments(project, block, bufio.NewReader(options.Input), options)
	if err != nil {
		t.Fatal(err)
	}
	if !proceed {
		t.Fatal("README absence stopped the bounded document migration")
	}
	if _, err := os.Stat(filepath.Join(project, "README.md")); !os.IsNotExist(err) {
		t.Fatalf("migration created or changed README: %v", err)
	}
}

func TestRunCommitsCurrentUntrackedScaffoldWithoutLaunching(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	installProjectInitResourcesForTest(t, root)
	writeTestFile(t, filepath.Join(project, ".specify", "marker"), "present\n")
	writeTestFile(t, filepath.Join(project, ".specify", "presets", "sdlc-standards", "preset.yml"), "present\n")
	writeTestFile(t, filepath.Join(root, "technologies", "GO.md"), "# Go\n")
	writeTestFile(t, filepath.Join(root, "presets", "sdlc-standards", "preset.yml"), "present\n")
	disabled := false
	config := resolvedConfig{Harness: "codex", ProjectType: "brownfield", Technologies: []string{"GO"}, InfraEnabled: &disabled}
	writeTestFile(t, filepath.Join(project, ".env"), strings.Join([]string{
		`SDLC_AGENT_HARNESS="codex"`,
		`SDLC_SPEC_PROVIDER=""`,
		`SDLC_SPEC_MODEL=""`,
		`SDLC_BUILD_PROVIDER=""`,
		`SDLC_BUILD_MODEL=""`,
		`SDLC_AUDIT_PROVIDER=""`,
		`SDLC_AUDIT_MODEL=""`,
		`SDLC_PROJECT_TYPE="brownfield"`,
		`SDLC_TECHNOLOGIES="GO"`,
		`SDLC_INFRA_ENABLED="false"`,
		`SDLC_INFRA_OWNER=""`,
		`SDLC_INFRA_CONTRACT=""`,
		"",
	}, "\n"))
	target := filepath.Join(project, ".specify", "templates", "overrides", "constitution-template.md")
	writeTestFile(t, target, string(renderConstitutionForTest(t, root, []Technology{{Name: "GO", Title: "Go", Path: filepath.Join(root, "technologies", "GO.md")}}, config)))

	var commands []string
	runner := func(name string, arguments []string, _ string, _ io.Reader, output, _ io.Writer) error {
		commands = append(commands, name+" "+strings.Join(arguments, " "))
		switch {
		case name == "git" && len(arguments) != 0 && arguments[0] == "status":
			fmt.Fprintln(output, "?? .specify/templates/overrides/constitution-template.md")
			return nil
		case name == "git" && len(arguments) != 0 && (arguments[0] == "add" || arguments[0] == "commit"):
			return nil
		default:
			return fmt.Errorf("unexpected command: %s %v", name, arguments)
		}
	}
	var output bytes.Buffer
	if err := Run(Options{
		ProjectRoot: project, SDLCRoot: root, UserConfigPath: filepath.Join(t.TempDir(), ".env"),
		Input: strings.NewReader(""), Output: &output, ErrorOutput: &bytes.Buffer{}, RunCommand: runner,
	}); err != nil {
		t.Fatal(err)
	}
	if len(commands) != 3 || !strings.Contains(commands[0], "git status") || !strings.Contains(commands[1], "git add") || !strings.Contains(commands[2], "git commit --only --quiet") {
		t.Fatalf("baseline checkpoint commands = %#v", commands)
	}
	if !strings.Contains(output.String(), "Committed constitution scaffold: .specify/templates/overrides/constitution-template.md") {
		t.Fatalf("checkpoint output = %q", output.String())
	}
}

func TestCommitConstitutionScaffoldCommitsOnlyGeneratedOverride(t *testing.T) {
	project := t.TempDir()
	runGitForTest(t, project, "init", "--quiet")
	runGitForTest(t, project, "config", "user.name", "SDLC test")
	runGitForTest(t, project, "config", "user.email", "sdlc-test@example.invalid")
	writeTestFile(t, filepath.Join(project, "README.md"), "# Project\n")
	runGitForTest(t, project, "add", "README.md")
	runGitForTest(t, project, "commit", "--quiet", "--message", "Initial commit")

	target := filepath.Join(project, ".specify", "templates", "overrides", "constitution-template.md")
	writeTestFile(t, target, "# Constitution scaffold\n")
	writeTestFile(t, filepath.Join(project, "notes.md"), "operator change\n")
	runGitForTest(t, project, "add", "notes.md")

	options := defaultOptions(Options{Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}})
	if err := commitConstitutionScaffold(project, target, options); err != nil {
		t.Fatal(err)
	}
	committed := runGitForTest(t, project, "show", "--pretty=format:", "--name-only", "HEAD")
	if strings.TrimSpace(committed) != ".specify/templates/overrides/constitution-template.md" {
		t.Fatalf("checkpoint commit files = %q", committed)
	}
	staged := runGitForTest(t, project, "diff", "--cached", "--name-only")
	if strings.TrimSpace(staged) != "notes.md" {
		t.Fatalf("unrelated staged files after checkpoint = %q", staged)
	}
}

func TestRunSnapshotsEveryResolvedGlobalDefaultIntoProject(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	installProjectInitResourcesForTest(t, root)
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
		`SDLC_PROJECT_TYPE="brownfield"`,
		`SDLC_TECHNOLOGIES="GO"`,
		`SDLC_INFRA_ENABLED="true"`,
		`SDLC_INFRA_OWNER="Exodan"`,
		`SDLC_INFRA_CONTRACT="~/code/exodan/deploy/docs/PROJECT-INTEGRATION.md"`,
		"",
	}, "\n"))

	var initialOutput bytes.Buffer
	if err := Run(Options{
		ProjectRoot: project, SDLCRoot: root, UserConfigPath: userConfig, NoLaunch: true,
		Input: strings.NewReader("greenfield\n"), Output: &initialOutput, ErrorOutput: &bytes.Buffer{}, RunCommand: cleanGitRunner,
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(initialOutput.String(), "Project type (greenfield or brownfield):") {
		t.Fatalf("user-level project type suppressed the project prompt: %q", initialOutput.String())
	}
	got, err := readManagedEnv(filepath.Join(project, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := readManagedEnv(userConfig)
	if err != nil {
		t.Fatal(err)
	}
	want[keyProjectType] = "greenfield"
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("project snapshot = %#v, want global defaults %#v", got, want)
	}

	writeTestFile(t, userConfig, strings.ReplaceAll(string(mustReadFile(t, userConfig)), "gpt-5.6-sol", "new-global-spec"))
	var rerunOutput bytes.Buffer
	if err := Run(Options{
		ProjectRoot: project, SDLCRoot: root, UserConfigPath: userConfig, NoLaunch: true,
		Input: strings.NewReader(""), Output: &rerunOutput, ErrorOutput: &bytes.Buffer{}, RunCommand: cleanGitRunner,
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
	sdlcRoot := t.TempDir()
	installProjectInitResourcesForTest(t, sdlcRoot)
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
	if err := launchConstitution(resolvedConfig{Harness: "codex", SpecProvider: "openai-codex", SpecModel: "gpt-5.6-sol"}, projectRoot, sdlcRoot, templatePath, options); err != nil {
		t.Fatal(err)
	}
	if command != "codex" || directory != projectRoot || len(arguments) != 3 || arguments[0] != "--model" || arguments[1] != "gpt-5.6-sol" {
		t.Fatalf("constitution launch = command %q, arguments %#v, directory %q", command, arguments, directory)
	}
	prompt := arguments[2]
	for _, required := range []string{
		"`" + templatePath + "` as editable scaffolding",
		"It has no authority before ratification",
		"This is a filtering exercise, not a summary",
		"It applies across unrelated future features",
		"Changing it would require a constitutional decision",
		"supported by authoritative project documentation",
		"feature requirements and acceptance criteria",
		"migration algorithms, schema procedures, commands, test gates",
		"The constitution may elevate a concise project-wide invariant",
		"Importance alone does not make something constitutional",
		"Use the generated `Specification Baseline` as a proposed structure",
		"Archived implementation plans are historical provenance, not design authority",
		"an authority map, not a summary of the system",
		"current approved requirements",
		"approved historical requirements that are not centralized",
		"regression evidence and traceability",
		"the project's precedence and supersession rule",
		"tests and code record evidence and implemented state but do not approve requirements",
		"consumed by later specification and audit commands",
		"Do not invent sources or placeholder paths",
		"retain the generated prospective-baseline statement",
		"requirements established under a legacy process",
		"The legacy record governs only requirements established through that process",
		"Approved Spec Kit feature specifications govern requirements established or changed through Spec Kit",
		"may supersede a legacy requirement only explicitly and must preserve its lineage",
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
		"review the entire assembled constitution",
		"runtime configuration, a feature requirement, detailed design",
		"durable across unrelated features and require a constitutional amendment",
		"Remove or correct any scaffold clause that fails this review",
		"the initialization template, is the candidate presented to the human for ratification",
		"Keep the core workflow's Sync Impact Report as the first line",
		"do not accumulate reports or create a separate changelog",
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

func renderConstitutionForTest(t *testing.T, sdlcRoot string, technologies []Technology, config resolvedConfig) []byte {
	t.Helper()
	layout := mustReadFile(t, filepath.Join("..", "..", "src", filepath.FromSlash(constitutionLayoutPath)))
	rendered, err := renderConstitution(layout, sdlcRoot, technologies, config)
	if err != nil {
		t.Fatal(err)
	}
	return rendered
}

func managedBlockForTest(t *testing.T) managedBlock {
	t.Helper()
	block, err := loadManagedBlock(filepath.Join("..", "..", "src"))
	if err != nil {
		t.Fatal(err)
	}
	return block
}

func installProjectInitResourcesForTest(t *testing.T, destination string) {
	t.Helper()
	for _, relative := range []string{
		legacyACBlockPath,
		legacyACContractPath,
		constitutionLayoutPath,
		constitutionPromptPath,
	} {
		contents := mustReadFile(t, filepath.Join("..", "..", "src", filepath.FromSlash(relative)))
		writeTestFile(t, filepath.Join(destination, filepath.FromSlash(relative)), string(contents))
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

func cleanGitRunner(name string, arguments []string, _ string, _ io.Reader, _, _ io.Writer) error {
	if name == "git" && len(arguments) != 0 && arguments[0] == "status" {
		return nil
	}
	return fmt.Errorf("unexpected command: %s %v", name, arguments)
}

func runGitForTest(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
	return string(output)
}

func initializeLegacyBrownfieldProject(t *testing.T) string {
	t.Helper()
	project := t.TempDir()
	runGitForTest(t, project, "init", "--quiet")
	runGitForTest(t, project, "config", "user.name", "SDLC test")
	runGitForTest(t, project, "config", "user.email", "sdlc-test@example.invalid")
	writeTestFile(t, filepath.Join(project, "README.md"), strings.Join([]string{
		"# Existing project",
		"",
		"[Implementation plan](docs/implementation_plan.md)",
		"",
	}, "\n"))
	writeTestFile(t, filepath.Join(project, "docs", "VISION.md"), "# Vision\n")
	writeTestFile(t, filepath.Join(project, "docs", "architecture.md"), "# Architecture\n")
	writeTestFile(t, filepath.Join(project, "docs", "ACs.md"), "\n## AC table\n")
	writeTestFile(t, filepath.Join(project, "docs", "implementation_plan.md"), "# Implementation Plan\n\nHistorical plan.\n")
	runGitForTest(t, project, "add", "README.md", "docs")
	runGitForTest(t, project, "commit", "--quiet", "--message", "Initial project documentation")
	return project
}
