---
name: test-code-audit
description: Self-review written tests before implementation in the current author context; never launch or act as an independent auditor.
---

Use only for a workflow-required or explicitly requested self-audit. Review in
the author's current context; do not delegate or select an independent role.
Read ~/.agents/sdlc/SELF-AUDIT.md and REVIEW.md's test-code criteria, with
READING's retained-context reuse. Inspect original candidate evidence rather
than only the authoring summary. This skill never loads the independent-auditor
entry point or launches the suspended audit harness.

Select the mode explicitly:

- **SDLC ticket:** use the approved specification and test definitions. Run the
  `ORG-SCHEMA.md` test-code readiness check using the installed validator. ATs and RTs
  must have observed intended RED or justified preservation GREEN; compilation
  and setup failures do not establish useful RED. OT/UT may remain AMBER.
- **Spike:** use the approved brief and current exploratory intent. Audit after
  the test package is written, accepting the caller's timing declaration. Apply
  all coding/TDD/testing standards without fixed-spec readiness or full delivery
  artefacts. Do not require a definition audit.
- **Non-SDLC:** use supplied project contracts and observable intended behaviour.
  Bypass automatic SDLC readiness and artefact requirements. Preserve any
  separately operator-requested validation without inventing a specification.

Check coverage of specified cases, meaningful behavioural assertions, real
invocation boundaries, determinism, credible doubles, failure/edge cases,
isolation and applicable testing rules. A CLI test must exercise CLI invocation
where that is the specified boundary; phantom helper calls and source greps do
not establish product behaviour. Identify missing implemented tests even if a
validation sheet claims coverage. Distinguish evidence from declarations.

Do not demand the new production implementation or all-GREEN results at this
stage. Minimal scaffolding may make tests executable but must not implement the
target behaviour. Do not require final product documentation at this gate or
reopen approved requirements. Refer material design/scope questions to the
operator rather than silently changing them.

Report every blocking finding together with location, evidence and correction.
Remediate within scope, rerun affected checks and reassess under SELF-AUDIT.
Record an explicitly internal result; never write audits.yaml or claim
independence. PASS approves the test package for otherwise-authorized implementation,
not closure, merge or deployment.
