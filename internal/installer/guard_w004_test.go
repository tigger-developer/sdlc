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
		{name: "unknown wrapper option", command: "sudo --mystery value rm obsolete.txt", blocked: true},
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

func TestW004CommandGuardEnvPathVectors(t *testing.T) {
	for _, path := range []string{".env", "config/.env", "./.env", `'config\.env'`} {
		blocked, _, _ := runGuard(t, map[string]any{
			"hook_event_name": "PreToolUse", "tool_name": "Bash",
			"tool_input": map[string]any{"command": "cat " + path},
		})
		if !blocked {
			t.Errorf("shell read was not blocked: %s", path)
		}
	}
	for _, command := range []string{
		"cat .env.example",
		"cat .env.local",
		`rg '\.env' README.md`,
	} {
		blocked, _, _ := runGuard(t, map[string]any{
			"hook_event_name": "PreToolUse", "tool_name": "Bash",
			"tool_input": map[string]any{"command": command},
		})
		if blocked {
			t.Errorf("permitted shell command was blocked: %s", command)
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

func TestW004CommandGuardAppliesEquivalentCommandPolicy(t *testing.T) {
	payloads := []struct {
		name    string
		payload func(string) map[string]any
	}{
		{name: "codex", payload: func(command string) map[string]any {
			return map[string]any{"hook_event_name": "PreToolUse", "tool_name": "Bash", "tool_input": map[string]any{"command": command}}
		}},
		{name: "claude", payload: func(command string) map[string]any {
			return map[string]any{"hook_event_name": "PreToolUse", "tool_name": "Bash", "tool_input": map[string]any{"command": command}}
		}},
		{name: "copilot", payload: func(command string) map[string]any {
			return map[string]any{"hookEventName": "preToolUse", "toolName": "shell", "toolArgs": map[string]any{"command": command}}
		}},
		{name: "hermes", payload: func(command string) map[string]any {
			return map[string]any{"hook_event_name": "pre_tool_call", "tool_name": "terminal", "tool_input": map[string]any{"command": command}}
		}},
	}
	blockedCommands := []string{
		"python -V", "python3 -V", "rm obsolete", "sed -n 1p README.md", "awk '{print $1}' README.md",
		"git commit --no-verify", "tool --no-hooks", "tool --no-pre-commit-hook", "chmod 777 target",
		"git remote add origin example", "git push --force origin master", "gh repo create example", "gh repo edit",
		"source config.sh", ". config.sh", "sudo -u builder rm obsolete",
	}
	allowedCommands := []string{
		"printf '%s\\n' python3", "echo source file", "echo git remote add example", "echo chmod 777 target",
		"echo --no-verify", "echo git push --force", "echo gh repo create", "rg python3 README.md",
		"command -v python3", "pythonista --version", "python3.14 --version",
	}
	for _, adapter := range payloads {
		t.Run(adapter.name, func(t *testing.T) {
			for _, command := range blockedCommands {
				blocked, _, _ := runGuard(t, adapter.payload(command))
				if !blocked {
					t.Errorf("command was not blocked: %s", command)
				}
			}
			for _, command := range allowedCommands {
				blocked, _, _ := runGuard(t, adapter.payload(command))
				if blocked {
					t.Errorf("permitted command was blocked: %s", command)
				}
			}
		})
	}
}

func TestW004CommandGuardAppliesEquivalentNativeReadPolicy(t *testing.T) {
	for _, test := range []struct {
		name    string
		payload func(tool string, input map[string]any) map[string]any
	}{
		{name: "codex", payload: func(tool string, input map[string]any) map[string]any {
			return map[string]any{"hook_event_name": "PreToolUse", "tool_name": tool, "tool_input": input}
		}},
		{name: "claude", payload: func(tool string, input map[string]any) map[string]any {
			return map[string]any{"hook_event_name": "PreToolUse", "tool_name": tool, "tool_input": input}
		}},
		{name: "copilot", payload: func(tool string, input map[string]any) map[string]any {
			return map[string]any{"hookEventName": "preToolUse", "toolName": tool, "toolArgs": input}
		}},
		{name: "hermes", payload: func(tool string, input map[string]any) map[string]any {
			return map[string]any{"hook_event_name": "pre_tool_call", "tool_name": tool, "tool_input": input}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, path := range []string{".env", "config/.env", "./.env", `config\.env`} {
				blocked, _, _ := runGuard(t, test.payload("read_file", map[string]any{"path": path}))
				if !blocked {
					t.Errorf("native read was not blocked: %s", path)
				}
			}
			for _, path := range []string{".env.example", ".env.local"} {
				blocked, _, _ := runGuard(t, test.payload("read_file", map[string]any{"path": path}))
				if blocked {
					t.Errorf("neighbouring native read was blocked: %s", path)
				}
			}
			blocked, _, _ := runGuard(t, test.payload("write_file", map[string]any{"path": ".env", "content": "example"}))
			if blocked {
				t.Error("native .env write was blocked")
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
		{payload: []byte(`{"unexpected":"shape"}`)},
	} {
		// #nosec G204 -- the command and repository fixture path are constants.
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
	// #nosec G204 -- the command and repository fixture path are constants.
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
