---
name: test-code-audit
description: Review written tests before implementation; use self-audit in delivery or findings-only independent audit when explicitly invoked in a fresh context.
---

Invoke at the required test-code step or on explicit request. An explicitly
invoked independent fresh-context auditor reads `~/.agents/sdlc/AUDIT.md`;
the author uses `~/.agents/sdlc/SELF-AUDIT.md`. Neither path loads MAIN or
launches another reviewer. Apply READING's reuse rule and REVIEW's test-code
profile. Inspect actual tests, helpers, execution-enabling scaffolding, relevant
existing implementation and recorded observations.

Select the mode explicitly:

- **SDLC ticket:** use the approved specification and test definitions. Run the
  `ORG-SCHEMA.md` test-code readiness check using the installed validator when
  authorized to execute it, or inspect its supplied result as a restricted
  independent auditor. ATs and RTs
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
Independent auditors return findings only under AUDIT. Self-auditing authors
may remediate within scope, rerun affected tests, reassess and record under
SELF-AUDIT. PASS approves tests only; the author may continue otherwise-authorized
implementation, not claim delivery approval. Never write `audits.yaml` or label
self-review independent.
