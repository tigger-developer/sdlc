---
name: deliver-change
description: Deliver a signed-off SDLC v3 specification through TDD and validation, using native goal-backed hands-off continuation where available.
---

Read `~/.agents/sdlc/MAIN.md`, `TESTING.md`, `CODING.md`, `DOCUMENTATION.md`,
`GIT.md`, the project profile, selected standards, the signed-off `spec.org`, and
its current audit evidence.

Admit delivery only when its item is `ACTIVE` in `docs/work.org` and `spec.org`
records definition-gate `PASS`, explicit operator sign-off, and linked audit
evidence that remains current. Do not begin by rerunning definition audits. A
material specification change resets the definition gate and sign-off to
`PENDING` and moves the work item to `REVIEW`.

In **HANDS-OFF** mode, a user-facing progress report is not a permitted
response boundary. Do not end the turn after reporting progress, checkpoints,
warnings, or partial completion. Continue with the next executable action in
the same invocation. Ending the turn while executable work remains is a
HANDS-OFF contract violation.

At the beginning of delivery, determine whether any information is required
from the operator before implementation can safely start. Ask that one bounded
set of questions once. If the operator deliberately selects `ATTENDED`, remain
attended and hand back when a human decision is required. Otherwise delivery
enters **HANDS-OFF** for this invocation. Declare `DELIVERY MODE: HANDS-OFF` and
advise the operator that any interruption or resumed interaction changes the
mode to **ATTENDED** until the operator invokes `HANDS-OFF` again. Hands-off
mode is not a global default: it is activated only by invoking this delivery
workflow, and it does not authorize changes to signed-off behaviour or other
human-only decisions.

On entering HANDS-OFF, read
`~/.agents/skills/deliver-change/references/native-goal.md` and resolve its
budgets before implementation. Activate the current harness's native goal
control where exposed and authorized; verify its result before announcing
`NATIVE GOAL: ACTIVE`. Do not treat a configured budget as an enforced budget.
If native activation needs an operator command, give that command during the
preflight; disclose the limitation rather than launching another agent or
silently substituting instructions for native continuation. ATTENDED skips
goal activation. Goal completion means ready for the agreed operator gate,
not permission to close the work item.

HANDS-OFF is the autonomous coding state. A final response, approval request,
decision request, blocker handback, or other transfer of control is prohibited
while safe, authorized, executable work remains. Progress and status reports
are non-terminal; after reporting one, continue with the next executable item.
An in-scope correction or instruction leaves HANDS-OFF active: incorporate it
and continue.

Hands-off delivery is a continuation contract. The cost of stopping
prematurely is high, especially when the operator may be unavailable for the
delivery window. A question, uncertainty, or routine interpretation is not a
terminal handback: make a sensible, reversible assumption within the signed-
off specification, design, standards, and safety boundaries and record it in
the specification's `Hands-off mode decision log` section. Add one level-two
Org subtree for this invocation, headed by its exact ISO 8601 timestamp and
tagged `:hands-off:review:`; put each material decision at level three beneath
that heading. A
progress report, commit, warning, completed test, audit result, or partial
milestone is likewise not terminal while safe, authorized, executable work
remains. Continue through TDD, remediation, documentation, validation, and
evidence recording.

Assume that no human is available during the hands-off delivery window. Do not
pause to wait for an answer; continue until a permitted stopping condition is
actually reached.

Before any final handback, re-read the signed-off specification, completion
state, decision log, audit evidence, and validation record. An unchecked item
with an executable next action means delivery remains active.

Valid reasons to stop are limited to a human-only decision that no sensible
reversible assumption can avoid, a genuine blocker, a failed mandatory test or
verification that cannot be safely remediated, or reaching the configured
maximum number of audit rounds without `PASS` or `PROVISIONAL PASS`. In those cases
return the exact state, evidence, assumptions, and next human action; do not
imply completion.

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

Apply the composite implementation gate through the installed audit harness
with `--phase audit --gate implementation`, using `audits.yaml` for the
retained session mapping. Remediate **all** findings before resuming that same
session; never resubmit a known FAIL unchanged or remediate findings piecemeal.
On timeout, follow the harness recovery diagnostic: in HANDS-OFF, resume
automatically within its limit, using the retained session. Timeout without a
verdict does not require an artificial edit or justify a handback by itself.
Do not invoke individual audit skills or re-audit the signed-off definition.
After effective
PASS, execute final OT and UT checks, record all results in `validation.org`,
update every affected project document, and return the change for operator
closure. Do not claim closure yourself.
