---
title: Shared Review Criteria and Profiles
version: 1
last-updated: 2026-10-24
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

## Candidate, readiness and review boundaries

- Assess the authorized scope and existing architecture, not a preferred
  redesign. Paired/emergency work uses recorded operator-authorized scope;
  do not demand an advance definition gate or retrospective pre-build review.
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
  invalidate the result for those changed inputs. Rerun affected checks and
  settle evidence before reassessment; preserve earlier results as history.
- Harness audit findings/state belong only to harness-owned `audits.yaml`.
  Duplicated audit state in reviewed artefacts is a blocking finding. Internal
  review evidence is separately owned and links, rather than copies, external
  findings. Manual fresh-context results are returned under AUDIT for separate
  recording, never inserted into harness-owned state. Never forge an external
  verdict or erase unresolved findings.
- Definition fails for any semantically identified annotation record/block,
  even resolved or empty, in the spec. Use the annotation tool's actual format,
  not keyword matching. Missing/conflicting dispositions in `feedback.yaml`
  also fail. Removing annotations without preserved resolution is not a remedy.
  This applies to retrospective definitions too.

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
Refer unresolved scope/design disputes and immediate safety risks to the operator.

PASS requires no unresolved blocking findings for the reviewed candidate.
Independent PROVISIONAL PASS is effective only after its exact mechanical
conditions are applied and verified. Self-audit records PASS or FAIL, resolving
mechanical conditions before PASS. Timeout, silence, malformed output, unavailable
providers or incomplete access are never PASS. Review provides evidence, not
human approval, a lifted hold or authority to close work.
