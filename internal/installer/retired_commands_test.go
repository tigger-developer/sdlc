package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// RT014.6: retire only recognized command remnants, without following links.
func TestRetiredCommandCleanupPreservesOtherPaths(t *testing.T) {
	root := t.TempDir()
	installHarnessExecutables(t, root)
	bin := filepath.Join(root, "commands")
	retired := filepath.Join(bin, "sdlc-install.sdlc-retired-3")
	writeFixtureFile(t, retired, "old command")
	for _, name := range []string{"personal.sdlc-v2-retired", "sdlc-audit.sdlc-v2-retired-notes", "sdlc-init.123.bak", "sdlc-audit"} {
		writeFixtureFile(t, filepath.Join(bin, name), "preserve")
	}
	directory := filepath.Join(bin, "sdlc-init.sdlc-v3-retired")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "old-target")
	writeFixtureFile(t, target, "preserve target")
	link := filepath.Join(bin, "sdlc-audit.sdlc-v2-retired")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	retirements, err := planRetiredCommandBackups(bin)
	if err != nil || len(retirements) != 2 {
		t.Fatalf("plan: %v (%v)", retirements, err)
	}
	if err := applyInstallation(installationPlan{retirements: retirements}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	assertFixtureContent(t, assertTrashed(t, retired), "old command")
	if got, err := os.Readlink(assertTrashed(t, link)); err != nil || got != target {
		t.Fatalf("link recovery: %s (%v)", got, err)
	}
	assertFixtureContent(t, target, "preserve target")
	for _, name := range []string{"personal.sdlc-v2-retired", "sdlc-audit.sdlc-v2-retired-notes", "sdlc-init.123.bak", "sdlc-audit"} {
		assertFixtureContent(t, filepath.Join(bin, name), "preserve")
	}
	if info, err := os.Stat(directory); err != nil || !info.IsDir() {
		t.Fatalf("directory changed: %v", err)
	}
	for _, dir := range []string{bin, filepath.Join(root, "missing")} {
		if remaining, err := planRetiredCommandBackups(dir); err != nil || len(remaining) != 0 {
			t.Fatalf("repeat plan: %v (%v)", remaining, err)
		}
	}
}
