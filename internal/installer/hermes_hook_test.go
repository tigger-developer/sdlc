package installer_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestHermesCommandPolicyHook(t *testing.T) {
	hookPath := filepath.Join("..", "..", "hooks", "hermes-command-policy")
	tests := []struct {
		name    string
		command string
		blocked bool
	}{
		{name: "python", command: "python -V", blocked: true},
		{name: "python3", command: "python3 -c 'print(1)'", blocked: true},
		{name: "absolute path", command: "/opt/homebrew/bin/python3 -V", blocked: true},
		{name: "pipeline", command: "printf x | python3 -c 'import sys'", blocked: true},
		{name: "compound", command: "printf x && python -V", blocked: true},
		{name: "shell wrapper", command: "/opt/homebrew/bin/bash -lc 'python3 -V'", blocked: true},
		{name: "nested shell wrapper compound", command: "bash -c 'printf x; python -V'", blocked: true},
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
			command := exec.Command(hookPath)
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
