---
name: code-audit
description: Review implementation and affected tests; use self-audit in delivery or findings-only independent audit when explicitly invoked in a fresh context.
---

Invoke at the required delivery-code step or on explicit request. An explicitly
invoked independent fresh-context auditor reads `~/.agents/sdlc/AUDIT.md`;
the author uses `~/.agents/sdlc/SELF-AUDIT.md`. Neither path loads MAIN or
launches another reviewer. Apply READING's reuse rule and REVIEW's delivery-code
profile, narrowed appropriately for spikes. Inspect the complete production
delta, affected tests, execution evidence and relevant unchanged interfaces,
not only an author summary or changed lines.

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
coverage, dependency risks, maintainability and approved design conformance.
Check state ownership, failure handling, concurrency and compatibility where
relevant. Passing tests do not prove design conformance or adequate coverage.
Do not use source-grep tests as behavioural evidence.

Identify unauthorized scope or unnecessary complexity in the delta without
reopening approved design choices. A blocking finding needs a concrete defect
or violated requirement/decision/standard. Optional refactoring and stylistic
preferences are suggestions, not reasons to fail. Report insufficient evidence
and material unresolved decisions honestly.

Report every blocking finding together with location, evidence and correction.
Independent auditors return findings only under AUDIT. Self-auditing authors
may remediate within scope, rerun affected checks, reassess and record under
SELF-AUDIT. PASS supports otherwise-authorized operator validation, not closure,
merge or deployment. Never edit `audits.yaml` or label self-review independent.
