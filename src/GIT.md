---
title: Git and Source-Control Standards
version: 9
last-updated: 2026-11-14
---

# Git and Source-Control Standards

Git provides recoverability and traceability. It is not a substitute for a
specification or verification evidence.

## Repository safety

- Confirm the repository and inspect the working tree before changing tracked
  files.
- Do not change project code outside a Git repository unless the human has
  explicitly chosen another recoverability mechanism.
- Preserve unrelated, unknown, and human-authored changes. Do not stage,
  rewrite, revert, or clean them.
- Stop before editing only when requested work overlaps changes whose ownership
  cannot be established.
- Never use destructive recovery commands against a broad path.
- Never directly read, enumerate, create, modify, or delete anything inside
  `.git`. Git internals are not application storage or an agent workspace. Use
  documented `git` commands for every repository operation. No exception
  process applies.
- Never bypass hooks or checks with `--no-verify`, disabled hooks, force flags,
  or an alternate Git directory.
- Never force-push or rewrite shared history without explicit authority for the
  exact branch.
- Verify unfamiliar Git flags before use. Any flag that weakens a safeguard
  requires explicit human authority.

## Commits

- Make each commit coherent, reviewable, and small enough to revert safely.
- Use an imperative subject that describes the outcome, not the activity.
- Explain important reasons, constraints, and compatibility effects in the body.
- Follow the project's established commit convention; use Conventional Commits
  when no convention exists.
- Do not include AI attribution or fabricated contribution trailers.
- Do not claim broad verification in a commit unless that verification was run
  against the committed state.
- When citing an identifier, include its descriptor: `FR-012 - repeated
  installation is a no-op`, not just `FR-012`.

Do not rely on commit-message keywords to close work automatically unless the
project explicitly adopts that behaviour. A merged change and a verified
outcome are separate facts.

## Make synchronization

For a project that uses Make, `make sync` is the canonical operator-facing
synchronization entry point. It performs this sequence and stops on the first
failure:

1. `git add -A`
2. If staged changes exist, `git commit -m "$(COMMIT_MESSAGE)"`
3. `git pull`
4. `git push`

`COMMIT_MESSAGE` defaults to the short neutral subject `chore: sync`. The target
must not infer a detailed message from the diff, demand a long message, amend an
existing commit, select a branch, rebase, stash, clean, force, bypass hooks, or
run unrelated generation or verification. When no staged changes exist, it
skips only the commit and still pulls and pushes through the configured
upstream and pull strategy.

The target stages the whole worktree by definition. An agent must inspect the
working tree first and may invoke `make sync` only when the human has explicitly
authorized synchronization and the agent can preserve every unrelated or
human-authored change. An operator invocation is the decision to include the
current worktree; the target must not try to outsmart that decision by filtering
or rewriting it.

Do not use `make sync` as the staged phase-synchronization mechanism. Phase
synchronization acts only on already committed phase work; `make sync` is an
explicit operator decision to stage the whole worktree.

## Branches and history

These branch rules apply only to full-SDLC work, not personal work or standalone
use of coding standards. The recorded primary branch is the sole write authority
for ticketing: allocation, lifecycle and status changes in `docs/work.org`, and
the ticket's canonical SDLC records under `tickets/`, including specifications,
feedback, `validation.org`, `audits.yaml`, sign-off and decision records. Edit
these records on primary, commit and push each coherent update through its
configured upstream before treating it as published or using it to advance a
feature-branch workflow. A feature branch's copies are snapshots, never a second
ticket authority. Do not create a feature branch during definition. Follow
`DELIVERY.md` for serialized ticket reservation and phase synchronization.

Feature branches are the default only for code delivery, starting with test
implementation after admission. Keep the established project checkout on its
recorded primary branch for ticketing and project tooling. For each admitted
ticket, create a separate shallow clone in the operating system's temporary
directory from the published primary baseline, then create or reuse that
ticket's feature branch there. Different tickets use different clones so their
builds can proceed concurrently without switching or dirtying the established
project checkout. The clone is an implementation workspace, not a ticket-record
writer. The task's filesystem authority must cover the intended primary checkout
and temporary implementation workspace; a clone does not grant wider access.
Where a shallow remote clone is unavailable, use another separately authorized
checkout and retain the same primary/implementation separation.

