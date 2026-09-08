package installer

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestW004InteractiveInstallLinksCanonicalSkillsAndRegistersGuards(t *testing.T) {
	root := t.TempDir()
	installHarnessExecutables(t, root, "claude", "codex", "copilot", "hermes")
	source := filepath.Join(root, "source")
	writeFixtureFile(t, filepath.Join(source, "README.md"), "# SDLC\n")
	writeFixtureFile(t, filepath.Join(source, "src", "MAIN.md"), "# Lean SDLC\n")
	writeFixtureFile(t, filepath.Join(source, "src", "prompts", "discover-project-authorities.md"), "Discover project authorities.\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "audit-code", "SKILL.md"), "---\nname: audit-code\ndescription: Review code.\n---\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "audit-test-code", "SKILL.md"), "---\nname: audit-test-code\ndescription: Review implemented tests.\n---\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "audit-test-definitions", "SKILL.md"), "---\nname: audit-test-definitions\ndescription: Review test definitions.\n---\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "define-change", "SKILL.md"), "---\nname: define-change\ndescription: Define a change.\n---\n")
	writeGuardFixture(t, filepath.Join(source, "hooks", "agent-command-guard.sh"))
	writeFixtureFile(t, filepath.Join(source, "templates", "codex-sdlc.rules.example"), codexPythonRulesStart+"\nprefix_rule(pattern=[\"python3\"], decision=\"forbidden\")\n"+codexPythonRulesEnd+"\n")
	writeCommandFixtures(t, source)

	for _, provider := range []string{"claude", "codex", "copilot", "hermes"} {
		if err := os.MkdirAll(filepath.Join(root, "."+provider), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, provider := range []string{"claude", "copilot", "hermes"} {
		writeFixtureFile(t, filepath.Join(root, "."+provider, "skills", "audit-code", "SKILL.md"), "old\n")
		writeFixtureFile(t, filepath.Join(root, "."+provider, "skills", "define-change", "SKILL.md"), "old\n")
	}
	writeFixtureFile(t, filepath.Join(root, ".copilot", "hooks", "sdlc-tool-guard.json"), "{}\n")
	writeFixtureFile(t, filepath.Join(root, ".agents", "sdlc", "skills", "audit-code", "SKILL.md"), "duplicate\n")
	writeFixtureFile(t, filepath.Join(root, ".agents", "skills", "audit-tests", "SKILL.md"), "obsolete\n")
	writeFixtureFile(t, filepath.Join(root, ".claude", "settings.json"), `{"personal":true,"hooks":{"PreToolUse":[{"matcher":"*","hooks":[{"type":"command","command":"bash ~/.agents/sdlc/hooks/agent-command-guard.sh","timeout":5}]}]}}`)
	writeFixtureFile(t, filepath.Join(root, ".hermes", "config.yaml"), "personal: true\nhooks:\n  pre_tool_call:\n    - command: bash ~/.agents/sdlc/hooks/agent-command-guard.sh\n      matcher: .*\n      timeout: 5\n")
	writeFixtureFile(t, filepath.Join(root, ".local", "bin", "sdlc-install"), "old installer link target\n")
	writeFixtureFile(t, filepath.Join(root, ".local", "bin", "sdlc-project-init"), "old initializer link target\n")
	writeFixtureFile(t, filepath.Join(root, ".local", "bin", "sdlc-audit"), "old audit runner\n")
	writeFixtureFile(t, filepath.Join(root, ".local", "bin", "sdlc-project-update"), "old updater\n")
	unrelatedPath := filepath.Join(root, ".claude", "unrelated.txt")
	writeFixtureFile(t, unrelatedPath, "operator-owned bytes\n")
	checkoutLink := filepath.Join(root, ".local", "bin", "sdlc-init")
	if err := os.Symlink(filepath.Join(source, "bin", "sdlc-init"), checkoutLink); err != nil {
		t.Fatal(err)
	}

	guardCountPath := filepath.Join(root, "guard-count")
	t.Setenv("SDLC_GUARD_COUNT_FILE", guardCountPath)
	var output bytes.Buffer
	if err := RunInteractive(source, root, "v3.0.0", strings.NewReader("yes\nyes\n"), &output); err != nil {
		t.Fatalf("install: %v\n%s", err, output.String())
	}
	for _, provider := range []string{"claude", "codex", "copilot", "hermes"} {
		if !strings.Contains(output.String(), "Harness "+provider+": interactive=READY; external-audit=READY") {
			t.Fatalf("missing %s readiness result:\n%s", provider, output.String())
		}
	}
	if count := strings.Count(readFixtureFile(t, guardCountPath), "invoked\n"); count != 4 {
		t.Fatalf("guard invocation count = %d, want 4", count)
	}
	assertFixtureContent(t, unrelatedPath, "operator-owned bytes\n")
	assertFixtureContent(t, filepath.Join(root, ".agents", "sdlc", "MAIN.md"), "# Lean SDLC\n")
	assertFixtureContent(t, filepath.Join(root, ".agents", "skills", "audit-code", "SKILL.md"), "---\nname: audit-code\ndescription: Review code.\n---\n")
	assertFixtureContent(t, filepath.Join(root, ".agents", "skills", "audit-test-code", "SKILL.md"), "---\nname: audit-test-code\ndescription: Review implemented tests.\n---\n")
	assertFixtureContent(t, filepath.Join(root, ".agents", "skills", "audit-test-definitions", "SKILL.md"), "---\nname: audit-test-definitions\ndescription: Review test definitions.\n---\n")
	assertFixtureContent(t, filepath.Join(root, ".agents", "skills", "define-change", "SKILL.md"), "---\nname: define-change\ndescription: Define a change.\n---\n")
	for _, name := range deployedCommandNames {
		deployed := filepath.Join(root, ".agents", "sdlc", "bin", name)
		assertFixtureContent(t, deployed, name+" executable\n")
		link := filepath.Join(root, ".local", "bin", name)
		target, err := os.Readlink(link)
		if err != nil {
			t.Fatalf("reading installed command link %s: %v", link, err)
		}
		if target != deployed {
			t.Fatalf("%s -> %s, want %s", link, target, deployed)
		}
	}
	checkoutLinkBackups, err := filepath.Glob(checkoutLink + ".*.bak")
	if err != nil || len(checkoutLinkBackups) != 1 {
		t.Fatalf("source-checkout link backups = %v, %v", checkoutLinkBackups, err)
	}
	oldTarget, err := os.Readlink(checkoutLinkBackups[0])
	if err != nil {
		t.Fatalf("reading source-checkout link backup: %v", err)
	}
	if oldTarget != filepath.Join(source, "bin", "sdlc-init") {
		t.Fatalf("source-checkout link backup target = %s", oldTarget)
	}
	for _, retired := range []string{"sdlc-audit", "sdlc-install", "sdlc-project-init", "sdlc-project-update"} {
		path := filepath.Join(root, ".local", "bin", retired)
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("retired command remains active at %s: %v", path, err)
		}
		backups, err := filepath.Glob(path + ".*.bak")
		if err != nil || len(backups) != 1 {
			t.Fatalf("retired command backups for %s = %v, %v", path, backups, err)
		}
	}
	if _, err := os.Lstat(filepath.Join(root, ".agents", "skills", "audit-tests")); !os.IsNotExist(err) {
		t.Fatalf("retired global audit-tests skill remains: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(root, ".agents", "sdlc", "skills")); !os.IsNotExist(err) {
		t.Fatalf("duplicate canonical-root skills directory remains: %v", err)
	}
	for _, provider := range []string{"claude", "copilot", "hermes"} {
		for _, skill := range []string{"audit-code", "define-change"} {
			path := filepath.Join(root, "."+provider, "skills", skill)
			target, err := os.Readlink(path)
			if err != nil {
				t.Fatalf("reading %s skill link %s: %v", provider, skill, err)
			}
			want := filepath.Join(root, ".agents", "skills", skill)
			if target != want {
				t.Fatalf("%s -> %s, want %s", path, target, want)
			}
		}
	}
	claude := readFixtureFile(t, filepath.Join(root, ".claude", "settings.json"))
	if !(strings.Contains(claude, `"personal": true`) || strings.Contains(claude, `"personal":true`)) || !strings.Contains(claude, "agent-command-guard") {
		t.Fatalf("Claude configuration did not preserve personal data and install guard: %s", claude)
	}
	hermes := readFixtureFile(t, filepath.Join(root, ".hermes", "config.yaml"))
	if !strings.Contains(hermes, "personal: true") || !strings.Contains(hermes, "agent-command-guard") {
		t.Fatalf("Hermes configuration did not preserve personal data and install guard: %s", hermes)
	}
	for _, path := range []string{
		filepath.Join(root, ".codex", "hooks.json"),
		filepath.Join(root, ".copilot", "hooks", "sdlc-tool-guard.json"),
	} {
		if contents := readFixtureFile(t, path); !strings.Contains(contents, "agent-command-guard") {
			t.Fatalf("guard absent from %s: %s", path, contents)
		}
	}
	copilotBackups, err := filepath.Glob(filepath.Join(root, ".copilot", "hooks", "sdlc-tool-guard.json.*.bak"))
	if err != nil || len(copilotBackups) != 1 || readFixtureFile(t, copilotBackups[0]) != "{}\n" {
		t.Fatalf("Copilot conflict backup = %v, %v", copilotBackups, err)
	}
	global := readFixtureFile(t, filepath.Join(root, ".agents", "sdlc.yaml"))
	if !strings.Contains(global, "version: 3") || !strings.Contains(global, "release: v3.0.0") {
		t.Fatalf("global configuration does not identify the schema and deployed release:\n%s", global)
	}

	writeFixtureFile(t, filepath.Join(source, "skills", "audit-code", "SKILL.md"), "updated canonical skill\n")
	var refreshOutput bytes.Buffer
	if err := RunInteractive(source, root, "v3.0.0", strings.NewReader("yes\n"), &refreshOutput); err != nil {
		t.Fatalf("refresh install: %v\n%s", err, refreshOutput.String())
	}
	assertFixtureContent(t, filepath.Join(root, ".agents", "skills", "audit-code", "SKILL.md"), "updated canonical skill\n")
	for _, provider := range []string{"claude", "copilot", "hermes"} {
		assertFixtureContent(t, filepath.Join(root, "."+provider, "skills", "audit-code", "SKILL.md"), "updated canonical skill\n")
	}
	var noOpOutput bytes.Buffer
	if err := RunInteractive(source, root, "v3.0.0", strings.NewReader("unexpected\n"), &noOpOutput); err != nil {
		t.Fatalf("all-provider no-op: %v\n%s", err, noOpOutput.String())
	}
	if !strings.Contains(noOpOutput.String(), "All detected SDLC copies are current.") || strings.Contains(noOpOutput.String(), "Deploy all") {
		t.Fatalf("all-provider no-op output = %q", noOpOutput.String())
	}
	assertFixtureContent(t, unrelatedPath, "operator-owned bytes\n")
}

func TestW004UnknownCopilotHookDeclinePreservesBytes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".copilot", "hooks", "sdlc-tool-guard.json")
	original := "{\"personal\":true}\n"
	writeFixtureFile(t, path, original)
	change, err := analyseCopilotConfiguration(filepath.Join(root, ".copilot"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if change == nil || !change.conflict {
		t.Fatalf("unknown hook was not classified as a conflict: %#v", change)
	}
	if err := offerConfigurationChanges([]*configurationChange{change}, strings.NewReader("no\n"), &bytes.Buffer{}); err == nil {
		t.Fatal("declined unknown conflict returned success")
	}
	assertFixtureContent(t, path, original)
	backups, err := filepath.Glob(path + ".*.bak")
	if err != nil || len(backups) != 0 {
		t.Fatalf("declined conflict backups = %v, %v", backups, err)
	}
}

func TestV3InteractiveInstallIsSilentNoOpAfterSynchronization(t *testing.T) {
	root := t.TempDir()
	installHarnessExecutables(t, root, "codex")
	source := filepath.Join(root, "source")
	writeFixtureFile(t, filepath.Join(source, "README.md"), "# SDLC\n")
	writeFixtureFile(t, filepath.Join(source, "src", "MAIN.md"), "# Lean SDLC\n")
	writeFixtureFile(t, filepath.Join(source, "skills", "audit-code", "SKILL.md"), "---\nname: audit-code\ndescription: Review code.\n---\n")
	writeGuardFixture(t, filepath.Join(source, "hooks", "agent-command-guard.sh"))
	writeCommandFixtures(t, source)
	if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, filepath.Join(source, "templates", "codex-sdlc.rules.example"), codexPythonRulesStart+"\n"+codexPythonRulesEnd+"\n")
	var first bytes.Buffer
	writeFixtureFile(t, filepath.Join(root, ".agents", "sdlc.yaml"), "# operator settings\nversion: 3\nrelease: v2.1.0\ndelivery:\n  branch_strategy: feature\n")
	if err := RunInteractive(source, root, "v3.0.0", strings.NewReader("yes\n"), &first); err != nil {
		t.Fatal(err)
	}
	global := readFixtureFile(t, filepath.Join(root, ".agents", "sdlc.yaml"))
	for _, expected := range []string{"# operator settings", "version: 3", "release: v3.0.0", "branch_strategy: feature"} {
		if !strings.Contains(global, expected) {
			t.Fatalf("global configuration lost %q:\n%s", expected, global)
		}
	}
	var second bytes.Buffer
	if err := RunInteractive(source, root, "v3.0.0", strings.NewReader("unexpected\n"), &second); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(second.String(), "All detected SDLC copies are current.") || strings.Contains(second.String(), "Deploy all") {
		t.Fatalf("second run was noisy or prompted:\n%s", second.String())
	}
}

