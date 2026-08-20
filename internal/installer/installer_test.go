package installer_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/tigger-developer/sdlc/internal/installer"
)

func TestDryRunReportsPlanWithoutWriting_RT3_1(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Input:     strings.NewReader(""),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertContains(t, output.String(), "DRY RUN")
	assertContains(t, output.String(), "would synchronize")
	assertContains(t, output.String(), "would create symlink")
	_, err = os.Lstat(filepath.Join(agentHome, "sdlc"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry run created destination: Lstat() error = %v", err)
	}
	_, err = os.Lstat(filepath.Join(filepath.Dir(agentHome), ".agents", "sdlc"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry run created live SDLC tree: Lstat() error = %v", err)
	}
}

func TestApplyCopiesCanonicalLiveTreeAndLinksClaude_RT3_2(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	commonHome := filepath.Join(filepath.Dir(agentHome), ".agents")
	liveSDLC := filepath.Join(commonHome, "sdlc")
	sharedSkills := filepath.Join(commonHome, "skills")
	writeFile(t, filepath.Join(liveSDLC, ".git", "stale"), []byte("old metadata\n"))
	writeFile(t, filepath.Join(liveSDLC, "stale.txt"), []byte("old live content\n"))
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertDirectoryNotSymlink(t, liveSDLC)
	assertFileEquals(t, filepath.Join(liveSDLC, "MAIN.md"), []byte("# SDLC\n"))
	assertFileEquals(t, filepath.Join(liveSDLC, "skills", "draft-design-issue", "references", "nested.md"), []byte("nested draft skill\n"))
	if _, err := os.Lstat(filepath.Join(liveSDLC, ".git")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("live SDLC tree contains staging Git metadata: Lstat() error = %v", err)
	}
	if _, err := os.Lstat(filepath.Join(liveSDLC, "stale.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("live SDLC tree contains stale content: Lstat() error = %v", err)
	}
	assertSymlinkTarget(t, filepath.Join(agentHome, "sdlc"), liveSDLC)
	assertSymlinkTarget(t, filepath.Join(agentHome, "commands", "build.md"), filepath.Join(liveSDLC, "commands", "build.md"))
	assertSymlinkTarget(t, filepath.Join(sharedSkills, "audit-code"), filepath.Join(liveSDLC, "skills", "audit-code"))
	assertSymlinkTarget(t, filepath.Join(sharedSkills, "draft-design-issue"), filepath.Join(liveSDLC, "skills", "draft-design-issue"))
	assertSymlinkTarget(t, filepath.Join(agentHome, "skills", "audit-code"), filepath.Join(sharedSkills, "audit-code"))
	assertSymlinkTarget(t, filepath.Join(agentHome, "skills", "draft-design-issue"), filepath.Join(sharedSkills, "draft-design-issue"))
	assertContains(t, output.String(), "installed")
}

func TestCodexApplyLinksLivePromptLibraryAndAllSkills_RT3_3(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	commonHome := filepath.Join(filepath.Dir(agentHome), ".agents")
	liveSDLC := filepath.Join(commonHome, "sdlc")
	assertSymlinkTarget(t, filepath.Join(agentHome, "sdlc"), liveSDLC)
	assertSymlinkTarget(t, filepath.Join(agentHome, "prompts-commands"), filepath.Join(liveSDLC, "commands"))
	sharedSkills := filepath.Join(commonHome, "skills")
	assertSymlinkTarget(t, filepath.Join(sharedSkills, "audit-code"), filepath.Join(liveSDLC, "skills", "audit-code"))
	assertSymlinkTarget(t, filepath.Join(sharedSkills, "draft-design-issue"), filepath.Join(liveSDLC, "skills", "draft-design-issue"))
}

func TestCopilotApplyLinksAllLiveSkills_RT3_4(t *testing.T) {
	source, agentHome := newFixture(t, ".copilot")

	err := installer.Run(installer.Options{
		Agent:     "copilot",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	commonHome := filepath.Join(filepath.Dir(agentHome), ".agents")
	liveSDLC := filepath.Join(commonHome, "sdlc")
	sharedSkills := filepath.Join(commonHome, "skills")
	assertSymlinkTarget(t, filepath.Join(agentHome, "prompts-commands"), filepath.Join(liveSDLC, "commands"))
	assertSymlinkTarget(t, filepath.Join(agentHome, "skills", "audit-code"), filepath.Join(sharedSkills, "audit-code"))
	assertSymlinkTarget(t, filepath.Join(agentHome, "skills", "draft-design-issue"), filepath.Join(sharedSkills, "draft-design-issue"))
}

func TestHermesApplyLinksAllLiveSkills_RT3_5(t *testing.T) {
	source, agentHome := newFixture(t, ".hermes")

	err := installer.Run(installer.Options{
		Agent:     "hermes",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	commonHome := filepath.Join(filepath.Dir(agentHome), ".agents")
	liveSDLC := filepath.Join(commonHome, "sdlc")
	sharedSkills := filepath.Join(commonHome, "skills")
	assertSymlinkTarget(t, filepath.Join(agentHome, "sdlc"), liveSDLC)
	assertSymlinkTarget(t, filepath.Join(agentHome, "skills", "audit-code"), filepath.Join(sharedSkills, "audit-code"))
	assertSymlinkTarget(t, filepath.Join(agentHome, "skills", "draft-design-issue"), filepath.Join(sharedSkills, "draft-design-issue"))
}

func TestApplyRefusesToReplaceExistingDestination(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	destination := filepath.Join(agentHome, "sdlc")
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	marker := filepath.Join(destination, "keep-me")
	if err := os.WriteFile(marker, []byte("authoritative\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want destination conflict")
	}
	assertContains(t, err.Error(), "already exists")
	data, readErr := os.ReadFile(marker)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if string(data) != "authoritative\n" {
		t.Fatalf("existing destination changed: marker = %q", data)
	}
}

func TestAgentIsInferredFromAgentHome(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "auto",
		AgentHome: agentHome,
		Source:    source,
		Input:     strings.NewReader(""),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertContains(t, output.String(), "Agent: codex")
}

func TestAgentIsInferredWhenCloneAlreadyLivesInAgentHome_RT3_6(t *testing.T) {
	root := t.TempDir()
	agentHome := filepath.Join(root, ".codex")
	source := filepath.Join(agentHome, "sdlc")
	writeFile(t, filepath.Join(source, "MAIN.md"), []byte("# SDLC\n"))
	writeFile(t, filepath.Join(source, "README.md"), []byte("# Quickstart\n"))
	writeFile(t, filepath.Join(source, "templates", "hermes-bootstrap.md"), []byte("Before acting in a new session, read `~/.agents/OPERATIONS.md` in full and state that you have done so. After context compaction, read it again before continuing. If you cannot read it in full, stop and explain why.\n"))
	writeFile(t, filepath.Join(source, "templates", "codex-sdlc.rules.example"), []byte("prefix_rule()\n"))
	writeFile(t, filepath.Join(source, "commands", "build.md"), []byte("# Build\n"))
	writeFile(t, filepath.Join(source, "skills", "audit-code", "SKILL.md"), []byte("# Audit\n"))
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:  "auto",
		Source: source,
		Input:  strings.NewReader(""),
		Output: &output,
	})
	if err == nil {
		t.Fatal("Run() error = nil, want staging-location conflict")
	}
	assertContains(t, err.Error(), "move the staging clone")
}

func TestApplyIsIdempotentForManagedLinks_RT3_7(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	options := installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	}

	if err := installer.Run(options); err != nil {
		t.Fatalf("first Run() error = %v", err)
	}
	var secondOutput bytes.Buffer
	options.Output = &secondOutput
	if err := installer.Run(options); err != nil {
		t.Fatalf("second Run() error = %v", err)
	}
	assertContains(t, secondOutput.String(), "already resolves")
	assertContains(t, secondOutput.String(), "already matches")
}

func TestApplyMigratesRecognizedStagingLinks_RT3_8(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	commonHome := filepath.Join(filepath.Dir(agentHome), ".agents")
	liveSDLC := filepath.Join(commonHome, "sdlc")
	sharedSkills := filepath.Join(commonHome, "skills")
	makeSymlink(t, source, filepath.Join(agentHome, "sdlc"))
	makeSymlink(t, filepath.Join(source, "commands", "build.md"), filepath.Join(agentHome, "commands", "build.md"))
	for _, skill := range []string{"audit-code", "draft-design-issue"} {
		stagingSkill := filepath.Join(source, "skills", skill)
		makeSymlink(t, stagingSkill, filepath.Join(sharedSkills, skill))
		makeSymlink(t, stagingSkill, filepath.Join(agentHome, "skills", skill))
	}

	if err := installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader("no\n"),
		Output:    &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertDirectoryNotSymlink(t, liveSDLC)
	assertSymlinkTarget(t, filepath.Join(agentHome, "sdlc"), liveSDLC)
	assertSymlinkTarget(t, filepath.Join(agentHome, "commands", "build.md"), filepath.Join(liveSDLC, "commands", "build.md"))
	for _, skill := range []string{"audit-code", "draft-design-issue"} {
		assertSymlinkTarget(t, filepath.Join(sharedSkills, skill), filepath.Join(liveSDLC, "skills", skill))
		assertSymlinkTarget(t, filepath.Join(agentHome, "skills", skill), filepath.Join(sharedSkills, skill))
	}
}

func TestStagingChangesRequireDeployment_RT3_9(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	options := installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Apply:     true,
		Input:     strings.NewReader(""),
		Output:    &bytes.Buffer{},
	}
	if err := installer.Run(options); err != nil {
		t.Fatalf("first Run() error = %v", err)
	}

	liveMain := filepath.Join(filepath.Dir(agentHome), ".agents", "sdlc", "MAIN.md")
	writeFile(t, filepath.Join(source, "MAIN.md"), []byte("# Changed staging SDLC\n"))
	assertFileEquals(t, liveMain, []byte("# SDLC\n"))

	if err := installer.Run(options); err != nil {
		t.Fatalf("second Run() error = %v", err)
	}
	assertFileEquals(t, liveMain, []byte("# Changed staging SDLC\n"))
}

func TestSharedDriftProducesOneConfirmation_RT4_1(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex")
	output := runInteractive(t, source, root, "no\nno\n")
	if strings.Count(output, "Deploy shared SDLC tree? [yes/no]:") != 1 {
		t.Fatalf("shared confirmation count was not one: %q", output)
	}
}

func TestDecliningSharedDeploymentLeavesLiveTreeUnchanged_RT4_2(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex")
	marker := filepath.Join(root, ".agents", "personal.txt")
	writeFile(t, marker, []byte("personal\n"))
	runInteractive(t, source, root, "no\nno\n")
	assertFileEquals(t, marker, []byte("personal\n"))
	if _, err := os.Lstat(filepath.Join(root, ".agents", "sdlc")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("declined shared deployment created live tree: %v", err)
	}
}

func TestAcceptingSharedDeploymentSynchronizesOwnedTree_RT4_3(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex")
	runInteractive(t, source, root, "yes\nno\n")
	live := filepath.Join(root, ".agents", "sdlc")
	assertDirectoryNotSymlink(t, live)
	assertFileEquals(t, filepath.Join(live, "MAIN.md"), []byte("# SDLC\n"))
	if _, err := os.Lstat(filepath.Join(live, ".git")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("root Git metadata was deployed: %v", err)
	}
}

func TestCurrentProviderAdaptersNeedNoConfirmation_RT4_5(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex")
	runInteractive(t, source, root, "yes\nyes\n")
	output := runInteractive(t, source, root, "")
	assertContains(t, output, "Provider codex adapters are current.")
	if strings.Contains(output, "[yes/no]") {
		t.Fatalf("current deployment requested confirmation: %q", output)
	}
}

func TestEachChangedProviderGetsOneConfirmation_RT4_6(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex", "hermes")
	runInteractive(t, source, root, "yes\nno\nno\n")
	output := runInteractive(t, source, root, "no\nno\n")
	for _, agent := range []string{"codex", "hermes"} {
		if strings.Count(output, "Install "+agent+" adapters? [yes/no]:") != 1 {
			t.Fatalf("%s confirmation count was not one: %q", agent, output)
		}
	}
}

func TestMixedProviderResponsesApplyOnlyAcceptedAdapters_RT4_7(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex", "hermes")
	runInteractive(t, source, root, "yes\nyes\nno\n")
	assertProviderBaseLink(t, root, "codex")
	if _, err := os.Lstat(filepath.Join(root, ".hermes", "sdlc")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("declined Hermes adapter changed: %v", err)
	}
}

func TestDecliningAllPromptsLeavesDestinationsUnchanged_RT4_8(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex", "hermes")
	marker := filepath.Join(root, ".hermes", "personal.txt")
	writeFile(t, marker, []byte("personal\n"))
	runInteractive(t, source, root, "no\nno\nno\n")
	assertFileEquals(t, marker, []byte("personal\n"))
	if _, err := os.Lstat(filepath.Join(root, ".codex", "sdlc")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("declined Codex adapter changed: %v", err)
	}
}

func TestInteractiveInstallationDoesNotChangeProviderConfiguration_RT4_9(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex", "hermes")
	codexConfig := filepath.Join(root, ".codex", "config.toml")
	hermesConfig := filepath.Join(root, ".hermes", "config.yaml")
	writeFile(t, codexConfig, []byte("approval_policy = \"personal\"\n"))
	writeFile(t, hermesConfig, []byte("personal: true\n"))
	runInteractive(t, source, root, "yes\nyes\nyes\n")
	assertFileEquals(t, codexConfig, []byte("approval_policy = \"personal\"\n"))
	assertFileEquals(t, hermesConfig, []byte("personal: true\n"))
}

func TestEverySourceSkillGetsACommonLiveLink_RT4_10(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex")
	writeFile(t, filepath.Join(source, "skills", "new-skill", "SKILL.md"), []byte("# New skill\n"))
	writeFile(t, filepath.Join(source, "skills", "new-skill", "nested", "reference.md"), []byte("nested\n"))
	runInteractive(t, source, root, "yes\nno\n")
	assertAllCommonSkillLinks(t, source, root)
}

func TestProviderAdapterTableCoversEveryDetectedAgent_RT4_11(t *testing.T) {
	source, root := newInteractiveFixture(t, "claude", "codex", "copilot", "hermes")
	runInteractive(t, source, root, "yes\nyes\nyes\nyes\nyes\n")
	for _, agent := range []string{"claude", "codex", "copilot", "hermes"} {
		assertProviderBaseLink(t, root, agent)
	}
	assertSymlinkTarget(t, filepath.Join(root, ".claude", "commands", "build.md"), filepath.Join(root, ".agents", "sdlc", "commands", "build.md"))
	assertSymlinkTarget(t, filepath.Join(root, ".codex", "prompts-commands"), filepath.Join(root, ".agents", "sdlc", "commands"))
	assertSymlinkTarget(t, filepath.Join(root, ".copilot", "prompts-commands"), filepath.Join(root, ".agents", "sdlc", "commands"))
	for _, agent := range []string{"claude", "copilot", "hermes"} {
		assertAllProviderSkillLinks(t, source, root, agent)
	}
}

func TestUnrelatedCommonAndProviderEntriesSurvive_RT4_12(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex")
	commonFile := filepath.Join(root, ".agents", "personal.txt")
	commonSkill := filepath.Join(root, ".agents", "skills", "personal", "SKILL.md")
	providerFile := filepath.Join(root, ".codex", "personal.txt")
	writeFile(t, commonFile, []byte("common\n"))
	writeFile(t, commonSkill, []byte("skill\n"))
	writeFile(t, providerFile, []byte("provider\n"))
	runInteractive(t, source, root, "yes\nyes\n")
	assertFileEquals(t, commonFile, []byte("common\n"))
	assertFileEquals(t, commonSkill, []byte("skill\n"))
	assertFileEquals(t, providerFile, []byte("provider\n"))
}

func TestRepeatedInteractiveInstallHasNoFilesystemWork_RT4_13(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex")
	runInteractive(t, source, root, "yes\nyes\n")
	output := runInteractive(t, source, root, "")
	if strings.Contains(output, "would synchronize") || strings.Contains(output, "would create symlink") || strings.Contains(output, "[yes/no]") {
		t.Fatalf("repeated install planned filesystem work: %q", output)
	}
}

func TestMatchingInteractiveAnalysisReportsCurrent_RT4_14(t *testing.T) {
	source, root := newInteractiveFixture(t, "codex")
	runInteractive(t, source, root, "yes\nyes\n")
	output := runInteractive(t, source, root, "")
	assertContains(t, output, "Shared deployment is current.")
	assertContains(t, output, "Provider codex adapters are current.")
}

func TestClaudeAnalysisNamesMissingCommandRestrictions(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	settings := []byte(`{"permissions":{"allow":["Read(**/*)"],"deny":[]},"custom":"preserve"}`)
	writeFile(t, filepath.Join(agentHome, "settings.json"), settings)
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Input:     strings.NewReader(""),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	for _, rule := range []string{"Bash(rm:*)", "Bash(sed:*)", "Bash(awk:*)", "Bash(source:*)", "Bash(python:*)", "Bash(python3:*)"} {
		assertContains(t, output.String(), rule)
	}
}

func TestClaudeConfigurationRequiresConfirmationAndPreservesUnknownFields(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	settingsPath := filepath.Join(agentHome, "settings.json")
	original := []byte(`{"permissions":{"allow":["Read(**/*)","Bash(sed:*)","Bash(python:*)","Bash(python3:*)"],"deny":[]},"custom":"preserve"}`)
	writeFile(t, settingsPath, original)
	var declinedOutput bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("no\n"),
		Output:    &declinedOutput,
	})
	if err != nil {
		t.Fatalf("declined Run() error = %v", err)
	}
	assertFileEquals(t, settingsPath, original)
	assertContains(t, declinedOutput.String(), "Configuration unchanged")

	var acceptedOutput bytes.Buffer
	err = installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &acceptedOutput,
	})
	if err != nil {
		t.Fatalf("accepted Run() error = %v", err)
	}

	var settings map[string]any
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if settings["custom"] != "preserve" {
		t.Fatalf("unknown field was not preserved: custom = %#v", settings["custom"])
	}
	permissions := settings["permissions"].(map[string]any)
	deny := permissions["deny"].([]any)
	for _, rule := range []string{"Bash(rm:*)", "Bash(sed:*)", "Bash(awk:*)", "Bash(source:*)", "Bash(python:*)", "Bash(python3:*)"} {
		if !containsAnyString(deny, rule) {
			t.Errorf("deny list missing %q: %#v", rule, deny)
		}
	}
	allow := permissions["allow"].([]any)
	for _, rule := range []string{"Bash(sed:*)", "Bash(python:*)", "Bash(python3:*)"} {
		if containsAnyString(allow, rule) {
			t.Errorf("allow list retains conflicting %q: %#v", rule, allow)
		}
	}
	if !containsAnyString(allow, "Read(**/*)") {
		t.Errorf("unrelated allow entry was not preserved: %#v", allow)
	}
	assertContains(t, acceptedOutput.String(), "Proposed configuration change")
	assertContains(t, acceptedOutput.String(), "Configuration updated")
	if strings.Contains(acceptedOutput.String(), "DRY RUN") {
		t.Fatalf("configuration-writing run was labelled dry-run: %q", acceptedOutput.String())
	}
}

