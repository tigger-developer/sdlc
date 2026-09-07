// Package harness defines fixed, capability-aware coding-agent invocations.
package harness

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Request is the bounded input to one harness start or resume operation.
type Request struct {
	Harness    string
	Model      string
	Provider   string
	Prompt     string
	Bundle     string
	ResultFile string
	SessionID  string
}

// Invocation is a fixed executable invocation. Args are never shell-evaluated.
type Invocation struct {
	Command string
	Args    []string
	Dir     string
	Stdin   string
}

// Result contains only the stable identity and final response.
type Result struct {
	SessionID string
	Response  string
}

// Input identifies one regular file copied into an isolated audit bundle.
type Input struct {
	Name   string
	Source string
}

// Bundle records the isolated path and its original content digests.
type Bundle struct {
	Path    string
	digests map[string]string
}

// BuildStart constructs the fixed start vector for a supported harness.
func BuildStart(request Request) (Invocation, error) {
	if err := validateRequest(request, false); err != nil {
		return Invocation{}, err
	}
	invocation := Invocation{Command: request.Harness, Dir: request.Bundle}
	switch request.Harness {
	case "codex":
		invocation.Args = []string{"exec", "-m", request.Model, "-s", "read-only", "-C", request.Bundle, "--json", "-o", request.ResultFile, "-"}
		invocation.Stdin = request.Prompt
	case "claude":
		invocation.Args = []string{"-p", "--output-format", "json", "--model", request.Model, "--session-id", request.SessionID, "--tools", "", "--permission-mode", "plan", request.Prompt}
	case "copilot":
		invocation.Args = []string{"-p", request.Prompt, "-s", "--output-format", "json", "--model", request.Model, "--name", request.SessionID, "--available-tools="}
	case "hermes":
		invocation.Args = []string{"-z", request.Prompt, "-m", request.Model, "--provider", request.Provider, "-t", "", "--pass-session-id", "--safe-mode", "--in", request.Bundle}
	}
	return invocation, nil
}

// BuildResume constructs the fixed resume vector for an existing identity.
func BuildResume(request Request) (Invocation, error) {
	if err := validateRequest(request, true); err != nil {
		return Invocation{}, err
	}
	invocation := Invocation{Command: request.Harness, Dir: request.Bundle}
	switch request.Harness {
	case "codex":
		invocation.Args = []string{"exec", "resume", "-m", request.Model, "--json", "-o", request.ResultFile, request.SessionID, "-"}
		invocation.Stdin = request.Prompt
	case "claude":
		invocation.Args = []string{"-p", "--output-format", "json", "--model", request.Model, "--resume", request.SessionID, "--tools", "", "--permission-mode", "plan", request.Prompt}
	case "copilot":
		invocation.Args = []string{"-p", request.Prompt, "-s", "--output-format", "json", "--model", request.Model, "--resume=" + request.SessionID, "--available-tools="}
	case "hermes":
		invocation.Args = []string{"-z", request.Prompt, "-m", request.Model, "--provider", request.Provider, "-t", "", "--resume", request.SessionID, "--safe-mode", "--in", request.Bundle}
	}
	return invocation, nil
}

func validateRequest(request Request, resume bool) error {
	request.Harness = strings.ToLower(strings.TrimSpace(request.Harness))
	switch request.Harness {
	case "codex", "claude", "copilot", "hermes":
	default:
		return fmt.Errorf("unsupported harness %q", request.Harness)
	}
	if strings.TrimSpace(request.Model) == "" {
		return errors.New("harness invocation requires a model")
	}
	if request.Harness == "hermes" && strings.TrimSpace(request.Provider) == "" {
		return errors.New("Hermes invocation requires a provider")
	}
	if (resume || request.Harness == "claude" || request.Harness == "copilot") && strings.TrimSpace(request.SessionID) == "" {
		return errors.New("harness invocation requires a stable session identity")
	}
	return nil
}

