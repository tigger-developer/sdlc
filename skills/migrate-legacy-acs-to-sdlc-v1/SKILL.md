---
name: migrate-legacy-acs-to-sdlc-v1
description: Archive and close a brownfield project's legacy GitHub tickets while reconciling its SDLC v1 acceptance-criteria ledger and documentation before Spec Kit initialization. Invoke only when the operator requests this migration.
metadata:
  preferred_provider: openai-codex
  preferred_model: gpt-5.6-luna
---

# Retire an SDLC v1 ticket system for Spec Kit

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/ISSUES.md`,
`~/.agents/sdlc/TESTING.md`, `~/.agents/sdlc/DOCUMENTATION.md`, and
`~/.agents/sdlc/GIT.md` in full. Consult project documentation only as needed
for the targeted staleness pass.

## Outcome and boundary

This is a fast archival and classification pass, not a historical review. It
produces:

1. an accurate, complete, and current `docs/ACs.md`, normally produced by
   augmenting the pre-existing ledger in place;
2. a lossless ticket snapshot under `docs/archive/migrated-tickets/`;
3. a concise `docs/ticket-migration.org` migration index;
4. targeted corrections to stale project documentation;
5. the unchanged move of the single file matching
   `docs/implementation?plan.md` to `docs/archive/` with the same basename,
   when the source exists; and
6. closure of every legacy issue after the archive is committed.

Do not invoke `sdlc-project-init`, create Spec Kit artefacts, change tests or
implementation, or migrate undelivered scope into `docs/ACs.md`. Do not inspect
old code, search Git history ticket by ticket, re-litigate design, reproduce
historical verification, or add migration comments to tickets.

Explicit invocation authorizes the defined final closure batch. It does not
authorize ticket-body, label, milestone, test-table, pull-request, or unrelated
GitHub mutations.

## Archive every issue once

Use the authenticated GitHub interface for the current repository. Retrieve
every actual issue, open and closed, with all comments and the timeline events
needed to identify linked code commits; exclude pull requests and do not iterate
an assumed numeric range. Download into an operating-system temporary staging
directory, verify complete issue, comment, and timeline pagination, then promote
the complete snapshot to `docs/archive/migrated-tickets/`.

Write one `<issue-number>.md` file per issue containing:

- issue number and descriptor, original URL, snapshot state and reason, labels,
  milestone, author, and creation and closure dates;
- the complete original issue body; and
- every comment in chronological order, with author and timestamp; and
- every code-commit reference recorded in the body, comments, or ticket
  timeline.

Write `manifest.json` in that directory with repository identity, snapshot
time, issue count, open and closed counts, pagination result, and filenames.
Preserve archived source text exactly; do not sanitize it.

The tracked archive is the immutable cache. Read only it for all later ticket
passes and never re-fetch individual issues. A purpose-created baseline issue
is the sole exception: create and archive it exactly once, then add it to the
manifest. If staging is incomplete or the destination contains conflicting
material, stop before closure.

## Start the durable migration index

Immediately after verifying and promoting the archive, create
`docs/ticket-migration.org` from
`~/.agents/sdlc/templates/migration/ticket-migration.org`. Read that template in
full and follow its embedded structure, syntax, and validation instructions. If
an interrupted migration already has an index, preserve its recorded state and
reconcile its structure rather than overwriting it.

Update the document after each completed evidence phase and immediately after
classifying each ticket. Incremental writes are required; per-ticket commits are
not. Give every ID a descriptor and link ticket titles to the local archived
files. Keep complete source text in the archive rather than duplicating it here.

## Build the live-RT delivery map first

1. Index the pre-existing `docs/ACs.md` without renumbering entries. Presence of
   an AC in this ledger is delivery evidence even when its archived ticket does
   not record test results.
2. Review every test definition in the maintained regression harness. Record
   its descriptor, path or test name, AC references, issue references, and
   membership in the supported whole-suite command. Do not infer this inventory
   solely from `docs/ACs.md`.
3. Run that whole suite, normally `make test`, at most once. Reuse current
   operator-supplied passing evidence when available.
4. If the suite fails, record its result in `docs/ticket-migration.org`,
   checkpoint only that index and the verified ticket archive, and stop. Do not
   diagnose the failure, classify tickets, change the AC ledger or project
   documentation, create a baseline issue, close tickets, or rerun individual
   or historical tests within this skill.
5. Only after a current passing result, build the RT-to-ticket-to-AC evidence
   map.

A passing suite makes each RT in its maintained harness current passing evidence
for the behaviour it verifies and every ticket and AC to which it directly
traces. Use this to correct legacy pending or unverified RT statuses. Reviewing
test definitions here is evidence inspection, not implementation investigation.
Do not classify any ticket as delivered, abandoned, or undelivered until this
complete live-RT map exists.

## Complete the live-RT delivery sweep

For every RT still in the maintained harness, confirm that:

- it traces to at least one AC;
- every cited AC exists in `docs/ACs.md`; and
- the ledger records the RT-to-AC relationship.

Recover a missing AC from its archived ticket when provenance exists. When a
ticket has no recorded passing test evidence but one or more of its RTs remain
in the passing maintained harness, classify the whole ticket as delivered.
Migrate every AC and treat its pending or unverified UTs and OTs as assumed
passing for migration. Mark every AC without independent passing evidence and
the delivered-ticket entry with the migration-heuristic footnote.

For orphan RTs, automatically create one pre-migration baseline issue for the
complete descriptor-bearing set, archive it once, and use its ticket number for
minimal observable ACs bounded by the maintained tests and current
documentation. Explicit invocation authorizes this issue creation; do not stop
for operator confirmation.

## Classify tickets after the RT sweep

Read [references/classification.md](references/classification.md) in full and
classify the archived tickets in ascending number order, using the completed
live-RT map first. Then apply pre-existing `docs/ACs.md` entries, archived test
results, ticket-linked code commits, and other explicit evidence. Apply
unambiguous AC additions, status corrections, supersessions, and traceability
repairs directly to `docs/ACs.md`. Add missing ACs at their correct identifier
positions so the ledger remains in sequence. Preserve identifiers, descriptors,
wording, relationships, statuses, provenance, and supersession lineage.

Treat a code-commit reference recorded on a ticket as delivery evidence unless
the archived record explicitly says that work was reverted, abandoned, or only
partial. Ticket state is not delivery evidence. Apart from identifying issues
that need final closure, use it only after the complete evidence pass finds no
delivery evidence: a closed ticket's scope is abandoned, while an open ticket's
scope is undelivered.

## Apply the near-complete heuristic automatically

Classify a ticket as delivered when:

- at least one test has a recorded passing result;
- every recorded test has a passing result except no more than two tests in
  total;
- every unresolved test is a unit test or operator test marked `PENDING` or
  otherwise unverified;
- no test is recorded as failed; and
- the archive contains no explicit abandonment, reversion, or known-incomplete
  evidence.

Treat those one or two tests as assumed passing for migration. If an AC's only
evidence is one of these assumed passes:

- mark the AC visibly in `docs/ACs.md` without changing its identifier;
- mark the test as an assumed pass using the existing table conventions; and
- add one footnote explaining that delivery was inferred because every other
  ticket test passed and the only unresolved evidence was one or two UTs or OTs
  whose results were not recorded.

The footnote must state that this is an inferred migration status, not a
contemporaneous test result, and identify whether inference came from the
near-complete rule or maintained-RT evidence for the ticket. Do not mark an AC
that has independent recorded passing evidence. Mark the ticket in the
delivered list with Org footnote `[fn:migration-heuristic]`.

When no delivery evidence remains after applying the evidence rules, use only
the ticket's snapshot state to classify its scope: closed means abandoned; open
means undelivered. Record an open bug under open defects. Record feature scope
under defined but undelivered features with its precise disposition. Do not
investigate the code merely to rescue the ticket from this classification.

Read one archived ticket at a time. As soon as it is classified, write its
disposition and evidence to `docs/ticket-migration.org`, update the completed and
next-ticket fields, then release its raw content before reading the next ticket.
Do not keep the complete ticket archive in model context.

## Reconcile project documents

Group the migrated ACs by affected product area, then compare each batch with
the approved design, architecture, project documentation, maintained RT
evidence, and external ownership contracts. When the batch clearly represents
a newer feature or change, update the affected documentation. When the agent is
unsure whether behaviour or documentation is current, inspect only the relevant
implementation as a tie-breaker. Do not inspect code AC by AC, search Git
history, or broaden the investigation beyond that uncertainty.

Match `docs/implementation?plan.md` so either the hyphenated or underscored
filename is accepted. If exactly one source matches and its same-basename path
under `docs/archive/` does not exist, move it unchanged with Git and repair
references made stale by the move. If more than one source matches or the
destination conflicts, ask the operator rather than overwriting anything.

Sanitize changed technical documentation except the lossless ticket snapshots.
Use proportionate link, format, and traceability checks; never create regression
tests that grep documentation.

## Finalize the Org migration index

Follow the canonical template's finalization instructions. Original GitHub URLs
remain in the lossless snapshots. Mark every heuristic-delivered ticket with
`[fn:migration-heuristic]` and define that footnote once in the Org document.
Set the migration phase to complete only after reconciliation and closure
recording finish. The archive is exhaustive; the Org file is the concise map a
later agent reads first.

## Preserve context continuity

Do not initiate or request context compaction merely because the migration is
batch-oriented or contains many tickets. Ticket count alone is not a reason to
compact. Never load the complete ticket archive into context at once; use the
archive as external memory and `docs/ticket-migration.org` as durable working
state.

Before any agent-initiated compaction that is genuinely necessary, first record
the current phase, completed ticket IDs or ranges, active assumptions, evidence,
unresolved items, and next action in `docs/ticket-migration.org`.

If the harness compacts automatically, continue without operator handback.
Reread this skill, the archive manifest, `docs/ACs.md`, and
`docs/ticket-migration.org`, then resume from the first unfinished ticket. Do not
re-fetch tickets, repeat completed analysis, rerun tests, or reconsider recorded
decisions without contradictory evidence.

## Commit and close in batches

Classify automatically wherever these rules apply. Do not stop for operator
confirmation of heuristics, orphan RTs, baseline-ticket creation, or closure.
Put genuinely contradictory or unclassifiable records in the human-review
section and continue; they do not prevent legacy ticket closure.

Commit the complete archive, AC ledger, Org index, documentation corrections,
and implementation-plan move before closing any issue. The authorized baseline
issue for orphan RTs is the only earlier GitHub mutation. Never commit per
ticket or AC. For an unusually large migration, use occasional fixed-size
recovery commits only when necessary.

After that durable commit succeeds, close every issue that was open in the
snapshot and any archived pre-migration baseline issue created during the
workflow. Attempt each closure once, without a comment, continue through the
batch, and collect failures. Do not re-fetch tickets after closure.

Update `docs/ticket-migration.org` with the closure date, totals, and exact
failures, then create one final closure-record commit. Only a recorded failed
closure may be retried later.

## Completion report

Report only:

- archived issue count and manifest path;
- the single suite result or current evidence reused;
- live-RT checksum totals, mapped and unresolved;
- migration-category totals;
- `docs/ACs.md` changes and heuristic-marked AC count;
- stale documents corrected and implementation-plan archival;
- archive and closure-record commits;
- tickets closed; and
- unresolved classifications or closure failures, with descriptors.

If the migration index or completion report does not fit in one terminal
screen, present `docs/ticket-migration.org` with `HTML_PREVIEW_TOOL` when that
variable names an available command; otherwise use an available text editor.
Presentation is not an approval gate.

The operator decides whether the project is ready for `sdlc-project-init`.
