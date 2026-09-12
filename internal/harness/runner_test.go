package harness

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestW004ExecuteUsesFixedVectorAndReturnsOnlyParsedResult(t *testing.T) {
	resultPath := filepath.Join(t.TempDir(), "result.txt")
	request := Request{Harness: "codex", Model: "model", Prompt: "prompt", Directory: t.TempDir(), ResultFile: resultPath}
	result, err := Execute(context.Background(), request, false, nil, func(_ context.Context, command string, args []string, directory string, stdin io.Reader, stdout, _ io.Writer) error {
		if command != "codex" || directory != request.Directory {
			t.Fatalf("invocation = %s in %s", command, directory)
		}
		contents, readErr := io.ReadAll(stdin)
		if readErr != nil || string(contents) != "prompt" {
			t.Fatalf("stdin = %q, %v", contents, readErr)
		}
		if err := os.WriteFile(resultPath, []byte("final"), 0o600); err != nil {
			return err
		}
		_, err := io.WriteString(stdout, "{\"type\":\"thread.started\",\"thread_id\":\"id\"}\n")
		return err
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if result != (Result{SessionID: "id", Response: "final"}) {
		t.Fatalf("result = %#v", result)
	}
}

func TestW004FixedStartAndResumeInvocations(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		start          []string
		resume         []string
		passesProvider bool
	}{
		{
			name:   "codex",
			start:  []string{"exec", "-m", "model", "-s", "read-only", "-C", "/project", "--skip-git-repo-check", "--json", "-o", "/result", "-"},
			resume: []string{"exec", "resume", "-m", "model", "--skip-git-repo-check", "--json", "-o", "/result", "session", "-"},
		},
		{
			name:   "claude",
			start:  []string{"-p", "--output-format", "stream-json", "--verbose", "--model", "model", "--session-id", "session", "--tools", "Read", "--permission-mode", "plan", "prompt"},
			resume: []string{"-p", "--output-format", "stream-json", "--verbose", "--model", "model", "--resume", "session", "--tools", "Read", "--permission-mode", "plan", "prompt"},
		},
		{
			name:   "copilot",
			start:  []string{"-p", "prompt", "-s", "--output-format", "json", "--model", "model", "--name", "session", "--available-tools="},
			resume: []string{"-p", "prompt", "-s", "--output-format", "json", "--model", "model", "--resume=session", "--available-tools="},
		},
		{
			name: "hermes", provider: "provider", passesProvider: true,
			start:  []string{"chat", "--quiet", "--query-file", "-", "-m", "model", "--provider", "provider", "-t", "none", "--safe-mode", "--in", "/project"},
			resume: []string{"chat", "--quiet", "--query-file", "-", "-m", "model", "--provider", "provider", "-t", "none", "--resume", "session", "--safe-mode", "--in", "/project"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := Request{Harness: test.name, Model: "model", Provider: test.provider, Prompt: "prompt", Directory: "/project", ResultFile: "/result", SessionID: "session"}
			start, err := BuildStart(request)
			if err != nil {
				t.Fatal(err)
			}
			resume, err := BuildResume(request)
			if err != nil {
				t.Fatal(err)
			}
			if start.Command != test.name || !reflect.DeepEqual(start.Args, test.start) {
				t.Fatalf("start = %s %#v, want %s %#v", start.Command, start.Args, test.name, test.start)
			}
			if !reflect.DeepEqual(resume.Args, test.resume) {
				t.Fatalf("resume = %#v, want %#v", resume.Args, test.resume)
			}
			joined := strings.Join(start.Args, " ")
			if strings.Contains(joined, "--provider") != test.passesProvider {
				t.Fatalf("provider argument mismatch: %s", joined)
			}
		})
	}
}

func TestW004SessionIdentityIsUUID(t *testing.T) {
	identity, err := NewSessionIdentity()
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(identity, "-")
	if len(parts) != 5 || len(identity) != 36 || parts[2][0] != '4' {
		t.Fatalf("session identity = %q, want UUID v4", identity)
	}
}

func TestW004ParseNativeFinalResponses(t *testing.T) {
	tests := []struct {
		name, stdout, resultFile, identity, response string
	}{
		{name: "codex", stdout: "{\"type\":\"thread.started\",\"thread_id\":\"codex-id\"}\n", resultFile: "final", identity: "codex-id", response: "final"},
		{name: "claude", stdout: `{"result":"final"}`, identity: "session", response: "final"},
		{name: "copilot", stdout: "{\"type\":\"assistant.message\",\"content\":\"final\"}\n", identity: "session", response: "final"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseResult(test.name, []byte(test.stdout), []byte(test.resultFile), "session")
			if err != nil {
				t.Fatal(err)
			}
			if result.SessionID != test.identity || result.Response != test.response {
				t.Fatalf("result = %#v", result)
			}
		})
	}
}

