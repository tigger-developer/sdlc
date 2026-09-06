---
name: audit-implementation
description: Run the combined SDLC v3 implemented-test and code gate with at most two contexts.
---

Read `~/.agents/sdlc/MAIN.md` and `AUDITS.md`, then apply the implementation-gate
contract to the active change.

In this delivery context, apply the `audit-tests` and `audit-code` criteria
together. Remediate, rerun affected automated tests, and repeat for no more than
five local rounds. Record each round in `audits.org`. Do not invoke the focused
skills through separate agents or scripts.

Only after both pass locally, spawn exactly one external Codex subagent. Give it
the signed-off specification, exact implementation delta, test evidence,
project profile, and selected standards. It runs both audits in that single
context and returns one composite verdict. Retain its agent identity. After
remediation, rerun affected tests and local reviews, then send the revision to
that same agent. Never replace it or exceed five failed gate rounds.

On PASS, execute and record pending OT and UT evidence, reconcile documentation,
and present the result for operator closure.
