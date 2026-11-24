package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoalConfigUsesInstalledInitializerWithoutMutatingProject(t *testing.T) {
	root := t.TempDir()
	profile := filepath.Join(root, ".sdlc", "project.yaml")
	if err := os.MkdirAll(filepath.Dir(profile), 0o700); err != nil {
		t.Fatal(err)
	}
	profileBody := []byte("version: 3\ndelivery:\n  goal:\n    max_turns: 11\n    max_token_budget: 120,000\n")
	if err := os.WriteFile(profile, profileBody, 0o600); err != nil {
		t.Fatal(err)
	}
	global := filepath.Join(t.TempDir(), "sdlc.yaml")
	if err := os.WriteFile(global, []byte("version: 3\ndelivery:\n  goal:\n    max_turns: 13\n    max_token_budget: 200,000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SDLC_GOAL_MAX_TURNS", "9")
	t.Setenv("SDLC_GOAL_MAX_TOKEN_BUDGET", "")
	args := []string{"goal-config", "--sdlc-root", "../../src", "--project", root, "--global-config", global, "--goal-max-turns", "7"}
	var output, diagnostics bytes.Buffer
	if err := runCommand(args, strings.NewReader(""), &output, &diagnostics); err != nil {
		t.Fatal(err)
	}
	var limits map[string]int64
	if err := json.Unmarshal(output.Bytes(), &limits); err != nil {
		t.Fatal(err)
	}
	if limits["max_turns"] != 7 || limits["max_token_budget"] != 120000 {
		t.Fatalf("goal limits = %+v", limits)
	}
	if current, err := os.ReadFile(profile); err != nil || !bytes.Equal(current, profileBody) {
		t.Fatalf("project profile changed: %q (%v)", current, err)
	}
	output.Reset()
	args = append(args, "--goal-max-token-budget", "100.000")
	if err := runCommand(args, strings.NewReader(""), &output, &diagnostics); err == nil || output.Len() != 0 {
		t.Fatalf("invalid budget accepted or output emitted: %q (%v)", output.String(), err)
	}
}

func TestGoalConfigHelpDoesNotRequireDeployedSchema(t *testing.T) {
	var output bytes.Buffer
	if err := runCommand([]string{"goal-config", "--help"}, strings.NewReader(""), &output, &output); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"usage: sdlc-init goal-config", "--goal-max-token-budget", "without initializing"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("goal-config help missing %q: %q", expected, output.String())
		}
	}
}