func TestW004RunnerRejectsInvalidOutcomes(t *testing.T) {
	for _, test := range []struct {
		name, stdout string
	}{
		{name: "codex", stdout: `{}`},
		{name: "claude", stdout: `{}`},
		{name: "copilot", stdout: `{}`},
	} {
		if _, err := ParseResult(test.name, []byte(test.stdout), nil, ""); err == nil {
			t.Errorf("%s malformed result succeeded", test.name)
		}
	}
	if _, err := BuildStart(Request{Harness: "missing", Model: "model"}); err == nil {
		t.Error("unsupported harness succeeded")
	} else {
		var incident *Incident
		if !errors.As(err, &incident) || incident.Kind != "capability-unsupported" {
			t.Fatalf("unsupported harness incident = %#v, %v", incident, err)
		}
	}
	if _, err := ParseTimeout("codex", 4*time.Minute); err == nil || !strings.Contains(err.Error(), "codex") {
		t.Fatalf("timeout incident = %v", err)
	}
}

func TestW004ExecuteReturnsTypedTimeoutAndExecutableIncidents(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	command := filepath.Join(bin, "codex")
	// #nosec G306 -- the fake harness must be executable.
	if err := os.WriteFile(command, []byte("#!/bin/sh\nsleep 10\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+"/bin:/usr/bin")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	request := Request{Harness: "codex", Model: "model", Prompt: "prompt", Directory: root, ResultFile: filepath.Join(root, "result")}
	if _, err := Execute(ctx, request, false, nil, nil, io.Discard); err == nil {
		t.Fatal("timed-out harness succeeded")
	} else {
		var incident *Incident
		if !errors.As(err, &incident) || incident.Kind != "timeout" {
			t.Fatalf("timeout incident = %#v, %v", incident, err)
		}
	}
	t.Setenv("PATH", bin)
	request.Harness = "claude"
	request.SessionID = "00000000-0000-4000-8000-000000000000"
	if _, err := Execute(context.Background(), request, false, nil, nil, io.Discard); err == nil {
		t.Fatal("missing executable succeeded")
	} else {
		var incident *Incident
		if !errors.As(err, &incident) || incident.Kind != "executable-unavailable" {
			t.Fatalf("executable incident = %#v, %v", incident, err)
		}
	}
}

func TestW004ExecuteReturnsTypedStartAndResumeFailures(t *testing.T) {
	for _, resume := range []bool{false, true} {
		request := Request{
			Harness: "claude", Model: "model", Prompt: "prompt", Directory: t.TempDir(),
			SessionID: "00000000-0000-4000-8000-000000000000",
		}
		_, err := Execute(context.Background(), request, resume, nil, func(context.Context, string, []string, string, io.Reader, io.Writer, io.Writer) error {
			return errors.New("provider exited unsuccessfully")
		}, io.Discard)
		if err == nil {
			t.Fatalf("resume=%t succeeded", resume)
		}
		var incident *Incident
		if !errors.As(err, &incident) {
			t.Fatalf("resume=%t error = %v", resume, err)
		}
		want := "start-failed"
		if resume {
			want = "resume-failed"
		}
		if incident.Kind != want || incident.SessionID != request.SessionID {
			t.Fatalf("resume=%t incident = %#v", resume, incident)
		}
	}
}

func TestW004ExecuteStartAndResumeAcrossHarnesses(t *testing.T) {
	for _, harnessName := range []string{"codex", "claude", "copilot", "hermes"} {
		for _, resume := range []bool{false, true} {
			t.Run(harnessName+map[bool]string{false: "-start", true: "-resume"}[resume], func(t *testing.T) {
				root := t.TempDir()
				request := matrixRequest(harnessName, root)
				result, err := Execute(context.Background(), request, resume, nil, successfulMatrixExecutor(t, request, resume), io.Discard)
				if err != nil {
					t.Fatal(err)
				}
				if result.SessionID == "" || result.Response != "final" {
					t.Fatalf("result = %#v", result)
				}
			})
		}
	}
}

func TestW004ExecuteFailureMatrixAcrossHarnesses(t *testing.T) {
	for _, harnessName := range []string{"codex", "claude", "copilot", "hermes"} {
		for _, resume := range []bool{false, true} {
			t.Run(harnessName+map[bool]string{false: "-start", true: "-resume"}[resume], func(t *testing.T) {
				request := matrixRequest(harnessName, t.TempDir())
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				<-ctx.Done()
				defer cancel()
				_, err := Execute(ctx, request, resume, nil, func(context.Context, string, []string, string, io.Reader, io.Writer, io.Writer) error {
					return context.DeadlineExceeded
				}, io.Discard)
				assertIncidentKind(t, err, "timeout")

				_, err = Execute(context.Background(), request, resume, nil, func(context.Context, string, []string, string, io.Reader, io.Writer, io.Writer) error {
					return errors.New("provider exited unsuccessfully")
				}, io.Discard)
				want := "start-failed"
				if resume {
					want = "resume-failed"
				}
				assertIncidentKind(t, err, want)

				_, err = Execute(context.Background(), request, resume, nil, emptyMatrixExecutor(t, request), io.Discard)
				assertIncidentKind(t, err, "response-empty")
			})
		}
	}
}

func TestW004HermesRequiresNativeSessionEnvelope(t *testing.T) {
	request := matrixRequest("hermes", t.TempDir())
	executor := func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, _ io.Writer) error {
		_, err := io.WriteString(stdout, "SESSION_ID: model-written-id\nfinal\n")
		return err
	}
	if _, err := Execute(context.Background(), request, false, nil, executor, io.Discard); err == nil {
		t.Fatal("Hermes response without native session envelope succeeded")
	} else {
		assertIncidentKind(t, err, "identity-missing")
	}
}

func matrixRequest(harnessName, root string) Request {
	request := Request{
		Harness: harnessName, Model: "model", Prompt: "prompt", Directory: root,
		SessionID: "native-session",
	}
	if harnessName == "codex" {
		request.ResultFile = filepath.Join(root, "result")
	}
	if harnessName == "hermes" {
		request.Provider = "provider"
	}
	return request
}

func successfulMatrixExecutor(t *testing.T, request Request, resume bool) Executor {
	t.Helper()
	return func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, stderr io.Writer) error {
		switch request.Harness {
		case "codex":
			if err := os.WriteFile(request.ResultFile, []byte("final"), 0o600); err != nil {
				return err
			}
			if !resume {
				_, _ = io.WriteString(stdout, "{\"type\":\"thread.started\",\"thread_id\":\"native-session\"}\n")
			}
		case "claude":
			_, _ = io.WriteString(stdout, `{"result":"final"}`)
		case "copilot":
			_, _ = io.WriteString(stdout, "{\"type\":\"assistant.message\",\"content\":\"final\"}\n")
		case "hermes":
			_, _ = io.WriteString(stderr, "session_id: native-session\n")
			_, _ = io.WriteString(stdout, "final\n")
		}
		return nil
	}
}

