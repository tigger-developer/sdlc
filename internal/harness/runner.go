// Package harness defines fixed, capability-aware coding-agent invocations.
package harness

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Executor runs one fixed argument vector without a shell.
type Executor func(ctx context.Context, command string, args []string, directory string, stdin io.Reader, stdout, stderr io.Writer) error

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

// Incident is a fail-closed harness operation failure with a stable kind.
type Incident struct {
	Kind      string
	Harness   string
	SessionID string
	Err       error
}

func (incident *Incident) Error() string {
	identity := ""
	if incident.SessionID != "" {
		identity = " (session " + incident.SessionID + ")"
	}
	return fmt.Sprintf("%s harness %s%s: %v", incident.Harness, incident.Kind, identity, incident.Err)
}

func (incident *Incident) Unwrap() error { return incident.Err }

func newIncident(kind, harness, sessionID string, err error) error {
	return &Incident{Kind: kind, Harness: harness, SessionID: sessionID, Err: err}
}

// NewSessionIdentity returns a UUID suitable for every supported harness.
func NewSessionIdentity() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("creating harness session identity: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
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
	sources map[string]string
}

// Execute runs one start or resume invocation and validates its bounded result.
func Execute(ctx context.Context, request Request, resume bool, bundle *Bundle, executor Executor, errorOutput io.Writer) (Result, error) {
	if executor == nil {
		executor = executeCommand
	}
	if errorOutput == nil {
		errorOutput = io.Discard
	}
	if bundle != nil {
		if err := bundle.Verify(); err != nil {
			return Result{}, err
		}
	}
	var invocation Invocation
	var err error
	if resume {
		invocation, err = BuildResume(request)
	} else {
		invocation, err = BuildStart(request)
	}
	if err != nil {
		return Result{}, err
	}
	var stdout bytes.Buffer
	if err := executor(ctx, invocation.Command, invocation.Args, invocation.Dir, strings.NewReader(invocation.Stdin), &stdout, errorOutput); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return Result{}, newIncident("timeout", request.Harness, request.SessionID, ctx.Err())
		}
		kind := "start-failed"
		if resume {
			kind = "resume-failed"
		}
		if errors.Is(err, exec.ErrNotFound) {
			kind = "executable-unavailable"
		}
		return Result{}, newIncident(kind, request.Harness, request.SessionID, err)
	}
	if bundle != nil {
		if err := bundle.Verify(); err != nil {
			return Result{}, err
		}
	}
	var resultFile []byte
	if request.ResultFile != "" {
		resultFile, err = os.ReadFile(request.ResultFile)
		if err != nil {
			return Result{}, fmt.Errorf("reading %s final response: %w", request.Harness, err)
		}
	}
	result, err := ParseResult(request.Harness, stdout.Bytes(), resultFile, request.SessionID)
	if err != nil {
		return Result{}, err
	}
	return result, nil
}

func executeCommand(ctx context.Context, command string, args []string, directory string, stdin io.Reader, stdout, stderr io.Writer) error {
	path, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("executable unavailable: %w", err)
	}
	process := exec.Command(path, args...)
	process.Dir = directory
	process.Stdin = stdin
	process.Stdout = stdout
	process.Stderr = stderr
	configureProcessGroup(process)
	if err := process.Start(); err != nil {
		return err
	}
	completed := make(chan error, 1)
	go func() { completed <- process.Wait() }()
	select {
	case err := <-completed:
		return err
	case <-ctx.Done():
		if err := terminateProcessGroup(process); err != nil {
			return fmt.Errorf("terminating timed-out harness: %w", err)
		}
		<-completed
		return ctx.Err()
	}
}

