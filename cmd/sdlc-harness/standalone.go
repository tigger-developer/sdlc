// ABOUTME: Selects standards-only audits and rejects temporary project evidence.
// ABOUTME: Keeps standalone verdicts separate from SDLC gate records.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var profilePresent = projectHasSDLCProfile

func projectHasSDLCProfile(root string) (bool, error) {
	path := filepath.Join(root, ".sdlc", "project.yaml")
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspecting SDLC project profile %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("SDLC project profile %q must be a regular file", path)
	}
	return true, nil
}

// This function is replaced only by local tests to provide a separate protected
// fixture root while their project fixtures occupy the operating system temp dir.
var temporaryAuditRoots = func() ([]string, error) {
	roots := []string{os.TempDir()}
	if runtime.GOOS != "windows" {
		roots = append(roots, "/tmp", "/var/tmp")
	}
	if runtime.GOOS == "darwin" {
		output, err := exec.Command("/usr/bin/getconf", "DARWIN_USER_TEMP_DIR").Output()
		if err != nil {
			return nil, fmt.Errorf("resolving macOS system temporary directory: %w", err)
		}
		roots = append(roots, strings.TrimSpace(string(output)))
	}
	for _, key := range []string{"TMPDIR", "TMP", "TEMP"} {
		if value := os.Getenv(key); value != "" {
			roots = append(roots, value)
		}
	}
	return roots, nil
}

func requireOriginalAuditPaths(project string, inputs []string, promptRegistry string) error {
	roots, err := temporaryAuditRoots()
	if err != nil {
		return err
	}
	paths := append([]string{project}, inputs...)
	if promptRegistry != "" {
		paths = append(paths, promptRegistry)
	}
	for _, raw := range paths {
		path := raw
		if !filepath.IsAbs(path) {
			path = filepath.Join(project, path)
		}
		if err := rejectTemporaryAuditPath(path, roots); err != nil {
			return err
		}
	}
	return nil
}

func rejectTemporaryAuditPath(path string, roots []string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("resolving audit path %q: %w", path, err)
		}
		resolved = absolute
	}
	for _, root := range roots {
		if root == "" {
			continue
		}
		canonicalRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			return fmt.Errorf("resolving system temporary directory %q: %w", root, err)
		}
		if withinPath(root, absolute) || withinPath(canonicalRoot, resolved) {
			return fmt.Errorf("audit refused temporary project or evidence path %q; review original project files", path)
		}
	}
	return nil
}

func withinPath(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func validateStandaloneVerdict(response string) error {
	fields := map[string][]string{}
	for _, line := range strings.Split(response, "\n") {
		for _, name := range []string{"AUDIT", "REVISION", "VERDICT"} {
			if value, found := strings.CutPrefix(line, name+":"); found {
				fields[name] = append(fields[name], strings.TrimSpace(value))
			}
		}
	}
	for _, name := range []string{"AUDIT", "REVISION", "VERDICT"} {
		if len(fields[name]) != 1 || fields[name][0] == "" {
			return fmt.Errorf("standalone audit response requires exactly one non-empty %s field", name)
		}
	}
	if fields["AUDIT"][0] != "standalone" {
		return fmt.Errorf("standalone audit response has unexpected AUDIT value %q", fields["AUDIT"][0])
	}
	switch fields["VERDICT"][0] {
	case "PASS", "PROVISIONAL PASS", "FAIL":
		return nil
	default:
		return fmt.Errorf("standalone audit response has unsupported VERDICT %q", fields["VERDICT"][0])
	}
}
