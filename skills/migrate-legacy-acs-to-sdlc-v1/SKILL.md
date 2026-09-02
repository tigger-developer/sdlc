---
name: migrate-legacy-acs-to-sdlc-v1
description: Prepare a brownfield project's SDLC v1 acceptance-criteria ledger and project documentation for Spec Kit initialization. Invoke only when the operator requests this readiness migration.
metadata:
  preferred_provider: openai-codex
  preferred_model: gpt-5.6-luna
---

# Prepare an SDLC v1 project for Spec Kit

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/ISSUES.md`,
`~/.agents/sdlc/TESTING.md`, `~/.agents/sdlc/DOCUMENTATION.md`, and
`~/.agents/sdlc/GIT.md` in full. Read the project's requirement, design,
architecture, testing, operational, user, and external-integration authorities.

## Outcome and boundary

This is a bounded readiness migration. Its repository deliverables are only:

1. an accurate, complete, and current `docs/ACs.md`;
2. corrections to stale project documentation; and
3. the unchanged move of `docs/implementation_plan.md` to
   `docs/archive/implementation_plan.md`, when the source exists.

Do not invoke `sdlc-project-init`, create Spec Kit artefacts, change tests or
implementation, add migration reports to the repository, or migrate undelivered
scope. Do not re-litigate a ticket's design, reproduce historical verification,
or investigate old implementation work merely to improve confidence beyond the
classification rules below.

Never edit a ticket body, labels, milestone, or test tables, and do not add
routine migration commentary. The only permitted ticket mutations are the
operator-authorized baseline creation and closure actions defined below.
Historical tickets remain available through the traceability preserved in
`docs/ACs.md`.

## Cache the complete issue record once

Use the authenticated GitHub interface for the current repository. Retrieve
the complete list of actual issues, open and closed; do not iterate an assumed
numeric range. Allocate a private directory through the operating system's
temporary-directory mechanism and cache every issue exactly once.

Write one `<issue-number>.md` file per issue. Each file contains:

- issue number and descriptor, open or closed state, URL, labels, milestone,
  author, and creation and closure dates;
- the complete issue body; and
- every comment in chronological order, with author and timestamp.

Write a manifest containing the repository identity, snapshot time, issue
count, and pagination result. Confirm that issue and comment pagination is
complete. Keep the snapshot outside the repository, report its exact path, and
use only that immutable cache for every subsequent ticket pass. Do not re-fetch
individual tickets. If the snapshot is incomplete, stop before changing files.

## Establish current evidence once

Before classifying tickets:

1. Index `docs/ACs.md` without renumbering or rewriting entries.
2. Index the maintained checked-in regression harness, including its AC and
   issue references and membership in the supported whole-suite command.
3. Run the supported whole-suite command, normally `make test`, at most once.
   If the operator has supplied current passing evidence, reuse it instead.
4. Record the result without diagnosing failures or rerunning individual or
   historical tests. A failed suite prevents inferred PASS updates but does not
   require debugging within this skill.

When the whole suite passes, the presence of a regression test in its maintained
harness is current passing evidence for that test and its traced AC. Use this to
correct legacy `PENDING` or unverified statuses in `docs/ACs.md`. Do not infer a
unit-test or operator-test PASS merely from the test's historical mention.

Read [references/classification.md](references/classification.md) in full and
apply it to the cached issues.

## Reconcile the three deliverables

Process cached issues in ascending issue-number order. Tickets without an AC
table are bug-fix history and require no AC migration. When a ticket contains
more than one AC table across its body and comments, do not select one; include
that ambiguity in the single operator-adjudication list.

Apply every unambiguous addition, status correction, supersession, and
traceability repair directly to `docs/ACs.md`. Preserve original AC and test
identifiers, descriptors, wording, relationships, statuses, issue provenance,
and supersession lineage. Never delete or silently rewrite historical
requirements.

## Run the regression checksum pass

After the ticket-led reconciliation, enumerate every regression test that is
still present in the maintained harness. For each live RT, confirm that:

- it has traceability to at least one AC;
- every cited AC exists in `docs/ACs.md` with its descriptor and source-ticket
  provenance; and
- the ledger records the RT-to-AC relationship.

When an AC is missing, recover its original wording and lineage from the cached
ticket body or comments and add it to `docs/ACs.md`. Never ordinarily
reverse-engineer a new requirement from test code. If a live RT has no
resolvable ticket or AC provenance, add it with its descriptor to the single
operator-adjudication list.

Partially completed work with maintained RT evidence may be adopted as the
pre-migration baseline when the operator authorizes it. Migrate only the ACs
traced to those live RTs; do not claim that the ticket's untested or unfinished
scope was delivered. Preserve the original ticket number for AC numbering.

For operator-approved live RTs that have no source ticket, create one
pre-migration baseline ticket explaining that it exists solely to establish AC
provenance for the migration. Use its ticket number for those AC identifiers.
In this narrow case, write the minimum observable AC supported by the maintained
test, current documentation, and operator's baseline decision. Do not expand the
requirement beyond the behaviour that evidence establishes. Create no ticket
until the operator authorizes the exact descriptor-bearing orphan-RT list.

When the one whole-suite run passed, record each live RT as current passing
evidence according to the classification rules. This is one static checksum
pass over the maintained harness, not permission to rerun tests, re-fetch
tickets, inspect historical implementation, or create intermediate commits.

Compare current project documentation with the reconciled AC ledger, maintained
regression evidence, current interfaces, approved architecture and design, and
external ownership contracts. Correct only stale, conflicting, missing, or
misleading material. Preserve useful history rather than modernizing prose for
its own sake.

If `docs/implementation_plan.md` exists and
`docs/archive/implementation_plan.md` does not, move it unchanged with Git and
repair references that would otherwise become stale. If both paths exist, or
the destination conflicts, ask the operator rather than overwriting either.

Sanitize changed technical documentation and perform proportionate link,
format, and traceability checks. Do not create permanent regression tests that
grep documentation.

## Adjudication, closure, and commit

Finish all unambiguous reconciliation before asking questions. Consolidate
only genuine AC-table ambiguities and evidence cases that the classification
rules cannot resolve into one short operator list. Every issue, AC, and test ID
must have an adjacent descriptor.

Create one coherent local migration commit after all AC, documentation, and
archival changes are assembled and sufficient operator decisions are available.
Never commit per ticket, AC, document, finding, or processing pass. If a long
report is necessary, write it only in the snapshot directory and open it with
`HTML_PREVIEW_TOOL` when configured, otherwise with an available text editor.

After the local deliverables are committed, present two concise,
descriptor-bearing closure lists where applicable:

1. Open tickets whose cached evidence says they were fully delivered and should
   already have been closed. If the operator authorizes that exact batch, close
   them without comments.
2. Partially completed tickets whose maintained RT-backed behaviour the operator
   adopted as the pre-migration baseline. If the operator authorizes that exact
   batch, close each with one concise comment stating that the RT-backed scope
   was carried into `docs/ACs.md` as the migration baseline and that any
   undelivered remainder must be specified afresh if pursued.

Do not re-fetch tickets before closure or change any other ticket field. Leave
wholly undelivered, uncertain, or unauthorized tickets untouched.

## Completion report

Report only:

- the snapshot path and issue count;
- the single whole-suite result, or the current evidence reused instead;
- the live-RT checksum count, including mapped and unresolved tests;
- changes to `docs/ACs.md`;
- any pre-migration baseline ticket created, with its descriptor;
- stale project documents corrected;
- implementation-plan archival;
- the local commit;
- fully delivered tickets closed silently, each with its descriptor;
- partially completed baseline tickets closed with a migration comment, each
  with its descriptor; and
- unresolved operator decisions or readiness blockers.

Do not report a catalogue of unchanged tickets. The operator decides whether
the resulting project is ready for `sdlc-project-init`.
