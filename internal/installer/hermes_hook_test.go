package installer_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHermesCommandPolicyHook(t *testing.T) {
	hookPath := filepath.Join("..", "..", "hooks", "agent-command-guard.sh")
	tests := []struct {
		name    string
		command string
		blocked bool
	}{
		{name: "python", command: "python -V", blocked: true},
		{name: "python3", command: "python3 -c 'print(1)'", blocked: true},
		{name: "rm", command: "rm obsolete.txt", blocked: true},
		{name: "sed", command: "sed -n '1p' README.md", blocked: true},
		{name: "awk", command: "awk '{print $1}' README.md", blocked: true},
		{name: "absolute path", command: "/opt/homebrew/bin/python3 -V", blocked: true},
		{name: "pipeline", command: "printf x | python3 -c 'import sys'", blocked: true},
		{name: "compound", command: "printf x && python -V", blocked: true},
		{name: "shell wrapper", command: "/opt/homebrew/bin/bash -lc 'python3 -V'", blocked: true},
		{name: "nested shell wrapper compound", command: "bash -c 'printf x; python -V'", blocked: true},
		{name: "environment wrapper", command: "env MODE=test python3 -V", blocked: true},
		{name: "control flow", command: "if true; then python3 -V; fi", blocked: true},
		{name: "command substitution", command: "printf '%s' $(python -V)", blocked: true},
		{name: "make entry point", command: "make test", blocked: false},
		{name: "argument", command: "printf '%s\\n' python3", blocked: false},
		{name: "substring executable", command: "pythonista --version", blocked: false},
		{name: "versioned executable", command: "python3.14 --version", blocked: false},
		{name: "filename", command: "rg python3 README.md", blocked: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, err := json.Marshal(map[string]any{
				"hook_event_name": "pre_tool_call",
				"tool_name":       "terminal",
				"tool_input":      map[string]string{"command": test.command},
			})
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			command := exec.Command("bash", hookPath)
			command.Stdin = bytes.NewReader(payload)
			output, err := command.Output()
			if err != nil {
				t.Fatalf("hook error = %v", err)
			}
			if !test.blocked {
				if len(bytes.TrimSpace(output)) != 0 {
					t.Fatalf("allowed command produced output: %s", output)
				}
				return
			}
			var response map[string]string
			if err := json.Unmarshal(output, &response); err != nil {
				t.Fatalf("hook returned invalid JSON %q: %v", output, err)
			}
			if response["decision"] != "block" {
				t.Fatalf("hook decision = %q, want block", response["decision"])
			}
		})
	}
}

func TestCommandPolicyHookRetainsClaudeExitContract(t *testing.T) {
	hookPath := filepath.Join("..", "..", "hooks", "agent-command-guard.sh")
	payload, err := json.Marshal(map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_input":      map[string]string{"command": "python3 -V"},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	command := exec.Command("bash", hookPath)
	command.Stdin = bytes.NewReader(payload)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	err = command.Run()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 2 {
		t.Fatalf("hook error = %v, want exit code 2", err)
	}
	if !strings.Contains(stderr.String(), "Blocked by agent-command-guard") {
		t.Fatalf("Claude block reason missing: %s", stderr.String())
	}
}