// BuildStart constructs the fixed start vector for a supported harness.
func BuildStart(request Request) (Invocation, error) {
	request.Harness = strings.ToLower(strings.TrimSpace(request.Harness))
	if err := validateRequest(request, false); err != nil {
		return Invocation{}, err
	}
	invocation := Invocation{Command: request.Harness, Dir: request.Bundle}
	switch request.Harness {
	case "codex":
		invocation.Args = []string{"exec", "-m", request.Model, "-s", "read-only", "-C", request.Bundle, "--skip-git-repo-check", "--json", "-o", request.ResultFile, "-"}
		invocation.Stdin = request.Prompt
	case "claude":
		invocation.Args = []string{"-p", "--output-format", "json", "--model", request.Model, "--session-id", request.SessionID, "--tools", "", "--permission-mode", "plan", request.Prompt}
	case "copilot":
		invocation.Args = []string{"-p", request.Prompt, "-s", "--output-format", "json", "--model", request.Model, "--name", request.SessionID, "--available-tools="}
	case "hermes":
		invocation.Args = []string{"-z", hermesPrompt(request.Prompt), "-m", request.Model, "--provider", request.Provider, "-t", "", "--pass-session-id", "--safe-mode", "--in", request.Bundle}
	}
	return invocation, nil
}

// BuildResume constructs the fixed resume vector for an existing identity.
func BuildResume(request Request) (Invocation, error) {
	request.Harness = strings.ToLower(strings.TrimSpace(request.Harness))
	if err := validateRequest(request, true); err != nil {
		return Invocation{}, err
	}
	invocation := Invocation{Command: request.Harness, Dir: request.Bundle}
	switch request.Harness {
	case "codex":
		invocation.Args = []string{"exec", "resume", "-m", request.Model, "--skip-git-repo-check", "--json", "-o", request.ResultFile, request.SessionID, "-"}
		invocation.Stdin = request.Prompt
	case "claude":
		invocation.Args = []string{"-p", "--output-format", "json", "--model", request.Model, "--resume", request.SessionID, "--tools", "", "--permission-mode", "plan", request.Prompt}
	case "copilot":
		invocation.Args = []string{"-p", request.Prompt, "-s", "--output-format", "json", "--model", request.Model, "--resume=" + request.SessionID, "--available-tools="}
	case "hermes":
		invocation.Args = []string{"-z", hermesPrompt(request.Prompt), "-m", request.Model, "--provider", request.Provider, "-t", "", "--resume", request.SessionID, "--safe-mode", "--in", request.Bundle}
	}
	return invocation, nil
}

func hermesPrompt(prompt string) string {
	return "Return your native session identity on the first output line exactly as SESSION_ID: <id>, followed by the requested response.\n\n" + prompt
}

func validateRequest(request Request, resume bool) error {
	request.Harness = strings.ToLower(strings.TrimSpace(request.Harness))
	switch request.Harness {
	case "codex", "claude", "copilot", "hermes":
	default:
		return newIncident("capability-unsupported", request.Harness, request.SessionID, fmt.Errorf("unsupported harness %q", request.Harness))
	}
	if strings.TrimSpace(request.Model) == "" {
		return newIncident("configuration-invalid", request.Harness, request.SessionID, errors.New("invocation requires a model"))
	}
	if request.Harness == "hermes" && strings.TrimSpace(request.Provider) == "" {
		return newIncident("configuration-invalid", request.Harness, request.SessionID, errors.New("Hermes invocation requires a provider"))
	}
	if (resume || request.Harness == "claude" || request.Harness == "copilot") && strings.TrimSpace(request.SessionID) == "" {
		return newIncident("identity-missing", request.Harness, request.SessionID, errors.New("invocation requires a stable session identity"))
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
			return Result{}, newIncident("response-malformed", harness, result.SessionID, fmt.Errorf("parsing Claude result: %w", err))
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
		return Result{}, newIncident("capability-unsupported", harness, result.SessionID, fmt.Errorf("unsupported harness %q", harness))
	}
	if result.SessionID == "" {
		return Result{}, newIncident("identity-missing", harness, "", errors.New("result omitted the stable session identity"))
	}
	if result.Response == "" {
		return Result{}, newIncident("response-empty", harness, result.SessionID, errors.New("result omitted the final response"))
	}
	return result, nil
}

