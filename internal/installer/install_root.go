package installer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
)

const installExplanationURL = "https://github.com/tigger-developer/sdlc#installing-under-optsdlc"

func ensureInstallRoot(root string, input io.Reader, output io.Writer) error {
	if input == nil {
		input = os.Stdin
	}
	if output == nil {
		output = os.Stdout
	}
	info, err := os.Lstat(root)
	if err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("SDLC installation root %q is not a regular directory", root)
		}
		return checkInstallRootWritable(root)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspecting SDLC installation root %q: %w", root, err)
	}
	if err = os.MkdirAll(root, 0o755); err == nil {
		if err := os.Chmod(root, 0o755); err != nil {
			return err
		}
		return checkInstallRootWritable(root)
	}
	if !errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("creating SDLC installation root %q: %w", root, err)
	}
	fmt.Fprintf(output, "%s prevented the current user from creating %s; sudo is required for directory setup.\nWhy limited sudo is requested: %s\n", installOSName(), root, installExplanationURL)
	reader, ok := input.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(input)
	}
	approved, confirmErr := confirm(reader, output, "Permit sudo mkdir and sudo chown for this directory only? [yes/no]: ")
	if confirmErr != nil {
		return confirmErr
	}
	if !approved {
		return fmt.Errorf("%s was not created; create and own it manually, then rerun the installer", root)
	}
	fmt.Fprintln(output, "Permission granted; setting up the directory with sudo.")
	current, err := user.Current()
	if err != nil {
		return fmt.Errorf("identifying current user for installation root ownership: %w", err)
	}
	for _, args := range [][]string{{"mkdir", root}, {"chown", current.Username, root}} {
		command := exec.Command("sudo", args...)
		command.Stdin = input
		command.Stdout = output
		command.Stderr = output
		if err := command.Run(); err != nil {
			return fmt.Errorf("sudo %s failed: %w", args[0], err)
		}
	}
	if err := os.Chmod(root, 0o755); err != nil {
		return fmt.Errorf("setting shared directory permissions on %q: %w", root, err)
	}
	return checkInstallRootWritable(root)
}

func installOSName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return runtime.GOOS
	}
}

func checkInstallRootWritable(root string) error {
	probe, err := os.CreateTemp(root, ".sdlc-write-check-")
	if err != nil {
		return fmt.Errorf("SDLC installation root %q exists but is not writable by the current user: %w", root, err)
	}
	name := probe.Name()
	if err := probe.Close(); err != nil {
		return err
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("removing installation write check %q: %w", name, err)
	}
	return nil
}

func ensureSharedDirectoryModes(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		mode := info.Mode().Perm()
		if mode&0o055 == 0o055 {
			return nil
		}
		if err := os.Chmod(path, mode|0o055); err != nil {
			return fmt.Errorf("making shared SDLC directory %q readable: %w", path, err)
		}
		return nil
	})
}