func TestClaudeConfigurationDoesNotReplaceSettingsSymlink(t *testing.T) {
	source, agentHome := newFixture(t, ".claude")
	realSettings := filepath.Join(filepath.Dir(agentHome), "dotfiles", "claude-settings.json")
	original := []byte(`{"permissions":{"deny":[]},"custom":"preserve"}`)
	writeFile(t, realSettings, original)
	settingsPath := filepath.Join(agentHome, "settings.json")
	if err := os.Symlink(realSettings, settingsPath); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "claude",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileEquals(t, realSettings, original)
	if info, err := os.Lstat(settingsPath); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("settings symlink was replaced: info = %#v, error = %v", info, err)
	}
	assertContains(t, output.String(), "symlink")
	assertContains(t, output.String(), "no automatic change")
}

func TestCodexConfigurationCreatesRulesOnlyAfterConfirmation(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	writeFile(t, filepath.Join(agentHome, "config.toml"), []byte("approval_policy = \"on-request\"\nsandbox_mode = \"workspace-write\"\n"))
	rulesPath := filepath.Join(agentHome, "rules", "sdlc.rules")
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	data, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", rulesPath, err)
	}
	assertContains(t, string(data), `pattern = ["sed"]`)
	assertContains(t, string(data), `pattern = ["python"]`)
	assertContains(t, string(data), `pattern = ["python3"]`)
	assertContains(t, output.String(), "Configuration updated")
}

