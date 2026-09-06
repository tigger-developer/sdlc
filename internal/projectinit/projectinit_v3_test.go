package projectinit

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestSchemaAndTechnologyDiscoveryAreDeterministic(t *testing.T) {
	root := testSDLCRoot(t)
	schema, err := LoadConfigSchema(root)
	if err != nil {
		t.Fatal(err)
	}
	if schema.Version != 3 {
		t.Fatalf("schema version = %d", schema.Version)
	}
	technologies, err := DiscoverTechnologies(filepath.Join(root, "technologies"))
	if err != nil {
		t.Fatal(err)
	}
	if len(technologies) < 2 {
		t.Fatal("expected discovered technology standards")
	}
	for index := 1; index < len(technologies); index++ {
		if technologies[index-1].Name >= technologies[index].Name {
			t.Fatalf("technologies not sorted: %#v", technologies)
		}
	}
	for _, field := range schema.Fields {
		if field.Key == "SDLC_AUDIT_TIMEOUT" {
			if err := field.ValidateValue("0s", technologies); err == nil {
				t.Fatal("zero audit timeout was accepted")
			}
			return
		}
	}
	t.Fatal("audit timeout field was not found")
}

func TestPromptFieldRendersQuestionBeforeDefault(t *testing.T) {
	var output bytes.Buffer
	value, explicit, err := promptField(
		bufio.NewReader(strings.NewReader("\n")),
		&output,
		ConfigField{Key: "SDLC_TEST_TIMEOUT", Type: "duration", Prompt: "Select timeout:"},
		"5m",
		"schema",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if value != "5m" || explicit {
		t.Fatalf("selection = %q, explicit = %v", value, explicit)
	}
	want := "Select timeout:\nDefault: 5m. Press Enter to inherit, or enter a project value.\nSelection: "
	if output.String() != want {
		t.Fatalf("prompt output = %q, want %q", output.String(), want)
	}
}

func TestSpecificationTemplateInternalLinksResolve(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile(filepath.Join(testSDLCRoot(t), "templates", "v3", "spec.org"))
	if err != nil {
		t.Fatal(err)
	}
	document := string(contents)
	links := regexp.MustCompile(`\[\[#([^]]+)\]`).FindAllStringSubmatch(document, -1)
	if len(links) == 0 {
		t.Fatal("specification template contains no internal links")
	}
	for _, link := range links {
		if !strings.Contains(document, ":CUSTOM_ID: "+link[1]+"\n") {
			t.Errorf("internal link target %q has no CUSTOM_ID", link[1])
		}
	}
}

func TestDetectSourceDistinguishesNewV1AndV2(t *testing.T) {
	root := t.TempDir()
	if source, _ := DetectSource(root); source != "new" {
		t.Fatalf("new source = %s", source)
	}
	writeProjectTestFile(t, filepath.Join(root, "docs", "ACs.org"), "#+TITLE: ACs\n")
	if source, _ := DetectSource(root); source != "v1" {
		t.Fatalf("v1 source = %s", source)
	}
	if err := os.Mkdir(filepath.Join(root, ".specify"), 0o755); err != nil {
		t.Fatal(err)
	}
	if source, _ := DetectSource(root); source != "v2" {
		t.Fatalf("v2 source = %s", source)
	}
}

func TestWriteProjectProfileOmitsInheritedGlobalDefaults(t *testing.T) {
	root := t.TempDir()
	schema, err := LoadConfigSchema(testSDLCRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	generation := projectGeneration{
		values: map[string]string{
			"SDLC_PROJECT_KIND":             "application",
			"SDLC_TECHNOLOGIES":             "GO,SHELL",
			"SDLC_PRODUCT_AUTHORITIES":      "docs/VISION.md,README.md",
			"SDLC_ARCHITECTURE_AUTHORITIES": "docs/architecture.md",
			"SDLC_REQUIREMENT_AUTHORITIES":  "docs/work.org",
			"SDLC_BRANCH_STRATEGY":          "feature",
			"SDLC_SPEC_MODEL":               "gpt-example",
			"SDLC_INFRA_ROLE":               "none",
		},
		explicit: map[string]bool{
			"SDLC_PROJECT_KIND":             true,
			"SDLC_TECHNOLOGIES":             true,
			"SDLC_PRODUCT_AUTHORITIES":      true,
			"SDLC_ARCHITECTURE_AUTHORITIES": true,
			"SDLC_INFRA_ROLE":               true,
		},
		source: "new", base: "master", archive: "sdlc_new_state_2026-09-06",
		migration: "sdlc-v3-migration-2026-09-06", date: "2026-09-06",
	}
	if err := writeProjectProfile(root, "v3-test", schema, generation); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(root, projectProfilePath))
	if err != nil {
		t.Fatal(err)
	}
	var profile map[string]any
	if err := yaml.Unmarshal(contents, &profile); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contents), "gpt-example") || strings.Contains(string(contents), "branch_strategy") {
		t.Fatalf("inherited defaults leaked into project profile:\n%s", contents)
	}
	if !strings.Contains(string(contents), "application") || !strings.Contains(string(contents), "GO") || !strings.Contains(string(contents), "docs/VISION.md") {
		t.Fatalf("project facts missing:\n%s", contents)
	}
	if !strings.Contains(string(contents), "requirements:\n        - docs/work.org") {
		t.Fatalf("deterministic work authority is missing from the YAML list:\n%s", contents)
	}
}

