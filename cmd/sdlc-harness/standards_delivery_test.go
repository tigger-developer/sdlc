// ABOUTME: Verifies mandatory standards at the provider boundary without metered calls.
// ABOUTME: Covers context reuse, changed standards and refusal before invocation.
package main

import (
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
			entry := harness.AuditEntry{SessionID: request.SessionID, StandardsContract: 1, Standards: request.Standards}
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

func TestSpikePackageIncludesAllTechnologyStandardsWithoutWorkflowDocs(t *testing.T) {
	root := installedStandardsRoot()
	paths, err := spikeAuditStandards(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	present := map[string]bool{}
	for _, path := range paths {
		present[path] = true
	}
	technologies, err := installedAuditDocuments(filepath.Join(root, "technologies"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range technologies {
		if !present[path] {
			t.Fatalf("spike omitted coding standard %s", path)
		}
	}
	for _, name := range []string{"SPIKE-AUDIT.md", "CODING.md", "GIT.md", "TESTING.md", "SECURITY.md", "ARCHITECTURE.md"} {
		if !present[filepath.Join(root, name)] {
			t.Fatalf("spike omitted %s", name)
		}
	}
	for _, name := range []string{"MAIN.md", "AUDITS.md", "ISSUES.md", "DOCUMENTATION.md"} {
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