func TestCodexConfigurationUpgradesRecognizedRulesAndIsIdempotent(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	rulesPath := filepath.Join(agentHome, "rules", "sdlc.rules")
	legacy := []byte("# Existing SDLC rules.\n" +
		"prefix_rule(\n    pattern = [\"rm\"],\n    decision = \"forbidden\",\n)\n\n" +
		"prefix_rule(\n    pattern = [\"sed\"],\n    decision = \"forbidden\",\n)\n\n" +
		"prefix_rule(\n    pattern = [\"awk\"],\n    decision = \"forbidden\",\n)\n\n" +
		"prefix_rule(\n    pattern = [\"source\"],\n    decision = \"forbidden\",\n)\n\n" +
		"# Unrelated local rule.\nprefix_rule(\n    pattern = [\"local-tool\"],\n    decision = \"prompt\",\n)\n")
	writeFile(t, rulesPath, legacy)

	var firstOutput bytes.Buffer
	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &firstOutput,
	})
	if err != nil {
		t.Fatalf("first Run() error = %v", err)
	}
	upgraded, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", rulesPath, err)
	}
	for _, expected := range []string{`pattern = ["python"]`, `pattern = ["python3"]`, `pattern = ["local-tool"]`, "BEGIN SDLC MANAGED PYTHON RULES"} {
		assertContains(t, string(upgraded), expected)
	}
	assertContains(t, firstOutput.String(), "Configuration updated")

	var secondOutput bytes.Buffer
	err = installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &secondOutput,
	})
	if err != nil {
		t.Fatalf("second Run() error = %v", err)
	}
	assertFileEquals(t, rulesPath, upgraded)
	assertContains(t, secondOutput.String(), "already contain the SDLC command restrictions")
}

