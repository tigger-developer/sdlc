// ABOUTME: Verifies observable provider progress, failure diagnostics and bounded output.
// ABOUTME: Uses local executors only; no metered provider is invoked.
package harness

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

type outputLog struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *outputLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *outputLog) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func TestW009ProgressArrivesBeforeProviderExit(t *testing.T) {
	var log outputLog
	r := matrixRequest("claude", t.TempDir())
	result, err := Execute(context.Background(), r, false, nil, func(_ context.Context, _ string, args []string, _ string, _ io.Reader, stdout, _ io.Writer) error {
		if !strings.Contains(strings.Join(args, " "), "--output-format stream-json --verbose") {
			t.Error("Claude invocation must request streaming output")
		}
		for _, part := range []string{`{"type":"assis`, `tant","message":{"content":"PRIVATE PAYLOAD"}}` + "\n"} {
			if _, e := io.WriteString(stdout, part); e != nil {
				return e
			}
		}
		if !strings.Contains(log.String(), "assistant") {
			t.Error("no provider event visible before exit")
		}
		if strings.Contains(log.String(), "PRIVATE PAYLOAD") {
			t.Error("progress leaked payload")
		}
		_, e := io.WriteString(stdout, `{"type":"result","is_error":false,"result":"final"}`)
		return e
	}, &log)
	if err != nil || result.Response != "final" {
		t.Fatalf("result=%#v error=%v", result, err)
	}
}

func TestW009ProviderErrorsSurviveUnsuccessfulExit(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "copilot", "hermes"} {
		for _, timedOut := range []bool{false, true} {
			t.Run(provider+map[bool]string{false: "-exit", true: "-timeout"}[timedOut], func(t *testing.T) {
				var log outputLog
				r := matrixRequest(provider, t.TempDir())
				ctx := context.Background()
				if timedOut {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, time.Nanosecond)
					defer cancel()
					<-ctx.Done()
				}
				result, err := Execute(ctx, r, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, _ io.Writer) error {
					message := `{"type":"error","message":"model not available"}`
					if provider == "claude" {
						message = `{"type":"result","is_error":true,"errors":["model not available"]}`
					}
					if provider == "hermes" {
						message = "model not available"
					}
					if _, e := io.WriteString(stdout, message); e != nil {
						return e
					}
					return errors.New("exit status 1")
				}, &log)
				if err == nil || result.Response != "" {
					t.Fatalf("unexpected result=%#v error=%v", result, err)
				}
				if !strings.Contains(log.String(), "model not available") {
					t.Fatalf("lost provider diagnostic: %s", log.String())
				}
				if timedOut {
					assertIncidentKind(t, err, "timeout")
				} else {
					assertIncidentKind(t, err, "start-failed")
				}
			})
		}
	}
}

func TestW009ClaudeErrorEnvelopeCannotSucceed(t *testing.T) {
	for _, output := range []string{
		`{"result":"provider refused","is_error":true}`,
		"{\"type\":\"system\",\"subtype\":\"init\"}\n" + `{"type":"result","is_error":true,"errors":["provider refused"]}`,
	} {
		if result, err := ParseResult("claude", []byte(output), nil, "session"); err == nil || result.Response != "" {
			t.Fatalf("error accepted: %#v %v", result, err)
		}
	}
}

func TestW009OutputCaptureIsBounded(t *testing.T) {
	r := matrixRequest("claude", t.TempDir())
	_, err := Execute(context.Background(), r, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, _ io.Writer) error {
		_, e := io.WriteString(stdout, strings.Repeat("x", 16*1024*1024+1))
		return e
	}, io.Discard)
	assertIncidentKind(t, err, "output-limit")
}

func TestW009FailureDiagnosticsCoverPlainAndPrettyOutput(t *testing.T) {
	for _, payload := range []string{
		"startup refused\n",
		"{\n  \"is_error\": true,\n  \"result\": \"startup refused\"\n}\n",
		`{"type":"turn.failed","error":{"message":"startup refused"}}` + "\n",
	} {
		var log outputLog
		r := matrixRequest("claude", t.TempDir())
		_, err := Execute(context.Background(), r, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, _ io.Writer) error {
			if _, e := io.WriteString(out, payload); e != nil {
				return e
			}
			return errors.New("exit status 1")
		}, &log)
		if err == nil || !strings.Contains(log.String(), "startup refused") {
			t.Fatalf("error=%v output=%s", err, log.String())
		}
	}
}

func TestW009ProgressIsBoundedWithoutSuppressingErrors(t *testing.T) {
	var log outputLog
	out := newSessionOutput(matrixRequest("claude", t.TempDir()), false)
	out.progress = &log
	for range 250 {
		if _, err := io.WriteString(out, "{\"type\":\"assistant\"}\n"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := io.WriteString(out, `{"type":"error","message":"late failure\u001b[2J"}`+"\n"); err != nil {
		t.Fatal(err)
	}
	if strings.Count(log.String(), "provider event") != 200 || !strings.Contains(log.String(), "late failure") || strings.ContainsRune(log.String(), '\x1b') {
		t.Fatalf("unbounded, lost or unsafe diagnostics: %s", log.String())
	}
}

func TestW009StreamAndLegacyClaudeResults(t *testing.T) {
	for _, payload := range []string{
		`{"result":"final"}`,
		"{\"type\":\"system\",\"subtype\":\"init\"}\n{\"type\":\"assistant\",\"message\":{}}\n" + `{"type":"result","is_error":false,"result":"final"}`,
	} {
		result, err := ParseResult("claude", []byte(payload), nil, "retained")
		if err != nil || result.SessionID != "retained" || result.Response != "final" {
			t.Fatalf("result=%#v error=%v", result, err)
		}
	}
	if _, err := ParseResult("claude", []byte("{\"type\":\"result\",\"result\":\"one\"}\n{\"type\":\"result\",\"result\":\"two\"}\n"), nil, "retained"); err == nil {
		t.Fatal("duplicate final accepted")
	}
}
