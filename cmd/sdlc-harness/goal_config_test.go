// ABOUTME: Exercises goal configuration at the command boundary without model calls.
// ABOUTME: Protects budget precedence and raw numeric spelling from unsafe coercion.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoalConfigDefaultsAndPrecedence(t *testing.T) {
	root := t.TempDir()
	global := filepath.Join(root, "global.yaml")
	project := filepath.Join(root, ".sdlc", "project.yaml")
	if err := os.Mkdir(filepath.Dir(project), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SDLC_GOAL_MAX_TURNS", "")
	t.Setenv("SDLC_GOAL_MAX_TOKEN_BUDGET", "")
	args := []string{"goal-config", "--sdlc-root", "../../src", "--project", root, "--global-config", global}
	check := func(extra []string, turns, tokens int64) {
		t.Helper()
		var output bytes.Buffer
		if err := run(append(append([]string{}, args...), extra...), nil, &output, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		var result struct {
			MaxTurns       int64 `json:"max_turns"`
			MaxTokenBudget int64 `json:"max_token_budget"`
		}
		if err := json.Unmarshal(output.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.MaxTurns != turns || result.MaxTokenBudget != tokens {
			t.Fatalf("limits = %s", output.Bytes())
		}
	}
	check(nil, 20, 100000)
	writeGoalFixture(t, global, "max_turns: 30\n    max_token_budget: 200,000")
	check(nil, 30, 200000)
	writeGoalFixture(t, project, "max_token_budget: \"150,000\"")
	check(nil, 30, 150000)
	t.Setenv("SDLC_GOAL_MAX_TOKEN_BUDGET", "120,000")
	check(nil, 30, 120000)
	check([]string{"--goal-max-token-budget", "110,000"}, 30, 110000)
	contents, err := os.ReadFile(project)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), `"150,000"`) {
		t.Fatal("project configuration was rewritten")
	}
}

func TestGoalConfigRejectsAmbiguousRawYAML(t *testing.T) {
	for _, scalar := range []string{"100.000", `"100.000"`, `"100,00"`, `"1,00,000"`, "1e5", "0x100", "0100", "+100", "-100", "0", "null", "true", "[100000]", "{value: 100000}", `"9,007,199,254,740,992"`} {
		t.Run(scalar, func(t *testing.T) {
			root := t.TempDir()
			global := filepath.Join(root, "global.yaml")
			writeGoalFixture(t, global, "max_token_budget: "+scalar)
			var output bytes.Buffer
			err := run([]string{"goal-config", "--sdlc-root", "../../src", "--project", root, "--global-config", global}, nil, &output, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), "max_token_budget") || output.Len() != 0 {
				t.Fatalf("error = %v, output = %s", err, output.Bytes())
			}
		})
	}
}

func TestGoalConfigRejectsBadShapeAndArguments(t *testing.T) {
	for _, body := range []string{"delivery:\n  goal: null\n", "delivery:\n  goal:\n    max_token_budget: 100\n    max_token_budget: 200\n", "delivery:\n  goal:\n    max_token_buget: 100\n", "delivery:\n  goal: {}\n---\ndelivery: {}\n"} {
		root := t.TempDir()
		global := filepath.Join(root, "global.yaml")
		if err := os.WriteFile(global, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		err := run([]string{"goal-config", "--sdlc-root", "../../src", "--project", root, "--global-config", global}, nil, &bytes.Buffer{}, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("accepted %q", body)
		}
	}
	if err := run([]string{"goal-config", "unexpected"}, nil, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("accepted extra argument")
	}
}

func TestGoalConfigRejectsInvalidOverridesWithoutFallingBack(t *testing.T) {
	root := t.TempDir()
	args := []string{"goal-config", "--sdlc-root", "../../src", "--project", root, "--global-config", filepath.Join(root, "absent.yaml")}
	for _, key := range []string{"SDLC_GOAL_MAX_TURNS", "SDLC_GOAL_MAX_TOKEN_BUDGET"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, "100.000")
			var output bytes.Buffer
			if err := run(args, nil, &output, &bytes.Buffer{}); err == nil || output.Len() != 0 {
				t.Fatalf("invalid environment budget produced %s: %v", output.Bytes(), err)
			}
		})
	}
	for _, flag := range []string{"--goal-max-turns", "--goal-max-token-budget"} {
		for _, value := range []string{"", "100.000", "100,00"} {
			var output bytes.Buffer
			if err := run(append(append([]string{}, args...), flag, value), nil, &output, &bytes.Buffer{}); err == nil || output.Len() != 0 {
				t.Fatalf("invalid flag budget produced %s: %v", output.Bytes(), err)
			}
		}
	}
}

func writeGoalFixture(t *testing.T, path, fields string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("version: 3\ndelivery:\n  goal:\n    "+fields+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}