func TestCodexConfigurationLeavesAmbiguousRulesUnchanged(t *testing.T) {
	source, agentHome := newFixture(t, ".codex")
	rulesPath := filepath.Join(agentHome, "rules", "sdlc.rules")
	original := []byte("prefix_rule(\n    pattern = [\"local-tool\"],\n    decision = \"prompt\",\n)\n")
	writeFile(t, rulesPath, original)
	var output bytes.Buffer

	err := installer.Run(installer.Options{
		Agent:     "codex",
		AgentHome: agentHome,
		Source:    source,
		Configure: true,
		Input:     strings.NewReader("yes\n"),
		Output:    &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileEquals(t, rulesPath, original)
	assertContains(t, output.String(), "cannot safely migrate")
}

func TestMalformedProviderConfigurationStopsWithContext(t *testing.T) {
	tests := []struct {
		name       string
		agent      string
		homeName   string
		configName string
		contents   string
	}{
		{name: "claude json", agent: "claude", homeName: ".claude", configName: "settings.json", contents: "{"},
		{name: "claude null", agent: "claude", homeName: ".claude", configName: "settings.json", contents: "null"},
		{name: "codex toml", agent: "codex", homeName: ".codex", configName: "config.toml", contents: "approval_policy = ["},
		{name: "hermes yaml", agent: "hermes", homeName: ".hermes", configName: "config.yaml", contents: "agent: ["},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source, agentHome := newFixture(t, test.homeName)
			writeFile(t, filepath.Join(agentHome, test.configName), []byte(test.contents))
			err := installer.Run(installer.Options{
				Agent:     test.agent,
				AgentHome: agentHome,
				Source:    source,
				Input:     strings.NewReader(""),
				Output:    &bytes.Buffer{},
			})
			if err == nil {
				t.Fatal("Run() error = nil, want parse error")
			}
			assertContains(t, err.Error(), test.configName)
		})
	}
}

