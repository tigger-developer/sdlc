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

Only after all three pass locally, resolve the canonical executable
`~/.agents/sdlc/bin/sdlc-harness` to its absolute filesystem path and start
exactly one external context through that path (never rely on `PATH`). Supply
the specification,
project profile, named authorities, applicable standards, and combined audit
criteria as exact `--input` files; send the bounded audit instruction on stdin.
The runner resolves the harness, provider where meaningful, model, and timeout
from project and global YAML configuration. Record the `SESSION_ID` it reports.
The external context runs all three audits in one turn and returns one composite
verdict. If
remediation is required, rerun affected local reviews and send the revised
candidate to that same absolute `sdlc-harness` path with `resume --phase audit`
and `--session SESSION_ID` using that same retained context. Never replace it or
exceed five failed gate rounds. If the configured timeout expires, interrupt
that retained agent, record a runner incident, and return it for operator
direction without treating it as a failed audit or creating a replacement.

On effective PASS, set the specification's definition-gate status to `PASS` and
move its `docs/work.org` item to `REVIEW`, then present the specification and
audit evidence for operator sign-off. After explicit sign-off, record the
authority and date and move its work item to `ACTIVE`. Do not begin
implementation without that recorded sign-off.