func TestW004DetectionUsesExecutableNotProviderHome(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	installHarnessExecutables(t, root, "hermes")
	agents, err := detectedAgents(root)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(agents, ",") != "hermes" {
		t.Fatalf("detected agents = %v, want executable-backed Hermes only", agents)
	}
}

func TestW004ConfigurationRestoreFailureReportsBothErrors(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.json")
	writeFixtureFile(t, path, "original\n")
	writeFailure := errors.New("write failed")
	restoreFailure := errors.New("restore failed")
	change := &configurationChange{path: path, contents: []byte("replacement\n"), mode: 0o600}
	err := applyConfigurationChangeWith(
		change,
		&bytes.Buffer{},
		func(string, []byte, os.FileMode) error { return writeFailure },
		func(string, string) error { return restoreFailure },
	)
	if !errors.Is(err, writeFailure) || !errors.Is(err, restoreFailure) || !strings.Contains(err.Error(), "restoring configuration backup") {
		t.Fatalf("configuration failure = %v", err)
	}
}

func writeFixtureFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	// #nosec G306 -- fixtures model tracked public project files.
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeCommandFixtures(t *testing.T, source string) {
	t.Helper()
	for _, name := range deployedCommandNames {
		path := filepath.Join(source, "bin", name)
		writeFixtureFile(t, path, name+" executable\n")
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range deployedInternalCommandNames {
		path := filepath.Join(source, "bin", name)
		writeFixtureFile(t, path, name+" executable\n")
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func installHarnessExecutables(t *testing.T, root string, names ...string) {
	t.Helper()
	bin := filepath.Join(root, "fake-bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		path := filepath.Join(bin, name)
		writeFixtureFile(t, path, "#!/bin/sh\nexit 0\n")
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
}

func writeGuardFixture(t *testing.T, path string) {
	t.Helper()
	contents := `#!/bin/sh
IFS= read -r payload
if [ -n "$SDLC_GUARD_COUNT_FILE" ]; then printf 'invoked\n' >> "$SDLC_GUARD_COUNT_FILE"; fi
case "$payload" in
*pre_tool_call*) printf '{"decision":"block","reason":"test"}\n'; exit 0 ;;
*) printf 'Blocked by agent-command-guard: test\n' >&2; exit 2 ;;
esac
`
	writeFixtureFile(t, path, contents)
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func readFixtureFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func assertFixtureContent(t *testing.T, path, expected string) {
	t.Helper()
	if actual := readFixtureFile(t, path); actual != expected {
		t.Fatalf("%s = %q, want %q", path, actual, expected)
	}
}
