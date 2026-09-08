---
name: audit-definition
description: Run the combined SDLC v3 specification, design, and test-definition gate with at most two contexts.
---

Read `~/.agents/sdlc/MAIN.md` and `AUDITS.md`, then apply the definition-gate
contract to the active `spec.org`.

Before the first external invocation in this session, run
`~/.agents/sdlc/bin/sdlc-harness --help` and follow its current start/resume,
stdin, evidence-input, output-channel, and `SESSION_ID` instructions.

In this authoring context, apply the `audit-spec`, `audit-design`, and
`audit-test-definitions` criteria together. Remediate **all** findings from a
failed round before repeating; never audit an unchanged candidate or remediate
findings piecemeal. Repeat for no more than five local rounds in this invoking
session; do not carry the count over from an earlier session. Record each round
in `audits.org`. Do not invoke the three focused skills through separate agents
or scripts.

Only after all three pass locally, start exactly one external context through
`~/.agents/sdlc/bin/sdlc-harness start --phase audit`. Supply the specification,
project profile, named authorities, applicable standards, and combined audit
criteria as exact `--input` files; send the bounded audit instruction on stdin.
The runner resolves the harness, provider where meaningful, model, and timeout
from project and global YAML configuration. Record the `SESSION_ID` it reports.
The external context runs all three audits in one turn and returns one composite
verdict. If remediation is required, address **all** findings, rerun affected
local reviews, and send the revised
candidate to `~/.agents/sdlc/bin/sdlc-harness resume --phase audit` with
`--session SESSION_ID` using that same retained context. Never replace it or
exceed five failed gate rounds. If the configured timeout expires, interrupt
that retained agent, record a runner incident, and return it for operator
direction without treating it as a failed audit or creating a replacement.

On effective PASS, set the specification's definition-gate status to `PASS` and
move its `docs/work.org` item to `REVIEW`, then present the specification and
audit evidence for operator sign-off. After explicit sign-off, record the
authority and date and move its work item to `ACTIVE`. Do not begin
implementation without that recorded sign-off.
