---
name: migrate-legacy-acs-to-sdlc-v1
description: Prepare a brownfield project's SDLC v1 requirements, tickets, tests, and documentation for Spec Kit initialization. Invoke only when the operator requests this readiness migration.
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
`~/.agents/sdlc/GIT.md` in full. Read the project's requirement and design
authorities, centralized acceptance-criteria record, testing documentation,
and supported test command.

Use this skill only for an operator-requested brownfield migration. It
prepares an accurate SDLC v1 project state before `sdlc-project-init` introduces
Spec Kit. It does not invoke `sdlc-project-init`, create feature specifications,
change tests or implementation, or migrate undelivered scope.

The evidence-gathering and proposal stages are read-only. Never close or
comment on a ticket, change an AC or its state, or update project documentation
without explicit operator authorization for that exact action. Authorization
may cover a clearly enumerated batch. Evidence, likelihood, an existing commit,
or a passing test never substitutes for operator authorization.

## Snapshot the issue record

Use the authenticated GitHub interface for the current repository. Retrieve
the complete list of actual issues, open and closed; do not iterate an assumed
numeric range. Allocate a private directory through the operating system's
temporary-directory mechanism and cache every issue once.

Write one `<issue-number>.md` file per issue. Each file must contain:

- issue number and descriptor, state, URL, author, labels, milestone, and
  creation and closure dates;
- the complete issue body;
- every comment in chronological order, with author and timestamp; and
- linked pull requests, commits, cross-references, or other recorded
  implementation evidence available from the issue timeline.

Write a manifest containing the repository identity, snapshot time, issue
count, and pagination result. Confirm that issue, comment, and timeline
pagination is complete. Keep the immutable snapshot outside the repository,
report its exact path, and use it for every subsequent pass. Read each cached
issue in full, but never concatenate the complete issue set into one model
context. If any page is incomplete, stop without classifying or proposing
mutations.

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

Read [references/classification.md](references/classification.md) in full.

## Assess open-ticket readiness

Classify every open ticket before changing anything. Candidate selection is
deliberately inclusive: a feature ticket with all or most AC evidence passing
may be presented as delivered or probably delivered, but an outstanding AC,
test, or documentation obligation must remain visible. A bug-fix ticket without
an AC table may be a delivered closure candidate when implementation and
regression evidence support that conclusion.

Treat partially delivered, ambiguous, and undelivered tickets as untouched
open scope. Do not edit them, close them, migrate their ACs, or create draft
specifications from them.

Present one proposed-action report containing:

- delivered and probably delivered ticket closure candidates;
- proposed AC additions, reconciliations, supersessions, retirements, or other
  state changes;
- project-documentation variances associated with each candidate;
- unresolved evidence and the exact adjudication required; and
- undelivered open scope explicitly marked `NO ACTION`.

Every ticket, AC, test, commit, or other identifier shown to the operator must
have its own adjacent descriptor. Never expect the operator to resolve an
unexplained identifier.

Ask in ordinary prose which exact ticket closures, AC changes, and documentation
updates are authorized. A batch answer authorizes only the enumerated actions.
Do not treat approval of a ticket closure as approval of an unlisted AC state
change, or vice versa.

If the proposed-action list or explanation is too long for a concise response,
write it to a Markdown file in the snapshot directory. If
`HTML_PREVIEW_TOOL` names an available command, open the exact report with that
command. Otherwise open it in an available text editor. Report the exact path;
never add the report to the repository.

## Apply an authorized closure batch

Process authorized tickets in ascending issue-number order so later decisions
retain their established precedence. For each authorized closure:

1. Re-fetch that exact ticket and compare its state, body, comments, and
   timeline with the snapshot. If it changed, exclude it from the batch and
   report the variance instead of acting on stale evidence.
2. Apply only its authorized AC changes to `docs/ACs.md`.
3. Check whether its delivered design or behaviour is accurately represented
   in the project documentation. Apply only the authorized documentation
   reconciliation.
4. Sanitize and verify changed documentation, then create a recoverability
   commit containing the local reconciliation.
5. Close only the ticket explicitly authorized for closure. Do not close it if
   an approved local reconciliation remains incomplete or uncommitted.

When changing the centralized AC record:

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
Do not change an operator-test status. Striking off an AC means preserving it
as superseded or retired with its reason and lineage; it never means deleting
it.

## Reconcile the complete historical record

After the authorized open-ticket batch, classify every cached issue in
ascending issue-number order and compare every AC-bearing ticket with
`docs/ACs.md`. Tickets without AC tables remain bug-fix history and need no AC
migration.

Process every unblocked ticket before asking for adjudication. Consolidate all
proposed changes and ambiguous cases into one report. For each, give every issue
and AC identifier with its descriptor, the available evidence, the exact
proposed change or ambiguity, and the decision required. Apply no historical AC
change until the operator authorizes it, individually or as an enumerated
batch.

The migration must be idempotent. A rerun against the same issue snapshot,
repository evidence, and operator decisions must produce no duplicate or
unexplained change.

## Review project documentation

As a final readiness check, compare the README, vision, architecture,
operations, testing, user guidance, and other declared project authorities with
the reconciled requirements, maintained regression evidence, current
interfaces, and external ownership boundaries. Identify stale, conflicting,
missing, or misleading material without rewriting documents merely to make
them look current.

Present the documentation variances and proposed corrections. Apply only the
operator-authorized corrections. Preserve useful historical decisions and mark
superseded material rather than silently deleting it. Sanitize changed
technical documentation and perform proportionate link, command, and format
checks.

## Report readiness

Report variances and summary counts, not a catalogue of unchanged or ignored
issues. Include:

- the snapshot path and issue count;
- ticket closures and AC changes authorized and applied, each with a
  descriptor;
- delivered candidates left open because authorization or evidence was absent;
- untouched undelivered scope;
- current, historically valid, superseded, and unresolved AC reconciliation;
- project documentation refreshed and remaining variances;
- human adjudications still required;
- files changed and resulting commits; and
- verification performed and anything not verified.

Conclude with the evidence supporting `sdlc-project-init` readiness and every
remaining blocker. The operator determines whether to proceed with
initialization.
