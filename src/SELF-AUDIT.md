---
title: Self-Audit in the Current Context
version: 2
last-updated: 2026-10-24
---

# Self-audit

Use this route for an explicitly requested self-review or a workflow-required
review in the author's existing context. Do not enter the independent-auditor
role, create another context or claim independence. Read `READING.md`,
`REVIEW.md` and the requested review skill's criteria, reusing retained standards.
This entry point does not load MAIN or activate SDLC in a non-SDLC project.

Automatic external auditing is suspended. Review internally from the outset;
do not first call the harness, probe provider health, reset auditor state or wait
for external recovery. After any terminal external incident under an authorized
future external route, this self-audit fallback also requires no further
permission. Harness recovery remains a separate operator concern.

Perform a deliberate review of the complete candidate after authoring, not a
summary of work performed. Remediate all valid findings within existing authority,
rerun affected checks and review the changed candidate. Preserve historical
external findings: address them or record an evidence-backed rejection, linking
the original record. Do not erase history or convert an external FAIL to PASS.
Stop for an unremediable defect or human-only decision, not provider unavailability.

## Evidence

Append dated entries, retaining prior reviews. Managed work uses
`internal-review.md` beside `spec.org`; spikes use the brief's Findings section;
non-SDLC review uses the existing review record or a clearly named project-local
review file. Review-only authority does not authorize candidate edits. If no
record-writing authority is available, return the evidence for the operator
rather than mutating files or claiming a recorded gate result.

```text
Review method: internal (current session; not independent)
Gate: definition | test-code | delivery-code
Mode: SDLC | paired | emergency | spike | non-SDLC
Revision: <reviewed commit and identified working-tree delta>
External review: suspended by standing policy
INTERNAL REVIEW: PASS | FAIL
```

Include actual standards/authorities, evidence, findings, remediation, checks and
limitations. Historical incident references are included only when applicable;
no new external diagnostic is required. Never edit `audits.yaml`, clear an
incident, forge an external envelope or copy external findings into this record.
Disclose the method and loss of independence at handback.

A self-audit satisfies a managed review prerequisite only with current recorded
internal PASS.
It permits the next otherwise-authorized step, not sign-off, lifted holds,
closure, merge or deployment. Tests and spike restrictions remain mandatory.
Changed candidates require reassessment under REVIEW; subsequent stages remain
internal without testing the suspended route.

## Review limits

Managed work follows REVIEW's five-round ceiling per ticket and gate, not a
configuration value that permits additional rounds. Count deliberate completed
reviews, including rejection of an incomplete submitted package. Ordinary author
checks and local pre-submission readiness checks do not consume rounds. Preserve
dated history across skill re-entry, compaction and reviewer/method changes;
changing method does not replenish the gate's allowance or relabel independent
evidence as internal. Keep each record's ownership and provenance intact.

At round five without effective PASS, record the hard gate failure and alert the
operator; do not run round six, reset the count or call the old harness. Standalone
and spike profiles retain their bounded invocation/brief budgets without inventing
managed gates. Transport incidents have their own bounded recovery and are not
semantic review rounds. Infrastructure configuration cannot raise the ceiling.
