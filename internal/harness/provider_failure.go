// ABOUTME: Recognizes narrow provider recovery diagnostics without interpreting audit prose.
// ABOUTME: Keeps live stderr and retains only a bounded diagnostic tail for classification.
package harness

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

type providerDiagnostics struct {
	mu         sync.Mutex
	writer     io.Writer
	tail       string
	request    Request
	pending    string
	identity   string
	err        error
	onActivity func()
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
			if body, found := strings.CutPrefix(line, "sdlc_failure: "); found {
				if d.err == nil {
					d.err = hermesFailure(body, d.request, d.identity)
				}
				continue
			}
			if line == "sdlc_response: content" || line == "sdlc_response: reasoning" {
				if d.onActivity != nil {
					d.onActivity()
				}
				continue
			}
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

func hermesFailure(body string, request Request, identity string) error {
	var failure struct {
		Version   int    `json:"version"`
		Reason    string `json:"reason"`
		Status    *int   `json:"status"`
		Retryable *bool  `json:"retryable"`
	}
	if len(body) > maxDiagnostic || json.Unmarshal([]byte(body), &failure) != nil || failure.Version != 1 {
		return newIncident("provider-rejected", request.Harness, identity, errors.New("invalid Hermes failure metadata; no verdict"))
	}
	kind := "provider-rejected"
	if failure.Retryable != nil && !*failure.Retryable {
		kind = "provider-permanent-error"
	}
	switch failure.Reason {
	case "model_not_found":
		kind = "model-unavailable"
	case "billing", "quota_exhausted":
		kind = "quota-exhausted"
	case "auth", "auth_permanent":
		kind = "authentication-failed"
	case "context_overflow", "payload_too_large":
		kind = "request-too-large"
	case "rate_limit", "upstream_rate_limit":
		kind = "rate-limited"
	case "overloaded", "server_error":
		kind = "provider-unavailable"
	case "ssl_cert_verification":
		kind = "tls-verification-failed"
	case "provider_policy_blocked", "content_policy_blocked", "format_error":
		kind = "request-rejected"
	case "timeout", "unknown":
	default:
		failure.Reason = "unknown"
	}
	detail := "Hermes provider failure: reason=" + failure.Reason
	if failure.Status != nil && *failure.Status >= 400 && *failure.Status <= 599 {
		detail += fmt.Sprintf(" HTTP=%d", *failure.Status)
	}
	if failure.Retryable != nil {
		detail += fmt.Sprintf(" retryable=%t", *failure.Retryable)
	}
	return newIncident(kind, request.Harness, identity, errors.New(detail))
}

func providerFailure(message string, request Request, resume bool) string {
	for _, line := range strings.Split(message, "\n") {
		line = strings.TrimSpace(line)
		if request.Harness == "hermes" {
			line = strings.TrimPrefix(line, "Hermes native adapter error: ")
		}
		// This native Claude diagnostic identifies the exact requested context.
		if resume && request.Harness == "claude" && request.SessionID != "" && strings.TrimSuffix(line, ".") == "No conversation found with session ID: "+request.SessionID {
			return "session-unavailable"
		}
		lower := strings.ToLower(line)
		if request.Harness == "codex" {
			var envelope struct {
				Error struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if json.Unmarshal([]byte(line), &envelope) == nil && strings.HasSuffix(envelope.Error.Message, "model is not supported when using Codex with a ChatGPT account.") {
				return "model-unavailable"
			}
		}
		if lower == "prompt is too long" || strings.HasPrefix(lower, "prompt is too long:") {
			return "request-too-large"
		}
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
