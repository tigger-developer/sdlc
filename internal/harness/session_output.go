// ABOUTME: Observes native session identities in provider output before a final verdict.
// ABOUTME: Checkpoints retained sessions without treating a generated name as a Codex or Hermes ID.
package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

type sessionOutput struct {
	buffer   bytes.Buffer
	request  Request
	identity string
	pending  []byte
	err      error
}

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
