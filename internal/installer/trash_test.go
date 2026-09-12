package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The command double moves exact test paths into a private recovery tree.
// No regression test may invoke the host's real Trash.
func installTrashDouble(t *testing.T, bin, root string) {
	t.Helper()
	t.Setenv("TEST_TRASH_ROOT", filepath.Join(root, "trash"))
	script := `#!/bin/sh
set -eu
dest="$TEST_TRASH_ROOT$1"
while [ -e "$dest" ] || [ -L "$dest" ]; do dest="$dest.next"; done
/bin/mkdir -p "${dest%/*}"
/bin/mv "$1" "$dest"
`
	writeFixtureFile(t, filepath.Join(bin, "trash"), script)
	if err := os.Chmod(filepath.Join(bin, "trash"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func assertTrashed(t *testing.T, path string) string {
	t.Helper()
	if matches, err := filepath.Glob(path + ".*.bak"); err != nil || len(matches) != 0 {
		t.Fatalf("managed artifact has adjacent backups: %v, %v", matches, err)
	}
	recovered := os.Getenv("TEST_TRASH_ROOT") + path
	if _, err := os.Lstat(recovered); err != nil {
		t.Fatal(err)
	}
	return recovered
}

// RT014.1: replaced managed data is recoverable without adjacent backup files.
func TestManagedReplacementUsesTrash(t *testing.T) {
	root := t.TempDir()
	installHarnessExecutables(t, root)
	source, destination := filepath.Join(root, "source"), filepath.Join(root, "live")
	writeFixtureFile(t, source, "new")
	writeFixtureFile(t, destination, "old")
	var output bytes.Buffer
	plan := installationPlan{syncs: []managedSync{{source: source, destination: destination, needsSync: true}}}
	if err := applyInstallation(plan, &output); err != nil {
		t.Fatal(err)
	}
	assertFixtureContent(t, destination, "new")
	assertFixtureContent(t, assertTrashed(t, destination), "old")
	if !strings.Contains(output.String(), "Trashed:") {
		t.Fatal(output.String())
	}
}

// RT014.3: configuration merges retain an exact adjacent recovery copy.
func TestConfigurationRetainsAdjacentBackup(t *testing.T) {
	root := t.TempDir()
	installHarnessExecutables(t, root)
	path := filepath.Join(root, "config.yaml")
	writeFixtureFile(t, path, "personal: old\n")
	change := &configurationChange{path: path, contents: []byte("personal: old\nmanaged: new\n"), mode: 0o600}
	if err := applyConfigurationChange(change, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	backups, err := filepath.Glob(path + ".*.bak")
	if err != nil || len(backups) != 1 {
		t.Fatalf("backups = %v, %v", backups, err)
	}
	assertFixtureContent(t, backups[0], "personal: old\n")
	assertFixtureContent(t, path, string(change.contents))
	if _, err := os.Lstat(os.Getenv("TEST_TRASH_ROOT") + path); !os.IsNotExist(err) {
		t.Fatalf("configuration was trashed: %v", err)
	}
}

func TestMissingSourceDoesNotTrashDestination(t *testing.T) {
	root := t.TempDir()
	installHarnessExecutables(t, root)
	path := filepath.Join(root, "live")
	writeFixtureFile(t, path, "old")
	err := synchronizeFile(managedSync{source: filepath.Join(root, "missing"), destination: path}, 0, &bytes.Buffer{})
	if err == nil {
		t.Fatal("missing source accepted")
	}
	assertFixtureContent(t, path, "old")
}

// RT014.2: missing, failed or dishonest trash cannot authorize replacement.
func TestTrashFailurePreservesDestination(t *testing.T) {
	for _, mode := range []string{"missing", "failure", "no-move"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			bin := filepath.Join(root, "bin")
			if err := os.MkdirAll(bin, 0o755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin)
			if mode != "missing" {
				body := "#!/bin/sh\nexit 1\n"
				if mode == "no-move" {
					body = "#!/bin/sh\nexit 0\n"
				}
				writeFixtureFile(t, filepath.Join(bin, "trash"), body)
				if err := os.Chmod(filepath.Join(bin, "trash"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			source, destination := filepath.Join(root, "source"), filepath.Join(root, "live")
			writeFixtureFile(t, source, "new")
			writeFixtureFile(t, destination, "old")
			plan := installationPlan{syncs: []managedSync{{source: source, destination: destination, needsSync: true}}}
			if err := applyInstallation(plan, &bytes.Buffer{}); err == nil {
				t.Fatal("replacement ignored trash failure")
			}
			assertFixtureContent(t, destination, "old")
		})
	}
}
