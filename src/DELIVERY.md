---
title: Delivery Execution and Record Responsibilities
version: 5
last-updated: 2026-10-29
---

# Delivery execution and records

For delivery authors only. MAIN selects this document for definition, delivery,
paired and emergency work. It is not an auditor reference. These provisions
were moved from shared standards without removing their obligations. Use AUDITS
for review orchestration and TESTING, DOCUMENTATION, GIT and ISSUES for criteria.
The three audit skills perform self-review only. An authorized independent review
uses INDEPENDENT-AUDITOR instead and returns evidence to the caller.

## Unattended authorization and acceptance

For an explicitly selected unattended delivery profile, definition sign-off
authorizes delivery but does not initiate execution. The execution environment
owns discovery and dispatch. Recheck current authority before starting work and
before acceptance or publication. Holds, cancellation and superseding definitions
override earlier approval; stop affected work and preserve its evidence rather
than continue under stale authority. This is not permission to cancel unrelated
work or undo human changes.

A profile promising independent review must satisfy AUDITS' independence contract;
it cannot silently fall back to self-review. Present current UT evidence and
recorded departures together for explicit final operator acceptance. Changes from
remediation or integration invalidate affected checks, reviews and UTs; repeat
those affected, not automatically every prior approval. Validate the actual
integration candidate before authorized merge. Readiness, acceptance, merge and
deployment remain distinct authorities.

## Hands-off delivery decision authority

This permission applies only to HANDS-OFF delivery of an agreed, signed-off
specification. It does not apply to definition, attended or paired delivery,
emergency work, or spikes. Spikes have their own approved brief and stack limits.

Follow the approved design. A departure is permitted only when evidence establishes
an implementation blocker and reasonable conforming alternatives have been
investigated. Difficulty, effort, inconvenience, unfamiliarity, aesthetic
preference, an easier alternative, deadline pressure or an exhausted allowance
are not grounds for deviation and do not expand authority. Do not confuse a failed
first attempt or a tool/setup fault with an infeasible design.

Choose the smallest localized, reversible adjustment that resolves the blocker.
A minor product decision is permitted only when it preserves the agreed purpose,
required outcomes and architecture. Judge significance by consequences, not code
size or implementation effort. Do not weaken acceptance criteria or tests to make
a failing candidate appear conformant.

Two red lines require operator agreement before the affected work proceeds:

- Architectural change: changing established component responsibilities, system
  or trust boundaries, state ownership, deployment topology, technology stack or
  fundamental integration contracts.
- Major product decisions: changing intended users, principal workflows,
  promised capabilities, required outcomes or material compatibility commitments.

Security, access, irreversible-data and spending restrictions remain binding.
Assess every departure together with all previous departures against the original
approved baseline, not a progressively rewritten design. Several minor decisions
must not cumulatively cross a red line. If they do, or their significance cannot
be established, stop the affected work for operator direction and continue other
safe authorized work. Do not implement a red-line change pending later approval.

Before implementation, append to the specification's Hands-off mode decision log:
identifier and timestamp; original design reference; blocker and evidence;
conforming alternatives investigated; chosen departure and rationale; user-visible
and compatibility effects; reversibility; and the cumulative assessment against
both red lines. Mark operator acceptance PENDING. Add affected tests, verification
results and review findings as they become available. Preserve the signed-off
baseline and earlier entries; do not rewrite approval to cover the departure.

Supply that log and the original baseline with the review candidate. REVIEW owns
assessment criteria. A permitted logged departure does not invalidate admission
or require renewed definition sign-off merely because it differs from the design.
A departure outside this permission requires the normal definition amendment and
renewed approval. Before final acceptance, merge or closure, obtain and record
explicit operator acceptance or rejection of every departure. Rejected departures
require remediation and affected checks/review before acceptance; audit PASS is
not operator acceptance. Reconcile accepted decisions into the current design
with links to preserved decision history.

## Delivery branch and phase synchronization

### Phase synchronization

While external auditing is suspended under `AUDITS.md`, a current recorded
internal PASS satisfies the audit prerequisite in this section. Do not call the
external harness to obtain a PASS before creating a delivery branch or
synchronizing phase checkpoints. Human approval and other admission rules remain.

