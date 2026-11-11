// ABOUTME: Verifies bare CLI deployment is isolated from user and provider state.
// ABOUTME: Covers preview, application, repeatability, and conflicting options.
package main

import (
	"bytes"
	"encoding/json"
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
	plugin := filepath.Join(destination, "plugins", "sdlc-command-guard-goose")
	manifest, err := os.ReadFile(filepath.Join(plugin, "plugin.json"))
	if err != nil || !json.Valid(manifest) {
		t.Fatalf("invalid plugin manifest: %q %v", manifest, err)
	}
	hooksPath := filepath.Join(plugin, "hooks", "hooks.json")
	hooks, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command   string `json:"command"`
				OnFailure string `json:"on_failure"`
			}
		}
	}
	if err := json.Unmarshal(hooks, &config); err != nil {
		t.Fatal(err)
	}
	events := config.Hooks["PreToolUse"]
	if len(events) != 1 || len(events[0].Hooks) != 1 {
		t.Fatalf("unexpected hook configuration: %s", hooks)
	}
	hook := events[0].Hooks[0]
	if hook.Command != `"/opt/sdlc/hooks/agent-command-guard.sh"` || hook.OnFailure != "block" {
		t.Fatalf("unexpected guard policy: %+v", hook)
	}
	for _, path := range []string{".local", ".codex", ".hermes", ".agents/skills", ".agents/plugins", ".config/goose"} {
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
	if err := os.WriteFile(hooksPath, []byte("operator-owned\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, mode := range [][]string{{"--bare"}, {"--bare", "--dry-run"}} {
		args := append(mode, "--source", source)
		if err := runAtRoot(args, strings.NewReader(""), &output, destination); err == nil || !strings.Contains(err.Error(), "conflicting Goose plugin") {
			t.Fatalf("expected conflict: %v", err)
		}
	}
	if data, err := os.ReadFile(hooksPath); err != nil || string(data) != "operator-owned\n" {
		t.Fatalf("conflicting plugin changed: %q %v", data, err)
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

func TestBareRetiresInventoriedSkills(t *testing.T) {
	root, source := newCLIFixture(t)
	bin := filepath.Join(root, "fake-bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	installCLITrashDouble(t, bin, root)
	t.Setenv("PATH", bin)
	put := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put(filepath.Join(source, "skills", "retire-me", "SKILL.md"), "owned")
	destination := filepath.Join(root, "opt", "sdlc")
	args := []string{"--bare", "--source", source}
	if err := runAtRoot(args, strings.NewReader(""), &bytes.Buffer{}, destination); err != nil {
		t.Fatal(err)
	}
	put(filepath.Join(destination, "skills", "unrelated", "SKILL.md"), "private")
	if err := os.Rename(filepath.Join(source, "skills", "retire-me"), filepath.Join(root, "saved-source")); err != nil {
		t.Fatal(err)
	}
	if err := runAtRoot(append(args, "--dry-run"), strings.NewReader(""), &bytes.Buffer{}, destination); err != nil {
		t.Fatal(err)
	}
	retired := filepath.Join(destination, "skills", "retire-me")
	if _, err := os.Stat(retired); err != nil {
		t.Fatal("preview retired skill", err)
	}
	for i := 0; i < 2; i++ {
		if err := runAtRoot(args, strings.NewReader(""), &bytes.Buffer{}, destination); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Lstat(retired); !os.IsNotExist(err) {
		t.Fatalf("retired skill remains: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(destination, "skills", "unrelated", "SKILL.md")); err != nil || string(data) != "private" {
		t.Fatalf("unrelated changed %q %v", data, err)
	}
}
