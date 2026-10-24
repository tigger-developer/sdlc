---
title: Self-Audit in the Current Context
version: 1
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

For managed work retain the effective `delivery.audit.max_rounds` limit
(explicit authorized override, process environment, project, global configuration,
then the schema default of five) separately per work item and gate. Count
completed internal PASS/FAIL reviews in the dated internal record, not ordinary
author checks, readiness refusals, cached reuse or provider incidents. Do not
reset the count on skill re-entry or compaction. Historical external rounds
remain in their own record and are not rewritten or silently replenished.
At the internal limit without PASS, report evidence and obtain explicit operator
authority for further rounds; do not call the harness's reset command. Standalone
and spike profiles retain their bounded invocation/brief budgets without
inventing persistent managed-gate state. Harness failure/retry counters apply
only to external transport and are not self-audit prerequisites.