func TestDistributedProviderExamplesParse(t *testing.T) {
	claudePath := filepath.Join("..", "..", "templates", "claude-settings.example.json")
	claudeData, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", claudePath, err)
	}
	var claudeSettings map[string]any
	if err := json.Unmarshal(claudeData, &claudeSettings); err != nil {
		t.Fatalf("Claude settings example is invalid JSON: %v", err)
	}

	codexPath := filepath.Join("..", "..", "templates", "codex-config.example.toml")
	var codexSettings map[string]any
	if _, err := toml.DecodeFile(codexPath, &codexSettings); err != nil {
		t.Fatalf("Codex config example is invalid TOML: %v", err)
	}
}

func newFixture(t *testing.T, homeName string) (string, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "canonical", "sdlc")
	agentHome := filepath.Join(root, homeName)
	if err := os.MkdirAll(filepath.Join(source, "templates"), 0o700); err != nil {
		t.Fatalf("MkdirAll(source) error = %v", err)
	}
	if err := os.MkdirAll(agentHome, 0o700); err != nil {
		t.Fatalf("MkdirAll(agentHome) error = %v", err)
	}
	writeFile(t, filepath.Join(source, "MAIN.md"), []byte("# SDLC\n"))
	writeFile(t, filepath.Join(source, "README.md"), []byte("# Quickstart\n"))
	writeFile(t, filepath.Join(source, "templates", "hermes-bootstrap.md"), []byte("Before acting in a new session, read `~/.agents/OPERATIONS.md` in full and state that you have done so. After context compaction, read it again before continuing. If you cannot read it in full, stop and explain why.\n"))
	writeFile(t, filepath.Join(source, "templates", "codex-sdlc.rules.example"), []byte("prefix_rule(\n    pattern = [\"sed\"],\n    decision = \"forbidden\",\n)\n\n# BEGIN SDLC MANAGED PYTHON RULES\nprefix_rule(\n    pattern = [\"python\"],\n    decision = \"forbidden\",\n)\n\nprefix_rule(\n    pattern = [\"python3\"],\n    decision = \"forbidden\",\n)\n# END SDLC MANAGED PYTHON RULES\n"))
	writeFile(t, filepath.Join(source, "commands", "build.md"), []byte("# Build\n"))
	writeFile(t, filepath.Join(source, "skills", "audit-code", "SKILL.md"), []byte("# Audit code\n"))
	writeFile(t, filepath.Join(source, "skills", "draft-design-issue", "SKILL.md"), []byte("# Draft design issue\n"))
	writeFile(t, filepath.Join(source, "skills", "draft-design-issue", "references", "nested.md"), []byte("nested draft skill\n"))
	writeFile(t, filepath.Join(source, ".git", "sentinel"), []byte("staging metadata\n"))
	return source, agentHome
}

