package projectinit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMergeLegacyAcceptanceCriteriaPreservesWorkAndConsolidatesAuthority(t *testing.T) {
	root := t.TempDir()
	writeProjectTestFile(t, filepath.Join(root, "docs", "work.org"), `#+TITLE: Existing Work

* Open defects

** TODO W041 - Preserve this work item :defect:

* Migration record

- *Historical requirements:* [[file:ACs.org][Legacy acceptance-criteria ledger]].
`)
	writeProjectTestFile(t, filepath.Join(root, "docs", "ACs.org"), `#+TITLE: Legacy acceptance criteria
#+DATE: 2026-09-03

* Ledger authority

This file records legacy requirements.

* Status vocabulary

- *HOLDING:* Current.

* Acceptance criteria

** Issue 7 - Reject invalid configuration

*** AC7.1 - Reject an invalid host definition

**** Requirement

Invalid hosts are rejected.

**** Provenance

- [[file:archive/migrated-tickets/7.md][#7 - Reject invalid configuration]]

* Migration notes and footnotes

[fn:migration] Historical evidence.
`)
	writeProjectTestFile(t, filepath.Join(root, projectProfilePath), `version: 3
authorities:
  product:
    - README.md
  requirements:
    - docs/work.org
    - docs/ACs.org
`)

	result, err := MergeLegacyAcceptanceCriteria(root)
	if err != nil {
		t.Fatal(err)
	}
	if !result.LedgerChanged || !result.ProfileChanged || !result.SourceRemoved || result.AcceptanceCriteria != 1 {
		t.Fatalf("merge result = %#v", result)
	}
	if exists(filepath.Join(root, "docs", "ACs.org")) {
		t.Fatal("separate legacy ledger remains after consolidation")
	}
	work, err := os.ReadFile(filepath.Join(root, "docs", "work.org"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"** TODO W041 - Preserve this work item :defect:",
		legacyLedgerHeading,
		":VISIBILITY: folded",
		":SOURCE_DATE: 2026-09-03",
		"** Ledger authority",
		"** Issue 7 - Reject invalid configuration",
		"*** AC7.1 - Reject an invalid host definition",
		"[[file:archive/migrated-tickets/7.md][#7 - Reject invalid configuration]]",
		"* Migration record",
		"[[#legacy-acceptance-criteria][Legacy Acceptance Criteria (SDLC v1)]]",
	} {
		if !strings.Contains(string(work), want) {
			t.Fatalf("consolidated work ledger lacks %q:\n%s", want, work)
		}
	}
	profile, err := os.ReadFile(filepath.Join(root, projectProfilePath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(profile), "requirements:\n        - docs/work.org") || strings.Contains(string(profile), "docs/ACs.org") {
		t.Fatalf("project profile did not consolidate requirement authority:\n%s", profile)
	}

	before := string(work)
	second, err := MergeLegacyAcceptanceCriteria(root)
	if err != nil {
		t.Fatal(err)
	}
	if second.LedgerChanged || second.ProfileChanged {
		t.Fatalf("idempotent rerun changed state: %#v", second)
	}
	after, err := os.ReadFile(filepath.Join(root, "docs", "work.org"))
	if err != nil || string(after) != before {
		t.Fatalf("idempotent work ledger = %q, %v", after, err)
	}
}

func TestMergeLegacyAcceptanceCriteriaRejectsDifferentEmbeddedCopy(t *testing.T) {
	root := t.TempDir()
	writeProjectTestFile(t, filepath.Join(root, "docs", "work.org"), legacyLedgerHeading+"\n:PROPERTIES:\n:VISIBILITY: folded\n:END:\n\nDifferent content.\n\n* Migration record\n")
	writeProjectTestFile(t, filepath.Join(root, "docs", "ACs.org"), `* Acceptance criteria

** Area
*** AC1.1 - Original content
`)

	_, err := MergeLegacyAcceptanceCriteria(root)
	if err == nil || !strings.Contains(err.Error(), "different") {
		t.Fatalf("conflicting merge error = %v", err)
	}
	if !exists(filepath.Join(root, "docs", "ACs.org")) {
		t.Fatal("conflicting source ledger was removed")
	}
}

func TestWriteWorkLedgerPreservesExistingContent(t *testing.T) {
	root := t.TempDir()
	writeProjectTestFile(t, filepath.Join(root, "docs", "work.org"), `#+TITLE: Existing Work

* Open defects

** TODO W009 - Existing defect :defect:
`)
	generation := projectGeneration{
		source: "v2", base: "master", archive: "sdlc_v2_state_2026-09-06",
		migration: "sdlc-v3-migration-2026-09-06",
	}
	if err := writeWorkLedger(testSDLCRoot(t), root, generation, testTime()); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(root, "docs", "work.org"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"#+TITLE: Existing Work",
		"** TODO W009 - Existing defect :defect:",
		"* Undelivered features",
		"* Active delivery",
		"* Human review",
		"* Closed work",
		"* Migration record",
		"- *Migration branch:* sdlc-v3-migration-2026-09-06",
	} {
		if !strings.Contains(string(contents), want) {
			t.Fatalf("amended work ledger lacks %q:\n%s", want, contents)
		}
	}
}

func testTime() time.Time {
	return time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
}
