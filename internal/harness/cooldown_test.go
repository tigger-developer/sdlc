// ABOUTME: Verifies provider failures and successful prose without parsing reset times.
// ABOUTME: Uses local executors only; no provider or global runtime state is accessed.
package harness

import (
	"context"
	"errors"
	"io"
	"testing"
)

func TestClaudeErrorEnvelopeRemainsIncident(t *testing.T) {
	for _, exitFailure := range []bool{false, true} {
		r := matrixRequest("claude", t.TempDir())
		_, err := Execute(context.Background(), r, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, _ io.Writer) error {
			if _, err := io.WriteString(out, `{"type":"result","is_error":true,"result":"Provider unavailable, no reset time supplied"}`); err != nil {
				return err
			}
			if exitFailure {
				return errors.New("exit status 1")
			}
			return nil
		}, io.Discard)
		if err == nil {
			t.Fatal("accepted provider error as a result")
		}
	}
}

func TestSuccessfulAuditProseDoesNotTriggerCooldown(t *testing.T) {
	r := matrixRequest("claude", t.TempDir())
	result, err := Execute(context.Background(), r, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, _ io.Writer) error {
		_, err := io.WriteString(out, `{"type":"result","is_error":false,"result":"You've hit your session limit · resets 6:20pm (Europe/Dublin)"}`)
		return err
	}, io.Discard)
	if err != nil || result.Response == "" {
		t.Fatalf("successful response rejected: %#v %v", result, err)
	}
}
