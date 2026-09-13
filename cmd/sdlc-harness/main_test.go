package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/harness"
)

// This fake records the actual process boundary; no metered provider is invoked.
func TestAuditEvidenceUsesOriginalPathsAndRetainsHashesOnResume(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
output=
previous=
for argument in "$@"; do
    if [ "$previous" = "-o" ]; then output="$argument"; fi
    previous="$argument"
done
pwd > "$PROBE_DIRECTORY"
cat > "$PROBE_PROMPT"
printf '%s' "$output" > "$PROBE_RESULT_PATH"
if [ "${PROBE_MUTATE:-}" = 1 ]; then printf 'mutated' > "$PROBE_SOURCE"; fi
printf 'GATE: implementation\nREVISION: candidate\nVERDICT: FAIL\nFix the described defect.\n' > "$output"
printf '{"type":"thread.started","thread_id":"native-session"}\n'
`
	// #nosec G306 -- this local fake is an executable, not a metered provider.
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PROBE_DIRECTORY", filepath.Join(root, "directory"))
	t.Setenv("PROBE_PROMPT", filepath.Join(root, "prompt"))
	t.Setenv("PROBE_RESULT_PATH", filepath.Join(root, "result-path"))
	t.Setenv("PROBE_MUTATE", "")
	source := filepath.Join(root, "spec.org")
	t.Setenv("PROBE_SOURCE", source)
	if err := os.WriteFile(source, []byte("original requirement"), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := filepath.Join(root, "prompts.yaml")
	if err := os.WriteFile(registry, []byte("version: 1\ngates:\n  implementation:\n    prompt: audit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(root, "audits.yaml")
	args := []string{"--phase", "audit", "--gate", "implementation", "--harness", "codex", "--model", "fixture", "--project", root, "--global-config", filepath.Join(root, "absent.yaml"), "--audit-prompts", registry, "--audit-record", record, "--work-item", "W005-evidence", "--input", source}
	for _, action := range []string{"start", "resume"} {
		if action == "resume" {
			if err := os.WriteFile(source, []byte("corrected requirement"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		if err := run(append([]string{action}, args...), strings.NewReader("review"), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		directory, err := os.ReadFile(filepath.Join(root, "directory"))
		if err != nil {
			t.Fatal(err)
		}
		// Resolve platform aliases such as /var -> /private/var before comparing.
		wantDirectory, err := filepath.EvalSymlinks(root)
		if err != nil {
			t.Fatal(err)
		}
		gotDirectory, err := filepath.EvalSymlinks(strings.TrimSpace(string(directory)))
		if err != nil || gotDirectory != wantDirectory {
			t.Fatalf("provider cwd = %s, want project %s (%v)", directory, root, err)
		}
		entry, found, err := harness.ReadAuditEntry(record, "W005-evidence", "implementation")
		if err != nil || !found {
			t.Fatalf("record = %#v, %v", entry, err)
		}
		manifest := entry.LatestEvidence()
		if len(manifest) != 1 || manifest[0].Path != source || len(manifest[0].SHA256) != 64 {
			t.Fatalf("audit evidence manifest = %#v", manifest)
		}
		if action == "resume" && (len(entry.History) != 2 || entry.History[0].Evidence[0].SHA256 == manifest[0].SHA256) {
			t.Fatal("resume did not retain distinct per-round evidence hashes")
		}
		prompt, err := os.ReadFile(filepath.Join(root, "prompt"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(prompt, []byte(source)) || bytes.Contains(prompt, []byte("corrected requirement")) {
			t.Fatalf("expected paths, not copied contents: %s", prompt)
		}
		change := "added"
		if action == "resume" {
			change = "changed"
		}
		if !bytes.Contains(prompt, []byte(`"change": "`+change+`"`)) {
			t.Fatalf("missing %s evidence delta: %s", change, prompt)
		}
		if entry.SessionID != "native-session" {
			t.Fatalf("session = %q", entry.SessionID)
		}
		assertResultRemoved(t, root)
	}
	before, _, err := harness.ReadAuditEntry(record, "W005-evidence", "implementation")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PROBE_MUTATE", "1")
	err = run(append([]string{"resume"}, args...), strings.NewReader("review with mutation probe"), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("mutation error = %v", err)
	}
	after, _, err := harness.ReadAuditEntry(record, "W005-evidence", "implementation")
	if err != nil || len(after.History) != len(before.History)+1 || !reflect.DeepEqual(before.History, after.History[:len(before.History)]) || after.Verdict != "" || after.History[len(after.History)-1].Incident == "" {
		t.Fatalf("invalid response lost prior history or gained a verdict: %#v (%v)", after, err)
	}
	assertResultRemoved(t, root)
}

func assertResultRemoved(t *testing.T, root string) {
	t.Helper()
	path, err := os.ReadFile(filepath.Join(root, "result-path"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(string(path)); !os.IsNotExist(err) {
		t.Fatalf("result file retained: %s (%v)", path, err)
	}
}

func TestW004HelpAndVersionExitSuccessfully(t *testing.T) {
	for _, arguments := range [][]string{{"-h"}, {"--help"}, {"start", "-h"}, {"--version"}} {
		var output bytes.Buffer
		var diagnostics bytes.Buffer
		if err := run(arguments, strings.NewReader(""), &output, &diagnostics); err != nil {
			t.Fatalf("%v returned %v: %s", arguments, err, diagnostics.String())
		}
		combined := output.String() + diagnostics.String()
		if strings.TrimSpace(combined) == "" {
			t.Fatalf("%v returned no help or version output", arguments)
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
	prompts := filepath.Join(project, "audits.yaml")
	if err := os.WriteFile(prompts, []byte("version: 1\ngates:\n  implementation:\n    prompt: audit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".sdlc", "project.yaml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	evidence := filepath.Join(project, "evidence.org")
	if err := os.WriteFile(evidence, []byte("evidence\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{
		{"start", "--phase", "audit", "--gate", "implementation", "--audit-prompts", prompts, "--project", project, "--global-config", filepath.Join(root, "absent.yaml"), "--input", evidence},
		{"resume", "--phase", "audit", "--gate", "implementation", "--audit-prompts", prompts, "--project", project, "--global-config", filepath.Join(root, "absent.yaml"), "--input", evidence, "--session", "native-session"},
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
	prompts := filepath.Join(project, "audits.yaml")
	if err := os.WriteFile(prompts, []byte("version: 1\ngates:\n  definition:\n    prompt: audit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	arguments := []string{"start", "--phase", "AUDIT", "--gate", "definition", "--audit-prompts", prompts, "--project", project, "--global-config", filepath.Join(root, "absent.yaml"), "--input", evidence}
	if err := run(arguments, strings.NewReader("audit this"), &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "composite verdict") {
		t.Fatalf("uppercase audit phase verdict validation = %v", err)
	}
}
