// ABOUTME: Stages and verifies complete installation trees beneath a writable root.
// ABOUTME: Retains previous contents and a restartable journal without replacing /opt entries.
package deployment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"syscall"
)

const housekeeping = ".installer"

type File struct {
	Source string
	Data   []byte
	Mode   os.FileMode
}
type Tree map[string]File

func (tree Tree) Add(path, source string, data []byte, mode os.FileMode) error {
	if !filepath.IsLocal(path) || path == "." || filepath.Clean(path) != path || path == housekeeping || len(path) > len(housekeeping) && path[:len(housekeeping)+1] == housekeeping+string(os.PathSeparator) {
		return fmt.Errorf("invalid deployment path %q", path)
	}
	if _, exists := tree[path]; exists {
		return fmt.Errorf("duplicate deployment path %q", path)
	}
	tree[path] = File{source, bytes.Clone(data), mode.Perm()}
	return nil
}
func (tree Tree) Copy(source, path string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("deployment source %q must be regular", source)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if info.Mode().Perm()&0o111 != 0 {
		mode = 0o755
	}
	return tree.Add(path, source, data, mode)
}
func (tree Tree) Directory(source, prefix string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		if name == ".DS_Store" || name == "Thumbs.db" || name == "desktop.ini" || len(name) > 2 && name[:2] == "._" {
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		return tree.Copy(path, filepath.Join(prefix, relative))
	})
}
func (tree Tree) Matches(root string) (bool, error) {
	if _, err := os.Lstat(root); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	seen := map[string]bool{}
	equal := true
	err := filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if relative == housekeeping {
			if e.IsDir() {
				return filepath.SkipDir
			}
			return fmt.Errorf("%s must be a directory", path)
		}
		if e.IsDir() {
			prefix := relative + string(os.PathSeparator)
			known := false
			for name := range tree {
				if len(name) > len(prefix) && name[:len(prefix)] == prefix {
					known = true
					break
				}
			}
			if !known {
				equal = false
			}
			return nil
		}
		desired, ok := tree[relative]
		if !ok || !e.Type().IsRegular() {
			equal = false
			return nil
		}
		info, err := e.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Equal(data, desired.Data) || info.Mode().Perm() != desired.Mode {
			equal = false
		}
		seen[relative] = true
		return nil
	})
	return equal && len(seen) == len(tree), err
}
func (tree Tree) Write(root string) error {
	for name, file := range tree {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, file.Data, file.Mode); err != nil {
			return err
		}
	}
	return nil
}

type journal struct {
	Transaction string
	Old, New    []string
	Adapters    []adapter
}

type adapter struct {
	Path    string
	Present bool
}
type Integration struct {
	Paths []string
	Apply func() error
}

func liveNames(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.Name() != housekeeping {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}
func exists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}
func lockRoot(root string) (*os.File, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("deployment root %q must be a real directory", root)
	}
	dir := filepath.Join(root, housekeeping)
	if info, err := os.Lstat(dir); err == nil && !info.IsDir() {
		return nil, fmt.Errorf("installer state %q must be a real directory", dir)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(dir, "lock")
	fd, err := syscall.Open(lockPath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), lockPath)
	if err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		closeErr := file.Close()
		return nil, errors.Join(fmt.Errorf("another installer holds %s: %w", root, err), closeErr)
	}
	return file, nil
}

