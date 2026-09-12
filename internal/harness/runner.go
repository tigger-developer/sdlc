// Package harness defines fixed, capability-aware coding-agent invocations.
package harness

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	Directory  string
	ResultFile string
	SessionID  string
	Evidence   []EvidenceFile
	// OnSession checkpoints a native identity as soon as the adapter observes it.
	OnSession func(string) error
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

// Execute runs one start or resume invocation and validates its bounded result.
func Execute(ctx context.Context, request Request, resume bool, evidence *Evidence, executor Executor, errorOutput io.Writer) (Result, error) {
	if executor == nil {
		executor = executeCommand
	}
	if errorOutput == nil {
		errorOutput = io.Discard
	}
	errorOutput = &lockedWriter{writer: errorOutput}
	if evidence != nil {
		if err := evidence.Verify(); err != nil {
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
	stdout := newSessionOutput(request, resume)
	stdout.progress = errorOutput
	if err := stdout.checkpoint(); err != nil {
		return Result{}, err
	}
	stopHeartbeat := make(chan struct{})
	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		reportHeartbeat(errorOutput, request.Harness, request.SessionID, stopHeartbeat)
	}()
	defer func() { <-heartbeatDone }()
	diagnostics := &providerDiagnostics{writer: errorOutput, request: request}
	executionErr := executor(ctx, invocation.Command, invocation.Args, invocation.Dir, strings.NewReader(invocation.Stdin), stdout, diagnostics)
	stdout.finish(executionErr != nil)
	if request.Harness == "hermes" {
		if diagnostics.identity != "" {
			stdout.identity = diagnostics.identity
		}
		if diagnostics.err != nil {
			stdout.err = diagnostics.err
		}
	}
	if err := executionErr; err != nil {
		close(stopHeartbeat)
		if stdout.err != nil {
			return Result{}, stdout.err
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return Result{}, newIncident("timeout", request.Harness, stdout.identity, ctx.Err())
		}
		kind := "start-failed"
		if resume {
			kind = "resume-failed"
		}
		if errors.Is(err, exec.ErrNotFound) {
			kind = "executable-unavailable"
		}
		if recovery := providerFailure(diagnostics.tail+"\n"+stdout.failureText+"\n"+stdout.plainFailure, request, resume); recovery != "" {
			kind = recovery
		}
		return Result{}, newIncident(kind, request.Harness, stdout.identity, err)
	}
	close(stopHeartbeat)
	if stdout.err != nil {
		return Result{}, stdout.err
	}
	if recovery := providerFailure(stdout.failureText, request, resume); recovery != "" {
		return Result{}, newIncident(recovery, request.Harness, stdout.identity, errors.New("provider rejected the invocation; see stderr"))
	}
	if evidence != nil {
		if err := evidence.Verify(); err != nil {
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
	var result Result
	if request.Harness == "hermes" {
		result = Result{SessionID: stdout.identity, Response: strings.TrimSpace(string(stdout.Bytes()))}
		// Hermes emits this native startup notice on stdout for an empty resolved
		// toolset. Keep it visible on stderr, outside the model's final response.
		const noToolsNotice = "Warning: Unknown toolsets: none\n"
		if response, found := strings.CutPrefix(result.Response, noToolsNotice); found {
			fmt.Fprint(errorOutput, noToolsNotice)
			result.Response = strings.TrimSpace(response)
		}
		if diagnostics.identity == "" {
			return Result{}, newIncident("identity-missing", request.Harness, "", errors.New("Hermes omitted native stderr session info"))
		}
		if result.Response == "" {
			return Result{}, newIncident("response-empty", request.Harness, result.SessionID, errors.New("Hermes omitted final response"))
		}
	} else {
		result, err = ParseResult(request.Harness, stdout.Bytes(), resultFile, stdout.identity)
	}
	if err != nil {
		return Result{}, err
	}
	if resume && result.SessionID != request.SessionID {
		return Result{}, newIncident("identity-mismatch", request.Harness, request.SessionID, errors.New("provider returned a different session"))
	}
	return result, nil
}

// Provider stderr and the heartbeat share one stream, including test buffers.
type lockedWriter struct {
	mu     sync.Mutex
	writer io.Writer
}

func (output *lockedWriter) Write(data []byte) (int, error) {
	output.mu.Lock()
	defer output.mu.Unlock()
	return output.writer.Write(data)
}

// reportHeartbeat gives the invoking agent visible liveness without implying
// provider progress or an audit verdict. Provider output still flows directly
// to errorOutput.
func reportHeartbeat(output io.Writer, harness, session string, stop <-chan struct{}) {
	started := time.Now()
	fmt.Fprintf(output, "sdlc-harness: external context started (harness=%s session=%s); waiting for final response\n", harness, session)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Fprintf(output, "sdlc-harness: external context still running (harness=%s session=%s elapsed=%s)\n", harness, session, time.Since(started).Round(time.Second))
		case <-stop:
			return
		}
	}
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
	invocation := Invocation{Command: request.Harness, Dir: request.Directory}
	switch request.Harness {
	case "codex":
		invocation.Args = []string{"exec", "-m", request.Model, "-s", "read-only", "-C", request.Directory, "--skip-git-repo-check", "--json", "-o", request.ResultFile, "-"}
		invocation.Stdin = request.Prompt
	case "claude":
		invocation.Args = []string{"-p", "--output-format", "stream-json", "--verbose", "--model", request.Model, "--session-id", request.SessionID, "--tools", "Read", "--permission-mode", "plan"}
		settings, err := claudeEvidenceSettings(request.Evidence)
		if err != nil {
			return Invocation{}, err
		}
		invocation.Args = append(invocation.Args, settings...)
		invocation.Args = append(invocation.Args, request.Prompt)
	case "copilot":
		invocation.Args = []string{"-p", request.Prompt, "-s", "--output-format", "json", "--model", request.Model, "--name", request.SessionID, "--available-tools="}
	case "hermes":
		invocation.Args = []string{"chat", "--quiet", "--query-file", "-", "-m", request.Model, "--provider", request.Provider, "-t", "none", "--safe-mode", "--in", request.Directory}
		invocation.Stdin = request.Prompt
	}
	return invocation, nil
}

// BuildResume constructs the fixed resume vector for an existing identity.
func BuildResume(request Request) (Invocation, error) {
	request.Harness = strings.ToLower(strings.TrimSpace(request.Harness))
	if err := validateRequest(request, true); err != nil {
		return Invocation{}, err
	}
	invocation := Invocation{Command: request.Harness, Dir: request.Directory}
	switch request.Harness {
	case "codex":
		invocation.Args = []string{"exec", "resume", "-m", request.Model, "--skip-git-repo-check", "--json", "-o", request.ResultFile, request.SessionID, "-"}
		invocation.Stdin = request.Prompt
	case "claude":
		invocation.Args = []string{"-p", "--output-format", "stream-json", "--verbose", "--model", request.Model, "--resume", request.SessionID, "--tools", "Read", "--permission-mode", "plan"}
		settings, err := claudeEvidenceSettings(request.Evidence)
		if err != nil {
			return Invocation{}, err
		}
		invocation.Args = append(invocation.Args, settings...)
		invocation.Args = append(invocation.Args, request.Prompt)
	case "copilot":
		invocation.Args = []string{"-p", request.Prompt, "-s", "--output-format", "json", "--model", request.Model, "--resume=" + request.SessionID, "--available-tools="}
	case "hermes":
		invocation.Args = []string{"chat", "--quiet", "--query-file", "-", "-m", request.Model, "--provider", request.Provider, "-t", "none", "--resume", request.SessionID, "--safe-mode", "--in", request.Directory}
		invocation.Stdin = request.Prompt
	}
	return invocation, nil
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
		var err error
		result.Response, err = parseClaudeResponse(stdout)
		if err != nil {
			return Result{}, newIncident("response-malformed", harness, result.SessionID, err)
		}
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
	if fields["VERDICT"][0] != "PASS" && fields["VERDICT"][0] != "PROVISIONAL PASS" && fields["VERDICT"][0] != "FAIL" {
		return newIncident("response-malformed", "audit", "", fmt.Errorf("unsupported VERDICT %q", fields["VERDICT"][0]))
	}
	return nil
}

// ParseTimeout returns the typed diagnostic used when a harness exceeds its bound.
func ParseTimeout(harness string, timeout time.Duration) (Result, error) {
	return Result{}, fmt.Errorf("%s harness exceeded the %s timeout", harness, timeout)
}
