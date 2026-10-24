---
name: code-audit
description: Self-review implementation and affected tests in the current author context; never launch or act as an independent auditor.
---

Use only for a workflow-required or explicitly requested self-audit. Review in
the author's current context; do not delegate or select an independent role.
Read ~/.agents/sdlc/SELF-AUDIT.md and REVIEW.md's delivery-code criteria, with
READING's retained-context reuse. Inspect original candidate evidence rather
than only the authoring summary. This skill never loads the independent-auditor
entry point or launches the suspended audit harness.

Select the mode explicitly:

- **SDLC ticket:** assess the approved requirements and design. Before review,
  required checks must have finished and passed, required AT/RT/OT records must be
  GREEN and the installed validator's delivery-code readiness check must pass.
  UTs may remain AMBER for the operator. Review affected product documentation.
- **Paired/emergency:** use recorded operator-authorized scope without imposing
  an advance definition gate. Required product documentation is current; the
  retrospective specification may follow under the governing workflow.
- **Spike:** use the brief and current exploratory intent after implementation
  is written, accepting the caller's timing declaration. All coding and testing
  standards apply, without full delivery gates/documentation requirements. No
  definition audit, merge or deployment is authorized by this review.
- **Non-SDLC:** assess supplied project context and coding/testing standards;
  bypass automatic SDLC readiness and spec/work/validation artefact requirements.
  Do not substitute invented requirements for missing context.

Review implemented tests first, then production code. Check correctness,
security, regressions, real invocation boundaries, meaningful assertions,
coverage, dependency risks, maintainability and approved design conformance,
including REVIEW's criteria for recorded hands-off departures.
Check state ownership, failure handling, concurrency and compatibility where
relevant. Passing tests do not prove design conformance or adequate coverage.
Do not use source-grep tests as behavioural evidence.

Identify unauthorized scope or unnecessary complexity in the delta without
reopening approved design choices. A blocking finding needs a concrete defect
or violated requirement/decision/standard. Optional refactoring and stylistic
preferences are suggestions, not reasons to fail. Report insufficient evidence
and material unresolved decisions honestly.

Report every blocking finding together with location, evidence and correction.
Remediate within scope, rerun affected checks and reassess under SELF-AUDIT.
Record an explicitly internal result; never write audits.yaml or claim
independence. PASS supports otherwise-authorized operator validation,
not closure, merge or deployment.