// Pending is read-only, including on roots that have never been installed.
func Pending(root string) (bool, error) {
	return exists(filepath.Join(root, housekeeping, "pending.json"))
}
func Recover(root string) (err error) {
	pending, err := Pending(root)
	if err != nil || !pending {
		return err
	}
	lock, err := lockRoot(root)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	return recoverLocked(root)
}
func readJournal(root string) (journal, error) {
	var j journal
	path := filepath.Join(root, housekeeping, "pending.json")
	info, err := os.Lstat(path)
	if err != nil {
		return j, err
	}
	if !info.Mode().IsRegular() {
		return j, fmt.Errorf("invalid recovery journal %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return j, err
	}
	if err = json.Unmarshal(data, &j); err != nil {
		return j, err
	}
	if !simpleName(j.Transaction) {
		return j, fmt.Errorf("invalid recovery transaction")
	}
	for _, name := range append(append([]string{}, j.Old...), j.New...) {
		if !simpleName(name) || name == housekeeping {
			return j, fmt.Errorf("invalid recovery entry %q", name)
		}
	}
	for _, entry := range j.Adapters {
		if !filepath.IsAbs(entry.Path) || filepath.Clean(entry.Path) != entry.Path || entry.Path == string(os.PathSeparator) {
			return j, fmt.Errorf("invalid adapter recovery path")
		}
	}
	return j, nil
}
func simpleName(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.Base(name) == name
}
func recoverLocked(root string) error {
	j, err := readJournal(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	transaction := filepath.Join(root, housekeeping, j.Transaction)
	failed := filepath.Join(transaction, "failed")
	if err := os.MkdirAll(failed, 0o700); err != nil {
		return err
	}
	old := map[string]bool{}
	for _, name := range j.Old {
		old[name] = true
	}
	for _, name := range j.New {
		if old[name] {
			continue
		}
		path := filepath.Join(root, name)
		present, err := exists(path)
		if err != nil {
			return err
		}
		if present {
			if err := moveAside(path, failed, name); err != nil {
				return err
			}
		}
	}
	for _, name := range j.Old {
		previous := filepath.Join(transaction, "previous", name)
		saved, err := exists(previous)
		if err != nil {
			return err
		}
		if !saved {
			continue
		}
		path := filepath.Join(root, name)
		present, err := exists(path)
		if err != nil {
			return err
		}
		if present {
			if err := moveAside(path, failed, name); err != nil {
				return err
			}
		}
		if err := os.Rename(previous, path); err != nil {
			return err
		}
	}
	for index, entry := range j.Adapters {
		present, err := exists(entry.Path)
		if err != nil {
			return err
		}
		if present {
			quarantine, err := os.MkdirTemp(filepath.Dir(entry.Path), ".installer-recovery-")
			if err != nil {
				return err
			}
			if err := os.Rename(entry.Path, filepath.Join(quarantine, filepath.Base(entry.Path))); err != nil {
				return err
			}
		}
		if entry.Present {
			if err := copyPath(filepath.Join(transaction, "adapters", fmt.Sprint(index)), entry.Path); err != nil {
				return err
			}
		}
	}
	return os.Rename(filepath.Join(root, housekeeping, "pending.json"), filepath.Join(transaction, "rolled-back.json"))
}
func moveAside(path, dir, name string) error {
	unique, err := os.MkdirTemp(dir, name+"-")
	if err != nil {
		return err
	}
	return os.Rename(path, filepath.Join(unique, name))
}

// Install never renames the root. Its short activation window is intentionally non-atomic.
// The optional validation callback runs against both the staged and activated tree.
func (tree Tree) Install(root string, validate func(string) error, integrations ...Integration) (backup string, err error) {
	lock, err := lockRoot(root)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	if err := recoverLocked(root); err != nil {
		return "", err
	}
	same, err := tree.Matches(root)
	if err != nil || same && len(integrations) == 0 {
		return "", err
	}
	transaction, err := os.MkdirTemp(filepath.Join(root, housekeeping), "deployment-")
	if err != nil {
		return "", err
	}
	stage := filepath.Join(transaction, "stage")
	if err := os.Mkdir(stage, 0o755); err != nil {
		return "", err
	}
	if err := tree.Write(stage); err != nil {
		return "", err
	}
	if validate != nil {
		if err := validate(stage); err != nil {
			return "", fmt.Errorf("staged validation: %w", err)
		}
	}
	matched, err := tree.Matches(stage)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", fmt.Errorf("staged contents differ from deployment mapping")
	}
	// This is provenance, not an ownership inventory used to retire individual files.
	type mapping struct{ Source, Target string }
	var mappings []mapping
	for name, file := range tree {
		mappings = append(mappings, mapping{file.Source, filepath.Join(root, name)})
	}
	sort.Slice(mappings, func(i, j int) bool { return mappings[i].Target < mappings[j].Target })
	mappingData, err := json.MarshalIndent(mappings, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(transaction, "mapping.json"), mappingData, 0o600); err != nil {
		return "", err
	}
	old, err := liveNames(root)
	if err != nil {
		return "", err
	}
	next, err := liveNames(stage)
	if err != nil {
		return "", err
	}
	sort.Strings(old)
	sort.Strings(next)
	j := journal{Transaction: filepath.Base(transaction), Old: old, New: next}
	seen := map[string]bool{}
	for _, integration := range integrations {
		for _, path := range integration.Paths {
			if seen[path] {
				continue
			}
			seen[path] = true
			present, err := exists(path)
			if err != nil {
				return "", err
			}
			if present {
				if err := copyPath(path, filepath.Join(transaction, "adapters", fmt.Sprint(len(j.Adapters)))); err != nil {
					return "", err
				}
			}
			j.Adapters = append(j.Adapters, adapter{path, present})
		}
	}
	data, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	previous := filepath.Join(transaction, "previous")
	if err := os.Mkdir(previous, 0o700); err != nil {
		return "", err
	}
	journalStage := filepath.Join(transaction, "journal.json")
	if err := os.WriteFile(journalStage, data, 0o600); err != nil {
		return "", err
	}
	if err := os.Rename(journalStage, filepath.Join(root, housekeeping, "pending.json")); err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, recoverLocked(root))
		}
	}()
	for _, name := range old {
		if err = os.Rename(filepath.Join(root, name), filepath.Join(previous, name)); err != nil {
			return "", err
		}
	}
	for _, name := range next {
		if err = os.Rename(filepath.Join(stage, name), filepath.Join(root, name)); err != nil {
			return "", err
		}
	}
	if validate != nil {
		if err = validate(root); err != nil {
			return "", fmt.Errorf("activated validation: %w", err)
		}
	}
	matched, err = tree.Matches(root)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", fmt.Errorf("activated contents differ from deployment mapping")
	}
	for _, integration := range integrations {
		if integration.Apply != nil {
			if err = integration.Apply(); err != nil {
				return "", err
			}
		}
	}
	matched, err = tree.Matches(root)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", fmt.Errorf("adapter application changed canonical contents")
	}
	if err = os.Rename(filepath.Join(root, housekeeping, "pending.json"), filepath.Join(transaction, "committed.json")); err != nil {
		return "", err
	}
	return previous, nil
}

// copyPath preserves adapter symlinks without following their targets.
func copyPath(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(source)
		if err != nil {
			return err
		}
		return os.Symlink(target, destination)
	}
	if info.IsDir() {
		if err := os.Mkdir(destination, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyPath(filepath.Join(source, entry.Name()), filepath.Join(destination, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("unsupported adapter type at %q", source)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, data, info.Mode().Perm())
}