func newInteractiveFixture(t *testing.T, agents ...string) (string, string) {
	t.Helper()
	source, firstHome := newFixture(t, "."+agents[0])
	root := filepath.Dir(firstHome)
	for _, agent := range agents[1:] {
		if err := os.MkdirAll(filepath.Join(root, "."+agent), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return source, root
}

func runInteractive(t *testing.T, source, root, input string) string {
	t.Helper()
	var output bytes.Buffer
	if err := installer.RunInteractive(source, root, strings.NewReader(input), &output); err != nil {
		t.Fatalf("Run() error = %v\n%s", err, output.String())
	}
	return output.String()
}

func assertProviderBaseLink(t *testing.T, root, agent string) {
	t.Helper()
	assertSymlinkTarget(t, filepath.Join(root, "."+agent, "sdlc"), filepath.Join(root, ".agents", "sdlc"))
}

func assertAllCommonSkillLinks(t *testing.T, source, root string) {
	t.Helper()
	for _, skill := range sourceSkillNames(t, source) {
		assertSymlinkTarget(t, filepath.Join(root, ".agents", "skills", skill), filepath.Join(root, ".agents", "sdlc", "skills", skill))
	}
}

func assertAllProviderSkillLinks(t *testing.T, source, root, agent string) {
	t.Helper()
	for _, skill := range sourceSkillNames(t, source) {
		assertSymlinkTarget(t, filepath.Join(root, "."+agent, "skills", skill), filepath.Join(root, ".agents", "skills", skill))
	}
}

func sourceSkillNames(t *testing.T, source string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(source, "skills"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func makeSymlink(t *testing.T, target, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatalf("Symlink(%q, %q) error = %v", target, path, err)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("output %q does not contain %q", got, want)
	}
}

func assertFileEquals(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("file %q = %q, want %q", path, got, want)
	}
}

func containsAnyString(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func assertSymlinkTarget(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("Readlink(%q) error = %v", path, err)
	}
	if got != want {
		t.Fatalf("symlink %q target = %q, want %q", path, got, want)
	}
}

func assertDirectoryNotSymlink(t *testing.T, path string) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat(%q) error = %v", path, err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("%q is not a regular directory: mode = %v", path, info.Mode())
	}
}
