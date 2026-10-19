---
name: test-code-audit
description: Review written tests internally against the specification or exploratory intent and testing standards before implementation.
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

Load this skill at the workflow's test-code audit step or on explicit request.
**Run in the current session.** Do not invoke the external harness or launch a
new model, agent or context.

Read `~/.agents/sdlc/AUDITS.md`, `TESTING.md`, `CODING.md`, `ARCHITECTURE.md`,
`SECURITY.md` and relevant technology standards. Reuse unchanged standards
already read in full. Inspect the actual test code, helpers, execution-enabling
scaffolding, relevant existing implementation and recorded test observations.

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

Record all blocking findings together with file/location, evidence and required
correction. Remediate and rerun affected tests, then review the changed package.
Follow `AUDITS.md`'s internal evidence contract: managed work uses a dated
`internal-review.md` entry; spikes use the brief's Findings section; non-SDLC
uses the established project review record or a clearly named local review file.

```text
Review method: internal (current session; not independent)
Gate: test-code
Mode: SDLC | spike | non-SDLC
Revision: <reviewed commit and identified working-tree delta>
External review: suspended by standing policy
INTERNAL REVIEW: PASS | FAIL
```

Include standards, evidence, findings, remediation, checks and limitations.
PASS requires no unresolved blocking findings and approves the tests only.
Continue to otherwise-authorized implementation; do not write `audits.yaml`,
claim independent review or grant delivery approval.
