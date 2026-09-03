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

1. an accurate, complete, and current `docs/ACs.md`;
2. a lossless ticket snapshot under `docs/archive/migrated-tickets/`;
3. a concise `docs/ticket-migration.org` migration index;
4. targeted corrections to stale project documentation;
5. the unchanged move of `docs/implementation_plan.md` to
   `docs/archive/implementation_plan.md`, when the source exists; and
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
every actual issue, open and closed, with all comments; exclude pull requests
and do not iterate an assumed numeric range. Download into an operating-system
temporary staging directory, verify complete issue and comment pagination, then
promote the complete snapshot to `docs/archive/migrated-tickets/`.

Write one `<issue-number>.md` file per issue containing:

- issue number and descriptor, original URL, snapshot state and reason, labels,
  milestone, author, and creation and closure dates;
- the complete original issue body; and
- every comment in chronological order, with author and timestamp.

Write `manifest.json` in that directory with repository identity, snapshot
time, issue count, open and closed counts, pagination result, and filenames.
Preserve archived source text exactly; do not sanitize it.

The tracked archive is the immutable cache. Read only it for all later ticket
passes and never re-fetch individual issues. A purpose-created baseline issue
is the sole exception: archive it exactly once after creation and add it to the
manifest. If staging is incomplete or the destination contains conflicting
material, stop before closure.

## Establish evidence once

1. Index `docs/ACs.md` without renumbering entries.
2. Index the maintained regression harness, including its AC and issue
   references and membership in the supported whole-suite command.
3. Run that whole suite, normally `make test`, at most once. Reuse current
   operator-supplied passing evidence when available.
4. Record the result. Do not diagnose a failure or rerun individual or
   historical tests within this skill.

A passing suite makes each RT in its maintained harness current passing evidence
for its traced AC. Use this to correct legacy pending or unverified RT statuses.

Read [references/classification.md](references/classification.md) in full and
classify the archived tickets in ascending number order. Apply unambiguous AC
additions, status corrections, supersessions, and traceability repairs directly
to `docs/ACs.md`. Preserve identifiers, descriptors, wording, relationships,
statuses, provenance, and supersession lineage.

## Run the live-RT checksum

For every RT still in the maintained harness, confirm that:

- it traces to at least one AC;
- every cited AC exists in `docs/ACs.md`; and
- the ledger records the RT-to-AC relationship.

Recover a missing AC from its archived ticket when provenance exists.
Partially completed tickets contribute only the ACs backed by maintained RTs;
their unfinished scope remains undelivered.

Put orphan RTs in the single operator-adjudication list with descriptors. If
the operator adopts them, create one pre-migration baseline issue for the whole
set, archive it once, and use its ticket number for minimal observable ACs
bounded by the maintained tests and current documentation.

## Apply the near-complete heuristic automatically

Classify a ticket as delivered when:

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
contemporaneous test result. Do not mark an AC that has independent recorded
passing evidence. Mark the ticket in the delivered list with Org footnote
`[fn:migration-heuristic]`.

## Reconcile project documents

Compare project documentation with the reconciled AC ledger, maintained RT
evidence, approved design and architecture, and external ownership contracts.
Correct clear stale or misleading material only. Do not modernize prose or
research old implementation merely to improve confidence.

If `docs/implementation_plan.md` exists and
`docs/archive/implementation_plan.md` does not, move it unchanged with Git and
repair references made stale by the move. If both paths exist or the destination
conflicts, ask the operator rather than overwriting either.

Sanitize changed technical documentation except the lossless ticket snapshots.
Use proportionate link, format, and traceability checks; never create regression
tests that grep documentation.

## Write the Org migration index

Create `docs/ticket-migration.org` with these sections in order:

1. **Open defects at migration:** one Org subtree per defect, containing its
   important current facts, relevant AC text, evidence, and disposition.
2. **Defined but undelivered features:** one Org subtree per ticket, summarizing
   its intended outcome and relevant ACs. Mark its disposition as abandoned at
   migration and state that any revival requires a new Spec Kit specification.
3. **Requires human review:** include only when classification remains
   unresolved; record the exact uncertainty and available evidence.
4. **Delivered tickets:** a simple bulleted list of every delivered ticket,
   including those open at migration but classified as delivered.

Use Org heading levels to nest the detail sections. Give every ID a descriptor
and link its title to the local archived ticket file. Original GitHub URLs remain
in the lossless snapshots. Mark every heuristic-delivered ticket with
`[fn:migration-heuristic]` and define that footnote once in the Org document.

Do not duplicate complete ticket text in the index. The archive is exhaustive;
the Org file is the concise map a later agent reads first.

## Commit and close in batches

Finish every unambiguous classification before asking questions. Consolidate
only genuine ambiguities into one descriptor-bearing operator list. An
unresolved item may remain in the human-review section and does not prevent its
legacy ticket from closing.

Commit the complete archive, AC ledger, Org index, documentation corrections,
and implementation-plan move before mutating GitHub. Never commit per ticket or
AC. For an unusually large migration, use occasional fixed-size recovery
commits only when necessary.

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

The operator decides whether the project is ready for `sdlc-project-init`.
