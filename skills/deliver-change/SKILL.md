---
name: deliver-change
description: Deliver a signed-off SDLC v3 specification through TDD and validation, using native goal-backed hands-off continuation where available.
---

Read `~/.agents/sdlc/MAIN.md`, `TESTING.md`, `CODING.md`, `ARCHITECTURE.md`,
`DOCUMENTATION.md`, `GIT.md`, the project profile, selected standards, the
signed-off `spec.org`, and its current audit evidence.

Admit delivery only with explicit operator sign-off in `spec.org`, an `ACTIVE`
work item in `docs/work.org`, and current `PASS` or effective `PROVISIONAL PASS`
evidence for the definition gate in `audits.yaml`. Check documented dependencies
and implementation holds; only explicit operator authority lifts a hold. If the
operator now lifts a hold, move the item to `ACTIVE` before admission once the
remaining conditions are satisfied.
Do not trust copied gate metadata or begin by rerunning definition audits. A
material specification change requires current audit evidence and renewed
operator approval: reset sign-off to `PENDING` and move the item to `REVIEW`,
preserving history. Never edit `audits.yaml` or copy its state into the spec.

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
advise the operator that pausing delivery or reopening decisions changes the
mode to **ATTENDED** until the operator invokes `HANDS-OFF` again; routine in-scope
corrections do not. Hands-off
mode is not a global default: it is activated only by invoking this delivery
workflow, and it does not authorize changes to signed-off behaviour or other
human-only decisions.

Include any required per-ticket live OT budget in that preflight, following
`TESTING.md`'s **Hands-off OT budget** rule. Record the approved bounds and
conservative expenditure estimates in the hands-off decision log; no billing
integration is needed. UTs and configured SDLC audits are outside this budget.

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

Read `~/.agents/sdlc/ORG-SCHEMA.md`. Create `validation.org` from its template
with one `:testdef:` heading per specified test; preserve IDs and existing history.
Run `~/.agents/sdlc/bin/sdlc-validate --help` before first use. Use it locally to
check the full inventory and correct schema/readiness errors before calling an auditor.

Write and execute the complete justified RT package before implementing the target
behaviour. Only minimal execution-enabling scaffolding is permitted under
`TESTING.md`; include it in test-code audit evidence. Record RED only for the intended observed failure; initially GREEN
preservation tests need their rationale. Mark deferred OT/UT execution AMBER with
a reason. Run the validator with `--spec <spec.org> --readiness-check=test-code`,
then the audit harness with `--gate test-code`, using the work item's
`audits.yaml`. Remediate until
PASS or effective PROVISIONAL PASS within the configured allowance before building.
If the approved spec has no RTs, skip this audit and proceed to implementation;
do not remove or reclassify approved RTs to bypass it.

Implement the authorized solution, obtain GREEN for every RT and execute bounded
OTs. Keep the specification synchronized for authorized routine details, without
changing signed-off behaviour. Finish the solution, tests and affected documentation,
not merely one file or edit. Run `sdlc-validate --readiness-check=delivery-code`
with the same spec, then invoke `--phase audit --gate delivery-code`: use the same normal command whether or not test-code was required. Each gate has its own verdict allowance. Include affected product
documentation; missing or materially stale required docs block the final gate.
A test-code PASS is not delivery approval; check `readiness_check` in audit evidence.
Before requesting another audit either stage, address **every** finding as a batch:
remediate valid findings and challenge incorrect or out-of-scope findings with
evidence under `~/.agents/sdlc/AUDITS.md`. Raise doubtful or unresolved disputes
to the operator; never silently waive findings or expand scope to obtain a PASS.
The harness handles timeouts and malformed-response recovery internally. An
intervention diagnostic requires operator recovery, not another agent retry loop.
Do not invoke individual audit skills or re-audit the signed-off definition.
After effective delivery-code PASS, return the results for operator validation.
In HANDS-OFF, defer UTs until all other delivery work is delivered: leave them AMBER for the
operator and list them at handback, alongside any genuinely blocked work. Do not
pause otherwise executable delivery for UT participation. Return the change for
operator validation and closure; do not claim closure yourself.

## Audit invocation boundary

Invoke the resolved absolute `~/.agents/sdlc/bin/sdlc-harness` path with
`--gate`, `--audit-record`, `--work-item`, and the complete `--input` list.
Use that same command for each review; do not select start/resume, pass session
IDs, override providers, or troubleshoot the auditor. The harness manages recovery.
Remediate every FAIL finding together. If it reports human intervention required,
report the diagnostic reference and continue other authorized project work; do not
inspect provider state or retry the infrastructure failure. Only after explicit
operator authorization, invoke `--reset` once for the selected gate without inputs,
then submit the normal audit without that flag. Never put resets in a retry loop.
