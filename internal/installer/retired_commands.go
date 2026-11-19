// ABOUTME: Identifies historical SDLC command retirements left in executable directories.
// ABOUTME: Plans exact recoverable removal without following symlinks or touching backups.
package installer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// These are names and markers emitted by earlier installers, not a general
// sweep of user scripts, .bak files, directories or unrelated SDLC-like names.
var retiredCommandName = regexp.MustCompile(`^sdlc-(audit|install|preview|project-init|project-update|init|merge-legacy-acs|harness)\.sdlc(-v[23])?-retired(-[0-9]+)?$`)

func planRetiredCommandBackups(root string) ([]managedRetirement, error) {
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading command retirement directory %q: %w", root, err)
	}
	var result []managedRetirement
	for _, entry := range entries {
		if !entry.IsDir() && retiredCommandName.MatchString(entry.Name()) && (entry.Type().IsRegular() || entry.Type()&os.ModeSymlink != 0) {
			result = append(result, managedRetirement{path: filepath.Join(root, entry.Name())})
		}
	}
	return result, nil
}

func planRetiredAuditLink(localBin, installRoot, userHome string) (*managedRetirement, error) {
	path := filepath.Join(localBin, "sdlc-audit")
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspecting retired audit link %q: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return nil, nil
	}
	target, err := os.Readlink(path)
	if err != nil {
		return nil, fmt.Errorf("reading retired audit link %q: %w", path, err)
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(localBin, target)
	}
	for _, owned := range []string{
		filepath.Join(installRoot, "bin", "sdlc-harness"),
		filepath.Join(userHome, ".agents", "sdlc", "bin", "sdlc-harness"),
	} {
		if sameLexicalPath(target, owned) {
			retirement := managedRetirement{path: path}
			return &retirement, nil
		}
	}
	return nil, nil
}
