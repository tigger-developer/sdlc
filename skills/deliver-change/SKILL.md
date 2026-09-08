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

At the beginning of delivery, determine whether any information is required
from the operator before implementation can safely start. Ask that one bounded
set of questions once. If the operator deliberately selects `ATTENDED`, remain
attended and hand back when a human decision is required. Otherwise delivery
enters **HANDS-OFF** for this invocation. Hands-off mode is not a global
default: it is activated only by invoking this delivery workflow, and it does
not authorize changes to signed-off behaviour or other human-only decisions.

Hands-off delivery is a continuation contract. The cost of stopping
prematurely is high, so a progress report, commit, warning, completed test,
audit result, or partial milestone is not a terminal handback while safe,
authorized, executable work remains. Continue through TDD, remediation,
documentation, validation, and evidence recording. Record every material
routine assumption in the specification's `Decision Log (Hands-off delivery
only)` section.

Valid reasons to stop include a required human decision or genuine blocker, a
failed mandatory test or verification that cannot be safely remediated, or
reaching the configured maximum number of audit rounds without `PASS` or
`PROVISIONAL`. In those cases return the exact state, evidence, assumptions,
and next human action; do not imply completion.

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
