---
name: definition-audit
description: Review a specification, test definitions and solution design internally in the current session; record findings and an internal result.
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

Load this skill as the replacement for a workflow's definition audit, or on an
explicit request. **Run in the current session.** Do not invoke the external
harness or launch a new model, agent or context. A separate context is started
only by the operator's manual route, never by this skill.

Read `~/.agents/sdlc/AUDITS.md`, `ISSUES.md`, `ARCHITECTURE.md`, `TESTING.md`,
`CODING.md`, `SECURITY.md` and relevant technology/format standards. Read the
project profile and its named authorities, the complete specification and any
processed-feedback evidence. Reuse standards already read in full and unchanged
in this context. Do not rely on the authoring summary alone.

Review the whole package:

- Requirements are bounded, observable, falsifiable, internally consistent and
  faithful to the authorized outcome. Scope exclusions and edge cases are clear.
- Every requirement has credible, traced verification. Test definitions specify
  actions and observable expectations at the real product boundary; a CLI is
  tested through CLI invocation. No source-grep substitutes for behaviour.
- Solution design resolves significant responsibilities, interfaces, state,
  failure handling and security decisions sufficiently for another agent to
  deliver. It fits project architecture without unnecessary machinery.
- Check scope expansion, over-engineering and template inflation. Optional
  improvements are not blocking findings. Do not invent requirements.
- Apply specification schema, title, Synopsis/detail consistency, traceability,
  annotation resolution and documentation rules. Do not demand execution results
  for tests that are only being defined.
- For paired/emergency retrospective review, assess the recorded authorized
  outcome and actual design; do not invent an advance-approval requirement.

List every concrete blocking finding together, with file/section, violated
requirement/standard or demonstrated defect, evidence and required correction.
Separate optional suggestions and unresolved human decisions. Remediate within
authorized scope, then reassess the changed package; do not silently waive
findings or redesign beyond authority.

Record under `AUDITS.md`'s internal evidence contract, normally in
`internal-review.md` beside the specification:

```text
Review method: internal (current session; not independent)
Gate: definition
Revision: <reviewed commit and identified working-tree delta>
External review: suspended by standing policy
INTERNAL REVIEW: PASS | FAIL
```

Include reviewed authorities/standards, findings, remediation, checks and
limitations. PASS requires no unresolved blocking findings. This result permits
the workflow's operator sign-off step; it is not operator approval. Do not
write `audits.yaml`, claim independent clearance or approve implementation.
