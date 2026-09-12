// ABOUTME: Observes native session identities in provider output before a final verdict.
// ABOUTME: Checkpoints retained sessions without treating a generated name as a Codex or Hermes ID.
package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type sessionOutput struct {
	buffer       bytes.Buffer
	request      Request
	identity     string
	pending      []byte
	err          error
	progress     io.Writer
	notices      int
	sawJSON      bool
	plainFailure string
}

const maxProviderOutput = 16 * 1024 * 1024
const maxDiagnostic = 4096
const maxProgressNotices = 200

func newSessionOutput(request Request, resume bool) *sessionOutput {
	output := &sessionOutput{request: request}
	if resume || request.Harness == "claude" || request.Harness == "copilot" {
		output.identity = request.SessionID
	}
	return output
}

func (output *sessionOutput) checkpoint() error {
	if output.identity != "" && output.request.OnSession != nil {
		return output.request.OnSession(output.identity)
	}
	return nil
}

func (output *sessionOutput) Write(data []byte) (int, error) {
	if len(data) > maxProviderOutput-output.buffer.Len() {
		output.err = newIncident("output-limit", output.request.Harness, output.identity, errors.New("provider stdout exceeds 16 MiB"))
		return 0, output.err
	}
	n, err := output.buffer.Write(data)
	if err != nil {
		return n, err
	}
	output.pending = append(output.pending, data...)
	for {
		index := bytes.IndexByte(output.pending, '\n')
		if index < 0 {
			break
		}
		line := output.pending[:index]
		output.pending = output.pending[index+1:]
		output.report(line)
		identity := nativeIdentity(output.request.Harness, line)
		if identity == "" {
			continue
		}
		if output.identity != "" && identity != output.identity {
			output.err = errors.New("provider returned conflicting native session identities")
			return n, output.err
		}
		if output.identity == identity {
			continue
		}
		output.identity = identity
		if output.err = output.checkpoint(); output.err != nil {
			return n, output.err
		}
	}
	return n, nil
}

// finish handles the final record even when a provider exits without a newline.
func (output *sessionOutput) finish(failed bool) {
	if !output.sawJSON && json.Valid(output.buffer.Bytes()) {
		output.report(output.buffer.Bytes())
		output.pending = nil
	}
	if len(output.pending) > 0 {
		output.report(output.pending)
		output.pending = nil
	}
	trimmed := bytes.TrimSpace(output.buffer.Bytes())
	structured := len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[')
	if failed && !output.sawJSON && !structured && output.plainFailure != "" {
		output.diagnostic("provider stdout on failure", output.plainFailure)
	}
}

// report forwards event metadata, never assistant prose, tool payloads or prompts.
// Error messages are the only provider text exposed by this stdout adapter.
func (output *sessionOutput) report(line []byte) {
	if output.progress == nil || len(bytes.TrimSpace(line)) == 0 {
		return
	}
	var event struct {
		Type    string          `json:"type"`
		Subtype string          `json:"subtype"`
		IsError bool            `json:"is_error"`
		Result  string          `json:"result"`
		Message json.RawMessage `json:"message"`
		Errors  []string        `json:"errors"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(line, &event); err != nil {
		// Never dump an unknown structured payload. Plain-text startup failures
		// may have no JSON envelope, and are useful only on unsuccessful exit.
		trimmed := bytes.TrimSpace(line)
		if trimmed[0] != '{' && trimmed[0] != '[' && !bytes.HasPrefix(trimmed, []byte("SESSION_ID:")) {
			if len(trimmed) > maxDiagnostic {
				trimmed = trimmed[len(trimmed)-maxDiagnostic:]
			}
			output.plainFailure += string(trimmed) + "\n"
			if len(output.plainFailure) > maxDiagnostic {
				output.plainFailure = output.plainFailure[len(output.plainFailure)-maxDiagnostic:]
			}
		}
		return
	}
	output.sawJSON = true
	if event.IsError || event.Type == "error" || event.Type == "turn.failed" || strings.HasPrefix(event.Subtype, "error_") {
		var message string
		_ = json.Unmarshal(event.Message, &message) // Only string error messages are safe to relay.
		if message == "" {
			message = event.Result
		}
		if message == "" {
			message = strings.Join(event.Errors, "; ")
		}
		if message == "" {
			var detail struct {
				Message string `json:"message"`
			}
			if json.Unmarshal(event.Error, &detail) == nil {
				message = detail.Message
			}
			if message == "" {
				_ = json.Unmarshal(event.Error, &message)
			} // A non-string error has no safe text to display.
		}
		if message == "" {
			message = event.Type + " " + event.Subtype
		}
		output.diagnostic("provider error", message)
		return
	}
	if event.Type == "" || output.notices >= maxProgressNotices {
		return
	}
	output.notices++
	output.diagnostic("provider event", event.Type)
	if output.notices == maxProgressNotices {
		output.diagnostic("progress", "event display limit reached; heartbeats and errors continue")
	}
}

func (output *sessionOutput) diagnostic(kind, message string) {
	if len(message) > maxDiagnostic {
		message = message[:maxDiagnostic] + " [truncated]"
	}
	// Quoting escapes control characters instead of executing terminal sequences.
	fmt.Fprintf(output.progress, "sdlc-harness: %s (%s): %q\n", kind, output.request.Harness, message)
}

func (output *sessionOutput) Bytes() []byte { return output.buffer.Bytes() }

func nativeIdentity(provider string, line []byte) string {
	var identity string
	switch provider {
	case "codex":
		var event struct {
			Type     string `json:"type"`
			ThreadID string `json:"thread_id"`
		}
		if json.Unmarshal(line, &event) == nil && event.Type == "thread.started" {
			identity = event.ThreadID
		}
	case "hermes":
		if strings.HasPrefix(string(line), "SESSION_ID:") {
			identity = strings.TrimSpace(strings.TrimPrefix(string(line), "SESSION_ID:"))
		}
	}
	if strings.ContainsAny(identity, " \t\r\n") {
		return ""
	}
	return identity
}
