package installer

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagedSkillLinkRejectsUnknownDirectory(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	writeFixtureFile(t, filepath.Join(source, "skills", "define-change", "SKILL.md"), "public\n")
	destination := filepath.Join(root, ".agents", "skills", "define-change")
	writeFixtureFile(t, filepath.Join(destination, "SKILL.md"), "personal\n")
	owner := skillLinkOwnership(source, root, "define-change", filepath.Join(source, "skills", "define-change"))
	_, err := planManagedLink(filepath.Join(root, "opt", "sdlc", "skills", "define-change"), destination, filepath.Dir(destination), owner)
	if err == nil || !strings.Contains(err.Error(), "not verified as SDLC-owned") {
		t.Fatalf("unknown skill replacement was accepted: %v", err)
	}
	assertFixtureContent(t, filepath.Join(destination, "SKILL.md"), "personal\n")
}

func TestUnverifiedRetiredSkillNeedsManualReview(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "skills", "audit-tests")
	writeFixtureFile(t, filepath.Join(path, "SKILL.md"), "personal\n")
	if err := rejectUnverifiedRetiredSkills(filepath.Dir(path), []string{"audit-tests"}); err == nil {
		t.Fatal("unknown retired skill was accepted")
	}
	assertFixtureContent(t, filepath.Join(path, "SKILL.md"), "personal\n")
}

func TestManagedCommandLinkRejectsUnrelatedFileAndLink(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, ".local", "bin", "sdlc-init")
	writeFixtureFile(t, destination, "personal command\n")
	target := filepath.Join(root, "opt", "sdlc", "bin", "sdlc-init")
	owner := linkOwnership{legacyTargets: []string{filepath.Join(root, ".agents", "sdlc", "bin", "sdlc-init")}}
	if _, err := planManagedLink(target, destination, filepath.Dir(destination), owner); err == nil {
		t.Fatal("unrelated real command was accepted")
	}
	if err := os.Remove(destination); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "personal-command"), destination); err != nil {
		t.Fatal(err)
	}
	if _, err := planManagedLink(target, destination, filepath.Dir(destination), owner); err == nil {
		t.Fatal("unrelated command link was accepted")
	}
}

func TestManagedLinkRechecksDestinationBeforeReplacement(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "bin", "sdlc-init")
	link, err := planManagedLink(filepath.Join(root, "opt", "sdlc", "bin", "sdlc-init"), destination, filepath.Dir(destination), linkOwnership{})
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, destination, "arrived after planning\n")
	if err := synchronizeLink(link, io.Discard); err == nil {
		t.Fatal("new unrelated destination was replaced")
	}
	assertFixtureContent(t, destination, "arrived after planning\n")
}

func TestTaggedReleaseProvesOldSkillDirectoryOnlyWhenUnchanged(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	oldSkill := filepath.Join(source, "skills", "define-change", "SKILL.md")
	writeFixtureFile(t, oldSkill, "old public skill\n")
	for _, arguments := range [][]string{
		{"init", "-q"},
		{"config", "user.name", "Test"},
		{"config", "user.email", "test@example.invalid"},
		{"add", "skills"},
		{"commit", "-qm", "old release"},
		{"tag", "-a", "v1.2.3", "-m", "old release"},
	} {
		command := exec.Command("git", append([]string{"-C", source}, arguments...)...)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", arguments, err, output)
		}
	}
	writeFixtureFile(t, oldSkill, "new public skill\n")
	writeFixtureFile(t, filepath.Join(root, ".agents", "sdlc.yaml"), "version: 3\nrelease: v1.2.3\n")
	destination := filepath.Join(root, ".agents", "skills", "define-change")
	writeFixtureFile(t, filepath.Join(destination, "SKILL.md"), "old public skill\n")
	owner := skillLinkOwnership(source, root, "define-change", filepath.Join(source, "skills", "define-change"))
	matched, err := installedSkillMatchesRelease(owner, destination)
	if err != nil || !matched {
		t.Fatalf("tagged public skill was not recognized: matched=%v err=%v", matched, err)
	}
	writeFixtureFile(t, filepath.Join(destination, ".DS_Store"), "other data\n")
	matched, err = installedSkillMatchesRelease(owner, destination)
	if err != nil || !matched {
		t.Fatalf("Finder metadata prevented recognition: matched=%v err=%v", matched, err)
	}
	writeFixtureFile(t, filepath.Join(destination, "operator.txt"), "other data\n")
	matched, err = installedSkillMatchesRelease(owner, destination)
	if err != nil || matched {
		t.Fatalf("directory with unrelated skill file was accepted: matched=%v err=%v", matched, err)
	}
}

func TestGoosePluginReplacementRequiresExactOldProjection(t *testing.T) {
	root := filepath.Join(t.TempDir(), "sdlc-command-guard-goose")
	writeFixtureFile(t, filepath.Join(root, "plugin.json"), `{"name":"sdlc-command-guard-goose","version":"1.0.0","description":"Apply the SDLC shell and protected-file policy before Goose tools run."}`)
	writeFixtureFile(t, filepath.Join(root, "hooks", "hooks.json"), `{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"\"${PLUGIN_ROOT}/../../sdlc/hooks/agent-command-guard.sh\"","timeout":10,"on_failure":"block"}]}]}}`)
	matched, err := installedGoosePluginMatches(root)
	if err != nil || !matched {
		t.Fatalf("old Goose plugin was not recognized: matched=%v err=%v", matched, err)
	}
	writeFixtureFile(t, filepath.Join(root, "operator.json"), `{"personal":true}`)
	matched, err = installedGoosePluginMatches(root)
	if err != nil || matched {
		t.Fatalf("Goose plugin with unrelated file was accepted: matched=%v err=%v", matched, err)
	}
}
