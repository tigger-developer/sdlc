// ABOUTME: Recognizes narrow provider recovery diagnostics without interpreting audit prose.
// ABOUTME: Keeps live stderr and retains only a bounded diagnostic tail for classification.
package harness

import (
	"encoding/json"
	"io"
	"strings"
	"sync"
)

type providerDiagnostics struct {
	mu       sync.Mutex
	writer   io.Writer
	tail     string
	request  Request
	pending  string
	identity string
	err      error
}

func (d *providerDiagnostics) Write(data []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.tail += string(data)
	if len(d.tail) > maxDiagnostic {
		d.tail = d.tail[len(d.tail)-maxDiagnostic:]
	}
	if d.request.Harness == "hermes" {
		d.pending += string(data)
		for {
			line, rest, found := strings.Cut(d.pending, "\n")
			if !found {
				break
			}
			d.pending = rest
			if !strings.HasPrefix(line, "session_id: ") {
				continue
			}
			identity := strings.TrimSpace(strings.TrimPrefix(line, "session_id: "))
			if identity == "" || strings.ContainsAny(identity, " \t\r\n") || identity == d.identity {
				continue
			}
			d.identity = identity
			if d.request.OnSession != nil {
				if err := d.request.OnSession(identity); err != nil {
					d.err = err
					return 0, err
				}
			}
		}
		if len(d.pending) > maxDiagnostic {
			d.pending = d.pending[len(d.pending)-maxDiagnostic:]
		}
	}
	return d.writer.Write(data)
}

func providerFailure(message string, request Request, resume bool) string {
	for _, line := range strings.Split(message, "\n") {
		line = strings.TrimSpace(line)
		// This native Claude diagnostic identifies the exact requested context.
		if resume && request.Harness == "claude" && request.SessionID != "" && strings.TrimSuffix(line, ".") == "No conversation found with session ID: "+request.SessionID {
			return "session-unavailable"
		}
		lower := strings.ToLower(line)
		// Claude can wrap the explicit authentication error in an HTTP diagnostic.
		if body, ok := strings.CutPrefix(line, "API Error: 401 "); ok {
			var envelope struct {
				Error struct {
					Type string `json:"type"`
				} `json:"error"`
			}
			if json.Unmarshal([]byte(body), &envelope) == nil && envelope.Error.Type == "authentication_error" {
				return "authentication-failed"
			}
		}
		if strings.HasPrefix(lower, "not logged in") || strings.HasPrefix(lower, "oauth token has expired") || strings.HasPrefix(lower, "your oauth token has expired") || strings.HasPrefix(lower, "authentication failed") || strings.HasPrefix(lower, "invalid api key") {
			return "authentication-failed"
		}
	}
	return ""
}
