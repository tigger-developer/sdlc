---
name: migrate-legacy-acs-to-sdlc-v1
description: Centralize legacy ticket acceptance criteria and test traceability in the SDLC v1 record required before a brownfield project adopts SDLC v2. Invoke only when the operator requests this migration.
metadata:
  preferred_provider: openai-codex
  preferred_model: gpt-5.6-luna
---

# Migrate legacy acceptance criteria to SDLC v1

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/ISSUES.md`,
`~/.agents/sdlc/TESTING.md`, and `~/.agents/sdlc/GIT.md` in full. Read the
project constitution, its named requirement and design authorities, the
centralized acceptance-criteria record, and the testing documentation.

Use this skill only for an operator-requested brownfield migration. It
centralizes acceptance criteria from the ticket-based SDLC v0.1 practice into
the SDLC v1 AC record. Completing that intermediate record gives SDLC v2 and
Spec Kit an authoritative brownfield baseline. The skill does not close or
comment on issues, migrate undelivered scope, create feature specifications,
change tests or implementation, or infer human approval.

## Snapshot the issue record

Use the authenticated GitHub interface for the current repository. Retrieve
the complete list of actual issues, open and closed; do not iterate an assumed
numeric range. Allocate a private directory through the operating system's
temporary-directory mechanism and cache every issue once.

Write one `<issue-number>.md` file per issue. Each file must contain:

- issue number and descriptor, state, URL, author, and creation and closure
  dates;
- the complete issue body;
- every comment in chronological order, with author and timestamp; and
- linked pull requests, commits, cross-references, or other recorded
  implementation evidence available from the issue timeline.

Confirm that comment and timeline pagination is complete. Keep the snapshot
outside the repository, report its exact path, and use it for every subsequent
pass. Read each cached issue in full, but never concatenate the complete issue
set into one model context. Never mutate GitHub during this skill.

## Build the evidence indexes

Before classifying tickets:

1. Index the existing centralized acceptance criteria without renumbering or
   rewriting them.
2. Index the current checked-in unit and regression tests, their AC and issue
   references, and whether they are part of the supported test command.
3. Record current whole-suite passing evidence supplied by the operator or
   produced by an already-authorized project command. Do not claim that a live
   regression test passes merely from its presence when no such evidence
   exists.
4. Record implementation evidence separately. A linked code commit supports
   implementation history but does not by itself prove approval or AC
   validity.

Read [references/classification.md](references/classification.md) in full, then
classify every cached issue in ascending issue-number order.

## Reconcile the centralized record

Ignore tickets containing no AC table; they are bug fixes for this migration.
For every unambiguous AC-bearing ticket:

- preserve the original AC and test identifiers, descriptors, wording,
  relationships, statuses, issue provenance, and implementation evidence;
- reconcile an existing or partially migrated entry in place rather than
  duplicating it;
- add missing valid ACs and their traceability to the centralized record;
- preserve superseded ACs as historical entries and identify their replacement
  when the evidence establishes one; and
- apply later issue-number precedence to genuine contradictions, while
  preserving the older requirement and its lineage as superseded.

Do not summarize, renumber, silently rewrite, or omit historical information.
Do not change an operator-test status. Make no project-documentation changes
outside the centralized AC record, and make no implementation, test, issue, or
feature-specification changes.

Process every unblocked ticket before asking for adjudication. Consolidate all
ambiguous cases into one report. For each, give the issue and AC identifiers
with descriptors, the available evidence, the exact ambiguity, and the
decision required. Do not migrate an ambiguous AC until the operator decides.

The migration must be idempotent. A rerun against the same issue snapshot,
repository evidence, and operator decisions must produce no duplicate or
unexplained change.

## Report

Report variances and summary counts, not a catalogue of unchanged or ignored
issues. Include:

- the snapshot path and issue count;
- AC-bearing tickets examined, each with a descriptor;
- current, historically valid, and superseded ACs migrated or reconciled;
- the count of tickets ignored as bug fixes without AC tables;
- human adjudications required;
- files changed and the resulting commit; and
- verification performed and anything not verified.
