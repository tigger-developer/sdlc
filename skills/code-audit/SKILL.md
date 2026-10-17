---
name: code-audit
description: Review implementation and affected tests internally against authorized scope and coding standards; record an internal delivery-code result.
---

The operator may also manually invoke this skill in an already-started fresh
session on any supported agent harness for a separate audit. Review directly
there; never invoke the suspended audit harness or launch another context.
For that explicit manual route, use `Review method: manual fresh-session audit`
instead of the in-session self-review label below, and record the actual reviewer
separation and limitations. Workflow invocation always stays in the current
session. A skill invocation alone does not establish independence.

For the manual fresh-session route, return findings without editing reviewed
artefacts unless remediation is explicitly authorized. In a delivery workflow,
the author may remediate within the existing scope and reassess. Any prohibition
below on claiming independence concerns author self-review; record genuine
reviewer separation accurately for a manual audit.

Load this skill at the workflow's implementation/code-code audit step or on
explicit request. **Run in the current session.** Do not invoke the external
harness or launch a new model, agent or context.

Read `~/.agents/sdlc/AUDITS.md`, `CODING.md`, `TESTING.md`, `ARCHITECTURE.md`,
`SECURITY.md` and relevant technology standards. For applicable documentation,
read `DOCUMENTATION.md` and its format standards. Reuse unchanged standards
already read in full. Inspect the complete production delta, affected tests,
execution evidence and relevant unchanged interfaces, rather than only an
authoring summary or changed lines.

Select the mode explicitly:

- **SDLC ticket:** assess the approved requirements and design. Before review,
  required checks must have finished and passed, required RT/OT records must be
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

Record every blocking finding together, with file/location, evidence and
required correction. Remediate within scope, rerun affected checks, then review
the changed candidate. Follow `AUDITS.md`'s internal evidence contract:

```text
Review method: internal (current session; not independent)
Gate: delivery-code
Mode: SDLC | paired | emergency | spike | non-SDLC
Revision: <reviewed commit and identified working-tree delta>
External review: suspended by standing policy
INTERNAL REVIEW: PASS | FAIL
```

Managed work uses a dated `internal-review.md` entry; spikes use their Findings
section; non-SDLC uses an established or clearly named local review record.
Include standards, evidence, findings, remediation, checks and limitations.
PASS requires no unresolved blocking findings. It permits otherwise-authorized
operator validation, not closure, merge or deployment by itself. Never edit
`audits.yaml` or claim this review was independent.
