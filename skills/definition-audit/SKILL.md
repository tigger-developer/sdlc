---
name: definition-audit
description: Review a specification, test definitions and solution design; use self-audit in delivery or findings-only independent audit when explicitly invoked in a fresh context.
---

Invoke only for a required definition review or an explicit request. Select the
method before reading standards: an explicitly invoked independent fresh-context
auditor reads `~/.agents/sdlc/AUDIT.md`; a delivery author reads
`~/.agents/sdlc/SELF-AUDIT.md`. Neither skill path loads MAIN. Never launch a
reviewer or invoke the suspended harness. Use READING's retained-context reuse
and REVIEW's definition profile, including the complete specification, project
authorities and processed feedback, not the authoring summary alone.

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
Separate optional suggestions and unresolved human decisions. An independent
auditor returns findings only. A self-auditing author may remediate within
authorized scope and reassess; neither may waive findings or redesign beyond
authority.

Return the definition result under AUDIT or record it under SELF-AUDIT, using
the actual method and shared verdict criteria. It supports the workflow's
operator sign-off step, not approval or implementation. Do not write
`audits.yaml` or describe self-review as independent clearance.
