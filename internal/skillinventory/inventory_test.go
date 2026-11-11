// ABOUTME: Exercises ownership, inventory validation and retirement boundaries.
// ABOUTME: Uses isolated fixtures without invoking a harness or touching live assets.
package skillinventory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRetirementOwnership(t *testing.T) {
	for _, links := range []bool{false, true} {
		t.Run(map[bool]string{false: "copies", true: "links"}[links], func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "source")
			destination := filepath.Join(root, "installed")
			put := func(path, text string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			put(filepath.Join(source, "old", "SKILL.md"), "owned")
			canonical := ""
			if links {
				canonical = source
				if err := os.MkdirAll(destination, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(source, "old"), filepath.Join(destination, "old")); err != nil {
					t.Fatal(err)
				}
			} else {
				put(filepath.Join(destination, "old", "SKILL.md"), "owned")
			}
			put(filepath.Join(destination, "unrelated", "SKILL.md"), "private")
			first, err := Build(source, destination, canonical, []string{"old"}, "test")
			if err != nil {
				t.Fatal(err)
			}
			put(first.Path, string(first.Contents))
			same, err := Build(source, destination, canonical, []string{"old"}, "test")
			if err != nil || same.Changed {
				t.Fatalf("repeat: %+v %v", same, err)
			}
			retired, err := Build(source, destination, canonical, nil, "test")
			if err != nil || len(retired.Retirements) != 1 {
				t.Fatalf("retirement: %+v %v", retired, err)
			}
			if links {
				if err := os.Rename(filepath.Join(destination, "old"), filepath.Join(root, "saved-link")); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(root, "elsewhere"), filepath.Join(destination, "old")); err != nil {
					t.Fatal(err)
				}
			} else {
				put(filepath.Join(destination, "old", "extra.txt"), "user addition")
			}
			if _, err := Build(source, destination, canonical, nil, "test"); err == nil {
				t.Fatal("accepted changed retired skill")
			}
			if err := retired.Retirements[0].Validate(); err == nil {
				t.Fatal("revalidation accepted changed skill")
			}
		})
	}
}

func TestRejectsInvalidInventory(t *testing.T) {
	for _, body := range []string{`{`, `{"version":2,"owner":"test","skills":{}}`, `{"version":1,"owner":"other","skills":{}}`, `{"version":1,"owner":"test","skills":{"../outside":"tree:bad"}}`} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ".test-inventory.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Build(root, root, "", nil, "test"); err == nil {
			t.Fatalf("accepted %s", body)
		}
	}
}
