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
var enforceAuditAnchor = verifyAuditAnchor

func verifyAuditAnchor(project string) (string, error) {
	invocation, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolving audit invocation directory: %w", err)
	}
	root := invocation
	command := exec.Command("git", "-C", invocation, "rev-parse", "--show-toplevel")
	if output, gitErr := command.CombinedOutput(); gitErr == nil {
		root = strings.TrimSpace(string(output))
	} else if !strings.Contains(string(output), "not a git repository") {
		return "", fmt.Errorf("resolving invocation Git root: %w: %s", gitErr, strings.TrimSpace(string(output)))
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	if project != "." {
		canonicalProject, resolveErr := filepath.EvalSymlinks(project)
		if resolveErr != nil {
			return "", resolveErr
		}
		if canonicalProject != canonicalRoot {
			return "", fmt.Errorf("audit project %q differs from invocation project %q; invoke the auditor from the original project", project, root)
		}
	}
	return canonicalRoot, nil
}

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
	for _, raw := range paths {
		path := raw
		if !filepath.IsAbs(path) {
			path = filepath.Join(project, path)
		}
		if err := rejectTemporaryAuditPath(path, roots); err != nil {
			return err
		}
		if err := requireProjectFile(project, path); err != nil {
			return err
		}
	}
	if promptRegistry != "" {
		if err := rejectTemporaryAuditPath(promptRegistry, roots); err != nil {
			return err
		}
	}
	return nil
}

func requireProjectFile(project, path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	canonicalProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		return err
	}
	canonicalPath, err := filepath.EvalSymlinks(path)
	if errors.Is(err, os.ErrNotExist) {
		parent, parentErr := filepath.EvalSymlinks(filepath.Dir(path))
		if parentErr != nil {
			return fmt.Errorf("resolving project path %q: %w", path, parentErr)
		}
		canonicalPath = filepath.Join(parent, filepath.Base(path))
	} else if err != nil {
		return fmt.Errorf("resolving project path %q: %w", path, err)
	}
	if !withinPath(project, absolute) || !withinPath(canonicalProject, canonicalPath) {
		return fmt.Errorf("audit path %q is outside the invocation project", path)
	}
	relative, err := filepath.Rel(project, absolute)
	if err != nil {
		return err
	}
	canonicalRelative, err := filepath.Rel(canonicalProject, canonicalPath)
	if err != nil {
		return err
	}
	if excludedAuditPath(relative) || excludedAuditPath(canonicalRelative) {
		return fmt.Errorf("audit path %q uses an excluded runtime or temporary directory", path)
	}
	return nil
}

func excludedAuditDirectory(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case ".git", ".agent", ".agents", ".claude", ".codex", ".copilot", ".hermes":
		return true
	}
	stem := strings.TrimPrefix(lower, ".")
	for _, word := range []string{"tmp", "temp", "temporary"} {
		if stem == word || strings.HasPrefix(stem, word+"-") || strings.HasPrefix(stem, word+"_") || strings.HasPrefix(stem, word+".") {
			return true
		}
	}
	return false
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