### Definition and ticket authority

Definition runs on the project's recorded primary branch (normally `master`
or `main`), never on a newly created feature branch. Verify the configured
primary branch; do not guess its name. Its `docs/work.org` is the single
source of truth for ticket allocation, the work register and lifecycle state.
Record and review the definition there, including its specification and feedback.
An implementation branch's copy is a snapshot, not a competing ticket authority.

Before definition edits, inspect the working tree and synchronize the primary
branch through its existing upstream. Preserve unrelated work; do not stash,
reset, discard or carry feature changes onto primary merely to reach it. A
separate clean checkout within the authorized workspace may be used instead.
Apply the ticket-reservation protocol below before assigning a work number.
Commit and publish the reviewed definition and operator sign-off on primary
before normal delivery admission. An unpublished definition is not dispatch-ready.
Branch configuration does not move ticket
creation or definition onto an implementation branch.

### Implementation branches

The configured strategy applies only when code delivery begins, including
writing tests, after the applicable admission requirements:

- `feature` is the default: synchronize primary, verify the published ticket
  and applicable definition baseline (signed off for normal delivery), then
  create and publish one feature branch from
  that revision before writing test or production code. Reuse an existing
  implementation branch for the same ticket after checking its definition
  baseline; do not create a second branch on resume.
- Explicit `current` retains the operator-selected implementation branch.
  It does not override primary-branch ticket authority.
- Emergency work follows `EMERGENCY.md`; its later ticket reconciliation
  still uses primary. Paired delivery retains its skeleton-before-code rule.

The initializer resolves the strategy from command line, process environment,
project `.sdlc/project.yaml`, user `~/.agents/sdlc.yaml`, then the schema
default, `feature`. It persists only explicit project overrides. Agents never
read `.env`. Documentation-only work does not create an implementation branch.

During delivery, keep code, tests and candidate-specific evidence on the
implementation branch. Update canonical `docs/work.org` on primary with narrow,
committed changes supported by that evidence; never merge code merely to update
the register or overwrite primary's register with a stale branch copy. At
integration, preserve primary's latest register and reconcile only the ticket's
authorized state changes. Approval and closure rules remain unchanged.

Synchronize each phase's branch through its configured upstream before edits
and after committed, reviewed phase checkpoints. The early reservation commit
below is allocation evidence, not a completed or approved definition phase.
Creating and publishing an implementation branch uses an existing configured
remote; it does not authorize new remotes, merges, deployment, branch deletion
or history rewriting. Follow `GIT.md` and the project's integration policy.

### Ticket reservation and collision prevention

Serialize writers sharing a checkout; never let concurrent agents edit or stage
its register together. Separate checkouts still require the publication check
below. A fetch alone does not reserve a number.

1. Synchronize primary and inspect its register, specifications, archives and
   relevant history under `ISSUES.md`. Choose the next never-used number.
2. Record a minimal reservation in primary's `docs/work.org`: ID, descriptor,
   intended specification path and a note that definition is pending, using the
   existing lifecycle vocabulary. Commit and push
   this narrow reservation before treating the ID as allocated or using it in
   a specification directory, implementation branch or external reference.
3. A successful non-force push establishes the reservation. If primary advanced
   and rejects the push, refresh it and recheck the namespace. Reconcile the
   unpublished reservation; if its candidate number was taken, choose the next
   never-used number and retry. Do not blindly replay a colliding reservation.
   A tentative unpublished number is not an assigned identifier.
4. Never reuse or silently renumber a published reservation. Retain abandoned
   reservations as history. If conflicting established identities already exist,
   stop allocation and request operator resolution with the conflicting evidence.

With no reachable upstream, do not claim a shared reservation or invent an
alternative allocator. Discussion may continue without an assigned ticket.
For an explicitly local-only project, primary remains authoritative and allocation
must be serialized by one writer; concurrent definitions must wait for the
reservation commit. No lock service or other infrastructure is implied here.

## Documentation during delivery and closure

### Delivery closure

This section governs SDLC-managed delivery. A standalone standards audit uses
DOCUMENTATION's review criteria and does not require gate, validation, sign-off or
closure records.

