package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/readiness"
	"gopkg.in/yaml.v3"
)

const automatedSpec = `* Requirements
** AC007.1 - Reject invalid input :ac:
* Tests
** AT007.1 - Exercise ticket-specific automation :testdef:
`

const automatedValidation = `#+TODO: AMBER RED | GREEN
* Results
** RED AT007.1 - Exercise ticket-specific automation :testdef:
:PROPERTIES:
:REVISION: fixture
:COMMAND: ./check-input
:EVIDENCE: Invalid input was accepted instead of rejected.
:END:
`

// Exercise the executable boundary without running the commands in test records.
func TestAutomatedTestReadinessCLI(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "sdlc-validate")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build validator: %v\n%s", err, output)
	}
	for _, tc := range []struct {
		name, mode, spec, validation, diagnostic string
		exit                                     int
	}{
		{"red before implementation", "test-code", automatedSpec, automatedValidation, "", 0},
		{"red blocks delivery", "delivery-code", automatedSpec, automatedValidation, "GREEN_REQUIRED", 1},
		{"green before implementation", "test-code", automatedSpec, strings.Replace(automatedValidation, "RED AT", "GREEN AT", 1), "", 0},
		{"green delivery", "delivery-code", automatedSpec, strings.Replace(automatedValidation, "RED AT", "GREEN AT", 1), "", 0},
		{"amber before implementation", "test-code", automatedSpec, strings.Replace(strings.Replace(automatedValidation, "RED AT", "AMBER AT", 1), ":COMMAND: ./check-input", ":REASON: Not executed.", 1), "AT_EXECUTION_REQUIRED", 1},
		{"amber blocks delivery", "delivery-code", automatedSpec, strings.Replace(strings.Replace(automatedValidation, "RED AT", "AMBER AT", 1), ":COMMAND: ./check-input", ":REASON: Not executed.", 1), "GREEN_REQUIRED", 1},
		{"missing state", "test-code", automatedSpec, strings.Replace(automatedValidation, "RED AT", "AT", 1), "MISSING_STATE", 1},
		{"missing command", "test-code", automatedSpec, strings.Replace(automatedValidation, ":COMMAND: ./check-input\n", "", 1), "MISSING_EVIDENCE", 1},
		{"missing revision", "delivery-code", automatedSpec, strings.Replace(strings.Replace(automatedValidation, "RED AT", "GREEN AT", 1), ":REVISION: fixture\n", "", 1), "MISSING_EVIDENCE", 1},
		{"missing counterpart", "test-code", automatedSpec, "#+TODO: AMBER RED | GREEN\n* Results\n", "MISSING_TEST", 1},
		{"untagged AT", "test-code", strings.Replace(automatedSpec, " :testdef:", "", 1), automatedValidation, "MISSING_TEST_TAG", 2},
		{"duplicate AT", "test-code", automatedSpec + "** AT007.1 - Duplicate :testdef:\n", automatedValidation, "DUPLICATE_ID", 2},
		{"unknown AT", "test-code", automatedSpec, strings.ReplaceAll(automatedValidation, "AT007.1", "AT007.2"), "UNKNOWN_TEST", 2},
		{"hyphenated ID", "test-code", strings.ReplaceAll(automatedSpec, "AT007.1", "AT-007.1"), strings.ReplaceAll(automatedValidation, "AT007.1", "AT-007.1"), "", 0},
		{"lowercase ID rejected", "test-code", strings.ReplaceAll(automatedSpec, "AT007.1", "at007.1"), automatedValidation, "INVALID_ID", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report := executeValidator(t, binary, tc.mode, tc.spec, tc.validation, tc.exit)
			if tc.diagnostic != "" {
				issues := append(report.Audit.Schema.Errors, report.Audit.Checks[tc.mode].Errors...)
				for _, issue := range issues {
					if issue.Code == tc.diagnostic {
						return
					}
				}
				t.Fatalf("missing %s: %+v", tc.diagnostic, issues)
			}
			if report.Audit.Inventory.Counts["at"] != 1 || report.Audit.Inventory.Counts["acceptance-criteria"] != 1 || report.Audit.Inventory.Counts["tests"] != 1 {
				t.Fatalf("AT and acceptance-criterion counts are not separate: %+v", report.Audit.Inventory.Counts)
			}
			if !report.Audit.Checks[tc.mode].Ready || !report.Audit.Checks[tc.mode].Evaluated {
				t.Fatalf("expected evaluated readiness: %+v", report.Audit.Checks)
			}
		})
	}
	t.Run("all four types counted separately", func(t *testing.T) {
		spec := automatedSpec + "** RT007.1 - Retained coverage :testdef:\n** OT007.1 - One-off verification :testdef:\n** UT007.1 - Human verification :testdef:\n"
		validation := strings.Replace(automatedValidation, "RED AT", "GREEN AT", 1) + `** GREEN RT007.1 - Retained coverage :testdef:
:PROPERTIES:
:REVISION: fixture
:COMMAND: make test
:EVIDENCE: Regression suite passed.
:END:
** GREEN OT007.1 - One-off verification :testdef:
:PROPERTIES:
:REVISION: fixture
:EVIDENCE: One-off verification succeeded.
:END:
** AMBER UT007.1 - Human verification :testdef:
:PROPERTIES:
:REASON: Awaiting operator.
:END:
`
		report := executeValidator(t, binary, "delivery-code", spec, validation, 0)
		for key, expected := range map[string]int{"acceptance-criteria": 1, "tests": 4, "at": 1, "rt": 1, "ot": 1, "ut": 1} {
			if report.Audit.Inventory.Counts[key] != expected {
				t.Fatalf("count %s: got %d, want %d", key, report.Audit.Inventory.Counts[key], expected)
			}
		}
	})
}

type cliReadinessReport struct {
	Audit struct {
		Schema    readiness.SchemaResult           `yaml:"schema"`
		Checks    map[string]readiness.CheckResult `yaml:",inline"`
		Inventory struct {
			Counts map[string]int `yaml:"counts"`
		} `yaml:"inventory"`
	} `yaml:"audit-readiness"`
}

func executeValidator(t *testing.T, binary, mode, spec, validation string, expectedExit int) cliReadinessReport {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{"spec.org": spec, "validation.org": validation}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// #nosec G204 -- binary is built from this package in t.TempDir; arguments are test-owned fixtures.
	command := exec.Command(binary, "--readiness-check="+mode, "--spec", filepath.Join(dir, "spec.org"))
	var out, diagnostics bytes.Buffer
	command.Stdout, command.Stderr = &out, &diagnostics
	err := command.Run()
	exit := 0
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			t.Fatal(err)
		}
		exit = exitError.ExitCode()
	}
	if exit != expectedExit || diagnostics.Len() != 0 {
		t.Fatalf("exit %d, want %d; stderr %s; stdout %s", exit, expectedExit, diagnostics.String(), out.String())
	}
	var report cliReadinessReport
	if err := yaml.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid readiness YAML: %v\n%s", err, out.String())
	}
	for name, expected := range files {
		actual, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(actual) != expected {
			t.Fatalf("validator changed %s: %v", name, err)
		}
	}
	return report
}
