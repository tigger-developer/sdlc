---
name: deliver-change
description: Deliver a signed-off SDLC v3 specification through automated TDD and recorded validation.
---

Read `~/.agents/sdlc/MAIN.md`, `TESTING.md`, `CODING.md`, `DOCUMENTATION.md`,
`GIT.md`, the project profile, selected standards, the signed-off `spec.org`, and
its current audit evidence.

Admit delivery only when its item is `ACTIVE` in `docs/work.org` and `spec.org`
records definition-gate `PASS`, explicit operator sign-off, and linked audit
evidence that remains current. Do not begin by rerunning definition audits. A
material specification change resets the definition gate and sign-off to
`PENDING` and moves the work item to `REVIEW`.

After admission, apply the implementation branch strategy from the project and
global configuration. Under `current`, remain on the operator-selected branch.
Under `feature`, pull the project's recorded primary branch, create and publish
one project-convention feature branch for this change, then deliver there. Do
not create the feature branch during definition or before checking admission.
Never guess a primary branch, remote, or naming convention.

Write every justified automated regression test before production code and
observe the expected RED result. Implement the smallest coherent solution,
observe GREEN, and keep the specification synchronized if implementation
reveals an authorized routine detail. Do not change signed-off behaviour without
operator authority.

Apply the complete implementation gate from `AUDITS.md` in this context: review
and remediate test code and production code locally for at most five rounds,
then use one retained external context for the composite independent audit. Do
not invoke focused auditors through separate agents or scripts. After effective
PASS, execute final OT and UT checks, record all results in `validation.org`,
update every affected project document, and return the change for operator
closure. Do not claim closure yourself.
