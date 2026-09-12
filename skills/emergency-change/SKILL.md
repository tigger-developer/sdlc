---
name: emergency-change
description: Execute an operator-authorized BYPASS-GATE-7 emergency change and reconcile durable evidence afterwards.
---

This skill is usable only when the operator includes the exact token
`BYPASS-GATE-7` in the same request. Never infer, suggest, or self-authorize it.

Read `~/.agents/sdlc/MAIN.md` and `EMERGENCY.md`, then the project profile and
applicable standards. Treat the bounded request as the temporary specification.
Select the required RT, UT, and OT evidence. For every justified automated test,
write it first and observe RED. Implement, observe GREEN, and run immediate
checks. Then reuse or create the durable work item: its ID, directory, bounded
request, actual change, and evidence must exist before the implementation audit.
Do not delay the fix for ticket creation or a definition audit. Documentation-only
changes do not require a ticket.

Audit implemented tests and production code together first. Afterwards, reconcile
the requirements, test definitions, solution design, validation, and affected
documentation, then run the retrospective definition review. Follow
`~/.agents/sdlc/AUDITS.md` with `Workflow: emergency` for audit-context markers,
separate retained gate sessions, and remediation limits. Refer behaviour or
design changes to the operator for normal delivery; correct documentary errors
without inventing advance approval. Present the evidence for operator closure.
