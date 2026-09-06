package configenv

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Load evaluates one shell environment file through the project-owned wrapper
// and returns only the explicitly allowed keys.
func Load(wrapper, path string, keys []string) (map[string]string, error) {
	values := map[string]string{}
	if path == "" {
		return values, nil
	}
	arguments := append([]string{wrapper, path}, keys...)
	command := exec.Command("bash", arguments...)
	var output bytes.Buffer
	var diagnostics bytes.Buffer
	command.Stdout = &output
	command.Stderr = &diagnostics
	if err := command.Run(); err != nil {
		message := bytes.TrimSpace(diagnostics.Bytes())
		if len(message) != 0 {
			return nil, fmt.Errorf("loading shell environment %q: %s", path, message)
		}
		return nil, fmt.Errorf("loading shell environment %q: %w", path, err)
	}

	parts := bytes.Split(output.Bytes(), []byte{0})
	if len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}
	if len(parts)%2 != 0 {
		return nil, errors.New("shell environment loader returned a malformed record")
	}
	allowed := make(map[string]bool, len(keys))
	for _, key := range keys {
		allowed[key] = true
	}
	for index := 0; index < len(parts); index += 2 {
		key := string(parts[index])
		if !allowed[key] {
			return nil, fmt.Errorf("shell environment loader returned unexpected key %q", key)
		}
		values[key] = string(parts[index+1])
	}
	return values, nil
}

// RemoveKeys rewrites one legacy environment file through the shell wrapper,
// removing only exact allowlisted assignments while preserving every other
// byte stream line. The Go process never receives unrelated values.
func RemoveKeys(wrapper, path string, keys []string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspecting shell environment %q: %w", path, err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".sdlc-env-migration-*")
	if err != nil {
		return fmt.Errorf("creating migrated environment beside %q: %w", path, err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		_ = temporary.Close()
		return err
	}
	arguments := append([]string{wrapper, "--filter", path}, keys...)
	command := exec.Command("bash", arguments...)
	command.Stdout = temporary
	var diagnostics bytes.Buffer
	command.Stderr = &diagnostics
	if err := command.Run(); err != nil {
		_ = temporary.Close()
		if message := bytes.TrimSpace(diagnostics.Bytes()); len(message) != 0 {
			return fmt.Errorf("filtering shell environment %q: %s", path, message)
		}
		return fmt.Errorf("filtering shell environment %q: %w", path, err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replacing migrated shell environment %q: %w", path, err)
	}
	return nil
}
