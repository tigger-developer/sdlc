---
name: pair-change
description: Run explicitly selected paired development with live operator validation and lean closure records.
---

Read `~/.agents/sdlc/MAIN.md` and `PAIRING.md`, then the project profile and
applicable standards. The operator's bounded objective and explicit iteration
instructions are the working specification. A question remains a question.

For code or script changes, reuse the applicable work item or create its skeleton
before coding. Append authorized requirements and decisions as the session
progresses. Documentation-only changes do not require a ticket.

Iterate implementation with the operator. Use automated TDD wherever justified,
and record each explicit visual, ergonomic, editorial, or operational validation
without replacing it with artificial automation. Audit each coherent code
increment promptly through `--gate delivery-code`, not each keystroke.
Before that review, follow `~/.agents/sdlc/ORG-SCHEMA.md` and run
`~/.agents/sdlc/bin/sdlc-validate --help`, then its delivery-code readiness check.
Every RT/OT must be GREEN; record outstanding UTs as AMBER. This route does not
require a pre-build test-code audit.
Update affected product documentation before that final delivery verdict; do not
require a polished retrospective specification before reviewing the code increment.
When the outcome stabilizes, consolidate `spec.org`, `validation.org`, and
affected documentation, then run the retrospective definition review.

Follow `~/.agents/sdlc/AUDITS.md` for the `Workflow: paired` audit-context marker,
separate retained gate sessions, and remediation limits. Refer proposed behaviour
or design changes to the operator for normal delivery; do not redesign during
wrap-up. Ask the operator to confirm the recorded user tests and close the work.
