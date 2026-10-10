// ABOUTME: Runs Hermes's native agent interface with private response activity callbacks.
// ABOUTME: Reuses the installed runtime without downloading or modifying its packages.
package harness

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed hermes_bridge.py
var hermesBridge []byte

func hermesInterpreter(command string) (string, error) {
	path, err := exec.LookPath(command)
	if err != nil {
		return "", err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	// Native console scripts and the fixed Homebrew libexec layout are supported.
	// Never execute shell wrapper contents to discover a runtime.
	for _, candidate := range []string{path, filepath.Join(filepath.Dir(filepath.Dir(path)), "libexec", "bin", "hermes")} {
		body, readErr := os.ReadFile(candidate)
		if readErr != nil {
			continue
		}
		first, _, _ := strings.Cut(string(body), "\n")
		interpreter := strings.TrimPrefix(first, "#!")
		if interpreter == first || !filepath.IsAbs(interpreter) || strings.ContainsAny(interpreter, " \t\r") || !strings.Contains(filepath.Base(interpreter), "python") {
			continue
		}
		if !strings.Contains(string(body), "hermes_cli.main") {
			continue
		}
		info, statErr := os.Stat(interpreter)
		if statErr != nil || !info.Mode().IsRegular() {
			continue
		}
		return interpreter, nil
	}
	return "", fmt.Errorf("Hermes native runtime unavailable: expected a hermes_cli console script or Homebrew libexec layout")
}

func executeHermesNative(ctx context.Context, command string, args []string, directory string, stdin io.Reader, stdout, stderr io.Writer) (runErr error) {
	interpreter, err := hermesInterpreter(command)
	if err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "sdlc-hermes-*")
	if err != nil {
		return err
	}
	defer func() {
		if cleanupErr := os.RemoveAll(scratch); cleanupErr != nil {
			runErr = errors.Join(runErr, fmt.Errorf("Hermes adapter temporary cleanup: %w", cleanupErr))
		}
	}()
	script := filepath.Join(scratch, "audit.py")
	if err := os.WriteFile(script, hermesBridge, 0600); err != nil {
		return err
	}
	nativeArgs := []string{"run", "--offline", "--no-project", "--no-config", "--no-env-file", "--no-cache", "--no-python-downloads", "--python", interpreter, script}
	// Retain the CLI's fixed arguments so synthetic boundary executors can use the
	// same model/provider/resume contract. The bridge accepts only this allowlist.
	nativeArgs = append(nativeArgs, args...)
	return executeCommand(ctx, "uv", nativeArgs, directory, stdin, stdout, stderr)
}
