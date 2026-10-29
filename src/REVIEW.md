---
title: Shared Review Criteria and Profiles
version: 7
last-updated: 2026-10-29
---

# Shared review criteria

Independent audit and self-audit apply these same criteria. This document does
not select a reviewer, authorize edits or start delivery. Follow `READING.md`;
do not load `MAIN.md`, `AUDITS.md` or transport instructions through this route.

## Select the narrow review package

| Purpose/profile | Required standards and evidence |
|---|---|
| Definition | `ISSUES.md`, `ARCHITECTURE.md`, `TESTING.md`, `CODING.md`, `SECURITY.md`, `DOCUMENTATION.md`; complete specification, named authorities and processed feedback |
| Test-code | `CODING.md`, `TESTING.md`, `ARCHITECTURE.md`, `SECURITY.md`, `GIT.md`; actual tests, helpers, execution-enabling scaffolding and observations |
| Delivery-code | Test-code standards plus `DOCUMENTATION.md`; complete production delta, affected tests, settled execution evidence and required product documentation |
| Standalone | Coding/testing, architecture, security, Git and documentation criteria; existing project contracts, not invented SDLC requirements |
| Spike | Coding/testing, architecture, security and Git criteria; approved brief and exploratory evidence, not product delivery artefacts |

Add technology standards selected by the project profile and those demonstrated
by the evidence. Include prerequisite standards such as JAVASCRIPT for NODE
and WEB for HUGO. Inspect relevant authorities if selection is ambiguous; never
omit a demonstrated technology merely because its profile declaration is absent.
For reviewed Markdown or Org documents, select their format standard; managed
specification/validation records also select `ORG-SCHEMA.md`. Format checks on
a spike brief do not create a product-documentation obligation. Add further
domain standards only where the evidence makes them applicable. All selected
standards must be read in full, with unchanged retained content reused.

A standard may describe delivery actions. Review their resulting evidence only
where the profile requires it; do not execute that workflow as an auditor.
Standalone review requires no work item, spec, gate/approval record, validation
record, test colour, traceability matrix or closure. Spike review requires no
definition gate, fixed-spec readiness, full workflow or product documentation.
Neither profile authorizes merge or deployment. Missing context is reported,
not replaced with invented requirements or a full PASS.

## Reject incomplete submissions

Before substantive review, independently establish that the supplied candidate
and evidence meet the named profile's prerequisites below. Reject an incomplete
submission with FAIL, identifying the missing prerequisites; do not become the
author's next-step planner, implement missing work or grant provisional success
for an unfinished package. Completeness is phase-specific: definition needs no
implementation, test-code needs no target production behaviour or final product
docs, and delivery-code may leave human UTs pending. A complete package may still
contain defects. A requester claim alone does not establish readiness.

## Candidate, readiness and review boundaries

- Assess the authorized scope and existing architecture, not a preferred
  redesign. Paired/emergency work uses recorded operator-authorized scope;
  do not demand an advance definition gate or retrospective pre-build review.
- Definition requires at least one meaningful UT under TESTING.md; a ceremonial
  check of automated results does not satisfy it.
- Definition reviews requirements, solution design and AT/RT/UT/OT definitions
  together. Do not demand execution results at that stage. Assess scope,
  over-engineering and template inflation, not length or aesthetic preferences.
- Managed test-code requires the specified AT/RT package to have executed with
  intended RED or justified preservation GREEN; OT/UT may be AMBER. Minimal
  scaffolding may enable execution but must not implement target behaviour.
  Compilation/setup failures do not establish RED. Final product docs and new
  production implementation are not prerequisites at this stage.
- Managed delivery-code requires finished, passing checks, including applicable
  lint/vulnerability checks, and current AT/RT/OT GREEN. UTs may remain AMBER.
  No checks run concurrently with the review or become GREEN before finishing.
  Required product docs must be current before this review, including in
  paired/emergency work; retrospective specification consolidation may follow.
- For managed implementation reviews, require the matching validator readiness
  result under `ORG-SCHEMA.md`. A restricted reviewer assesses supplied output;
  the author runs the validator. Definition has no execution-readiness check.
  Readiness proves record structure, not truth, completeness or test quality.
- Review the complete candidate and relevant unchanged interfaces. Brownfield
  evidence includes named requirements/design, history, maintained regressions
  and affected implementation. Report material conflicts or omissions; do not
  copy the whole baseline into the change.
- Tests must observe the real public boundary. Source greps, declarations,
  phantom helper calls and assertions mirroring implementation are not product
  evidence. Passing tests alone establish neither coverage nor design fit.
- Candidate inputs freeze for review. Changed code, tests, docs or validation
  invalidate the result for those changed inputs. Current affected checks and
  settled evidence are required before reassessment; preserve earlier results as history.
- `audits.yaml` is the sole canonical audit record under `AUDIT-RECORD.md`,
  for self-audit and independent review. Do not duplicate findings, verdicts or
  counters in specifications or parallel review documents. Independent reviewers
  return findings to the authorized caller for recording. Preserve provenance
  and unresolved findings; never fabricate separation or erase earlier results.
