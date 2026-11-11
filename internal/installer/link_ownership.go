package installer

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
)

type linkOwnership struct {
	kind          string
	source        string
	userHome      string
	skill         string
	legacyTargets []string
}

func skillLinkOwnership(source, userHome, skill, currentDirectory string) linkOwnership {
	return linkOwnership{
		kind: "skill", source: source, userHome: userHome, skill: skill,
		legacyTargets: []string{currentDirectory, filepath.Join(userHome, ".agents", "sdlc", "skills", skill)},
	}
}

func verifyOwnedLinkReplacement(link managedLink, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		current, err := os.Readlink(link.destination)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(current) {
			current = filepath.Join(filepath.Dir(link.destination), current)
		}
		for _, expected := range link.ownership.legacyTargets {
			if sameLexicalPath(current, expected) {
				return nil
			}
		}
		return fmt.Errorf("refusing to replace link %q: target %q is not a known SDLC installation; move it aside manually", link.destination, current)
	}
	if info.IsDir() {
		var owned bool
		var err error
		switch link.ownership.kind {
		case "skill":
			owned, err = installedSkillMatchesRelease(link.ownership, link.destination)
		case "goose-plugin":
			owned, err = installedGoosePluginMatches(link.destination)
		}
		if err != nil {
			return err
		}
		if owned {
			return nil
		}
	}
	return fmt.Errorf("refusing to replace %q: its contents are not verified as SDLC-owned; move it aside manually", link.destination)
}

func installedSkillMatchesRelease(owner linkOwnership, destination string) (bool, error) {
	current, err := directoryContents(filepath.Join(owner.source, "skills", owner.skill))
	if err != nil {
		return false, err
	}
	installed, err := directoryContents(destination)
	if err != nil {
		return false, err
	}
	if equalDirectoryContents(current, installed) {
		return true, nil
	}
	release, err := installedLegacyRelease(owner.userHome)
	if err != nil || release == "" {
		return false, err
	}
	archived, err := taggedSkillContents(owner.source, release, owner.skill)
	if err != nil {
		return false, err
	}
	return equalDirectoryContents(archived, installed), nil
}

type directoryEntryContents struct {
	directories map[string]bool
	files       map[string][]byte
}

func directoryContents(root string) (directoryEntryContents, error) {
	result := directoryEntryContents{directories: map[string]bool{}, files: map[string][]byte{}}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			result.directories[relative] = true
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("cannot verify non-regular skill entry %q", path)
		}
		result.files[relative], err = os.ReadFile(path)
		return err
	})
	return result, err
}

func equalDirectoryContents(a, b directoryEntryContents) bool {
	if !reflect.DeepEqual(a.directories, b.directories) || len(a.files) != len(b.files) {
		return false
	}
	for path, content := range a.files {
		if !bytes.Equal(content, b.files[path]) {
			return false
		}
	}
	return true
}

func installedLegacyRelease(userHome string) (string, error) {
	path := filepath.Join(userHome, ".agents", "sdlc.yaml")
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading legacy SDLC release %q: %w", path, err)
	}
	_, root, err := globalConfigurationDocument(path, content)
	if err != nil {
		return "", err
	}
	value, _ := releaseMappingValue(root, "release")
	if value == nil || !semanticReleasePattern.MatchString(value.Value) {
		return "", nil
	}
	return value.Value, nil
}

func taggedSkillContents(source, release, skill string) (directoryEntryContents, error) {
	result := directoryEntryContents{directories: map[string]bool{".": true}, files: map[string][]byte{}}
	path := filepath.ToSlash(filepath.Join("skills", skill))
	archive, err := exec.Command("git", "-C", source, "archive", "--format=tar", release, path).Output()
	if err != nil {
		return result, fmt.Errorf("verifying old SDLC skill %q against release %s: %w", skill, release, err)
	}
	reader := tar.NewReader(bytes.NewReader(archive))
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return result, err
		}
		relative := strings.TrimPrefix(header.Name, path+"/")
		if relative == header.Name {
			continue
		}
		relative = filepath.Clean(relative)
		if relative == "." {
			continue
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
			return result, fmt.Errorf("invalid archived skill path %q", header.Name)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			result.directories[relative] = true
		case tar.TypeReg, tar.TypeRegA:
			result.files[relative], err = io.ReadAll(reader)
			if err != nil {
				return result, err
			}
		default:
			return result, fmt.Errorf("unsupported archived skill entry %q", header.Name)
		}
	}
	return result, nil
}

func installedGoosePluginMatches(root string) (bool, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false, err
	}
	if len(entries) != 2 || entries[0].Name() != "hooks" || !entries[0].IsDir() || entries[1].Name() != "plugin.json" || !entries[1].Type().IsRegular() {
		return false, nil
	}
	hookEntries, err := os.ReadDir(filepath.Join(root, "hooks"))
	if err != nil {
		return false, err
	}
	if len(hookEntries) != 1 || hookEntries[0].Name() != "hooks.json" || !hookEntries[0].Type().IsRegular() {
		return false, nil
	}
	manifest := map[string]any{}
	hooks := map[string]any{}
	manifestBytes, err := os.ReadFile(filepath.Join(root, "plugin.json"))
	if err != nil {
		return false, err
	}
	hookBytes, err := os.ReadFile(filepath.Join(root, "hooks", "hooks.json"))
	if err != nil {
		return false, err
	}
	if json.Unmarshal(manifestBytes, &manifest) != nil || json.Unmarshal(hookBytes, &hooks) != nil {
		return false, nil
	}
	expectedManifest := map[string]any{
		"name": "sdlc-command-guard-goose", "version": "1.0.0",
		"description": "Apply the SDLC shell and protected-file policy before Goose tools run.",
	}
	if !reflect.DeepEqual(manifest, expectedManifest) {
		return false, nil
	}
	for _, command := range []string{`"${PLUGIN_ROOT}/../../sdlc/hooks/agent-command-guard.sh"`, `"/opt/sdlc/hooks/agent-command-guard.sh"`} {
		expectedHooks := map[string]any{"hooks": map[string]any{"PreToolUse": []any{map[string]any{"hooks": []any{map[string]any{
			"type": "command", "command": command,
			"timeout": float64(10), "on_failure": "block",
		}}}}}}
		if reflect.DeepEqual(hooks, expectedHooks) {
			return true, nil
		}
	}
	return false, nil
}

func rejectUnverifiedRetiredSkills(root string, names []string) error {
	for _, name := range names {
		path := filepath.Join(root, name)
		if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}
		return fmt.Errorf("retired skill %q needs manual provenance review before it can be moved aside", path)
	}
	return nil
}