// ValidateCompositeVerdict enforces the audit gate's minimal result envelope.
func ValidateCompositeVerdict(response string) error {
	fields := map[string][]string{}
	for _, line := range strings.Split(response, "\n") {
		for _, key := range []string{"GATE", "REVISION", "VERDICT"} {
			prefix := key + ":"
			if strings.HasPrefix(line, prefix) {
				fields[key] = append(fields[key], strings.TrimSpace(strings.TrimPrefix(line, prefix)))
			}
		}
	}
	for _, key := range []string{"GATE", "REVISION", "VERDICT"} {
		if len(fields[key]) != 1 || fields[key][0] == "" {
			return newIncident("response-malformed", "audit", "", fmt.Errorf("composite verdict requires exactly one non-empty %s field", key))
		}
	}
	if fields["GATE"][0] != "definition" && fields["GATE"][0] != "implementation" {
		return newIncident("response-malformed", "audit", "", fmt.Errorf("unsupported GATE %q", fields["GATE"][0]))
	}
	if fields["VERDICT"][0] != "PASS" && fields["VERDICT"][0] != "PROVISIONAL" && fields["VERDICT"][0] != "FAIL" {
		return newIncident("response-malformed", "audit", "", fmt.Errorf("unsupported VERDICT %q", fields["VERDICT"][0]))
	}
	return nil
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
	bundle := Bundle{Path: path, digests: map[string]string{}, sources: map[string]string{}}
	fail := func(cause error) (Bundle, error) {
		if cleanupErr := bundle.Cleanup(); cleanupErr != nil {
			return Bundle{}, fmt.Errorf("%w; cleaning incomplete audit bundle: %v", cause, cleanupErr)
		}
		return Bundle{}, cause
	}
	type manifestEntry struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Bytes       int    `json:"bytes"`
		SHA256      string `json:"sha256"`
	}
	const maximumInputBytes = 16 * 1024 * 1024
	const maximumBundleBytes = 64 * 1024 * 1024
	seen := map[string]bool{}
	totalBytes := 0
	manifest := make([]manifestEntry, 0, len(inputs))
	for _, input := range inputs {
		if filepath.Base(input.Name) != input.Name || input.Name == "." || input.Name == "" {
			return fail(fmt.Errorf("invalid bundle input name %q", input.Name))
		}
		if input.Name == "manifest.json" || seen[input.Name] {
			return fail(fmt.Errorf("duplicate or reserved bundle input name %q", input.Name))
		}
		seen[input.Name] = true
		info, statErr := os.Lstat(input.Source)
		if statErr != nil {
			return fail(fmt.Errorf("inspecting bundle input %q: %w", input.Source, statErr))
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fail(fmt.Errorf("bundle input %q must be a regular non-symlink file", input.Source))
		}
		contents, readErr := os.ReadFile(input.Source)
		if readErr != nil {
			return fail(fmt.Errorf("reading bundle input %q: %w", input.Source, readErr))
		}
		if len(contents) > maximumInputBytes || totalBytes+len(contents) > maximumBundleBytes {
			return fail(fmt.Errorf("bundle input %q exceeds the bounded evidence size", input.Source))
		}
		totalBytes += len(contents)
		destination := filepath.Join(path, input.Name)
		if writeErr := os.WriteFile(destination, contents, 0o400); writeErr != nil {
			return fail(fmt.Errorf("writing bundle input %q: %w", destination, writeErr))
		}
		checksum := digest(contents)
		bundle.digests[input.Name] = checksum
		bundle.sources[input.Name] = input.Source
		manifest = append(manifest, manifestEntry{Source: input.Source, Destination: input.Name, Bytes: len(contents), SHA256: checksum})
	}
	manifestBytes, err := json.MarshalIndent(map[string]any{"version": 1, "files": manifest}, "", "  ")
	if err != nil {
		return fail(fmt.Errorf("rendering audit bundle manifest: %w", err))
	}
	manifestBytes = append(manifestBytes, '\n')
	if err := os.WriteFile(filepath.Join(path, "manifest.json"), manifestBytes, 0o400); err != nil {
		return fail(fmt.Errorf("writing audit bundle manifest: %w", err))
	}
	bundle.digests["manifest.json"] = digest(manifestBytes)
	if err := os.Chmod(path, 0o500); err != nil {
		return fail(fmt.Errorf("making audit bundle read-only: %w", err))
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
		if source, exists := bundle.sources[name]; exists {
			sourceContents, err := os.ReadFile(source)
			if err != nil {
				return fmt.Errorf("verifying audit source %q: %w", source, err)
			}
			if got := digest(sourceContents); got != want {
				return fmt.Errorf("audit source %q changed", source)
			}
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
