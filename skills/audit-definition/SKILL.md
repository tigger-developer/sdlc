---
name: audit-definition
description: Run the combined SDLC v3 specification, design, and test-definition gate with at most two contexts.
---

Read `~/.agents/sdlc/MAIN.md` and `AUDITS.md`, then apply the definition-gate
contract to the active `spec.org`.

In this authoring context, apply the `audit-spec`, `audit-design`, and
`audit-tests` criteria together. Remediate and repeat for no more than five
local rounds. Record each round in `audits.org`. Do not invoke the three focused
skills through separate agents or scripts.

Only after all three pass locally, spawn exactly one external Codex subagent.
Give it the specification, project profile, named authorities, applicable
standards, and the combined audit criteria. It runs all three audits in that
single context and returns one composite verdict. Retain its agent identity. If
remediation is required, rerun affected local reviews and send the revised
candidate back to that same agent with a follow-up task. Never replace it or
exceed five failed gate rounds.

On PASS, present the specification and audit evidence for operator sign-off.
Do not begin implementation without that sign-off.