// ParseResult extracts the native stable identity and final response.
func ParseResult(harness string, stdout, resultFile []byte, preassignedIdentity string) (Result, error) {
	harness = strings.ToLower(strings.TrimSpace(harness))
	result := Result{SessionID: preassignedIdentity}
	switch harness {
	case "codex":
		scanner := bufio.NewScanner(bytes.NewReader(stdout))
		for scanner.Scan() {
			var event map[string]any
			if json.Unmarshal(scanner.Bytes(), &event) == nil && event["type"] == "thread.started" {
				result.SessionID, _ = event["thread_id"].(string)
			}
		}
		result.Response = strings.TrimSpace(string(resultFile))
	case "claude":
		var envelope struct {
			Result string `json:"result"`
		}
		if err := json.Unmarshal(stdout, &envelope); err != nil {
			return Result{}, fmt.Errorf("parsing Claude result: %w", err)
		}
		result.Response = strings.TrimSpace(envelope.Result)
	case "copilot":
		scanner := bufio.NewScanner(bytes.NewReader(stdout))
		for scanner.Scan() {
			var event map[string]any
			if json.Unmarshal(scanner.Bytes(), &event) != nil {
				continue
			}
			kind, _ := event["type"].(string)
			if kind == "assistant.message" || kind == "assistant" {
				result.Response, _ = event["content"].(string)
			}
		}
		result.Response = strings.TrimSpace(result.Response)
	case "hermes":
		lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
		if len(lines) != 0 && strings.HasPrefix(lines[0], "SESSION_ID:") {
			result.SessionID = strings.TrimSpace(strings.TrimPrefix(lines[0], "SESSION_ID:"))
			result.Response = strings.TrimSpace(strings.Join(lines[1:], "\n"))
		}
	default:
		return Result{}, fmt.Errorf("unsupported harness %q", harness)
	}
	if result.SessionID == "" {
		return Result{}, fmt.Errorf("%s result omitted the stable session identity", harness)
	}
	if result.Response == "" {
		return Result{}, fmt.Errorf("%s result omitted the final response", harness)
	}
	return result, nil
}

// ParseTimeout returns the typed diagnostic used when a harness exceeds its bound.
func ParseTimeout(harness string, timeout time.Duration) (Result, error) {
	return Result{}, fmt.Errorf("%s harness exceeded the %s timeout", harness, timeout)
}

// CreateBundle copies exact regular-file inputs into a read-only temporary directory.
func CreateBundle(parent string, inputs []Input) (Bundle, error) {
	if len(inputs) == 0 {
		return Bundle{}, errors.New("audit bundle requires at least one input")
	}
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return Bundle{}, fmt.Errorf("creating bundle parent: %w", err)
	}
	path, err := os.MkdirTemp(parent, "sdlc-audit-")
	if err != nil {
		return Bundle{}, fmt.Errorf("creating audit bundle: %w", err)
	}
	bundle := Bundle{Path: path, digests: map[string]string{}}
	for _, input := range inputs {
		if filepath.Base(input.Name) != input.Name || input.Name == "." || input.Name == "" {
			return Bundle{}, fmt.Errorf("invalid bundle input name %q", input.Name)
		}
		info, statErr := os.Lstat(input.Source)
		if statErr != nil {
			return Bundle{}, fmt.Errorf("inspecting bundle input %q: %w", input.Source, statErr)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return Bundle{}, fmt.Errorf("bundle input %q must be a regular non-symlink file", input.Source)
		}
		contents, readErr := os.ReadFile(input.Source)
		if readErr != nil {
			return Bundle{}, fmt.Errorf("reading bundle input %q: %w", input.Source, readErr)
		}
		destination := filepath.Join(path, input.Name)
		if writeErr := os.WriteFile(destination, contents, 0o400); writeErr != nil {
			return Bundle{}, fmt.Errorf("writing bundle input %q: %w", destination, writeErr)
		}
		bundle.digests[input.Name] = digest(contents)
	}
	if err := os.Chmod(path, 0o500); err != nil {
		return Bundle{}, fmt.Errorf("making audit bundle read-only: %w", err)
	}
	return bundle, nil
}

// Verify confirms that every bundled byte remains unchanged.
func (bundle Bundle) Verify() error {
	for name, want := range bundle.digests {
		contents, err := os.ReadFile(filepath.Join(bundle.Path, name))
		if err != nil {
			return fmt.Errorf("verifying bundle input %q: %w", name, err)
		}
		if got := digest(contents); got != want {
			return fmt.Errorf("audit bundle input %q changed", name)
		}
	}
	return nil
}

// Cleanup removes only the exact temporary bundle created by CreateBundle.
func (bundle Bundle) Cleanup() error {
	if bundle.Path == "" {
		return nil
	}
	if err := os.Chmod(bundle.Path, 0o700); err != nil {
		return fmt.Errorf("unlocking audit bundle for cleanup: %w", err)
	}
	for name := range bundle.digests {
		if err := os.Chmod(filepath.Join(bundle.Path, name), 0o600); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unlocking bundle input for cleanup: %w", err)
		}
	}
	return os.RemoveAll(bundle.Path)
}

func digest(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