func TestLegacyConfigurationDropsUnsupportedProviderModelPair(t *testing.T) {
	var diagnostics bytes.Buffer
	legacy := map[string]string{
		"SDLC_AUDIT_PROVIDER": "nous",
		"SDLC_AUDIT_MODEL":    "z-ai/glm-example",
		"SDLC_BUILD_PROVIDER": "openai-codex",
		"SDLC_BUILD_MODEL":    "gpt-example",
	}
	normalized := normalizeLegacyConfiguration(legacy, &diagnostics)
	if normalized["SDLC_AUDIT_PROVIDER"] != "" || normalized["SDLC_AUDIT_MODEL"] != "" {
		t.Fatalf("unsupported audit pair remains: %#v", normalized)
	}
	if normalized["SDLC_BUILD_MODEL"] != "gpt-example" {
		t.Fatalf("supported build model was removed: %#v", normalized)
	}
	if !strings.Contains(diagnostics.String(), "audit provider") {
		t.Fatalf("missing migration warning: %s", diagnostics.String())
	}
}

func TestArchiveSpecKitPreservesFilesAndRemovesActiveCopy(t *testing.T) {
	root := t.TempDir()
	writeProjectTestFile(t, filepath.Join(root, ".specify", "specs", "001", "spec.md"), "unfinished\n")
	writeProjectTestFile(t, filepath.Join(root, ".agents", "skills", "speckit-constitution", "SKILL.md"), "old\n")
	writeProjectTestFile(t, filepath.Join(root, ".agents", "skills", "project-helper", "SKILL.md"), "keep\n")
	if _, err := archiveSpecKit(root); err != nil {
		t.Fatal(err)
	}
	archived := filepath.Join(root, "docs", "archive", "sdlc-v2", ".specify", "specs", "001", "spec.md")
	contents, err := os.ReadFile(archived)
	if err != nil || string(contents) != "unfinished\n" {
		t.Fatalf("archived content = %q, %v", contents, err)
	}
	if exists(filepath.Join(root, ".specify")) || exists(filepath.Join(root, ".agents", "skills", "speckit-constitution")) {
		t.Fatal("active Spec Kit artefacts remain")
	}
	if !exists(filepath.Join(root, ".agents", "skills", "project-helper", "SKILL.md")) {
		t.Fatal("unrelated project skill was removed")
	}
	if !exists(filepath.Join(root, "docs", "archive", "sdlc-v2", "integrations", ".agents", "skills", "speckit-constitution", "SKILL.md")) {
		t.Fatal("project-local Spec Kit skill was not archived")
	}
}

func TestImportLegacyWorkCreatesDescriptorBearingWorkItems(t *testing.T) {
	root := t.TempDir()
	writeProjectTestFile(t, filepath.Join(root, "docs", "work.org"), "* Open defects\n\n* Undelivered features\n\n* Human review\n")
	writeProjectTestFile(t, filepath.Join(root, "docs", "ticket-migration.org"), `* Open defects at migration

** [[file:archive/migrated-tickets/7.md][#7 - Reject invalid host]]
*** Disposition
Still reproducible.

* Defined but undelivered features

** [[file:archive/migrated-tickets/12.md][#12 - Add status page]]
*** Relevant legacy criteria
**** AC12.1 - Status is visible

* Requires human review

- None.

* Delivered tickets
`)
	if err := importLegacyWork(root); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(root, "docs", "work.org"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(contents)
	for _, want := range []string{
		"** TODO W007 - Reject invalid host :defect:",
		"** TODO W012 - Add status page :feature:",
		":PRIORITY: unassigned",
		":CREATED: unknown",
		"**** Disposition",
		"***** AC12.1 - Status is visible",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("work ledger lacks %q:\n%s", want, text)
		}
	}
}