Explicit implementation branch overrides remain valid. Exact-token
`BYPASS-GATE-7` emergencies default to primary unless the operator specifies
another implementation branch. Neither override moves ticket authority off
primary. Documentation-only work does not require a feature branch. Branch
selection authorizes neither merge, deployment nor closure. Read-only work
requires no branch switch.

An explicit operator instruction to merge a named feature branch into primary
authorizes that merge. State a known risk briefly, then carry out the merge;
do not refuse, ask for confirmation again, or substitute a pull request. The
instruction does not authorize deployment, release publication or ticket
closure, and it does not turn unresolved audit findings into a PASS.

For an authorized merge, first verify that the established primary checkout is
clean and still on the recorded primary branch. Use a separate temporary
integration clone, not the established project checkout, to pull current
primary from its configured upstream, merge the named feature branch and
inspect changes to `docs/work.org` and `tickets/`. Do not merge a stale branch
copy over newer primary ticket records. Reconcile any ticket-record changes
against primary's published history; do not select a whole side of a conflict
without checking its contents. Pull primary again before the non-force push;
if it advanced, reconcile and recheck the ticket paths. After the push, update
the established primary checkout with a fast-forward-only pull. Stop on a
conflict needing judgement, a rejected push, or an established checkout that
is dirty or off primary; report the exact state without rewriting history.
Project-specific merge or integration requirements still apply.

A shallow implementation checkout without the current primary history must use
an authorized, synchronized primary checkout for ticket edits, or hand the
evidence to an authorized primary-branch publisher. It does not acquire ticket
authority by being the implementation target; do not advance on an unpublished
ticket-record update.

- Resolve the primary branch from repository configuration and follow its
  existing branch naming convention. Do not assume the name `main` or `master`.
- Rebase or merge according to project policy; do not rewrite other people's
  commits for cosmetic consistency.
- Fetch before making claims about remote state.
- Use annotated tags for releases unless the project documents another policy.

For release version numbers, follow the project's documented policy and any
required or established best-practice convention for its language, framework,
or product type. If none applies, use [Semantic Versioning (SemVer)](https://semver.org/)
for a declared public interface, including its `MAJOR.MINOR.PATCH` numbering
and compatibility-based increment rules. Follow the project's applicable tag
convention for whether a Git tag includes a `v` prefix; the prefix is not part
of the SemVer version number.

Obtain the operator's explicit approval before selecting a major version
increment. A breaking change or an applicable numbering convention does not
itself authorize that increment.

## Hooks

Repository hooks must not evade organization-wide or user-wide hooks. If a
project supplies a hook, chain to the configured global hook path or document a
safe equivalent that preserves its checks.

Treat hook failure as evidence to investigate. Do not edit the hook, suppress
its output, or change environment variables merely to obtain a successful
commit.

## Tracked and generated files

- Ignore credentials, local environment files, editor state, build output,
  caches, and provider runtime directories.
- Track reproducibility files such as lockfiles and dependency checksums unless
  the ecosystem explicitly defines them as application-local output.
- In Go modules, track `go.sum`; it records dependency checksums and does not
  belong in `.gitignore`.
- Review large or generated files before staging. Use the project's large-file
  mechanism where appropriate.
- A staged or work-generated file larger than the project's documented limit,
  or 100 MiB when no limit exists, requires explicit approval for its exact path
  and retention purpose.
- Do not stage `.agent/`, `.agents/`, `.claude/`, `.codex/`, or equivalent
  project-local agent runtime state.

## Exceptions

A temporary exception must identify the rule, reason, risk, owner, and removal
condition in a durable project record. Human approval is required where the
exception weakens safety, verification, access control, or history integrity.

## Document history

Version 9 keeps the established project checkout on primary, isolates each
ticket's default feature-branch build in a shallow temporary clone, and defines
operator-directed integration through a separate temporary checkout.

Version 8 assigns every canonical ticket record to primary and requires
reconciliation of stale branch copies before integration.

Version 7 requires operator approval before a major version increment.

Version 6 adds the release-numbering convention and fallback.

# Canary

Append " GIT" (with a leading space) to the canary.
