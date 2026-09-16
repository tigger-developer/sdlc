package readiness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const specFixture = `#+TITLE: Example
* Acceptance Criteria
** AC007.1 - Reject bad input :ac:
- Required outcome.
* Test Definitions
** RT007.1 - Reject invalid input :testdef:
- Expected: rejected.
** OT-007.2 - Verify integration :testdef:
- Expected: accepted.
** UT007.3 - Readable output :testdef:
- Expected: operator confirmation.
`

const validationFixture = `#+TODO: AMBER RED | GREEN
* Results
** GREEN RT007.1 - Reject invalid input :testdef:
:PROPERTIES:
:REVISION: abc123
:COMMAND: make test
:EVIDENCE: Suite result: invalid input rejected.
:END:
** GREEN OT-007.2 - Verify integration :testdef:
:PROPERTIES:
:REVISION: abc123
:EVIDENCE: Bounded integration check succeeded.
:END:
** AMBER UT007.3 - Readable output :testdef:
:PROPERTIES:
:REASON: Awaiting operator validation.
:END:
`

func checkFixture(t *testing.T, spec, validation, mode string) Report {
	t.Helper()
	dir := t.TempDir()
	specPath, validationPath := filepath.Join(dir, "spec.org"), filepath.Join(dir, "validation.org")
	for path, content := range map[string]string{specPath: spec, validationPath: validation} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	report, err := Check(specPath, validationPath, mode)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func TestReadinessInventoryAndStages(t *testing.T) {
	report := checkFixture(t, specFixture, validationFixture, "delivery-code")
	if report.ExitCode() != 0 || report.Audit.Inventory.Counts.Tests != 3 || report.Audit.Inventory.Counts.ACs != 1 {
		t.Fatalf("report: %+v", report)
	}
	red := strings.Replace(validationFixture, "GREEN RT", "RED RT", 1)
	if r := checkFixture(t, specFixture, red, "test-code"); r.ExitCode() != 0 {
		t.Fatalf("RED should qualify for test-code: %+v", r)
	}
	if r := checkFixture(t, specFixture, red, "delivery-code"); r.ExitCode() != 1 {
		t.Fatalf("RED should block delivery: %+v", r)
	}
	amberOT := strings.Replace(validationFixture, "GREEN OT", "AMBER OT", 1)
	amberOT = strings.Replace(amberOT, ":EVIDENCE: Bounded", ":REASON: Not yet executed\n:EVIDENCE: Bounded", 1)
	if r := checkFixture(t, specFixture, amberOT, "test-code"); r.ExitCode() != 0 {
		t.Fatal("AMBER OT and UT should qualify for test-code with reasons")
	}
	if r := checkFixture(t, specFixture, amberOT, "delivery-code"); r.ExitCode() != 1 {
		t.Fatal("OT must be GREEN")
	}
}

func TestSchemaAndEligibilityFailures(t *testing.T) {
	for _, tc := range []struct {
		name, spec, validation, code string
		exit                         int
	}{
		{"no state", specFixture, strings.Replace(validationFixture, "GREEN RT", "RT", 1), "MISSING_STATE", 1},
		{"amber RT", specFixture, strings.Replace(strings.Replace(validationFixture, "GREEN RT", "AMBER RT", 1), ":COMMAND: make test", ":REASON: Not yet executed", 1), "RT_EXECUTION_REQUIRED", 1},
		{"missing test", specFixture, strings.Split(validationFixture, "** AMBER UT")[0], "MISSING_TEST", 1},
		{"untagged", strings.Replace(specFixture, " :testdef:", "", 1), validationFixture, "MISSING_TEST_TAG", 2},
		{"duplicate", specFixture + "\n** RT007.1 - Duplicate :testdef:\n", validationFixture, "DUPLICATE_ID", 2},
		{"duplicate result", specFixture, validationFixture + "\n** RED RT007.1 - Duplicate :testdef:\n", "DUPLICATE_ID", 2},
		{"bad state", specFixture, strings.Replace(validationFixture, "GREEN RT", "PASS RT", 1), "INVALID_STATE", 2},
		{"no title", strings.Replace(specFixture, "RT007.1 - Reject invalid input", "RT007.1", 1), validationFixture, "MISSING_TITLE", 2},
		{"bad id", specFixture + "\n** RTwrong - Bad :testdef:\n", validationFixture, "INVALID_ID", 2},
		{"wrong todo", specFixture, strings.Replace(validationFixture, "AMBER RED | GREEN", "AMBER | RED GREEN", 1), "INVALID_STATES", 2},
		{"late todo", specFixture, "* Results\n" + validationFixture, "INVALID_STATES", 2},
		{"untagged AC", strings.Replace(specFixture, " :ac:", "", 1), validationFixture, "INVALID_AC_TAG", 2},
		{"missing evidence", specFixture, strings.ReplaceAll(validationFixture, ":REVISION: abc123\n", ""), "MISSING_EVIDENCE", 1},
		{"unknown test", specFixture, validationFixture + "\n** AMBER RT008.1 - Unknown :testdef:\n", "UNKNOWN_TEST", 2},
		{"empty spec", "#+TITLE: Empty\n", validationFixture, "EMPTY_INVENTORY", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := checkFixture(t, tc.spec, tc.validation, "test-code")
			if r.ExitCode() != tc.exit {
				t.Fatalf("exit %d, expected %d: %+v", r.ExitCode(), tc.exit, r)
			}
			issues := append(append([]Issue{}, r.Audit.Schema.Errors...), r.Audit.Checks["test-code"].Errors...)
			found := false
			for _, issue := range issues {
				if issue.Code == tc.code {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing %s: %+v", tc.code, issues)
			}
			if tc.exit == 2 && r.Audit.Checks["test-code"].Evaluated {
				t.Fatal("schema failure evaluated readiness")
			}
		})
	}
}

func TestDeliveryWithoutRegressionTests(t *testing.T) {
	spec := "* Acceptance Criteria\n** AC001.1 - Required outcome :ac:\n* Test Definitions\n** OT001.1 - Check integration :testdef:\n** UT001.2 - Confirm presentation :testdef:\n"
	validation := "#+TODO: AMBER RED | GREEN\n* Results\n** GREEN OT001.1 - Check integration :testdef:\n:PROPERTIES:\n:REVISION: fixture\n:EVIDENCE: Bounded check passed\n:END:\n** AMBER UT001.2 - Confirm presentation :testdef:\n:PROPERTIES:\n:REASON: Awaiting operator\n:END:\n"
	r := checkFixture(t, spec, validation, "delivery-code")
	if r.ExitCode() != 0 || r.Audit.Inventory.Counts.Tests != 2 {
		t.Fatalf("OT/UT-only delivery should be eligible: %+v", r)
	}
	validation = strings.Replace(validation, "GREEN OT", "OT", 1)
	if r := checkFixture(t, spec, validation, "delivery-code"); r.ExitCode() != 1 {
		t.Fatalf("no-RT delivery still requires OT execution: %+v", r)
	}
}

func TestIgnoresExamplesAndHistoryButRejectsBrokenBlocks(t *testing.T) {
	example := "\n#+begin_src org\n** RT007.1 - Not a record :testdef:\n#+end_src\n"
	history := "\n*** Run 2026-09-16\nPreviously RED; RT007.1 failed before the change.\n"
	r := checkFixture(t, specFixture+example, validationFixture+history, "delivery-code")
	if r.ExitCode() != 0 {
		t.Fatalf("example/history counted: %+v", r)
	}
	r = checkFixture(t, specFixture+"\n#+begin_example\n", validationFixture, "delivery-code")
	if r.ExitCode() != 2 {
		t.Fatal("unclosed block accepted")
	}
}