func TestWriteWorkLedgerAllocatesMigratedV2WorkAboveHistoricIdentifiers(t *testing.T) {
	root := t.TempDir()
	writeProjectTestFile(t, filepath.Join(root, "docs", "ACs.org"), "*** AC130.1 - Existing historical requirement\n")
	generation := projectGeneration{
		source: "v2", base: "master", archive: "sdlc_v2_state_2026-09-06",
		migration: "sdlc-v3-migration-2026-09-06", legacyV2: []string{"001-old-feature"},
		migratedV2: []migratedWorkCandidate{{
			Path: "docs/archive/sdlc-v2/specs/001-old-feature/spec.md", Descriptor: "Old feature",
			Disposition: "unresolved", Priority: "unassigned", Created: "unknown", Evidence: "Approval and delivery are not recorded.",
		}},
	}
	if err := writeWorkLedger(testSDLCRoot(t), root, generation, time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(root, "docs", "work.org"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"** REVIEW W131 - Old feature :migration:", ":LEGACY_ID: 001", ":TYPE: migration",
		":PRIORITY: unassigned", ":SOURCE: docs/archive/sdlc-v2/specs/001-old-feature/spec.md",
		":CREATED: unknown", ":DISPOSITION: unresolved", "*Migration evidence:* Approval and delivery are not recorded.",
	} {
		if !strings.Contains(string(contents), want) {
			t.Fatalf("work ledger lacks %q:\n%s", want, contents)
		}
	}
	if strings.Contains(string(contents), "* Migrated v2 work requiring disposition") {
		t.Fatalf("work ledger added a non-template top-level section:\n%s", contents)
	}
	template, err := os.ReadFile(filepath.Join(testSDLCRoot(t), "templates", "v3", "work.org"))
	if err != nil {
		t.Fatal(err)
	}
	preamble := strings.SplitN(string(template), "* Open defects\n", 2)[0]
	if !strings.HasPrefix(string(contents), preamble+"* Open defects\n") {
		t.Fatalf("work ledger did not preserve the canonical template preamble:\n%s", contents)
	}
	humanReview := strings.Index(string(contents), "* Human review\n")
	closedWork := strings.Index(string(contents), "* Closed work\n")
	item := strings.Index(string(contents), "** REVIEW W131 - Old feature")
	if humanReview < 0 || item < humanReview || closedWork < item {
		t.Fatalf("unresolved migrated work was not placed under Human review:\n%s", contents)
	}
}

func TestMigratedWorkPlacementUsesCanonicalSections(t *testing.T) {
	tests := []struct {
		disposition string
		state       string
		section     string
	}{
		{"delivered", "DONE", "Closed work"},
		{"approved-undelivered", "TODO", "Undelivered features"},
		{"abandoned", "ABANDONED", "Closed work"},
		{"unresolved", "REVIEW", "Human review"},
	}
	for _, test := range tests {
		state, section := migratedWorkPlacement(test.disposition)
		if state != test.state || section != test.section {
			t.Errorf("placement for %s = %s/%s, want %s/%s", test.disposition, state, section, test.state, test.section)
		}
	}
}

func TestValidateMigratedWorkRequiresExactArchivedCoverage(t *testing.T) {
	root := t.TempDir()
	path := "docs/archive/sdlc-v2/specs/001-example/spec.md"
	writeProjectTestFile(t, filepath.Join(root, filepath.FromSlash(path)), "# Example\n")
	available := map[string]bool{path: true}

	if _, err := validateMigratedWork(nil, []string{"001-example"}, root, available); err == nil || !strings.Contains(err.Error(), "omits archived Spec Kit specification") {
		t.Fatalf("missing-specification error = %v", err)
	}
	candidate := migratedWorkCandidate{
		Path: path, Descriptor: "Example", Disposition: "unresolved",
		Priority: "unassigned", Created: "unknown", Evidence: "Approval is not recorded.",
	}
	if _, err := validateMigratedWork([]migratedWorkCandidate{candidate, candidate}, []string{"001-example"}, root, available); err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatalf("duplicate-specification error = %v", err)
	}
}