func emptyMatrixExecutor(t *testing.T, request Request) Executor {
	t.Helper()
	return func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, stderr io.Writer) error {
		switch request.Harness {
		case "codex":
			if err := os.WriteFile(request.ResultFile, nil, 0o600); err != nil {
				return err
			}
			_, _ = io.WriteString(stdout, "{\"type\":\"thread.started\",\"thread_id\":\"native-session\"}\n")
		case "claude":
			_, _ = io.WriteString(stdout, `{}`)
		case "copilot":
			_, _ = io.WriteString(stdout, `{}`)
		case "hermes":
			_, _ = io.WriteString(stderr, "session_id: native-session\n")
		}
		return nil
	}
}

func assertIncidentKind(t *testing.T, err error, want string) {
	t.Helper()
	var incident *Incident
	if !errors.As(err, &incident) || incident.Kind != want {
		t.Fatalf("incident = %#v, %v; want %s", incident, err, want)
	}
}

func TestW004CompositeVerdictValidation(t *testing.T) {
	for _, valid := range []string{
		"GATE: implementation\nREVISION: abc123\nVERDICT: PASS\n",
		"GATE: implementation\nREVISION: abc123\nVERDICT: PROVISIONAL PASS\n",
	} {
		if err := ValidateCompositeVerdict(valid); err != nil {
			t.Fatal(err)
		}
	}
	for _, invalid := range []string{
		"VERDICT: PASS\n",
		"GATE: implementation\nREVISION: abc123\nVERDICT: MAYBE\n",
		"GATE: implementation\nREVISION: abc123\nVERDICT: PROVISIONAL\n",
		"GATE: implementation\nREVISION: abc123\nVERDICT: PASS\nVERDICT: FAIL\n",
	} {
		if err := ValidateCompositeVerdict(invalid); err == nil {
			t.Fatalf("malformed verdict succeeded: %q", invalid)
		}
	}
}
