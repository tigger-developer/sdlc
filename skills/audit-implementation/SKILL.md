---
name: audit-implementation
description: Run the combined SDLC v3 implemented-test and code gate with at most two contexts.
---

Read `~/.agents/sdlc/MAIN.md` and `AUDITS.md`, then apply the implementation-gate
contract to the active change.

Before the first external invocation in this session, run
`~/.agents/sdlc/bin/sdlc-harness --help` and follow its current start/resume,
stdin, evidence-input, output-channel, and `SESSION_ID` instructions.

In this delivery context, apply the `audit-test-code` and `audit-code` criteria
together. Remediate **all** findings from a failed round before repeating;
never audit an unchanged candidate or remediate findings piecemeal. Rerun
affected automated tests and repeat for no more than five local rounds in this
invoking session; do not carry the count over from an earlier session. Record
each round in `audits.org`. Do not invoke the focused skills through separate
agents or scripts.

Only after both pass locally, start exactly one external context through
`~/.agents/sdlc/bin/sdlc-harness start --phase audit`. Supply the signed-off
specification, exact implementation delta, test evidence, project profile, and
selected standards as exact `--input` files; send the bounded audit instruction
on stdin. The runner resolves the harness, provider where meaningful, model,
and timeout from project and global YAML configuration. Record the `SESSION_ID`
it reports. The external context runs both audits in one turn. After addressing
**all** findings, rerun affected tests and local reviews, then send the revision to
that same context with `~/.agents/sdlc/bin/sdlc-harness resume --phase audit`
and `--session SESSION_ID`. Never replace it or exceed five failed gate rounds. If the
configured timeout expires, interrupt that retained agent, record a runner
incident, and return it for operator direction without treating it as a failed
audit or creating a replacement.

On PASS, execute and record pending OT and UT evidence, reconcile documentation,
and present the result for operator closure.
