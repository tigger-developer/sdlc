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
	request := Request{Harness: "codex", Model: "model", Prompt: "prompt", Bundle: t.TempDir(), ResultFile: resultPath}
	result, err := Execute(context.Background(), request, false, nil, func(_ context.Context, command string, args []string, directory string, stdin io.Reader, stdout, _ io.Writer) error {
		if command != "codex" || directory != request.Bundle {
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
			start:  []string{"exec", "-m", "model", "-s", "read-only", "-C", "/bundle", "--json", "-o", "/result", "-"},
			resume: []string{"exec", "resume", "-m", "model", "--json", "-o", "/result", "session", "-"},
		},
		{
			name:   "claude",
			start:  []string{"-p", "--output-format", "json", "--model", "model", "--session-id", "session", "--tools", "", "--permission-mode", "plan", "prompt"},
			resume: []string{"-p", "--output-format", "json", "--model", "model", "--resume", "session", "--tools", "", "--permission-mode", "plan", "prompt"},
		},
		{
			name:   "copilot",
			start:  []string{"-p", "prompt", "-s", "--output-format", "json", "--model", "model", "--name", "session", "--available-tools="},
			resume: []string{"-p", "prompt", "-s", "--output-format", "json", "--model", "model", "--resume=session", "--available-tools="},
		},
		{
			name: "hermes", provider: "provider", passesProvider: true,
			start:  []string{"-z", hermesPrompt("prompt"), "-m", "model", "--provider", "provider", "-t", "", "--pass-session-id", "--safe-mode", "--in", "/bundle"},
			resume: []string{"-z", hermesPrompt("prompt"), "-m", "model", "--provider", "provider", "-t", "", "--resume", "session", "--safe-mode", "--in", "/bundle"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := Request{Harness: test.name, Model: "model", Provider: test.provider, Prompt: "prompt", Bundle: "/bundle", ResultFile: "/result", SessionID: "session"}
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
		{name: "hermes", stdout: "SESSION_ID: hermes-id\nfinal\n", identity: "hermes-id", response: "final"},
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
		{name: "hermes", stdout: "final"},
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
	request := Request{Harness: "codex", Model: "model", Prompt: "prompt", Bundle: root, ResultFile: filepath.Join(root, "result")}
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

func TestW004ImmutableBundleDetectsMutation(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.org")
	if err := os.WriteFile(source, []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	bundle, err := CreateBundle(filepath.Join(root, "bundles"), []Input{{Name: "spec.org", Source: source}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := bundle.Cleanup(); err != nil {
			t.Error(err)
		}
	}()
	if err := bundle.Verify(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(bundle.Path, "spec.org")
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := bundle.Verify(); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("mutation verification = %v", err)
	}
}

func TestW004ImmutableBundleDetectsSourceMutation(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.org")
	if err := os.WriteFile(source, []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	bundle, err := CreateBundle(filepath.Join(root, "bundles"), []Input{{Name: "spec.org", Source: source}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := bundle.Cleanup(); err != nil {
			t.Error(err)
		}
	}()
	if err := os.WriteFile(source, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := bundle.Verify(); err == nil || !strings.Contains(err.Error(), "audit source") {
		t.Fatalf("source mutation verification = %v", err)
	}
}

func TestW004CompositeVerdictValidation(t *testing.T) {
	valid := "GATE: implementation\nREVISION: abc123\nVERDICT: PASS\n"
	if err := ValidateCompositeVerdict(valid); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{
		"VERDICT: PASS\n",
		"GATE: implementation\nREVISION: abc123\nVERDICT: MAYBE\n",
		"GATE: implementation\nREVISION: abc123\nVERDICT: PASS\nVERDICT: FAIL\n",
	} {
		if err := ValidateCompositeVerdict(invalid); err == nil {
			t.Fatalf("malformed verdict succeeded: %q", invalid)
		}
	}
}

func TestW004IncompleteBundleIsRemoved(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "bundles")
	_, err := CreateBundle(parent, []Input{{Name: "../invalid", Source: "unused"}})
	if err == nil {
		t.Fatal("invalid bundle input succeeded")
	}
	entries, readErr := os.ReadDir(parent)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("incomplete bundle was retained: %v", entries)
	}
}

func TestW004BundleManifestAndDuplicateRejection(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.org")
	if err := os.WriteFile(source, []byte("evidence\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	bundle, err := CreateBundle(filepath.Join(root, "bundles"), []Input{{Name: "spec.org", Source: source}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := bundle.Cleanup(); err != nil {
			t.Error(err)
		}
	}()
	manifest, err := os.ReadFile(filepath.Join(bundle.Path, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{source, `"destination": "spec.org"`, `"bytes": 9`, `"sha256"`} {
		if !strings.Contains(string(manifest), expected) {
			t.Fatalf("manifest lacks %q:\n%s", expected, manifest)
		}
	}
	if _, err := CreateBundle(filepath.Join(root, "duplicates"), []Input{{Name: "spec.org", Source: source}, {Name: "spec.org", Source: source}}); err == nil {
		t.Fatal("duplicate bundle destination succeeded")
	}
}