func TestDiscoverAuthorityChoicesIncludesStemVariantsAndSelectsCanonicalPaths(t *testing.T) {
	available := map[string]bool{
		"README.md":                          true,
		"docs/VISION.md":                     true,
		"docs/proposal/V2_VISION_DRAFT.md":   true,
		"docs/v1/VISION.md":                  true,
		"docs/ARCHITECTURE.md":               true,
		"docs/architecture-v2.org":           true,
		"docs/v1/architecture.md":            true,
		"docs/revision-notes.md":             true,
		"vendor/example/README.txt":          true,
		"docs/archive/README-background.org": true,
	}

	product := discoverAuthorityChoices(available, "product")
	wantProduct := map[string]bool{
		"README.md":                          true,
		"docs/VISION.md":                     true,
		"docs/archive/README-background.org": false,
		"docs/proposal/V2_VISION_DRAFT.md":   false,
		"docs/v1/VISION.md":                  false,
	}
	if len(product) != len(wantProduct) {
		t.Fatalf("product candidates = %#v", product)
	}
	for _, choice := range product {
		if selected, ok := wantProduct[choice.Path]; !ok || selected != choice.Selected {
			t.Errorf("unexpected product choice %#v", choice)
		}
	}

	architecture := discoverAuthorityChoices(available, "architecture")
	wantArchitecture := map[string]bool{
		"docs/ARCHITECTURE.md":     true,
		"docs/architecture-v2.org": false,
		"docs/v1/architecture.md":  false,
	}
	if len(architecture) != len(wantArchitecture) {
		t.Fatalf("architecture candidates = %#v", architecture)
	}
	for _, choice := range architecture {
		if selected, ok := wantArchitecture[choice.Path]; !ok || selected != choice.Selected {
			t.Errorf("unexpected architecture choice %#v", choice)
		}
	}
}

func TestRequirementAuthoritiesAreDeterministic(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		available map[string]bool
		want      string
		wantError string
	}{
		{"new project", "new", map[string]bool{"docs/work.org": true}, "docs/work.org", ""},
		{"migrated Org ledger", "v1", map[string]bool{"docs/work.org": true, "docs/ACs.org": true}, "docs/work.org,docs/ACs.org", ""},
		{"unconverted Markdown ledger", "v1", map[string]bool{"docs/work.org": true, "docs/ACs.md": true}, "docs/work.org,docs/ACs.md", ""},
		{"conflicting ledgers", "v1", map[string]bool{"docs/work.org": true, "docs/ACs.org": true, "docs/ACs.md": true}, "", "both docs/ACs.org and docs/ACs.md"},
		{"missing v1 ledger", "v1", map[string]bool{"docs/work.org": true}, "", "legacy migration requires docs/ACs.org or docs/ACs.md"},
		{"missing work ledger", "new", map[string]bool{}, "", "docs/work.org is missing"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := deterministicRequirementAuthorities(test.available, test.source)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("error = %v, want containing %q", err, test.wantError)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("authorities = %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestAuthorityChooserRefreshesAfterDocumentMove(t *testing.T) {
	project := t.TempDir()
	writeProjectTestFile(t, filepath.Join(project, "ARCHITECTURE.md"), "# Architecture\n")
	var inventories int
	runner := func(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
		if name != "git" || !containsArgument(arguments, "ls-files") {
			return fmt.Errorf("unexpected command: %s %v", name, arguments)
		}
		inventories++
		if inventories == 1 {
			_, err := io.WriteString(output, "ARCHITECTURE.md\x00")
			return err
		}
		if err := os.MkdirAll(filepath.Join(project, "docs"), 0o755); err != nil {
			return err
		}
		if err := os.Rename(filepath.Join(project, "ARCHITECTURE.md"), filepath.Join(project, "docs", "ARCHITECTURE.md")); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		_, err := io.WriteString(output, "docs/ARCHITECTURE.md\x00")
		return err
	}
	var output bytes.Buffer
	options := defaultOptions(Options{
		ProjectRoot: project,
		Input:       strings.NewReader("r\n\n"),
		Output:      &output,
		ErrorOutput: &bytes.Buffer{},
		RunCommand:  runner,
	})
	field := ConfigField{Key: "SDLC_ARCHITECTURE_AUTHORITIES", Prompt: "Architecture authority paths", DiscoveryCategory: "architecture"}
	got, err := promptDiscoveredAuthorityField(options, field, project)
	if err != nil {
		t.Fatal(err)
	}
	if got != "docs/ARCHITECTURE.md" {
		t.Fatalf("authority after refresh = %q", got)
	}
	for _, want := range []string{"[ ] 1. ARCHITECTURE.md", "[x] 1. docs/ARCHITECTURE.md", "'r' to rescan"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("chooser output lacks %q:\n%s", want, output.String())
		}
	}
}

func TestRunInitializesOnceOnMigrationBranch(t *testing.T) {
	project := t.TempDir()
	runGitTest(t, project, "init")
	runGitTest(t, project, "config", "user.name", "Test Operator")
	runGitTest(t, project, "config", "user.email", "operator@example.invalid")
	writeProjectTestFile(t, filepath.Join(project, "README.md"), "# Project\n")
	writeProjectTestFile(t, filepath.Join(project, ".gitignore"), ".env\n")
	writeProjectTestFile(t, filepath.Join(project, ".env"), "PRIVATE_TOKEN=secret\nSDLC_PROJECT_TYPE=greenfield\n")
	runGitTest(t, project, "add", "README.md", ".gitignore")
	runGitTest(t, project, "commit", "-m", "initial")
	options := Options{
		ProjectRoot:  project,
		SDLCRoot:     testSDLCRoot(t),
		Overrides:    v3TestOverrides(t),
		SDLCRevision: "v3-test",
		Input:        strings.NewReader("\n\n\nn\nn\n"),
		Output:       &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{},
		RunCommand: localOnlyTestRunner,
		Now:        func() time.Time { return time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC) },
	}
	if err := Run(options); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(project, projectProfilePath)) || !exists(filepath.Join(project, "docs", "work.org")) {
		t.Fatal("v3 project artefacts missing")
	}
	profile, err := os.ReadFile(filepath.Join(project, projectProfilePath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(profile), "requirements:\n        - docs/work.org") || strings.Contains(string(profile), "authorities:\n    requirements:\n        - README.md") {
		t.Fatalf("new-project requirement authorities = %s", profile)
	}
	environment, err := os.ReadFile(filepath.Join(project, ".env"))
	if err != nil || string(environment) != "PRIVATE_TOKEN=secret\n" {
		t.Fatalf("legacy environment cleanup = %q, %v", environment, err)
	}
	branch := strings.TrimSpace(runGitTest(t, project, "branch", "--show-current"))
	if branch != "sdlc-v3-migration-2026-09-06" {
		t.Fatalf("branch = %s", branch)
	}
	if err := Run(options); err == nil || !strings.Contains(err.Error(), "runs exactly once") {
		t.Fatalf("second initialization error = %v", err)
	}
}

