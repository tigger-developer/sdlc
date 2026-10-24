---
title: Delivery Review Gates
version: 15
last-updated: 2026-10-24
---

# Delivery review gates

This is the delivery agent's gate contract. An explicitly invoked independent
fresh-context auditor stops here and follows `INDEPENDENT-AUDITOR.md`, without loading MAIN
or delivery workflows. Read `READING.md` and reuse retained standards.

## Active invocation policy

Automatic external auditing is suspended until explicitly re-enabled. Delivery
reviews use `definition-audit`, `test-code-audit` and `code-audit` in the current
session through `SELF-AUDIT.md`. These three skills are self-audit only; none
launches or acts as an independent reviewer. Do not call the suspended harness,
probe provider health, reset auditor state or wait for recovery.
These skills are authorized by a required workflow review or an explicit audit
request, not by their mere availability in a non-SDLC project.

An explicitly requested independent audit uses INDEPENDENT-AUDITOR.md directly,
not one of the self-audit skills. A manual fresh session or an isolated native
sub-agent may perform it. This does not re-enable the external audit harness.

### Calling an independent reviewer

Only an explicit independent-audit request or an expressly authorized workflow
permits this route. Ordinary self-audit requirements do not. The caller:

- Starts one fresh reviewer with no inherited author conversation or reasoning.
  Supplies the entry-point path, review purpose/profile, permitted project paths,
  exact candidate, requirements/brief and original evidence. Includes unresolved
  prior findings as evidence, not the desired verdict or a coached conclusion.
- Uses a native isolation control only when available and authorized. If it
  cannot avoid inherited author context, reports that limitation and uses an
  authorized self-review without calling it independent. It does not launch a
  child CLI or bypass a harness restriction to simulate isolation.
- Settles required checks and freezes the candidate before review. The reviewer
  receives read-only scope; the caller owns remediation, records and workflow
  progression. No reviewer launches another reviewer.
- Records actual separation, findings and limitations separately from internal
  review and harness-owned state. Rechecks affected evidence after remediation.
  Re-review uses the same authorized scope and configured review allowance; a
  standalone invocation is one review unless further rounds were authorized.

The reviewer reads none of this caller orchestration. Its complete operating
instructions are in INDEPENDENT-AUDITOR; it selects engineering criteria only.

Read `REVIEW.md` for shared criteria, candidate validity, evidence ownership and
verdict semantics. Read `SELF-AUDIT.md` for internal records, continuation and
review limits. The current contracts supersede historical external invocation
and PASS requirements. Preserve still-current historical external verdicts;
do not repeat a review merely to change its method. A successful provider probe
never re-enables external auditing.

## Composite gates

| Gate | Candidate and prerequisite | Next otherwise-authorized step |
|---|---|---|
| Definition | Complete specification, design and traced AT/RT/UT/OT definitions; shared definition criteria | Operator sign-off, not implementation |
| Test-code | Written AT/RT package and minimal execution scaffolding; intended RED or justified preservation GREEN; recorded OT/UT state | Implementation within signed-off scope |
| Delivery-code | Production delta, affected tests, settled AT/RT/OT GREEN and required current product docs; UT may be AMBER | Operator validation, not closure |

Apply ARCHITECTURE at all gates. Review a coherent package, not each file or
keystroke. Ordinary author checks are not additional review rounds. Before an
implementation review, read the installed validator's `--help` once and run the
matching readiness check under `ORG-SCHEMA.md`. Readiness refusals and historical
valid FAIL findings are not bypassed. Definition has no execution-readiness gate.

If the approved specification has neither AT nor RT, the definition review must
justify OT/UT evidence; skip test-code but retain delivery-code. Never remove or
reclassify approved tests to evade review. A test-code PASS approves tests only;
old unlabelled implementation rounds are not current-stage evidence.

Normal delivery uses the signed-off definition. Paired work records a skeleton
before code and reviews each coherent increment; emergency work fixes first,
records the durable scope/evidence, then reviews. Both may proceed directly to
delivery-code and consolidate a retrospective definition afterwards. Do not
invent advance definition approval or retrospective pre-build test-code gates.
Required product documentation is current before delivery-code in every route;
the later retrospective specification does not defer it.

## Findings, continuation and human boundaries

Address every finding as a batch: remediate valid findings and substantiate
challenges to incorrect or out-of-scope findings. Reassess changed candidates
and rerun affected checks. Do not expand scope to obtain PASS, split known
findings across rounds, silently waive them or replace a reviewer to evade FAIL.
Refer unresolved disputes or behaviour/design changes outside DELIVERY.md's
Hands-off delivery decision authority to the operator. Apply REVIEW's departure
criteria to permitted logged changes; do not require the auditor to load DELIVERY.
Correct retrospective documentary defects within the existing review allowance without
redesigning the delivered behaviour. Report immediate safety risks promptly.

A recorded current internal PASS under SELF-AUDIT satisfies that exact review
prerequisite, including definition admission to operator review. It is not an
independent external PASS. Continue safe authorized work after it; provider
unavailability alone is not a blocker. Stop for a human-only decision that no
safe reversible assumption resolves, an unremediable mandatory failure or the
applicable review limit. Human sign-off, holds, test evidence, scope, closure
and spike non-merge/non-deployment boundaries remain unchanged.

Only the operator approves definitions, accepts departures, authorizes architectural
or major product changes and closes delivered work. DELIVERY.md permits bounded
minor departures during hands-off delivery pending final operator acceptance.
`BYPASS-GATE-7` waives only advance definition
admission, retaining applicable TDD, delivery review and retrospective evidence.

## Evidence ownership and transport

Self-audit records belong in the locations defined by SELF-AUDIT. Manual
fresh-context findings are returned for recording as a separate manual review,
with actual separation, candidate, evidence and limitations; never relabel them
as author self-review. Managed manual-review records live beside the spec, not
in harness-owned `audits.yaml`. Workflow consumption still checks current
readiness and applicable gate criteria before accepting any result.
An operator-requested manual independent PASS that meets those same criteria
satisfies the review prerequisite. Apply and verify any mechanical conditions
before accepting PROVISIONAL PASS. Workflow references to internal PASS describe
the default suspended-harness route; they do not require a duplicate self-audit
of an already accepted manual result. No method replaces human sign-off.

Only the harness writes external findings, verdicts, revisions, round numbers,
remediation and session identities to `audits.yaml`. Other artefacts link that
record, never duplicate its state. Do not clear incidents, forge envelopes or
rewrite external history. Candidate changes invalidate relevant results;
previous verdicts remain historical evidence only for their reviewed candidate.

The preserved invocation, cache, timeout, reset, transport and recovery details
are in `EXTERNAL-AUDIT.md` and `HARNESS.md`, loaded only for authorized transport
maintenance or re-enablement. They are not part of the active review package.
External test-code/delivery-code context reuse and counters remain transport
responsibilities; self-audit never operates those controls.

# Canary

Append " AUDIT" to the governing delivery canary after reading this contract.
This is not a claim of independent review.

## Document history

This revision separates independent review, author self-audit and delivery-only
responsibilities. Earlier rules remain in their canonical owners and Git history.

Version 13 separates independent audit, self-audit and transport from delivery
orchestration. Every prior version-12 provision and its document history is
preserved in EXTERNAL-AUDIT; shared active safeguards now reside in REVIEW and
SELF-AUDIT. The former full-library auditor route is superseded by narrow profiles.
