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

func TestRetiredAuditLinkRequiresKnownInstallerTarget(t *testing.T) {
	for _, tc := range []struct {
		name, target string
		owned        bool
	}{
		{"current", "current", true},
		{"legacy", "legacy", true},
		{"unrelated", "unrelated", false},
		{"regular file", "file", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			localBin := filepath.Join(root, ".local", "bin")
			if err := os.MkdirAll(localBin, 0o700); err != nil {
				t.Fatal(err)
			}
			installRoot := filepath.Join(root, "opt", "sdlc")
			path := filepath.Join(localBin, "sdlc-audit")
			switch tc.target {
			case "file":
				writeFixtureFile(t, path, "operator-owned\n")
			default:
				target := filepath.Join(root, "other", "tool")
				if tc.target == "current" {
					target = filepath.Join(installRoot, "bin", "sdlc-harness")
				} else if tc.target == "legacy" {
					target = filepath.Join(root, ".agents", "sdlc", "bin", "sdlc-harness")
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			}
			retirement, err := planRetiredAuditLink(localBin, installRoot, root)
			if err != nil || (retirement != nil) != tc.owned {
				t.Fatalf("retirement=%v, err=%v; owned=%v", retirement, err, tc.owned)
			}
		})
	}
}