func TestRunMigratesV2WithoutRenumberingOrDeletingUnrelatedIntegrations(t *testing.T) {
	project := t.TempDir()
	runGitTest(t, project, "init")
	runGitTest(t, project, "config", "user.name", "Test Operator")
	runGitTest(t, project, "config", "user.email", "operator@example.invalid")
	writeProjectTestFile(t, filepath.Join(project, "README.md"), "# Project\n")
	writeProjectTestFile(t, filepath.Join(project, "docs", "ACs.org"), "*** AC130.1 - Existing behaviour\n")
	writeProjectTestFile(t, filepath.Join(project, ".specify", "memory", "constitution.md"), "legacy constitution\n")
	writeProjectTestFile(t, filepath.Join(project, "specs", "001-old-feature", "spec.md"), "unfinished feature\n")
	writeProjectTestFile(t, filepath.Join(project, ".agents", "skills", "speckit-plan", "SKILL.md"), "old integration\n")
	writeProjectTestFile(t, filepath.Join(project, ".agents", "skills", "project-helper", "SKILL.md"), "unrelated\n")
	runGitTest(t, project, "add", "-A")
	runGitTest(t, project, "commit", "-m", "v2 state")
	runner := func(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
		if name == "codex" {
			writeProjectTestFile(t, argumentValue(arguments, "--output-last-message"), `version: 1
migrated_work:
  - path: docs/archive/sdlc-v2/specs/001-old-feature/spec.md
    descriptor: Old feature
    disposition: unresolved
    priority: unassigned
    created: unknown
    evidence: Approval and delivery are not recorded.
warnings: []
`)
			return nil
		}
		return localOnlyTestRunner(name, arguments, directory, input, output, errorOutput)
	}

	options := Options{
		ProjectRoot: project, SDLCRoot: testSDLCRoot(t), Overrides: v3TestOverrides(t),
		SDLCRevision: "v3-test", Input: strings.NewReader("\n\n\nn\nn\n"),
		Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{},
		RunCommand: runner,
		Now:        func() time.Time { return time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC) },
	}
	if err := Run(options); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"docs/archive/sdlc-v2/.specify/memory/constitution.md",
		"docs/archive/sdlc-v2/specs/001-old-feature/spec.md",
		"docs/archive/sdlc-v2/integrations/.agents/skills/speckit-plan/SKILL.md",
		".agents/skills/project-helper/SKILL.md",
	} {
		if !exists(filepath.Join(project, path)) {
			t.Fatalf("expected preserved path %s", path)
		}
	}
	if exists(filepath.Join(project, ".specify")) || exists(filepath.Join(project, "specs")) {
		t.Fatal("active Spec Kit directories remain")
	}
	work, err := os.ReadFile(filepath.Join(project, "docs", "work.org"))
	if err != nil || !strings.Contains(string(work), "W131 - Old feature") || !strings.Contains(string(work), ":LEGACY_ID: 001") {
		t.Fatalf("migrated work ledger = %q, %v", work, err)
	}
	profile, err := os.ReadFile(filepath.Join(project, projectProfilePath))
	if err != nil || !strings.Contains(string(profile), "requirements:\n        - docs/work.org\n        - docs/ACs.org") {
		t.Fatalf("v2 requirement authorities = %q, %v", profile, err)
	}
	archived := runGitTest(t, project, "show", "sdlc_v2_state_2026-09-06:.specify/memory/constitution.md")
	if archived != "legacy constitution\n" {
		t.Fatalf("archive branch content = %q", archived)
	}
}

