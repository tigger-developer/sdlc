---
name: audit-definition
description: Run the combined SDLC v3 specification, design, and test-definition gate with at most two contexts.
---

Read `~/.agents/sdlc/MAIN.md` and `AUDITS.md`, then apply the definition-gate
contract to the active `spec.org`.

In this authoring context, apply the `audit-spec`, `audit-design`, and
`audit-test-definitions` criteria together. Remediate and repeat for no more
than five local rounds. Record each round in `audits.org`. Do not invoke the
three focused skills through separate agents or scripts.

Only after all three pass locally, spawn exactly one external Codex subagent.
Give it the specification, project profile, named authorities, applicable
standards, and the combined audit criteria. It runs all three audits in that
single context and returns one composite verdict. Use the audit model and
timeout resolved from the project and global YAML configuration. Retain its
agent identity. If
remediation is required, rerun affected local reviews and send the revised
candidate back to that same agent with a follow-up task. Never replace it or
exceed five failed gate rounds. If the configured timeout expires, interrupt
that retained agent, record a runner incident, and return it for operator
direction without treating it as a failed audit or creating a replacement.

On effective PASS, set the specification's definition-gate status to `PASS` and
move its `docs/work.org` item to `REVIEW`, then present the specification and
audit evidence for operator sign-off. After explicit sign-off, record the
authority and date and move its work item to `ACTIVE`. Do not begin
implementation without that recorded sign-off.
