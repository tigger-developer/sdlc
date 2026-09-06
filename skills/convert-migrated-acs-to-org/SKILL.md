---
name: convert-migrated-acs-to-org
description: Losslessly convert an SDLC v1 docs/ACs.md ledger to canonical foldable docs/ACs.org during an authorized migration.
metadata:
  preferred_provider: openai-codex
  preferred_model: gpt-5.6-luna
---

# Convert a completed migration ledger to Org

If `~/.agents/sdlc/MAIN.md` is absent or unreadable, report that exact path;
never search for another copy.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/ISSUES.md`,
`~/.agents/sdlc/DOCUMENTATION.md`, `~/.agents/sdlc/GIT.md`, and
`~/.agents/sdlc/templates/migration/ACs.org` in full.

## Preconditions

This is a structural conversion, not a requirements review. It handles both a
direct SDLC v1 ledger and a project whose ticket migration finished before
`docs/ACs.org` became canonical.

Always require:

- `docs/ACs.md` is a regular file and `docs/ACs.org` is absent;
- the Git working tree permits an isolated conversion; and
- no current document other than the source ledger claims a conflicting
  requirements authority.

When `docs/ticket-migration.org` or `docs/archive/migrated-tickets/` exists,
require both: the manifest and numbered archive are complete, the migration
index records closure as complete, and the authenticated GitHub interface
reports zero open issues. Retrieve the complete open set, excluding pull
requests. When neither migration artefact exists, treat this as a direct v1
ledger conversion and do not invent a ticket archive or GitHub claim.

If any precondition fails, make no changes. Report the exact failed condition
and the descriptor of every open issue returned.

Inspect the Git working tree before editing. Stop only if the conversion would
overlap changes not made in this workflow.

## Lossless conversion

Read `docs/ACs.md` in full. Build a source inventory containing every:

- AC identifier, descriptor, and full requirement;
- source ticket or other provenance reference;
- shipped version, migration date, status, and qualification;
- test identifier, descriptor, type, status, and evidence note;
- supersession, conflict, footnote, glossary entry, and explanatory note; and
- heading or table grouping that carries meaning.

Render the candidate using the canonical Org template. Preserve all inventoried
information. Normalize Markdown structure and syntax into Org, but do not
summarize, silently correct, reinterpret, merge, split, renumber, or omit source
material. Keep ACs in identifier sequence. Use heading nesting for foldable
structure and native Org links, verbatim spans, and footnotes.

The authority section must say that `docs/ACs.org` is the sole authority for
requirements established under the legacy ticket-led process. Archived tickets,
comments, and `docs/ticket-migration.org` retain disposition, provenance, and
rationale only; they are not current requirement or AC authorities. Signed-off
SDLC v3 specifications govern later requirements.

Stage the candidate outside the tracked source path until it passes:

1. Pandoc Org parsing;
2. a field-by-field reconciliation against the source inventory;
3. an outline cross-check against the canonical example: every AC is a
   level-three heading containing both its identifier and a useful descriptor,
   such as `*** AC29.1 - Reject an invalid host definition`; identifier-only AC
   headings fail validation;
4. identifier, requirement, test, status, provenance, supersession, footnote,
   glossary, and note completeness checks; and
5. a check that instructional template comments and placeholders are absent.

Counts are supporting evidence only; they do not replace field reconciliation.
Do not modify the lossless files under `docs/archive/migrated-tickets/`.

After validation, use `git mv docs/ACs.md docs/ACs.org` and replace that renamed
file with the validated candidate. Update current tracked documentation and
configuration references from `docs/ACs.md` to `docs/ACs.org`, excluding the
immutable ticket archive and historical quotations whose old path is material.
Verify that no `docs/ACs.md` remains and that no current non-archived document
still presents it as the active ledger.

Commit the conversion and reference updates together with the message
`docs: normalize migrated AC ledger`. Do not change requirements, tests, code,
ticket state, archived tickets, or active SDLC v3 specifications.

## Report

Report:

- the AC and test totals reconciled, with descriptors for any exceptions;
- the Org parser result;
- the current references updated;
- confirmation that `docs/ACs.md` is absent;
- confirmation that archived tickets were unchanged; and
- the commit identifier and descriptor.
