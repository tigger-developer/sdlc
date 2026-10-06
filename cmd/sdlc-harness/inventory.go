// ABOUTME: Lists the invocation project's visible files for audit scope review.
// ABOUTME: Excludes provider runtime and scratch directories without copying contents.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const maximumInventoryFiles = 10000
const maximumInventoryBytes = 1024 * 1024

var inventoryProject = auditProjectInventory

func auditProjectInventory(root string) (string, error) {
	paths, err := gitProjectFiles(root)
	if err != nil {
		return "", err
	}
	if paths == nil {
		paths, err = filesystemProjectFiles(root)
		if err != nil {
			return "", err
		}
	}
	sort.Strings(paths)
	var listing strings.Builder
	listing.WriteString("\n\nInvocation project file inventory (paths only; omitted content must not be assumed reviewed):\n")
	for _, path := range paths {
		if strings.ContainsAny(path, "\r\n") {
			return "", fmt.Errorf("project inventory path %q contains a line break", path)
		}
		if listing.Len()+len(path)+2 > maximumInventoryBytes {
			return "", errors.New("project inventory exceeds 1 MiB; narrow the project before auditing")
		}
		listing.WriteString("- ")
		listing.WriteString(path)
		listing.WriteByte('\n')
	}
	return listing.String(), nil
}

func gitProjectFiles(root string) ([]string, error) {
	command := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel")
	output, err := command.CombinedOutput()
	if err != nil && strings.Contains(string(output), "not a git repository") {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("checking project Git root: %w: %s", err, strings.TrimSpace(string(output)))
	}
	canonical, err := filepath.EvalSymlinks(strings.TrimSpace(string(output)))
	if err != nil {
		return nil, err
	}
	if canonical != root {
		return nil, fmt.Errorf("project inventory root %q differs from invocation project %q", canonical, root)
	}
	output, err = exec.Command("git", "-C", root, "ls-files", "--cached", "--others", "--exclude-standard", "-z").Output()
	if err != nil {
		return nil, fmt.Errorf("listing project files: %w", err)
	}
	paths := make([]string, 0)
	for _, path := range strings.Split(string(output), "\x00") {
		if path == "" || excludedAuditPath(path) {
			continue
		}
		paths = append(paths, path)
		if len(paths) > maximumInventoryFiles {
			return nil, errors.New("project inventory exceeds 10000 files")
		}
	}
	return paths, nil
}

func filesystemProjectFiles(root string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() && excludedAuditDirectory(entry.Name()) {
			return filepath.SkipDir
		}
		if entry.IsDir() || excludedAuditPath(relative) {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		paths = append(paths, relative)
		if len(paths) > maximumInventoryFiles {
			return errors.New("project inventory exceeds 10000 files")
		}
		return nil
	})
	return paths, err
}

func excludedAuditPath(path string) bool {
	parts := strings.Split(filepath.Clean(path), string(os.PathSeparator))
	for _, part := range parts[:len(parts)-1] {
		if excludedAuditDirectory(part) {
			return true
		}
	}
	return false
}