func TestRunMigratesV1ThroughTicketSkillBeforeCreatingProfile(t *testing.T) {
	project := t.TempDir()
	runGitTest(t, project, "init")
	runGitTest(t, project, "config", "user.name", "Test Operator")
	runGitTest(t, project, "config", "user.email", "operator@example.invalid")
	writeProjectTestFile(t, filepath.Join(project, "README.md"), "# Project\n")
	writeProjectTestFile(t, filepath.Join(project, "docs", "ACs.md"), "# Acceptance criteria\n\nAC7.1 - Existing result\n")
	runGitTest(t, project, "add", "-A")
	runGitTest(t, project, "commit", "-m", "v1 state")

	var skillCalls int
	runner := func(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
		if name == "gh" {
			_, err := io.WriteString(output, `[{"number":7}]`)
			return err
		}
		if name == "codex" {
			skillCalls++
			if !strings.Contains(strings.Join(arguments, " "), "migrate-legacy-acs-to-sdlc-v1") {
				return fmt.Errorf("unexpected skill invocation: %v", arguments)
			}
			writeProjectTestFile(t, filepath.Join(project, "docs", "ACs.org"), "#+TITLE: Acceptance Criteria\n\n*** AC7.1 - Existing result\n")
			writeProjectTestFile(t, filepath.Join(project, "docs", "ticket-migration.org"), "#+TITLE: Ticket Migration\n")
			if err := os.Remove(filepath.Join(project, "docs", "ACs.md")); err != nil {
				return err
			}
			return nil
		}
		command := exec.Command(name, arguments...)
		command.Dir = directory
		command.Stdin = input
		command.Stdout = output
		command.Stderr = errorOutput
		return command.Run()
	}

	options := Options{
		ProjectRoot: project, SDLCRoot: testSDLCRoot(t), Overrides: v3TestOverrides(t),
		SDLCRevision: "v3-test", Input: strings.NewReader("yes\n\n\n\nn\nn\n"),
		Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}, RunCommand: runner,
		Now: func() time.Time { return time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC) },
	}
	if err := Run(options); err != nil {
		t.Fatal(err)
	}
	if skillCalls != 1 {
		t.Fatalf("legacy migration skill calls = %d, want 1", skillCalls)
	}
	if !exists(filepath.Join(project, "docs", "ACs.org")) || exists(filepath.Join(project, "docs", "ACs.md")) {
		t.Fatal("legacy AC ledger was not converted")
	}
	profile, err := os.ReadFile(filepath.Join(project, projectProfilePath))
	if err != nil || !strings.Contains(string(profile), "requirements:\n        - docs/work.org\n        - docs/ACs.org") {
		t.Fatalf("v1 requirement authorities = %q, %v", profile, err)
	}
	archived := runGitTest(t, project, "show", "sdlc_v1_state_2026-09-06:docs/ACs.md")
	if !strings.Contains(archived, "AC7.1 - Existing result") {
		t.Fatalf("archive branch lost v1 ledger: %q", archived)
	}
}

