// ABOUTME: Verifies bare CLI deployment is isolated from user and provider state.
// ABOUTME: Covers preview, application, repeatability, and conflicting options.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBareDeploymentDoesNotTouchHomes(t *testing.T) {
	root, source := newCLIFixture(t)
	bin := filepath.Join(root, "fake-bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	installCLITrashDouble(t, bin, root)
	t.Setenv("PATH", bin)
	t.Setenv("HOME", filepath.Join(root, "home"))
	legacy := filepath.Join(root, "home", ".agents", "sdlc.yaml")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte("invalid: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(root, "opt", "sdlc")
	var output bytes.Buffer
	previous := buildRelease
	buildRelease = "v4.0.16"
	t.Cleanup(func() { buildRelease = previous })
	if err := runAtRoot([]string{"--bare", "--dry-run", "--source", source}, strings.NewReader(""), &output, destination); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		t.Fatalf("preview created root: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := runAtRoot([]string{"--bare", "--source", source}, strings.NewReader(""), &output, destination); err != nil {
			t.Fatalf("%v\n%s", err, output.String())
		}
	}
	for _, path := range []string{"docs/MAIN.md", "hooks/agent-command-guard.sh", "bin/sdlc-guard-shell", "sdlc.yaml"} {
		if _, err := os.Stat(filepath.Join(destination, path)); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{".local", ".codex", ".hermes", ".agents/skills"} {
		if _, err := os.Lstat(filepath.Join(root, "home", path)); !os.IsNotExist(err) {
			t.Fatalf("home changed: %s %v", path, err)
		}
	}
	data, err := os.ReadFile(legacy)
	if err != nil || string(data) != "invalid: [" {
		t.Fatalf("legacy modified: %q %v", data, err)
	}
	if strings.Contains(output.String(), "[yes/no]") || strings.Contains(output.String(), "Detected agent") {
		t.Fatal(output.String())
	}
}

func TestBareRejectsProviderFlags(t *testing.T) {
	_, source := newCLIFixture(t)
	for _, args := range [][]string{{"--agent", "hermes"}, {"--configure"}} {
		flags := append([]string{"--bare", "--source", source}, args...)
		if err := runAtRoot(flags, strings.NewReader(""), &bytes.Buffer{}, filepath.Join(t.TempDir(), "sdlc")); err == nil {
			t.Fatalf("accepted %v", flags)
		}
	}
}
