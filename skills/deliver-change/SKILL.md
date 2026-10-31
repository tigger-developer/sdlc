---
name: deliver-change
description: Deliver a signed-off SDLC v3 specification through TDD and validation, using native goal-backed hands-off continuation where available.
---

## Initialization prerequisite

Before creating or editing any ticket or candidate, require a readable
`.sdlc/project.yaml` declaring version 3 and no project-root `specs/`
directory. If either condition fails, stop this skill and tell the operator to
run `sdlc-init` from the project root. Do not allocate an ID, edit a record,
start a review or write code until initialization/migration succeeds. Do not
work around the check, migrate by hand or infer permission to invoke the
initializer. Recheck the profile and layout before resuming.


Follow `~/.agents/sdlc/READING.md`; reuse required texts already read in full,
unchanged and retained in context. Reading lists do not request duplicate loads.

Read `~/.agents/sdlc/MAIN.md`, `TESTING.md`, `CODING.md`, `ARCHITECTURE.md`,
`DOCUMENTATION.md`, `GIT.md`, the project profile, selected standards, the
signed-off `spec.org`, and its current audit evidence.

Admit delivery only when all three prerequisites hold: explicit operator
sign-off in `spec.org`, an `ACTIVE` work item in `docs/work.org`, and accepted
current definition-review evidence under `AUDITS.md`. That evidence may be
internal PASS or an accepted isolated independent result in `audits.yaml`,
or a still-current historical external PASS linked by that record. Review method never substitutes
for sign-off or ACTIVE state. Do not repeat an accepted review merely to change
its method; no new external verdict or failure is required.
Check documented dependencies and implementation holds; only explicit operator
authority lifts a hold. If the
operator now lifts a hold, move the item to `ACTIVE` before admission once the
remaining conditions are satisfied.
Do not trust copied gate metadata or begin by rerunning definition audits. A
material specification change outside DELIVERY.md's Hands-off delivery decision
authority requires current audit evidence and renewed
operator approval: reset sign-off to `PENDING` and move the item to `REVIEW`,
preserving history. Append review results to `audits.yaml` under `AUDIT-RECORD.md`; do not copy
its state into the spec or a parallel ledger.

## Delivery modes and continuation

This section is the canonical continuation contract for delivery and, by
explicit reference, signed-off hands-off spikes. It does not activate either
workflow by itself. Definition keeps its own interruption rules.

At the beginning of delivery, determine whether any information is required
from the operator before implementation can safely start. Ask that one bounded
set of questions once. If the operator deliberately selects `ATTENDED`, remain
attended and hand back when a human decision is required. Otherwise delivery
enters **HANDS-OFF** for this invocation. Declare `DELIVERY MODE: HANDS-OFF` and
advise the operator that pausing delivery or reopening decisions changes the
mode to **ATTENDED** until the operator invokes `HANDS-OFF` again; routine in-scope
corrections do not. Hands-off
mode is not a global default: it is activated only by invoking this delivery
workflow. Delivery-only departure authority is defined in DELIVERY.md; spikes
do not inherit it. Human-only decisions remain outside the shared continuation
contract.

Include any required per-ticket live OT budget in that preflight, following
`DELIVERY.md`'s **Hands-off one-off test budget** rule. Record the approved bounds and
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

HANDS-OFF assumes no human is available. Do not end the turn, request approval,
wait for an answer or otherwise transfer control while safe, authorized,
executable work remains. Questions, uncertainty, routine interpretation,
progress reports, commits, warnings, test/audit results and partial milestones
are not stopping conditions. Continue after them through TDD, remediation,
documentation, validation and evidence recording. Routine in-scope corrections
leave the mode active; incorporate them and continue. Premature handback is a
contract violation with a material cost, not a safer default.

