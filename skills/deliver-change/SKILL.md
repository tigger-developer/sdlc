---
name: deliver-change
description: Deliver a signed-off SDLC v3 specification through automated TDD and recorded validation.
---

Read `~/.agents/sdlc/MAIN.md`, `TESTING.md`, `CODING.md`, `DOCUMENTATION.md`,
`GIT.md`, the project profile, selected standards, the signed-off `spec.org`, and
its current audit evidence.

Confirm the definition gate and operator sign-off are current. Write every
justified automated regression test before production code and observe the
expected RED result. Implement the smallest coherent solution, observe GREEN,
and keep the specification synchronized if implementation reveals an authorized
routine detail. Do not change signed-off behaviour without operator authority.

Run the implementation gate before final OT or UT execution. Record all results
in `validation.org`, update every affected project document, and return the
change for operator closure. Do not claim closure yourself.
