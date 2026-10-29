---
title: Self-Audit in the Current Context
version: 3
last-updated: 2026-10-29
---

# Self-audit

Use this route for an explicitly requested self-review or a workflow-required
review in the author's existing context. Do not enter the independent-auditor
role, create another context or claim independence. Read `READING.md`,
`REVIEW.md` and the requested review skill's criteria, reusing retained standards.
This entry point does not load MAIN or activate SDLC in a non-SDLC project.

The old external audit harness is retired. Review internally without probing or
attempting to recover it. Independent external review is a separate explicitly
authorized fresh-context route, not a fallback invocation of the old harness.

Perform a deliberate review of the complete candidate after authoring, not a
summary of work performed. Remediate all valid findings within existing authority,
rerun affected checks and review the changed candidate. Preserve historical
external findings: address them or record an evidence-backed rejection, linking
the original record. Do not erase history or convert an external FAIL to PASS.
Stop for an unremediable defect or human-only decision, not provider unavailability.

## Evidence

Append dated entries, retaining prior reviews. Managed work uses
`audits.yaml` beside `spec.org`, under `AUDIT-RECORD.md`. Ticketed spikes use
the same single record, linked from the brief; do not duplicate review results
in Findings. Non-SDLC review uses one `audits.yaml` at the authorized review
location without inventing managed gates. Review-only authority does not authorize candidate edits. If no
record-writing authority is available, return the evidence for the operator
rather than mutating files or claiming a recorded gate result.

```text
Review method: self (current session; not independent)
Gate: definition | test-code | delivery-code
Mode: SDLC | paired | emergency | spike | non-SDLC
Revision: <reviewed commit and identified working-tree delta>
Verdict: PASS | FAIL | PROVISIONAL PASS (only under REVIEW)
```

Include actual standards/authorities, evidence, findings, remediation, checks and
limitations in the canonical YAML record; the text above describes report fields,
not a second file format. Preserve legacy evidence as linked history. Do not
create or continue parallel review documents. Disclose that this is self-review,
not independent review.

A self-audit satisfies a managed review prerequisite only with current recorded
internal PASS.
It permits the next otherwise-authorized step, not sign-off, lifted holds,
closure, merge or deployment. Tests and spike restrictions remain mandatory.
Changed candidates require reassessment under REVIEW; subsequent stages remain
internal unless the operator explicitly requests independent review.

## Review limits

Managed work follows REVIEW's five-round ceiling per ticket and gate, not a
configuration value that permits additional rounds. Count deliberate completed
reviews, including rejection of an incomplete submitted package. Ordinary author
checks and local pre-submission readiness checks do not consume rounds. Preserve
dated history across skill re-entry, compaction and reviewer/method changes;
changing method does not replenish the gate's allowance or relabel independent
evidence as internal. Keep each record's ownership and provenance intact.

At round five without effective PASS, stop autonomous reviews and alert the
operator. In an interactive session the operator may direct further self-audit
or independent external audit and authorize a reset under REVIEW. Record a new
cycle in `audits.yaml` before continuing; preserve earlier rounds and findings.
Standalone and spike profiles retain their authorized bounds without inventing
managed gates. Neither infrastructure nor the agent may reset autonomously.