func TestRunClassifiesV2WorkAndResolvesAuthoritiesAfterArchival(t *testing.T) {
	project := t.TempDir()
	runGitTest(t, project, "init")
	runGitTest(t, project, "config", "user.name", "Test Operator")
	runGitTest(t, project, "config", "user.email", "operator@example.invalid")
	writeProjectTestFile(t, filepath.Join(project, "README.md"), "# Project\n")
	writeProjectTestFile(t, filepath.Join(project, "docs", "VISION.md"), "# Vision\n")
	writeProjectTestFile(t, filepath.Join(project, "docs", "architecture.md"), "# Architecture\n")
	writeProjectTestFile(t, filepath.Join(project, "docs", "requirements.md"), "# Requirements\n")
	writeProjectTestFile(t, filepath.Join(project, ".specify", "memory", "constitution.md"), "legacy constitution\n")
	writeProjectTestFile(t, filepath.Join(project, "specs", "001-delivered-feature", "spec.md"), "# Delivered feature\n")
	writeProjectTestFile(t, filepath.Join(project, "specs", "002-unresolved-feature", "spec.md"), "# Unresolved feature\n")
	runGitTest(t, project, "add", "-A")
	runGitTest(t, project, "commit", "-m", "v2 state")

	overrides := v3TestOverrides(t)
	delete(overrides, "SDLC_PRODUCT_AUTHORITIES")
	delete(overrides, "SDLC_ARCHITECTURE_AUTHORITIES")
	delete(overrides, "SDLC_REQUIREMENT_AUTHORITIES")
	var discoveryCalls int
	runner := func(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
		if name == "codex" {
			discoveryCalls++
			if argumentValue(arguments, "--sandbox") != "read-only" || argumentValue(arguments, "--model") != overrides["SDLC_AUDIT_MODEL"] || !containsArgument(arguments, "--ephemeral") {
				return fmt.Errorf("unsafe or misconfigured authority discovery invocation: %v", arguments)
			}
			if exists(filepath.Join(project, ".specify")) {
				return errors.New("authority discovery ran before v2 archival")
			}
			if !exists(filepath.Join(project, "docs", "archive", "sdlc-v2", ".specify", "memory", "constitution.md")) {
				return errors.New("authority discovery ran before archived evidence was available")
			}
			work, err := os.ReadFile(filepath.Join(project, "docs", "work.org"))
			if err != nil {
				return fmt.Errorf("authority discovery ran before the work ledger was created: %w", err)
			}
			if !strings.Contains(string(work), "Archived Spec Kit work") {
				return errors.New("authority discovery ran before archived v2 work was indexed")
			}
			prompt, err := io.ReadAll(input)
			if err != nil {
				return err
			}
			if strings.Contains(string(prompt), "authorities:") || !strings.Contains(string(prompt), "docs/archive/sdlc-v2/specs/001-delivered-feature/spec.md") || !strings.Contains(string(prompt), "docs/archive/sdlc-v2/specs/002-unresolved-feature/spec.md") {
				return fmt.Errorf("unexpected discovery prompt: %s", prompt)
			}
			outputPath := argumentValue(arguments, "--output-last-message")
			if outputPath == "" {
				return fmt.Errorf("missing migrated-work proposal output path: %v", arguments)
			}
			writeProjectTestFile(t, outputPath, `version: 1
migrated_work:
  - path: docs/archive/sdlc-v2/specs/001-delivered-feature/spec.md
    descriptor: Delivered feature
    disposition: delivered
    priority: P1
    created: 2026-09-01
    evidence: The operator authorized delivery and validation is recorded.
  - path: docs/archive/sdlc-v2/specs/002-unresolved-feature/spec.md
    descriptor: Unresolved feature
    disposition: unresolved
    priority: unassigned
    created: unknown
    evidence: No operator approval or delivery is recorded.
warnings: []
`)
			return nil
		}
		command := exec.Command(name, arguments...)
		command.Dir = directory
		command.Stdin = input
		command.Stdout = output
		command.Stderr = errorOutput
		return command.Run()
	}

	options := Options{
		ProjectRoot: project, SDLCRoot: testSDLCRoot(t), Overrides: overrides,
		SDLCRevision: "v3-test", Input: strings.NewReader("n\n\n\nn\n"),
		Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{}, RunCommand: runner,
		Now: func() time.Time { return time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC) },
	}
	if err := Run(options); err != nil {
		t.Fatal(err)
	}
	if discoveryCalls != 1 {
		t.Fatalf("authority discovery calls = %d, want 1", discoveryCalls)
	}
	contents, err := os.ReadFile(filepath.Join(project, projectProfilePath))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"product:\n        - README.md\n        - docs/VISION.md", "architecture:\n        - docs/architecture.md", "requirements:\n        - docs/work.org"} {
		if !strings.Contains(string(contents), want) {
			t.Fatalf("project profile lacks authority list %q:\n%s", want, contents)
		}
	}
	if strings.Contains(string(contents), "docs/archive/sdlc-v2/specs/") || strings.Contains(string(contents), "docs/requirements.md") {
		t.Fatalf("project profile enumerates migrated or non-canonical requirement documents:\n%s", contents)
	}
	work, err := os.ReadFile(filepath.Join(project, "docs", "work.org"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"** DONE W001 - Delivered feature :migration:",
		"** REVIEW W002 - Unresolved feature :migration:",
		":PRIORITY: P1", ":CREATED: 2026-09-01", ":DISPOSITION: delivered",
	} {
		if !strings.Contains(string(work), want) {
			t.Fatalf("work ledger lacks %q:\n%s", want, work)
		}
	}
	if exists(initializationWorkspacePath(project)) {
		t.Fatal("successful initialization left its temporary working directory behind")
	}
}

