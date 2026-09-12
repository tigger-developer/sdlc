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
