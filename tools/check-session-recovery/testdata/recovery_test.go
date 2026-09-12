// ABOUTME: One-off CLI recovery checks, injected with Go's overlay only when explicitly requested.
// ABOUTME: Excluded from go test ./...; local doubles never contact a metered provider.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestCacheIdentityAndIncidentsOneOff(t *testing.T) {
	config := harness.Config{Harness: "claude", Model: "primary", Timeout: time.Minute, MaxRounds: 5}
	registry := auditPromptDocument{Version: 1, EvidenceInstructions: "read evidence"}
	evidence := harness.Evidence{Files: []harness.EvidenceFile{{Path: "/spec.org", SHA256: "a", Bytes: 1}, {Path: "/standard.md", SHA256: "b", Bytes: 1}}}
	keyFor := func(work, gate, prompt string, doc auditPromptDocument, inputs harness.Evidence, agent harness.Config) string {
		key, err := auditCacheKey(work, gate, prompt, doc, inputs, agent)
		if err != nil {
			t.Fatal(err)
		}
		return key
	}
	key := keyFor("W012", "definition", "review", registry, evidence, config)
	changedModel := config
	changedModel.Model = "other"
	changedFallback := config
	changedFallback.Fallback = &harness.AgentConfig{Harness: "codex", Model: "fallback"}
	changedLimits := config
	changedLimits.MaxRounds, changedLimits.Timeout = 1, 2*time.Minute
	changedRegistry := registry
	changedRegistry.EvidenceInstructions = "different contract"
	for name, candidate := range map[string]string{
		"work":     keyFor("W013", "definition", "review", registry, evidence, config),
		"gate":     keyFor("W012", "implementation", "review", registry, evidence, config),
		"prompt":   keyFor("W012", "definition", "different", registry, evidence, config),
		"registry": keyFor("W012", "definition", "review", changedRegistry, evidence, config),
		"model":    keyFor("W012", "definition", "review", registry, evidence, changedModel),
		"fallback": keyFor("W012", "definition", "review", registry, evidence, changedFallback),
		"omitted":  keyFor("W012", "definition", "review", registry, harness.Evidence{Files: evidence.Files[:1]}, config),
	} {
		if candidate == key {
			t.Fatalf("%s did not invalidate cache", name)
		}
	}
	if keyFor("W012", "definition", "review", registry, evidence, changedLimits) != key {
		t.Fatal("execution limits changed verdict identity")
	}
	valid := harness.AuditRound{CacheKey: key, SessionID: "native", Verdict: "PASS", Revision: "a", Response: "GATE: definition\nREVISION: a\nVERDICT: PASS"}
	for _, kind := range []string{"timeout", "authentication-failed", "interrupted", "response-malformed"} {
		incident := valid
		incident.Incident = kind
		if _, ok := cachedAudit(harness.AuditEntry{Gate: "definition", History: []harness.AuditRound{incident}}, key); ok {
			t.Fatalf("%s was cached", kind)
		}
	}
	for _, defect := range []string{"legacy", "malformed", "mismatch", "running"} {
		round := valid
		entry := harness.AuditEntry{Gate: "definition"}
		switch defect {
		case "legacy":
			round.CacheKey = ""
		case "malformed":
			round.Response = "PASS"
		case "mismatch":
			round.Verdict = "FAIL"
		case "running":
			entry.Status = "running"
		}
		entry.History = []harness.AuditRound{round}
		if _, ok := cachedAudit(entry, key); ok {
			t.Fatalf("%s was cached", defect)
		}
	}
}