- Definition fails for any semantically identified annotation record/block,
  even resolved or empty, in the spec. Use the annotation tool's actual format,
  not keyword matching. Missing/conflicting dispositions in `feedback.yaml`
  also fail. Removing annotations without preserved resolution is not a remedy.
  This applies to retrospective definitions too.

## Assessing recorded hands-off departures

For hands-off delivery of an agreed specification only, assess the original
approved design and cumulative decision log. A localized reversible departure
may resolve an evidenced implementation blocker while preserving required
outcomes and architecture. Difficulty, effort, convenience, preference or budget
pressure alone do not justify it. Require investigated conforming alternatives,
consequences, verification and an explicit pending operator disposition.
Architectural changes and major product decisions are outside that permission;
security, access, irreversible-data and spending restrictions still apply.
Assess combined effects against the original baseline. Do not reject an otherwise
permitted departure solely because it differs from that baseline, or award PASS
when its authority or evidence is missing. Reviewer acceptance of compliance is
not human acceptance of the decision. Do not load delivery workflow to assess it.

For spikes, assess the agreed stack and architecture in the brief. Packages within
that stack may support exploration; new runtimes, languages, package/build
ecosystems, databases or services require prior approved amendment, including
when introduced transitively by a package. A non-merge branch does not waive
these boundaries. Neither profile grants the other profile's permissions.

## Five-round limit per ticket and gate

For managed work, each ticket has a separate allowance of at most five audit
rounds for each gate: define (definition), write-tests (test-code) and deliver
(delivery-code). The initial assessment is round one, leaving at most four
further assessments. Counts do not carry between gates or between tickets.
They do persist across sessions, reviewer changes, candidate revisions, retries
and re-entry into the same gate; returning to a gate does not replenish it.

A round is a submitted candidate receiving a substantive assessment, including
an auditor's rejection of an incomplete submission. It is not an agent message
or tool call. Ordinary author checks, local readiness refusals that prevent
submission, and transport failures without an assessment are not semantic audit
rounds. PROVISIONAL PASS, where allowed, consumes a round; its exact mechanical
conditions must be verified before it is effective. A new substantive assessment
is another round, not a free clarification or renamed verification.

If there is no effective PASS at the fifth round, halt autonomous progression
and alert the operator with the ticket, gate, candidate, findings and history.
An unattended agent must not start another round or reset its own allowance.
Changing reviewers, methods, candidates or ticket labels does not reset it.

In an interactive session, the operator may request self-audit or independent
external audit and authorize a reset. Explicit direction to continue auditing
after being informed of exhaustion is reset authority; no special token or
maintainer change is required. Record that instruction before the next review.
A reset starts a new five-round cycle for the named ticket and gate only.
Retain all previous cycles and unresolved findings. It neither awards PASS nor
waives criteria, readiness, holds, sign-off or independence requirements.
A generic audit request without exhaustion/reset context does not silently
reset an exhausted count. For unattended work, record explicit operator reset
authority before resuming; infrastructure cannot replenish rounds on its own.

Use `AUDIT-RECORD.md` for the structured record and reset provenance. The
five-round limit bounds autonomous work; it is not a prohibition on live
operator-directed review. External audit means an explicitly invoked independent
fresh-context reviewer, not the retired audit harness.

The five-round ceiling is process policy, not an infrastructure override. The
implementing project may choose timeouts and bounded transport recovery, which
must not create a semantic verdict or replenish review rounds. Exhausted recovery
also requires an actionable operator alert; never retry indefinitely because no
assessment has arrived. Standalone reviews and spikes retain their separately
authorized bounds without inventing these managed-ticket gates.

## Findings and verdicts

Report all concrete blocking findings together: location, violated requirement,
decision or standard (or demonstrated defect), evidence and required correction.
Separate optional improvements and human-only decisions. No scope expansion to
obtain PASS, silent waivers, fabricated evidence or unsupported causal claims.
Review security, state ownership, failure handling, concurrency, compatibility,
dependency risk and maintainability where relevant.

Evaluate evidence-backed challenges and explicitly withdraw unsupported or
out-of-scope findings; explain why retained findings remain valid. A challenge
alone is not remediation. Do not split known findings across rounds, omit
evidence to evade a finding or seek another reviewer merely for a better result.
Refer unresolved scope/design disputes, departures outside the bounded permission
above and immediate safety risks to the operator.

PASS requires no unresolved blocking findings for the reviewed candidate.
PROVISIONAL PASS is allowed only for work under live operator supervision: a
human is present in the session and supervising the work, regardless of who
initiated the build or audit. Both requester and auditor must confirm this status.
The requester supplies it; the auditor acknowledges it when issuing the verdict.
Missing, ambiguous or ended supervision means provisional success is unavailable.

Without live supervision, only PASS or FAIL semantic verdicts are permitted. The
auditor returns FAIL for a would-be PROVISIONAL PASS and lists the outstanding
corrections and verification. The requester independently rejects any provisional
result received without current supervision as FAIL, preserving the original
response and recording the effective disposition. Never advance on it or promote
it automatically. Under live supervision, exact mechanical conditions must still
be applied and verified before effective PASS. Provisional review is not human
acceptance or merge authority. Timeout, silence, malformed output, unavailable
providers or incomplete access are never PASS. Review provides evidence, not
human approval, a lifted hold or authority to close work.
