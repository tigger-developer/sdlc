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

* Work items

** TODO W041 - Preserve this work item :defect:

* Legacy Acceptance Criteria (SDLC v1)                         :legacy:
:PROPERTIES:
:CUSTOM_ID: legacy-acceptance-criteria
:VISIBILITY: folded
:END:

{{LEGACY_ACCEPTANCE_CRITERIA}}

* Migration record

- *Historical requirements:* [[file:ACs.org][Legacy acceptance-criteria ledger]].
`)
	writeProjectTestFile(t, filepath.Join(root, "docs", "ACs.org"), `#+TITLE: Legacy acceptance criteria
#+DATE: 2026-09-03

* Ledger authority

This file records legacy requirements.

Requirements established or changed through Spec Kit are governed by approved
=specs/*/spec.md= artefacts.

Requirements established or changed through SDLC v3 are governed by signed-off
=specs/*/spec.org= artefacts.

* Status vocabulary

- *HOLDING:* Current.

* Acceptance criteria

** Issue 7 - Reject invalid configuration

*** AC7.1 - Reject an invalid host definition

**** Requirement

Invalid hosts are rejected.

**** Provenance

- [[file:archive/migrated-tickets/7.md][#7 - Reject invalid configuration]]

**** Status

*HOLDING:* Current requirement confirmed during migration.

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
		"** Legacy ledger authority at migration",
		"** Issue 7 - Reject invalid configuration",
		"*** HOLD AC7.1 - Reject an invalid host definition",
		"**** Status qualification",
		"Current requirement confirmed during migration.",
		"[[file:archive/migrated-tickets/7.md][#7 - Reject invalid configuration]]",
		"* Migration record",
		"[[#legacy-acceptance-criteria][Legacy Acceptance Criteria (SDLC v1)]]",
	} {
		if !strings.Contains(string(work), want) {
			t.Fatalf("consolidated work ledger lacks %q:\n%s", want, work)
		}
	}
	if strings.Contains(string(work), "**** Status\n") || strings.Contains(string(work), legacyLedgerPlaceholder) {
		t.Fatalf("consolidated work ledger retained duplicate status or placeholder:\n%s", work)
	}
	if strings.Contains(strings.Join(strings.Fields(string(work)), " "), obsoleteSpecKitAuthority) {
		t.Fatalf("consolidated work ledger retained obsolete Spec Kit authority:\n%s", work)
	}
	if !strings.Contains(string(work), "Requirements established or changed through SDLC v3") {
		t.Fatalf("consolidated work ledger dropped current v3 authority:\n%s", work)
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

**** Status

HOLDING
`)

	_, err := MergeLegacyAcceptanceCriteria(root)
	if err == nil || !strings.Contains(err.Error(), "different") {
		t.Fatalf("conflicting merge error = %v", err)
	}
	if !exists(filepath.Join(root, "docs", "ACs.org")) {
		t.Fatal("conflicting source ledger was removed")
	}
}

func TestMergeLegacyAcceptanceCriteriaNormalizesExistingMergedLedger(t *testing.T) {
	root := t.TempDir()
	writeProjectTestFile(t, filepath.Join(root, "docs", "work.org"), `#+TITLE: Existing Work

* Legacy Acceptance Criteria (SDLC v1)
:PROPERTIES:
:CUSTOM_ID: legacy-acceptance-criteria
:VISIBILITY: folded
:END:

** Ledger authority

Historical authority.

Requirements established or changed through Spec Kit are governed by approved
=specs/*/spec.md= artefacts.

** Status vocabulary

Historical vocabulary.

** Area

*** AC1.1 - Existing requirement

**** Status

HOLDING

* Migration record
`)

	result, err := MergeLegacyAcceptanceCriteria(root)
	if err != nil {
		t.Fatal(err)
	}
	if !result.LedgerChanged || result.SourceRemoved {
		t.Fatalf("normalization result = %#v", result)
	}
	contents, err := os.ReadFile(filepath.Join(root, "docs", "work.org"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		legacyLedgerHeading,
		"** Legacy ledger authority at migration",
		"** Legacy status vocabulary at migration",
		"*** HOLD AC1.1 - Existing requirement",
	} {
		if !strings.Contains(string(contents), want) {
			t.Fatalf("normalized ledger lacks %q:\n%s", want, contents)
		}
	}
	if strings.Contains(string(contents), "**** Status\n") {
		t.Fatalf("normalized ledger retained the duplicate Status field:\n%s", contents)
	}
	if strings.Contains(strings.Join(strings.Fields(string(contents)), " "), obsoleteSpecKitAuthority) {
		t.Fatalf("normalized ledger retained obsolete Spec Kit authority:\n%s", contents)
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
		"#+TYP_TODO: PENDING FAILING SUPERSEDED | HOLD ASSUMED_PASS",
		"#+TAGS: feature defect maintenance migration legacy",
		"** TODO W009 - Existing defect :defect:",
		"* Work items",
		"* Migration record",
		"- *Migration branch:* sdlc-v3-migration-2026-09-06",
	} {
		if !strings.Contains(string(contents), want) {
			t.Fatalf("amended work ledger lacks %q:\n%s", want, contents)
		}
	}
	for _, absent := range []string{"* Open defects\n", "* Undelivered features\n", "* Active delivery\n", "* Human review\n", "* Closed work\n", legacyLedgerPlaceholder} {
		if strings.Contains(string(contents), absent) {
			t.Fatalf("amended work ledger retained obsolete structure %q:\n%s", absent, contents)
		}
	}
}

func testTime() time.Time {
	return time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
}