func TestRunReportsInterruptedInitializationWorkspace(t *testing.T) {
	project := t.TempDir()
	runGitTest(t, project, "init")
	runGitTest(t, project, "config", "user.name", "Test Operator")
	runGitTest(t, project, "config", "user.email", "operator@example.invalid")
	writeProjectTestFile(t, filepath.Join(project, "README.md"), "# Project\n")
	runGitTest(t, project, "add", "README.md")
	runGitTest(t, project, "commit", "-m", "initial")
	workspace := initializationWorkspacePath(project)
	writeProjectTestFile(t, filepath.Join(workspace, "authority-proposal.yaml"), "version: 1\n")

	err := Run(Options{
		ProjectRoot: project, SDLCRoot: testSDLCRoot(t), Overrides: v3TestOverrides(t),
		Input: strings.NewReader(""), Output: &bytes.Buffer{}, ErrorOutput: &bytes.Buffer{},
	})
	if err == nil || !strings.Contains(err.Error(), "previous initialization did not complete") || !strings.Contains(err.Error(), workspace) {
		t.Fatalf("interrupted initialization error = %v", err)
	}
}

func TestChooseTicketMigrationDeclineContinues(t *testing.T) {
	project := t.TempDir()
	writeProjectTestFile(t, filepath.Join(project, "docs", "ACs.md"), "# Acceptance criteria\n")
	var output bytes.Buffer
	options := defaultOptions(Options{
		ProjectRoot: project,
		Input:       strings.NewReader("n\n"),
		Output:      &output,
		ErrorOutput: &bytes.Buffer{},
		RunCommand: func(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
			if name != "gh" {
				return fmt.Errorf("unexpected command: %s %v", name, arguments)
			}
			_, err := io.WriteString(output, `[{"number":7}]`)
			return err
		},
	})
	run, err := chooseTicketMigration(options, project)
	if err != nil {
		t.Fatalf("declining optional ticket migration returned an error: %v", err)
	}
	if run {
		t.Fatal("declined ticket migration was selected")
	}
}

func TestInitializationWorkspaceIsIgnoredByGit(t *testing.T) {
	project := t.TempDir()
	runGitTest(t, project, "init")
	writeProjectTestFile(t, filepath.Join(project, ".sdlc", ".gitignore"), "local-cache/\n")
	workspace := initializationWorkspacePath(project)
	if err := createInitializationWorkspace(testSDLCRoot(t), project, workspace); err != nil {
		t.Fatal(err)
	}
	writeProjectTestFile(t, filepath.Join(workspace, "interrupted.yaml"), "version: 1\n")
	command := exec.Command("git", "check-ignore", "--quiet", ".sdlc/.init/interrupted.yaml")
	command.Dir = project
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("temporary initialization state is not ignored: %v: %s", err, output)
	}
	ignore, err := os.ReadFile(filepath.Join(project, ".sdlc", ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if string(ignore) != "local-cache/\n.init/\n" {
		t.Fatalf("initialization ignore file = %q", ignore)
	}
}

func argumentValue(arguments []string, name string) string {
	for index := 0; index+1 < len(arguments); index++ {
		if arguments[index] == name {
			return arguments[index+1]
		}
	}
	return ""
}

func containsArgument(arguments []string, want string) bool {
	for _, argument := range arguments {
		if argument == want {
			return true
		}
	}
	return false
}

func v3TestOverrides(t *testing.T) map[string]string {
	t.Helper()
	schema, err := LoadConfigSchema(testSDLCRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	overrides := map[string]string{
		"SDLC_PROJECT_KIND":             "application",
		"SDLC_TECHNOLOGIES":             "GO",
		"SDLC_PRODUCT_AUTHORITIES":      "README.md",
		"SDLC_ARCHITECTURE_AUTHORITIES": "README.md",
		"SDLC_BRANCH_STRATEGY":          "current",
		"SDLC_SPEC_HARNESS":             "codex",
		"SDLC_SPEC_PROVIDER":            "openai",
		"SDLC_BUILD_HARNESS":            "codex",
		"SDLC_BUILD_PROVIDER":           "openai",
		"SDLC_AUDIT_HARNESS":            "codex",
		"SDLC_AUDIT_PROVIDER":           "openai",
		"SDLC_AUDIT_TIMEOUT":            "5m",
		"SDLC_INFRA_ROLE":               "none",
	}
	for _, field := range schema.Fields {
		if strings.HasSuffix(field.Key, "_MODEL") {
			overrides[field.Key] = field.Default
		}
	}
	return overrides
}

func testSDLCRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "src"))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func writeProjectTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	// #nosec G306 -- fixtures model tracked public project files.
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGitTest(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(arguments, " "), err, output)
	}
	return string(output)
}

func localOnlyTestRunner(name string, arguments []string, directory string, input io.Reader, output, errorOutput io.Writer) error {
	if name == "codex" {
		return errors.New("regression test attempted to invoke metered Codex")
	}
	command := exec.Command(name, arguments...)
	command.Dir = directory
	command.Stdin = input
	command.Stdout = output
	command.Stderr = errorOutput
	return command.Run()
}
