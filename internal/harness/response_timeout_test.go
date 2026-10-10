// ABOUTME: Exercises response deadlines using synthetic native provider events.
// ABOUTME: Separates model activity from initialization and process liveness.
package harness

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

func TestResponseDeadlines(t *testing.T) {
	for _, tc := range []struct{ name, event, want string }{
		{"initialization is not a response", `{"type":"system","subtype":"init"}`, "response-start-timeout"},
		{"reasoning starts idle deadline", `{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"thinking_delta","thinking":"working"}}}`, "response-idle-timeout"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := Request{Harness: "claude", Model: "fixture", SessionID: "native", ResponseStartTimeout: 80 * time.Millisecond, ResponseIdleTimeout: 15 * time.Millisecond}
			_, err := Execute(context.Background(), request, false, nil, func(ctx context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, _ io.Writer) error {
				if _, err := io.WriteString(stdout, tc.event+"\n"); err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
					return errors.New("watchdog did not cancel provider")
				}
			}, io.Discard)
			var incident *Incident
			if !errors.As(err, &incident) || incident.Kind != tc.want {
				t.Fatalf("got %v, want %s", err, tc.want)
			}
		})
	}
}

func TestHermesReasoningKeepsResponseAlive(t *testing.T) {
	request := Request{Harness: "hermes", Provider: "fixture", Model: "fixture", SessionID: "native", ResponseStartTimeout: 100 * time.Millisecond, ResponseIdleTimeout: 200 * time.Millisecond}
	result, err := Execute(context.Background(), request, false, nil, func(ctx context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, stderr io.Writer) error {
		if _, err := io.WriteString(stderr, "session_id: native\nsdlc_response: reasoning\n"); err != nil {
			return err
		}
		ticker := time.NewTicker(25 * time.Millisecond)
		defer ticker.Stop()
		finish := time.NewTimer(250 * time.Millisecond)
		defer finish.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				if _, err := io.WriteString(stderr, "sdlc_response: reasoning\n"); err != nil {
					return err
				}
			case <-finish.C:
				_, err := io.WriteString(stdout, "VERDICT: PASS\n")
				return err
			}
		}
	}, io.Discard)
	if err != nil || result.Response != "VERDICT: PASS" {
		t.Fatalf("active Hermes response=%#v err=%v", result, err)
	}
}

func TestLateSuccessCannotEscapeResponseDeadline(t *testing.T) {
	request := Request{Harness: "hermes", Provider: "fixture", Model: "fixture", SessionID: "native", ResponseStartTimeout: 20 * time.Millisecond}
	_, err := Execute(context.Background(), request, false, nil, func(ctx context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, stderr io.Writer) error {
		<-ctx.Done()
		if _, err := io.WriteString(stderr, "session_id: native\n"); err != nil {
			return err
		}
		_, err := io.WriteString(stdout, "VERDICT: PASS\n")
		return err
	}, io.Discard)
	var incident *Incident
	if !errors.As(err, &incident) || incident.Kind != "response-start-timeout" {
		t.Fatalf("late response accepted: %v", err)
	}
}
