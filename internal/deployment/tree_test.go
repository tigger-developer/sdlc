// ABOUTME: Tests clean replacement, rollback and interruption recovery in isolated roots.
// ABOUTME: Exercises real renames and a failed validation without accessing live installations.
package deployment

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
func content(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != want {
		t.Fatalf("%s: %q %v", path, data, err)
	}
}
func absent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("expected absent %s: %v", path, err)
	}
}
func candidate(t *testing.T) Tree {
	t.Helper()
	tree := Tree{}
	if err := tree.Add("docs/current.md", "fixture", []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	return tree
}
func TestCleanReplacementAndNoOp(t *testing.T) {
	root := t.TempDir()
	put(t, filepath.Join(root, "docs", "obsolete.md"), "old")
	put(t, filepath.Join(root, "profiles", "helper", "SOUL.md"), "legacy")
	tree := candidate(t)
	backup, err := tree.Install(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	content(t, filepath.Join(root, "docs", "current.md"), "new")
	absent(t, filepath.Join(root, "profiles"))
	absent(t, filepath.Join(root, "docs", "obsolete.md"))
	content(t, filepath.Join(backup, "docs", "obsolete.md"), "old")
	content(t, filepath.Join(backup, "profiles", "helper", "SOUL.md"), "legacy")
	if backup, err := tree.Install(root, nil); err != nil || backup != "" {
		t.Fatalf("repeat %q %v", backup, err)
	}
}
func TestValidationFailureRestoresPreviousTree(t *testing.T) {
	for _, activation := range []bool{false, true} {
		t.Run(map[bool]string{false: "stage", true: "activation"}[activation], func(t *testing.T) {
			root := t.TempDir()
			put(t, filepath.Join(root, "docs", "old.md"), "old")
			_, err := candidate(t).Install(root, func(path string) error {
				if (path == root) == activation {
					return errors.New("validation failure")
				}
				return nil
			})
			if err == nil {
				t.Fatal("expected failure")
			}
			content(t, filepath.Join(root, "docs", "old.md"), "old")
			absent(t, filepath.Join(root, "docs", "current.md"))
			if pending, err := Pending(root); err != nil || pending {
				t.Fatalf("unresolved recovery %v %v", pending, err)
			}
		})
	}
}
func TestInterruptedReplacementRecovery(t *testing.T) {
	for _, partialBackup := range []bool{false, true} {
		root := t.TempDir()
		txn := filepath.Join(root, housekeeping, "deployment-test")
		put(t, filepath.Join(txn, "previous", "docs", "old.md"), "old")
		if partialBackup {
			put(t, filepath.Join(root, "config"), "settings")
		} else {
			put(t, filepath.Join(txn, "previous", "config"), "settings")
			put(t, filepath.Join(root, "docs", "current.md"), "new")
			put(t, filepath.Join(root, "extra"), "new")
		}
		data, err := json.Marshal(journal{Transaction: "deployment-test", Old: []string{"docs", "config"}, New: []string{"docs", "extra"}})
		if err != nil {
			t.Fatal(err)
		}
		put(t, filepath.Join(root, housekeeping, "pending.json"), string(data))
		if err := Recover(root); err != nil {
			t.Fatal(err)
		}
		if err := Recover(root); err != nil {
			t.Fatal(err)
		}
		content(t, filepath.Join(root, "docs", "old.md"), "old")
		content(t, filepath.Join(root, "config"), "settings")
		absent(t, filepath.Join(root, "extra"))
		absent(t, filepath.Join(root, "docs", "current.md"))
	}
}
func TestConcurrentInstallerIsRejected(t *testing.T) {
	root := t.TempDir()
	lock, err := lockRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	_, installErr := candidate(t).Install(root, nil)
	closeErr := lock.Close()
	if installErr == nil || closeErr != nil {
		t.Fatalf("lock %v %v", installErr, closeErr)
	}
}

func TestAdapterFailureRollsBackTreeAndAdapters(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	put(t, filepath.Join(root, "docs", "old.md"), "old")
	config := filepath.Join(home, "config.yaml")
	put(t, config, "operator configuration")
	link := filepath.Join(home, "AGENTS.md")
	_, err := candidate(t).Install(root, nil, Integration{Paths: []string{config, link}, Apply: func() error {
		if err := os.WriteFile(config, []byte("modified"), 0o644); err != nil {
			return err
		}
		if err := os.Symlink(filepath.Join(root, "docs", "current.md"), link); err != nil {
			return err
		}
		return errors.New("adapter verification failed")
	}})
	if err == nil {
		t.Fatal("expected failure")
	}
	content(t, config, "operator configuration")
	absent(t, link)
	content(t, filepath.Join(root, "docs", "old.md"), "old")
	absent(t, filepath.Join(root, "docs", "current.md"))
}
