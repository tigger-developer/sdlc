---
name: emergency-change
description: Execute an operator-authorized BYPASS-GATE-7 emergency change and reconcile durable evidence afterwards.
---

This skill is usable only when the operator includes the exact token
`BYPASS-GATE-7` in the same request. Never infer, suggest, or self-authorize it.

Read `~/.agents/sdlc/MAIN.md` and `EMERGENCY.md`, then the project profile and
applicable standards. Treat the bounded request as the temporary specification.
Select the required RT, UT, and OT evidence. For every justified automated test,
write it first and observe RED. Implement, observe GREEN, and run the combined
implementation gate.

Afterwards, reconcile the actual requirement, test definitions, solution design,
validation evidence, and affected project documentation into the normal v3
artefacts. Do not impose a retrospective definition gate. Present the evidence
and anything left unresolved for operator closure.
