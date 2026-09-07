package installer

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestW004CommandGuardCommandPosition(t *testing.T) {
	tests := []struct {
		name    string
		command string
		blocked bool
	}{
		{name: "direct", command: "python3 -V", blocked: true},
		{name: "absolute", command: "/usr/bin/sed -n 1p README.md", blocked: true},
		{name: "assignment", command: "MODE=test rm obsolete.txt", blocked: true},
		{name: "env wrapper", command: "env MODE=test rm obsolete.txt", blocked: true},
		{name: "command wrapper", command: "command rm obsolete.txt", blocked: true},
		{name: "sudo option value", command: "sudo -u builder rm obsolete.txt", blocked: true},
		{name: "pipeline", command: "printf x | awk '{print $1}'", blocked: true},
		{name: "compound", command: "printf x && python -V", blocked: true},
		{name: "shell command string", command: "bash -lc 'printf x; python3 -V'", blocked: true},
		{name: "data executable", command: "printf '%s\\n' python3", blocked: false},
		{name: "data source", command: "echo source file", blocked: false},
		{name: "data remote", command: "echo git remote add example", blocked: false},
		{name: "data chmod", command: "echo chmod 777 target", blocked: false},
		{name: "data bypass", command: "echo --no-verify", blocked: false},
		{name: "data force", command: "echo git push --force", blocked: false},
		{name: "data github", command: "echo gh repo create", blocked: false},
		{name: "search", command: "rg python3 README.md", blocked: false},
		{name: "command inspection", command: "command -v python3", blocked: false},
		{name: "substring", command: "pythonista --version", blocked: false},
		{name: "versioned", command: "python3.14 --version", blocked: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			blocked, _, _ := runGuard(t, map[string]any{
				"hook_event_name": "PreToolUse",
				"tool_name":       "Bash",
				"tool_input":      map[string]any{"command": test.command},
			})
			if blocked != test.blocked {
				t.Fatalf("blocked = %t, want %t", blocked, test.blocked)
			}
		})
	}
}

func TestW004CommandGuardExactRestrictions(t *testing.T) {
	blocked := []string{
		"git commit --no-verify", "tool --no-hooks", "tool --no-pre-commit-hook",
		"chmod 777 target", "git remote add origin example", "git push --force origin main",
		"git push --force-with-lease origin main", "gh repo create", "gh repo edit",
		"source config.sh", ". config.sh",
	}
	for _, command := range blocked {
		got, _, _ := runGuard(t, map[string]any{
			"hook_event_name": "PreToolUse", "tool_name": "Bash",
			"tool_input": map[string]any{"command": command},
		})
		if !got {
			t.Errorf("command was not blocked: %s", command)
		}
	}
}

func TestW004CommandGuardNativePayloads(t *testing.T) {
	tests := []struct {
		name       string
		payload    map[string]any
		hermesJSON bool
	}{
		{name: "codex", payload: map[string]any{"hook_event_name": "PreToolUse", "tool_name": "Bash", "tool_input": map[string]any{"command": "rm x"}}},
		{name: "claude", payload: map[string]any{"hook_event_name": "PreToolUse", "tool_name": "Bash", "tool_input": map[string]any{"command": "rm x"}}},
		{name: "copilot", payload: map[string]any{"hookEventName": "preToolUse", "toolName": "shell", "toolArgs": map[string]any{"command": "rm x"}}},
		{name: "hermes", payload: map[string]any{"hook_event_name": "pre_tool_call", "tool_name": "terminal", "tool_input": map[string]any{"command": "rm x"}}, hermesJSON: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			blocked, stdout, stderr := runGuard(t, test.payload)
			if !blocked {
				t.Fatal("expected block")
			}
			if test.hermesJSON {
				var response map[string]string
				if err := json.Unmarshal(stdout, &response); err != nil || response["decision"] != "block" {
					t.Fatalf("Hermes response = %q, error = %v", stdout, err)
				}
			} else if !strings.Contains(string(stderr), "Blocked by agent-command-guard") {
				t.Fatalf("stderr = %q", stderr)
			}
		})
	}
}

func TestW004CommandGuardRejectsMalformedRecognizedPayload(t *testing.T) {
	for _, test := range []struct {
		payload    []byte
		hermesJSON bool
	}{
		{payload: []byte("not-json")},
		{payload: []byte(`{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{}}`)},
		{payload: []byte(`{"hook_event_name":"pre_tool_call","tool_name":"terminal","tool_input":{}}`), hermesJSON: true},
	} {
		command := exec.Command("bash", filepath.Join("..", "..", "hooks", "agent-command-guard.sh"))
		command.Stdin = bytes.NewReader(test.payload)
		output, err := command.CombinedOutput()
		if test.hermesJSON {
			if err != nil || !strings.Contains(string(output), `"decision":"block"`) {
				t.Fatalf("Hermes malformed payload result = %q, %v", output, err)
			}
		} else if err == nil {
			t.Fatalf("payload unexpectedly succeeded: %s", test.payload)
		}
	}
}

func runGuard(t *testing.T, payload map[string]any) (bool, []byte, []byte) {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("bash", filepath.Join("..", "..", "hooks", "agent-command-guard.sh"))
	command.Stdin = bytes.NewReader(encoded)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	var exitError *exec.ExitError
	blocked := errors.As(err, &exitError) && exitError.ExitCode() == 2
	if err == nil && strings.Contains(stdout.String(), `"decision":"block"`) {
		blocked = true
	}
	return blocked, stdout.Bytes(), stderr.Bytes()
}
