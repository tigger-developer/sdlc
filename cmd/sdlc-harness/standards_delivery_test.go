// ABOUTME: Verifies mandatory standards at the provider boundary without metered calls.
// ABOUTME: Covers context reuse, changed standards and refusal before invocation.
package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tigger-developer/sdlc/internal/harness"
)

func TestMandatoryStandardsDeliveredToEveryProviderOnce(t *testing.T) {
	project, root := t.TempDir(), t.TempDir()
	standard, source := filepath.Join(root, "TESTING.md"), filepath.Join(project, "tests.go")
	for path, text := range map[string]string{standard: "MANDATORY_TESTING_RULE", source: "PROJECT_TEST_CODE"} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	evidence, err := harness.CaptureEvidence(project, []string{source, standard})
	if err != nil {
		t.Fatal(err)
	}
	for _, provider := range []string{"codex", "claude", "copilot", "hermes"} {
		t.Run(provider, func(t *testing.T) {
			runner := recoveryRun{request: harness.Request{Prompt: "Audit.", Directory: project}, evidence: evidence,
				standardsRoot: root, requiredStandards: []string{standard}, diagnostics: io.Discard}
			config := harness.Config{Harness: provider, Provider: "fixture", Model: "fixture", Timeout: time.Second}
			request, err := runner.prepare(harness.AuditEntry{}, config, false, "")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(request.Prompt, "MANDATORY_TESTING_RULE") != 1 {
				t.Fatalf("mandatory contents missing or duplicated: %s", request.Prompt)
			}
			entry := harness.AuditEntry{SessionID: request.SessionID, StandardsContract: reviewStandardsContract, Standards: request.Standards}
			retained, err := runner.prepare(entry, config, true, "")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(retained.Prompt, "MANDATORY_TESTING_RULE") {
				t.Fatal("unchanged standards delivered again")
			}
			if err := os.WriteFile(standard, []byte("CHANGED_TESTING_RULE"), 0o600); err != nil {
				t.Fatal(err)
			}
			runner.evidence, err = harness.CaptureEvidence(project, []string{source, standard})
			if err != nil {
				t.Fatal(err)
			}
			changed, err := runner.prepare(entry, config, true, "")
			if err != nil || strings.Count(changed.Prompt, "CHANGED_TESTING_RULE") != 1 {
				t.Fatalf("changed standard delivery: %v", err)
			}
			fresh, err := runner.prepare(entry, config, false, "provider-replaced")
			if err != nil || !strings.Contains(fresh.Prompt, "CHANGED_TESTING_RULE") {
				t.Fatalf("replacement context delivery: %v", err)
			}
			if err := os.WriteFile(standard, []byte("MANDATORY_TESTING_RULE"), 0o600); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLegacyWorkflowContextIsReplacedBeforeReview(t *testing.T) {
	for name, contract := range map[string]int{"full-library": 1, "mixed-shared-standards": 2} {
		t.Run(name, func(t *testing.T) {

			project, root, bin := t.TempDir(), t.TempDir(), t.TempDir()
			standard := filepath.Join(root, "TESTING.md")
			if err := os.WriteFile(standard, []byte("CURRENT_REVIEW_RULE"), 0o600); err != nil {
				t.Fatal(err)
			}
			probe := filepath.Join(project, "arguments")
			if err := os.WriteFile(filepath.Join(project, "result"), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("REVIEW_ARGUMENTS", probe)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			script := `#!/bin/sh
printf '%s\n' "$@" > "$REVIEW_ARGUMENTS"
output=
previous=
for argument in "$@"; do
  if [ "$previous" = "-o" ]; then output="$argument"; fi
  previous="$argument"
done
/bin/cat >/dev/null
printf 'AUDIT: standalone\nREVISION: candidate\nVERDICT: PASS\n' > "$output"
printf '{"type":"thread.started","thread_id":"new-review-context"}\n'
`
			// #nosec G306 -- private synthetic provider never contacts a hosted model.
			if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			evidence, err := harness.CaptureEvidence(project, []string{standard})
			if err != nil {
				t.Fatal(err)
			}
			var diagnostics bytes.Buffer
			runner := recoveryRun{audit: true, evidence: evidence, standardsRoot: root,
				requiredStandards: []string{standard}, diagnostics: &diagnostics,
				request: harness.Request{Prompt: "Audit.", Directory: project, ResultFile: filepath.Join(project, "result")},
				config:  harness.Config{Harness: "codex", Model: "fixture", Timeout: time.Second, MaxFailures: 1}}
			legacy := harness.AuditEntry{SessionID: "old-workflow-context", StandardsContract: contract, Standards: evidence.Files}
			result, err := runner.execute(legacy, true)
			if err != nil || !strings.Contains(result.Response, "VERDICT: PASS") {
				t.Fatalf("replacement review: %v, %s", err, diagnostics.String())
			}
			arguments, err := os.ReadFile(probe)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(arguments, []byte("old-workflow-context")) || bytes.Contains(arguments, []byte("resume")) {
				t.Fatalf("legacy context was resumed: %s", arguments)
			}
			if !strings.Contains(diagnostics.String(), "round budget unchanged") {
				t.Fatal("replacement did not report preserved budget")
			}

		})
	}
}

func TestMandatoryStandardCannotBeOmittedFromRequest(t *testing.T) {
	root := t.TempDir()
	runner := recoveryRun{standardsRoot: root, requiredStandards: []string{filepath.Join(root, "TESTING.md")}}
	_, err := runner.prepare(harness.AuditEntry{}, harness.Config{Harness: "hermes"}, false, "")
	if err == nil || !strings.Contains(err.Error(), "required audit standard") {
		t.Fatalf("missing standard accepted: %v", err)
	}
}

func TestStandardsRequestDoesNotPartiallyConsumeInvalidBatch(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "GO.md")
	if err := os.WriteFile(path, []byte("Go rules"), 0o600); err != nil {
		t.Fatal(err)
	}
	evidence, err := harness.CaptureEvidence(root, []string{path})
	if err != nil {
		t.Fatal(err)
	}
	runner := recoveryRun{standardsRoot: root, evidence: evidence}
	seen := map[string]bool{}
	if _, _, err := runner.requestedStandardPaths(`STANDARDS-REQUEST: ["GO.md","missing.md"]`, seen); err == nil {
		t.Fatal("invalid batch accepted")
	}
	if seen["GO.md"] {
		t.Fatal("invalid request consumed an undelivered standard")
	}
}

func TestAuditRejectsCallerRoutingOverrideBeforeProvider(t *testing.T) {
	err := run([]string{"--audit-prompts", "caller.yaml"}, strings.NewReader("audit"), io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "routing and standards are harness-owned") {
		t.Fatalf("caller routing override accepted: %v", err)
	}
}

func TestInstalledAuditAliasUsesCanonicalStandardsRoot(t *testing.T) {
	root := t.TempDir()
	command := filepath.Join(root, "sdlc", "bin", "sdlc-harness")
	alias := filepath.Join(root, "commands", "sdlc-audit")
	for _, directory := range []string{filepath.Dir(command), filepath.Dir(alias)} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(command, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(command, alias); err != nil {
		t.Fatal(err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(canonicalRoot, "sdlc", "prompts", "audits.yaml")
	for _, executable := range []string{command, alias} {
		if got := auditRegistryForExecutable(executable); got != want {
			t.Fatalf("registry for %s = %s; want %s", executable, got, want)
		}
	}
}

func TestAuditRejectsCallerStandardAndItsAlias(t *testing.T) {
	project := t.TempDir()
	root := filepath.Join(project, "installed")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	standard := filepath.Join(root, "TESTING.md")
	if err := os.WriteFile(standard, []byte("rules"), 0o600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(project, "alias.md")
	if err := os.Symlink(standard, alias); err != nil {
		t.Fatal(err)
	}
	old := installedStandardsRoot
	installedStandardsRoot = func() string { return root }
	t.Cleanup(func() { installedStandardsRoot = old })
	for _, input := range []string{standard, alias} {
		err := run([]string{"--project", project, "--input", input}, strings.NewReader("audit"), io.Discard, io.Discard)
		if err == nil || !strings.Contains(err.Error(), "caller-supplied audit standard") {
			t.Fatalf("caller standard accepted: %v", err)
		}
	}
}

func TestCachedVerdictRequiresStandardsDeliveryProof(t *testing.T) {
	standard := harness.EvidenceFile{Path: "TESTING.md", SHA256: "current", Bytes: 10}
	round := harness.AuditRound{Round: 1, SessionID: "native", CacheKey: "key", Revision: "candidate", Verdict: "PASS", Response: "GATE: definition\nREVISION: candidate\nVERDICT: PASS\n"}
	entry := harness.AuditEntry{Gate: "definition", History: []harness.AuditRound{round}}
	if _, ok := cachedAudit(entry, "key", standard); ok {
		t.Fatal("legacy verdict without delivery proof was reused")
	}
	entry.History[0].Standards = []harness.EvidenceFile{standard}
	if _, ok := cachedAudit(entry, "key", standard); !ok {
		t.Fatal("verified verdict was not reused")
	}
	entry.History[0].Standards[0].SHA256 = "old"
	if _, ok := cachedAudit(entry, "key", standard); ok {
		t.Fatal("stale standards proof was reused")
	}
}

func TestSDLCGoEvidenceRequiresGoStandardWithoutProfileDeclaration(t *testing.T) {
	project := t.TempDir()
	if err := os.Mkdir(filepath.Join(project, ".sdlc"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".sdlc", "project.yaml"), []byte("version: 3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	root := installedStandardsRoot()
	paths, err := managedRequiredStandards(project, "test-code", root, "tests.go")
	if err != nil {
		t.Fatal(err)
	}
	goPresent, testingPresent := false, false
	for _, path := range paths {
		goPresent = goPresent || path == filepath.Join(root, "technologies", "GO.md")
		testingPresent = testingPresent || path == filepath.Join(root, "TESTING.md")
	}
	if !goPresent || !testingPresent {
		t.Fatal("Go test review omitted mandatory Go or testing rules")
	}
}

func TestSpikePackageSelectsEvidenceTechnologyWithoutWorkflowDocs(t *testing.T) {
	root := installedStandardsRoot()
	paths, err := spikeAuditStandards(root, []string{"experiment.go"})
	if err != nil {
		t.Fatal(err)
	}
	present := map[string]bool{}
	for _, path := range paths {
		present[path] = true
	}
	for _, name := range []string{"INDEPENDENT-AUDITOR.md", "READING.md", "REVIEW.md", "CODING.md", "GIT.md", "TESTING.md", "SECURITY.md", "ARCHITECTURE.md", "technologies/GO.md"} {
		if !present[filepath.Join(root, name)] {
			t.Fatalf("spike omitted %s", name)
		}
	}
	for _, name := range []string{"AUDIT.md", "STANDALONE-AUDIT.md", "SPIKE-AUDIT.md", "MAIN.md", "DELIVERY.md", "ADVICE.md", "AUDITS.md", "SELF-AUDIT.md", "EXTERNAL-AUDIT.md", "HARNESS.md", "ISSUES.md", "DOCUMENTATION.md", "technologies/RUST.md"} {
		if present[filepath.Join(root, name)] {
			t.Fatalf("spike included workflow/documentation standard %s", name)
		}
	}
}

func TestSDLCStandardsIncludeEvidenceFormatsAndTechnologyDependencies(t *testing.T) {
	project := t.TempDir()
	if err := os.Mkdir(filepath.Join(project, ".sdlc"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".sdlc", "project.yaml"), []byte("standards:\n  technologies: [NODE, HUGO]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	root := installedStandardsRoot()
	paths, err := managedRequiredStandards(project, "test-code", root, "README.md", "plan.org")
	if err != nil {
		t.Fatal(err)
	}
	present := map[string]int{}
	for _, path := range paths {
		present[path]++
	}
	for _, name := range []string{"DOCUMENTATION.md", "MARKDOWN.md", "ORGMODE.md", "SECURITY.md", "technologies/NODE.md", "technologies/JAVASCRIPT.md", "technologies/HUGO.md", "technologies/WEB.md"} {
		if present[filepath.Join(root, name)] != 1 {
			t.Fatalf("required standard %s count=%d", name, present[filepath.Join(root, name)])
		}
	}
}