Definition identifies affected documentation in the handoff. Update it during
delivery and include it in the final delivery-code audit: verify usage,
configuration, interfaces and architecture wherever the change affects them,
and compliance with DOCUMENTATION's voice, purpose and presentation rules.
Missing or materially stale required documentation blocks that gate; optional
editorial improvements do not. Test-code review does not require completed product
documentation. Paired/emergency work updates product docs after the code change
but before the final delivery verdict; retrospective spec consolidation may follow.

Before closing normal, paired, or `BYPASS-GATE-7` work, reconcile every
document made inaccurate, incomplete, or misleading by the change. Record the
documentation outcome as updated or not affected, with the reason when it is
not affected. Known stale documentation blocks closure.

For an emergency change, replace the temporary specification with durable
project authority after verification. Update the specification, design,
operational guidance, and user documentation only where the delivered change
affects them; do not create duplicative prose that merely restates the code.

## Test review sequence

`REVIEW.md` defines gate-specific criteria: definition assesses proposed tests;
test-code assesses the written AT/RT package and intended RED or justified
preservation GREEN; delivery-code assesses production code, changed tests and
settled AT/RT/OT GREEN. The author must verify the reason for RED. Planned tests
are not implementation evidence; implemented tests do not retroactively repair
an inadequate definition. Paired/emergency review uses recorded authorized scope
without inventing a retrospective pre-build gate. Delivery authors follow
`AUDITS.md` for sequence and skills; independent auditors use `INDEPENDENT-AUDITOR.md` and do
not load delivery orchestration. During automatic-harness suspension the author
uses SELF-AUDIT, including retrospective definition review, without an external
attempt or provider context.

Before test-code review, the agent may add only the minimum declarations,
interfaces or placeholder bodies needed to compile and execute the specified ATs and RTs.
This scaffolding must not implement the target behaviour and must be included in
test-code audit evidence. Compilation and setup failures do not qualify as RED.

If the approved specification contains neither ATs nor RTs, skip the test-code gate. The
definition audit must establish why OT/UT evidence is appropriate. Do not remove
or reclassify approved ATs or RTs merely to bypass the gate. Implement, execute required
OTs and invoke delivery-code directly; UTs may await the operator. The validator's
inventory supplies separate AT and RT counts; it does not authorize changing the test strategy.

Define the expected evidence for one-off and user tests before implementation
where practical. They do not follow TDD and do not require a pre-change failure.
Execute final OTs before the delivery-code audit; every AT, RT and OT must be GREEN.
UTs may remain AMBER until operator participation. Earlier diagnostic executions
count only if their evidence remains current for the candidate. Paired development
uses the live user-validation contract below.

## Validation records and user approval

### Test execution records and readiness

Every delivery has `validation.org`, with one current heading per AT, RT, OT and UT
defined in `spec.org`. Follow `~/.agents/sdlc/ORG-SCHEMA.md` and the canonical
templates: both documents tag tests `:testdef:`, and only validation headings carry
execution states. Declare `#+TODO: AMBER RED | GREEN`; an absent state is schema-valid
but audit-ineligible. Preserve superseded runs below the same test node.

The deterministic `~/.agents/sdlc/bin/sdlc-validate` reports schema errors, inventory
and readiness as YAML. Read its `--help` and run it before the corresponding audit.
When explicitly re-enabled, the harness repeats the same checks before spending
a round. Schema failures and
ineligibility never invoke a model. Test-code requires ATs/RTs RED/GREEN and a recorded
state for every OT/UT; delivery-code requires AT/RT/OT GREEN, permitting outstanding UTs only.

Do not infer execution from implementation claims. Record tested revision,
environment, dated observations and evidence. For UTs, record the method as user
validation and a reviewer role only where relevant, not a personal name or
conversation summary; follow DOCUMENTATION.md. A broken
fixture is AMBER; AT/RT RED requires an executed intended assertion failure before
implementing the target behaviour. Initially GREEN preservation tests need no artificial RED.
Readiness checks verify recorded structure, not evidence truth or adequate coverage.
Closure requires current GREEN for every required test, including operator UTs.

