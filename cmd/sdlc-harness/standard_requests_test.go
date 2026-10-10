package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestStandardsRequestUsesCapturedInventory(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	standard := filepath.Join(root, "ORG-SCHEMA.md")
	core := filepath.Join(root, "CODING.md")
	if err := os.WriteFile(core, []byte("CORE_CODING_RULE\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(standard, []byte("SCHEMA_RULE_MARKER\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	spec := filepath.Join(project, "spec.org")
	if err := os.WriteFile(spec, []byte("PROJECT_SPEC_MARKER\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	evidence, err := harness.CaptureEvidence(project, []string{spec, standard, core})
	if err != nil {
		t.Fatal(err)
	}
	runner := recoveryRun{request: harness.Request{Prompt: "Audit."}, evidence: evidence, standardsRoot: root, requiredStandards: []string{core}}
	prepared, err := runner.prepare(harness.AuditEntry{}, harness.Config{Harness: "hermes"}, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prepared.Prompt, "PROJECT_SPEC_MARKER") || strings.Contains(prepared.Prompt, "SCHEMA_RULE_MARKER") {
		t.Fatal("initial prompt failed to separate project content from available standards")
	}
	seen := map[string]bool{}
	paths, requested, err := runner.requestedStandardPaths(`STANDARDS-REQUEST: ["ORG-SCHEMA.md"]`, seen)
	if err != nil || !requested || len(paths) != 1 || paths[0] != standard {
		t.Fatalf("request = %v, %t, %v", paths, requested, err)
	}
	contents, err := evidence.ContentPromptFor(paths)
	if err != nil || !strings.Contains(contents, "SCHEMA_RULE_MARKER") {
		t.Fatalf("requested contents = %q, %v", contents, err)
	}
	seen["ORG-SCHEMA.md"] = true
	for _, invalid := range []string{
		`STANDARDS-REQUEST: ["ORG-SCHEMA.md"]`,
		`STANDARDS-REQUEST: ["../spec.org"]`,
		`STANDARDS-REQUEST: []`,
		`STANDARDS-REQUEST: ["MAIN.md"]`,
	} {
		if _, _, err := runner.requestedStandardPaths(invalid, seen); err == nil {
			t.Fatalf("accepted invalid standards request %q", invalid)
		}
	}
}

func TestToolsDisabledAuditFetchesStandardBeforeVerdict(t *testing.T) {
	project := t.TempDir()
	root := t.TempDir()
	bin := t.TempDir()
	standard := filepath.Join(root, "ORG-SCHEMA.md")
	core := filepath.Join(root, "CODING.md")
	spec := filepath.Join(project, "spec.org")
	for path, body := range map[string]string{standard: "SCHEMA_RULE_MARKER\n", core: "CORE_CODING_RULE\n", spec: "PROJECT_SPEC_MARKER\n"} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	script := `#!/bin/sh
prompt=$(/bin/cat)
case "$prompt" in
  *SCHEMA_RULE_MARKER*)
    printf '%s' "$prompt" > "$FETCHED_PROMPT"
    printf 'session_id: standards-session\n' >&2
    printf 'AUDIT: standalone\nREVISION: fixture\nVERDICT: PASS\n'
    ;;
  *)
    printf '%s' "$prompt" > "$INITIAL_PROMPT"
    printf 'session_id: standards-session\n' >&2
    printf 'STANDARDS-REQUEST: ["ORG-SCHEMA.md"]\n'
    ;;
esac
`
	// #nosec G306 -- private executable fixture never invokes a hosted provider.
	if err := os.WriteFile(filepath.Join(bin, "hermes"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("INITIAL_PROMPT", filepath.Join(project, "initial-prompt"))
	t.Setenv("FETCHED_PROMPT", filepath.Join(project, "fetched-prompt"))
	evidence, err := harness.CaptureEvidence(project, []string{spec, standard, core})
	if err != nil {
		t.Fatal(err)
	}
	config := harness.Config{Harness: "hermes", Provider: "fixture", Model: "fixture", Timeout: 5 * time.Second}
	runner := recoveryRun{request: harness.Request{Prompt: "Audit.", Directory: project}, evidence: evidence, standardsRoot: root, requiredStandards: []string{core}, diagnostics: os.Stderr}
	request, err := runner.prepare(harness.AuditEntry{}, config, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := prepareHermesSDKFixture(); err != nil {
		t.Fatal(err)
	}
	result, err := runner.invoke(context.Background(), harness.AuditEntry{}, config, request, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Response, "VERDICT: PASS") {
		t.Fatalf("final result = %q", result.Response)
	}
	initial, err := os.ReadFile(filepath.Join(project, "initial-prompt"))
	if err != nil {
		t.Fatal(err)
	}
	fetched, err := os.ReadFile(filepath.Join(project, "fetched-prompt"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(initial), "SCHEMA_RULE_MARKER") || !strings.Contains(string(fetched), "SCHEMA_RULE_MARKER") {
		t.Fatal("standard was preloaded or absent from the requested response")
	}
	if err := evidence.Verify(); err != nil {
		t.Fatal(err)
	}
}
