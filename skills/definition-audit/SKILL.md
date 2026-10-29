---
name: definition-audit
description: Self-review a specification, design and test definitions in the current author context; never launch or act as an independent auditor.
---

Use only for a workflow-required or explicitly requested self-audit. Review in
the author's current context; do not delegate or select an independent role.
Read ~/.agents/sdlc/SELF-AUDIT.md and REVIEW.md's definition criteria, with
READING's retained-context reuse. Inspect original candidate evidence rather
than only the authoring summary. This skill never loads the independent-auditor
entry point or launches the retired audit harness.

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

List every concrete blocking finding together with location, violated criterion,
evidence and correction. Remediate within authority, rerun affected checks and
reassess. Record the actual internal method and result under SELF-AUDIT. Refer
human-only decisions to the operator; never waive findings or expand scope.
A current internal PASS supports operator sign-off, not implementation approval.
Record the result in audits.yaml under AUDIT-RECORD.md; never claim independence.