The explicit legacy ticket-migration skill has one historical-record
exception. It may infer delivery when either its near-complete rule has at least
one recorded pass and leaves at most two unrecorded unit or operator tests, or a
ticket without recorded passing evidence retains at least one RT in the
maintained passing harness. It must mark every AC supported only by inference,
identify the applicable heuristic in a
footnote, and distinguish an assumed migration pass from contemporaneous test
evidence. This creates no precedent for active delivery.

An audit verdict remains historical evidence for the revision and scope it
assessed. If later remediation changes relevant code, the earlier implementation
gate PASS is no longer current for completion. Rerun the affected automated tests
and affected one-off tests before the delivery-code audit; repeat materially affected
user tests when the operator is available. Unaffected results remain current.

### Paired user validation

Explicit operator validation during paired development is first-class user-test
evidence. Do not replace it with an automated imitation of the visual or
subjective judgement that just occurred.

Maintain a provisional ledger during the session. For each validation, retain:

- a plain-language description of the observed behaviour;
- the revision or working-tree state reviewed;
- material viewing conditions such as viewport, browser, or device; and
- whether a later iteration superseded it.

At closure, present the current ledger once and ask whether it may be recorded
as the user tests for the change. On approval, write it to the active change's
`validation.org`. A later change invalidates only entries whose observed behaviour it
materially affects.

Add automated regression coverage only when the behaviour is objective and
stable, regression risk justifies retention, the test uses an appropriate user
boundary, and it adds evidence beyond the paired validation. Report useful gaps
at closure; do not create brittle tests merely to eliminate the report.

## Hands-off one-off test budget

### Hands-off OT budget

When hands-off delivery needs cost-incurring live OTs, establish one
**operator-approved total budget for the ticket** during preflight. Reuse an
existing explicit authorization covering those operations; otherwise ask once,
not per call. Record the amount, currency, covered operations, retry and resource
lifetime limits, and estimated expenditure in the spec's hands-off decision log.
Resuming delivery does not reset the allowance.

**Conservative estimates are sufficient**, using expected unit costs and
operation counts, including retries and resource lifetime. Do not build billing
integrations, query account balances, or add budget-management tooling. Never
start an operation whose cost cannot reasonably be bounded within the remaining
allowance. If the budget is missing, exhausted or insufficient, defer that OT,
continue other unblocked work, and report outstanding verification at handback.
Do not imply a deferred test passed or authorize further spending yourself.

**UTs are outside this budget and hands-off execution**: leave them AMBER for
operator participation after all other delivery work, or report any genuine
remaining blockers. Configured SDLC audits remain separately authorized under
their existing limits. An OT budget never permits metered checks in RT packs.

## Work-item and specification timing

Every code or script change requires a work item; documentation-only changes do
not. Reuse an applicable item or allocate one under the project's numbering
scheme. Normal delivery requires the agreed durable specification before code.
Paired delivery creates a skeleton before code; emergency delivery creates it
immediately after the bounded fix. Their operator-authorized working
specifications follow `PAIRING.md` and `EMERGENCY.md`.

## Processing definition feedback

During definition, process annotations as one coherent batch. Check the referenced
revision and passage, answer questions as questions, and apply authorized changes
to the detailed requirements, tests, design and derived Synopsis. Explain rejected
suggestions; never silently reject an explicit operator instruction or infer new
authority from an annotation. Conflicting or ambiguous feedback remains unresolved.

For each resolved annotation, preserve its stable identity, author, original text,
referenced passage and reviewed revision in `feedback.yaml`. Record the disposition
(applied, answered or rejected), reason, processing date, affected sections and
resulting specification hash. Retain earlier entries; create this file only when
feedback is processed. This defines required information, not a finalized YAML
schema. Do not fabricate missing provenance or design a competing annotation format.

Record the disposition before removing the resolved annotation from `spec.org`.
Verify the files have not changed underneath the agent and commit the document
and log together. On interruption, reconcile recorded identities and actual edits
before continuing; never apply the same annotation twice or drop new feedback.
Unresolved annotations stay visible. The previewer writes annotations only;
the drafting agent reconciles them into the specification body.