Make sensible reversible assumptions only within the signed-off specification
or brief, design, standards and safety boundaries. Do not infer new authority over
product behaviour, scope, architecture, security,
access, persisted data, external contracts or irreversible decisions. For delivery
of an agreed specification, apply DELIVERY.md's Hands-off delivery decision
authority as the sole limited-departure exception. Spikes instead apply their
skill's agreed-stack and exploration boundaries. Record material assumptions in the
`Hands-off mode decision log`: one level-two Org subtree per invocation, headed
by its exact ISO 8601 timestamp and tagged `:hands-off:review:`, with each
decision at level three.

Before any final handback, re-read the signed-off specification (or spike brief),
completion state, decision log, review evidence and applicable validation record.
A spike uses its brief's Evidence and Findings rather than delivery artefacts. An unchecked item
with an executable next action means delivery remains active.

Valid reasons to stop are limited to a human-only decision that no sensible
reversible assumption can avoid, a genuine blocker, a failed mandatory test or
verification that cannot be safely remediated, or reaching the configured
review limit without an effective PASS under `SELF-AUDIT.md` or the explicitly
enabled external contract. In those cases
return the exact state, evidence, assumptions, and next human action; do not
imply completion.

## Deliver the approved change

After admission, verify the ticket reservation and approved definition published
on primary. Apply `GIT.md` and `DELIVERY.md`: default to creating and publishing
an implementation feature branch from that baseline before writing tests or
production code. Reuse a verified implementation branch on resume. An explicit
`current` override affects implementation only. Keep ticket allocation and
canonical `docs/work.org` updates on primary; never allocate from a branch copy.
Never guess a primary branch, remote or naming convention.

Read `~/.agents/sdlc/ORG-SCHEMA.md`. Create `validation.org` from its template
with one `:testdef:` heading per specified test; preserve IDs and existing history.
Run `~/.agents/sdlc/bin/sdlc-validate --help` before first use. Use it locally to
check the full inventory and correct schema/readiness errors before calling an auditor.

Write and execute the complete justified AT/RT package before implementing the target
behaviour. Only minimal execution-enabling scaffolding is permitted under
`TESTING.md`; include it in test-code audit evidence. Record RED only for the intended observed failure; initially GREEN
preservation tests need their rationale. Mark deferred OT/UT execution AMBER with
a reason. Run the validator with `--spec <spec.org> --readiness-check=test-code`,
then load `~/.agents/skills/test-code-audit/SKILL.md` in this session. Remediate
findings and record current internal PASS before building.
If the approved spec has neither ATs nor RTs, skip this audit and proceed to implementation;
do not remove or reclassify approved ATs or RTs to bypass it.

Implement the authorized solution, obtain GREEN for every AT/RT and execute bounded
OTs. Keep the specification synchronized for authorized routine details, without
changing signed-off behaviour. Finish the solution, tests and affected documentation,
not merely one file or edit. Run `sdlc-validate --readiness-check=delivery-code`
with the same spec, then load `~/.agents/skills/code-audit/SKILL.md` in this
session whether or not test-code was required. Include affected product
documentation; missing or materially stale required docs block the final gate.
A test-code PASS is not delivery approval; check `readiness_check` in audit evidence.
Before requesting another audit either stage, address **every** finding as a batch:
remediate valid findings and challenge incorrect or out-of-scope findings with
evidence under `~/.agents/sdlc/AUDITS.md`. Raise doubtful or unresolved disputes
to the operator; never silently waive findings or expand scope to obtain a PASS.
The old external harness is retired. Use self-review by default; launch an
independent reviewer only with explicit authority under `AUDITS.md`. Do not re-audit the signed-off definition without a material change.
After current delivery-code internal PASS under
`AUDITS.md`, return the results for operator validation, explicitly naming the
review method and any loss of independence.
In HANDS-OFF, defer UTs until all other delivery work is delivered: leave them AMBER for the
operator and list them at handback, alongside any genuinely blocked work. Do not
pause otherwise executable delivery for UT participation. Return the change for
operator validation and closure; do not claim closure yourself.