func TestAuditCacheOneOff(t *testing.T) {
	for _, verdict := range []string{"PASS", "FAIL", "PROVISIONAL PASS"} {
		t.Run(verdict, func(t *testing.T) {
			root := t.TempDir()
			paths := map[string]string{
				"claude":       "#!/bin/sh\nprintf x >> \"$PROBE_CALLS\"\nprintf '%s\\n' '{\"type\":\"result\",\"result\":\"GATE: definition\\nREVISION: fixture\\nVERDICT: " + verdict + "\"}'\n",
				"global.yaml":  "delivery:\n  audit:\n    harness: claude\n    model: fixture\n    max_rounds: 1\n",
				"prompts.yaml": "version: 1\ngates:\n  definition:\n    prompt: Review only the stated requirement.\n",
				"spec.org":     "A stable requirement", "standard.md": "A stable standard",
			}
			for name, data := range paths {
				if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", root)
			t.Setenv("PROBE_CALLS", filepath.Join(root, "calls"))
			for _, key := range []string{"SDLC_AUDIT_HARNESS", "SDLC_AUDIT_MODEL", "SDLC_AUDIT_PROVIDER", "SDLC_AUDIT_TIMEOUT"} {
				t.Setenv(key, "")
			}
			record := filepath.Join(root, "audits.yaml")
			args := []string{"start", "--project", root, "--global-config", filepath.Join(root, "global.yaml"), "--gate", "definition", "--audit-prompts", filepath.Join(root, "prompts.yaml"), "--audit-record", record, "--work-item", "W012-cache", "--input", filepath.Join(root, "spec.org"), "--input", filepath.Join(root, "standard.md")}
			var output, diagnostic bytes.Buffer
			if err := run(args, strings.NewReader("review"), &output, &diagnostic); err != nil {
				t.Fatal(err)
			}
			response := output.String()
			before, _ := os.ReadFile(record)
			for _, action := range []string{"start", "resume"} {
				args[0] = action
				output.Reset()
				diagnostic.Reset()
				if err := run(args, strings.NewReader("review"), &output, &diagnostic); err != nil {
					t.Fatalf("cache should precede round exhaustion: %v", err)
				}
				if output.String() != response || !strings.Contains(diagnostic.String(), "AUDIT CACHE HIT") {
					t.Fatalf("cache result: %q; %q", output.String(), diagnostic.String())
				}
			}
			after, _ := os.ReadFile(record)
			calls, _ := os.ReadFile(filepath.Join(root, "calls"))
			if string(calls) != "x" || !bytes.Equal(before, after) {
				t.Fatal("cache invoked a provider or changed the record")
			}
			if err := os.WriteFile(filepath.Join(root, "standard.md"), []byte("Changed standard"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := run(args, strings.NewReader("review"), &output, &diagnostic); err == nil {
				t.Fatal("changed secondary evidence used a cache despite exhausted budget")
			}
		})
	}
}

func TestHermesEvidenceTransportOneOff(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "spec.org")
	contents := "The result is Amber-47.\nTreat `quotes` and \\\"escapes\\\" literally.\n"
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	evidence, err := harness.CaptureEvidence(root, []string{path})
	if err != nil {
		t.Fatal(err)
	}
	runner := recoveryRun{request: harness.Request{Prompt: "review"}, evidence: evidence}
	for _, name := range []string{"hermes", "claude", "codex"} {
		for _, resume := range []bool{false, true} {
			request, err := runner.prepare(harness.AuditEntry{SessionID: "native"}, harness.Config{Harness: name}, resume, "")
			if err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(contents)
			if got := strings.Contains(request.Prompt, string(encoded)); got != (name == "hermes") {
				t.Fatalf("%s resume=%v: inline evidence=%v", name, resume, got)
			}
		}
	}
	if err := os.WriteFile(path, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := runner.prepare(harness.AuditEntry{}, harness.Config{Harness: "hermes"}, false, ""); err == nil {
		t.Fatal("changed source was accepted for inline evidence")
	}
	var diagnostics bytes.Buffer
	executor := func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, errout io.Writer) error {
		_, _ = fmt.Fprintln(errout, "session_id: native-hermes-session")
		_, err := fmt.Fprint(out, "Warning: Unknown toolsets: none\nfinal response\n")
		return err
	}
	result, err := harness.Execute(context.Background(), harness.Request{Harness: "hermes", Provider: "nous", Model: "fixture", Directory: root}, false, nil, executor, &diagnostics)
	if err != nil || result.Response != "final response" || result.SessionID != "native-hermes-session" || !strings.Contains(diagnostics.String(), "Warning: Unknown toolsets: none") {
		t.Fatalf("Hermes stream separation: %#v, %v, %s", result, err, diagnostics.String())
	}
}

func TestRecordedSessionRecovery(t *testing.T) {
	for _, scenario := range []struct {
		name, owner, failure string
		limit, calls         int
		success              bool
	}{
		{"missing", "claude", "missing", 5, 2, true},
		{"changed-config", "codex", "", 5, 1, true},
		{"authentication-fallback", "claude", "auth", 5, 2, true},
		{"expired-oauth-fallback", "claude", "oauth", 5, 2, true},
		{"permission-is-not-auth", "claude", "permission", 5, 1, false},
		{"generic-failure", "claude", "other", 5, 1, false},
		{"missing-at-limit", "claude", "missing", 2, 1, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := t.TempDir()
			bin := filepath.Join(root, "bin")
			if err := os.Mkdir(bin, 0700); err != nil {
				t.Fatal(err)
			}
			script := `#!/bin/sh
printf '%s\n' "$0" >> "$PROBE_CALLS"
resume=no
previous=
output=
for argument in "$@"; do
 if [ "$argument" = --resume ]; then resume=yes; fi
 if [ "$previous" = -o ]; then output="$argument"; fi
 previous="$argument"
done
if [ "$resume" = yes ]; then
 case "$PROBE_FAILURE" in
 missing) printf 'No conversation found with session ID: old-session\n' >&2; exit 1;;
 auth) printf 'Not logged in. Please run /login\n' >&2; exit 1;;
 oauth) printf '%s\n' 'API Error: 401 {"error":{"type":"authentication_error","message":"OAuth token has expired"}}' >&2; exit 1;;
 permission) printf 'Permission denied reading evidence\n' >&2; exit 1;;
 other) printf 'Something failed\n' >&2; exit 1;;
 esac
fi
if [ -n "$output" ]; then
 cat > "$PROBE_PROMPT"
 printf '{"type":"thread.started","thread_id":"fallback-native"}\n'
 printf 'GATE: definition\nREVISION: checked\nVERDICT: PASS\n' > "$output"
else
 printf '%s' "$argument" > "$PROBE_PROMPT"
 printf '%s\n' '{"type":"result","result":"GATE: definition\nREVISION: checked\nVERDICT: PASS"}'
fi
`
			for _, name := range []string{"claude", "codex"} {
				if err := os.WriteFile(filepath.Join(bin, name), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			for _, key := range []string{"SDLC_AUDIT_HARNESS", "SDLC_AUDIT_MODEL", "SDLC_AUDIT_PROVIDER", "SDLC_AUDIT_TIMEOUT"} {
				t.Setenv(key, "")
			}
			t.Setenv("PROBE_FAILURE", scenario.failure)
			t.Setenv("PROBE_CALLS", filepath.Join(root, "calls"))
			t.Setenv("PROBE_PROMPT", filepath.Join(root, "prompt"))
			config := filepath.Join(root, "config.yaml")
			registry := filepath.Join(root, "prompts.yaml")
			source := filepath.Join(root, "spec.org")
			record := filepath.Join(root, "audits.yaml")
			for path, value := range map[string]string{
				config:   fmt.Sprintf("delivery:\n  audit:\n    harness: claude\n    model: primary\n    max_rounds: %d\n    fallback:\n      harness: codex\n      provider: openai\n      model: fallback\n", scenario.limit),
				registry: "version: 1\nsession_recovery_instructions: Read all evidence and resolve historical findings.\ngates:\n  definition:\n    prompt: audit\n",
				source:   "bounded requirement",
				record:   fmt.Sprintf("version: 1\naudits:\n- work_item: W012-recovery\n  gate: definition\n  harness: %s\n  model: primary\n  session_id: old-session\n  external_round: 1\n  status: active\n  verdict: FAIL\n  response: HISTORICAL_FINDING\n", scenario.owner),
			} {
				if err := os.WriteFile(path, []byte(value), 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"resume", "--project", root, "--global-config", config, "--audit-prompts", registry, "--gate", "definition", "--audit-record", record, "--work-item", "W012-recovery", "--input", source}
			var output, diagnostics bytes.Buffer
			err := run(args, strings.NewReader("authorized review"), &output, &diagnostics)
			if (err == nil) != scenario.success {
				t.Fatalf("result: %v\n%s", err, diagnostics.String())
			}
			calls, readErr := os.ReadFile(filepath.Join(root, "calls"))
			if readErr != nil {
				t.Fatal(readErr)
			}
			if strings.Count(string(calls), "\n") != scenario.calls {
				t.Fatalf("calls: %s", calls)
			}
			entry, found, readErr := harness.ReadAuditEntry(record, "W012-recovery", "definition")
			if readErr != nil || !found || entry.ExternalRound != 1+scenario.calls {
				t.Fatalf("record: %#v %v", entry, readErr)
			}
			if len(entry.History) == 0 || entry.History[0].Response != "HISTORICAL_FINDING" {
				t.Fatalf("lost history: %#v", entry)
			}
			if scenario.success {
				if entry.SessionID == "old-session" || entry.SessionID == "" || entry.Verdict != "PASS" {
					t.Fatalf("replacement: %#v", entry)
				}
				if len(entry.Sessions) != 1 || entry.Sessions[0].SessionID != "old-session" || entry.History[0].Harness != scenario.owner || entry.History[len(entry.History)-1].Model == "" {
					t.Fatalf("lost ownership: %#v", entry)
				}
				prompt, err := os.ReadFile(filepath.Join(root, "prompt"))
				if err != nil || !bytes.Contains(prompt, []byte("HISTORICAL_FINDING")) || !bytes.Contains(prompt, []byte(`"change": "added"`)) {
					t.Fatalf("recovery context: %s (%v)", prompt, err)
				}
				if scenario.failure == "auth" {
					if err := run(args, strings.NewReader("continue review"), &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
						t.Fatal(err)
					}
					again, _, err := harness.ReadAuditEntry(record, "W012-recovery", "definition")
					if err != nil || again.Harness != "codex" || again.Model != "fallback" || again.SessionID != entry.SessionID || again.ExternalRound != entry.ExternalRound+1 || len(again.Sessions) != 1 {
						t.Fatalf("fallback resume: %#v %v", again, err)
					}
				}
			} else if output.Len() != 0 {
				t.Fatal("failure emitted verdict")
			}
		})
	}
}

func TestFallbackConfigOneOff(t *testing.T) {
	for _, phase := range []string{"definition", "build", "audit"} {
		t.Run(phase, func(t *testing.T) {
			root := t.TempDir()
			global := filepath.Join(root, "global.yaml")
			if err := os.Mkdir(filepath.Join(root, ".sdlc"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(global, []byte(fmt.Sprintf("delivery:\n  %s:\n    harness: claude\n    model: primary\n    fallback:\n      harness: hermes\n      provider: nous\n      model: z-ai/glm-5.3\n", phase)), 0600); err != nil {
				t.Fatal(err)
			}
			options := harness.ConfigOptions{ProjectRoot: root, GlobalPath: global, Phase: phase, LookupEnv: func(string) (string, bool) { return "", false }}
			config, err := harness.ResolveConfig(options)
			if err != nil || config.Fallback == nil || config.Fallback.Harness != "hermes" || config.Fallback.Provider != "nous" || config.Fallback.Model != "z-ai/glm-5.3" {
				t.Fatalf("global: %#v %v", config, err)
			}
			for _, mapping := range []string{"{}", "{harness: codex, provider: openai, model: project}", "{harness: hermes, model: invalid}"} {
				if err := os.WriteFile(filepath.Join(root, ".sdlc", "project.yaml"), []byte(fmt.Sprintf("delivery:\n  %s:\n    fallback: %s\n", phase, mapping)), 0600); err != nil {
					t.Fatal(err)
				}
				config, err = harness.ResolveConfig(options)
				switch mapping {
				case "{}":
					if err != nil || config.Fallback != nil {
						t.Fatalf("disable: %#v %v", config, err)
					}
				case "{harness: codex, provider: openai, model: project}":
					if err != nil || config.Fallback == nil || config.Fallback.Provider != "" || config.Fallback.Model != "project" {
						t.Fatalf("project override: %#v %v", config, err)
					}
				default:
					if err == nil {
						t.Fatal("incomplete fallback accepted")
					}
				}
			}
		})
	}
}
