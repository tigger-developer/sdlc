package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestW004HelpAndVersionExitSuccessfully(t *testing.T) {
	for _, arguments := range [][]string{{"-h"}, {"--help"}, {"start", "-h"}, {"--version"}} {
		var output bytes.Buffer
		var diagnostics bytes.Buffer
		if err := run(arguments, strings.NewReader(""), &output, &diagnostics); err != nil {
			t.Fatalf("%v returned %v: %s", arguments, err, diagnostics.String())
		}
		combined := output.String() + diagnostics.String()
		if arguments[0] != "--version" && !strings.Contains(combined, "Internal SDLC helper") {
			t.Fatalf("%v help lacks command purpose: %s", arguments, combined)
		}
	}
}

func TestW004StartAndResumeThroughInternalCLI(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
output=
previous=
for argument in "$@"; do
    if [ "$previous" = "-o" ]; then output="$argument"; fi
    previous="$argument"
done
printf 'GATE: implementation\nREVISION: abc123\nVERDICT: PASS\n' > "$output"
printf '{"type":"thread.started","thread_id":"native-session"}\n'
`
	command := filepath.Join(bin, "codex")
	// #nosec G306 -- the fake harness must be executable.
	if err := os.WriteFile(command, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(filepath.Join(project, ".sdlc"), 0o700); err != nil {
		t.Fatal(err)
	}
	config := "version: 3\ndelivery:\n  audit:\n    harness: codex\n    model: test-model\n    timeout: 5s\n"
	if err := os.WriteFile(filepath.Join(project, ".sdlc", "project.yaml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	evidence := filepath.Join(project, "evidence.org")
	if err := os.WriteFile(evidence, []byte("evidence\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{
		{"start", "--project", project, "--global-config", filepath.Join(root, "absent.yaml"), "--input", evidence},
		{"resume", "--project", project, "--global-config", filepath.Join(root, "absent.yaml"), "--input", evidence, "--session", "native-session"},
	} {
		var output bytes.Buffer
		var diagnostics bytes.Buffer
		if err := run(arguments, strings.NewReader("audit this"), &output, &diagnostics); err != nil {
			t.Fatalf("%v returned %v: %s", arguments, err, diagnostics.String())
		}
		if !strings.Contains(output.String(), "VERDICT: PASS") || !strings.Contains(diagnostics.String(), "SESSION_ID: native-session") {
			t.Fatalf("%v output = %q; diagnostics = %q", arguments, output.String(), diagnostics.String())
		}
	}
}

func TestW004InternalCLIRejectsIncompleteRequests(t *testing.T) {
	for _, arguments := range [][]string{{}, {"unknown"}, {"start", "extra"}, {"start"}} {
		if err := run(arguments, strings.NewReader("prompt"), &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
			t.Fatalf("%v succeeded", arguments)
		}
	}
}

func TestW004StartRejectsSuppliedSession(t *testing.T) {
	err := run([]string{"start", "--session", "agent-owned-id"}, strings.NewReader("prompt"), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "start must not receive --session") {
		t.Fatalf("start with supplied session error = %v", err)
	}
}

func TestW004UppercaseAuditPhaseStillValidatesVerdict(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
output=
previous=
for argument in "$@"; do
    if [ "$previous" = "-o" ]; then output="$argument"; fi
    previous="$argument"
done
printf 'not an audit verdict\n' > "$output"
printf '{"type":"thread.started","thread_id":"native-session"}\n'
`
	command := filepath.Join(bin, "codex")
	// #nosec G306 -- the fake harness must be executable.
	if err := os.WriteFile(command, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(filepath.Join(project, ".sdlc"), 0o700); err != nil {
		t.Fatal(err)
	}
	config := "version: 3\ndelivery:\n  audit:\n    harness: codex\n    model: test-model\n    timeout: 5s\n"
	if err := os.WriteFile(filepath.Join(project, ".sdlc", "project.yaml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	evidence := filepath.Join(project, "evidence.org")
	if err := os.WriteFile(evidence, []byte("evidence\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	arguments := []string{"start", "--phase", "AUDIT", "--project", project, "--global-config", filepath.Join(root, "absent.yaml"), "--input", evidence}
	if err := run(arguments, strings.NewReader("audit this"), &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "composite verdict") {
		t.Fatalf("uppercase audit phase verdict validation = %v", err)
	}
}
