// ABOUTME: Exercises native usage-limit detection without calling a provider.
// ABOUTME: Keeps ordinary audit prose distinct from provider failures.
package harness

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

func TestClaudeUsageLimitIncident(t *testing.T) {
	for _, exitFailure := range []bool{false, true} {
		r := matrixRequest("claude", t.TempDir())
		_, err := Execute(context.Background(), r, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, _ io.Writer) error {
			if _, err := io.WriteString(out, `{"type":"result","is_error":true,"result":"You've hit your session limit · resets 6:20pm (Europe/Dublin)"}`); err != nil {
				return err
			}
			if exitFailure {
				return errors.New("exit status 1")
			}
			return nil
		}, io.Discard)
		assertIncidentKind(t, err, "rate-limited")
		var incident *Incident
		if !errors.As(err, &incident) || incident.RetryAt.IsZero() {
			t.Fatalf("missing reset deadline: %v", err)
		}
	}
}

func TestClaudeResetDeadline(t *testing.T) {
	now := time.Date(2026, 9, 20, 14, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name, message, provider, want string
	}{
		{"daylight-saving", "You've hit your session limit · resets 6:20pm (Europe/Dublin)", "claude", "2026-09-20T17:20:00Z"},
		{"next-day", "You've hit your session limit · resets 4:40am (Europe/Dublin)", "claude", "2026-09-21T03:40:00Z"},
		{"invalid-zone", "You've hit your session limit · resets 6:20pm (Not/AZone)", "claude", ""},
		{"invalid-clock", "You've hit your session limit · resets 25:20pm (Europe/Dublin)", "claude", ""},
		{"missing-zone", "You've hit your session limit · resets 6:20pm", "claude", ""},
		{"other-provider", "You've hit your session limit · resets 6:20pm (Europe/Dublin)", "hermes", ""},
		{"quoted-prose", "Example: You've hit your session limit · resets 6:20pm (Europe/Dublin)", "claude", ""},
		{"credits", "Billing or credits exhausted: HTTP 402", "claude", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := claudeReset(tc.message, tc.provider, now)
			if tc.want == "" {
				if !got.IsZero() {
					t.Fatalf("unexpected deadline %s", got)
				}
			} else if got.Format(time.RFC3339) != tc.want {
				t.Fatalf("deadline %s, want %s", got, tc.want)
			}
		})
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
