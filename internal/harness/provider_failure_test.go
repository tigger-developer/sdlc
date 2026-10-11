// ABOUTME: Verifies native failure metadata prevents failed turns supplying verdicts.
// ABOUTME: Keeps provider errors typed even when later diagnostics exceed the retained tail.
package harness

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestHermesStructuredFailure(t *testing.T) {
	for _, tc := range []struct {
		reason, kind string
		retryable    string
	}{
		{"model_not_found", "model-unavailable", "false"},
		{"quota_exhausted", "quota-exhausted", "false"},
		{"billing", "quota-exhausted", "false"},
		{"auth_permanent", "authentication-failed", "false"},
		{"format_error", "request-rejected", "false"},
		{"server_error", "provider-unavailable", "true"},
		{"context_overflow", "request-too-large", "true"},
		{"unknown", "provider-rejected", "null"},
	} {
		t.Run(tc.reason, func(t *testing.T) {
			executor := func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, stderr io.Writer) error {
				fmt.Fprintln(stderr, "session_id: native-fixture")
				line := fmt.Sprintf("sdlc_failure: {\"version\":1,\"reason\":%q,\"retryable\":%s,\"status\":403}\n", tc.reason, tc.retryable)
				for _, part := range []string{line[:10], line[10:], strings.Repeat("diagnostic\n", maxDiagnostic)} {
					fmt.Fprint(stderr, part)
				}
				fmt.Fprintln(stdout, "VERDICT: PASS")
				return nil // A failure cannot become a verdict even after a clean process exit.
			}
			result, err := Execute(context.Background(), Request{Harness: "hermes", Model: "fixture", Provider: "fixture"}, false, nil, executor, io.Discard)
			var incident *Incident
			if !errors.As(err, &incident) || incident.Kind != tc.kind || incident.SessionID != "native-fixture" || result.Response != "" {
				t.Fatalf("result=%+v err=%v", result, err)
			}
		})
	}
}
