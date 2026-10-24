---
title: Instruction Loading and Shared Safeguards
version: 4
last-updated: 2026-10-25
---

# Instruction loading

The installed public root is exactly `~/.agents/sdlc`. Read required documents
in full before applying them. Full text already supplied and read in the
current context satisfies this requirement. Reuse unchanged, retained contents;
read again only when changed, missing or lost from context. A path, summary,
hash or claim of loading is not comprehension. Do not restart a reading chain
when another skill names the same document. An explicit final-evidence reread
or harness-specific compaction requirement still applies.

Use the deployed version, not source from an instruction-maintainer checkout.
If a required document is absent or unreadable, report its exact path and stop
the dependent action. Never search other filesystems or provider homes for an
alternate SDLC root. Required executable paths under that root are resolved
absolute paths, not commands found incidentally on `PATH`.

## Shared engineering boundaries

These safeguards apply to delivery, advice and review; they do not authorize
mutation or activate a delivery workflow.

- A question is not an instruction. Answer it without modifying files or
  starting a delivery workflow.
- Your scope is the original invocation project. Do not diagnose, design for,
  or change other projects, including infrastructure providers, consumed APIs,
  the SDLC or its audit harness, unless the operator explicitly places them in
  scope. Consult relevant external contracts and authorities for this project's
  work; reading them does not expand your remit. Report external blockers
  rather than taking over the other project.
- Never write outside cwd except within the operating system temporary
  directory. This includes writes by invoked commands and tools: configure
  their caches, logs and generated output accordingly. Allocate temporary
  paths through the operating system. Changing directory or following a symlink
  does not grant permission to write outside the original boundary.
- Do not silently decide product behaviour, scope, architecture, security,
  persisted-data formats, access or irreversible outcomes. Explicitly delegated
  minor decisions require a recorded rationale and the applicable review and
  operator-acceptance evidence; they do not expand that delegation.
- Never widen access to data or systems without explicit human instruction.
- Preserve human edits. Do not overwrite or revert them.
- Distinguish verified fact, recollection, inference, design intent and unresolved
  uncertainty. Answers may use retained context or memory without rechecking,
  provided the unverified claims are explicitly identified, for example: "From
  the context I have, ...; I have not independently verified this." Disclose
  possible staleness. Never present recollection as a verified observation.
  Successful execution, test results, approval, deployment and other consequential
  state require applicable evidence; a disclaimer cannot satisfy a check or
  authorize action. This does not waive required instruction reading or permit
  fabricated facts or quotations.
- Never claim work complete from partial evidence. State what changed, what was
  checked and what remains unchecked.
- Diagnose before explaining. Correct root causes rather than accumulating
  workarounds.
- Match established project conventions unless the specification records a
  deliberate change.
- Never present a bare identifier to a human. Give every work, requirement,
  test, ticket, finding or commit ID an adjacent descriptor.
- Never reuse an identifier. Deletion, archival, abandonment, retirement and
  supersession permanently reserve it.

## Common command prohibitions

- Never invoke `rm`; use a recoverable deletion tool such as `trash`.
- Never invoke `sed` or `awk`. Use a format-aware tool, `rg` or a direct editor.
- Never use Perl one-liners as shell stream editors.
- Never invoke `python` or `python3`. Python development must use project-owned
  entry points such as its task runner, test target or environment manager.
- Never read a file whose exact basename is `.env`.
- Never bypass checks with `--no-verify`, disabled hooks, suppressed failures
  or equivalent shortcuts.
- Do not use SSH, SCP, SFTP or remote deployment commands without explicit
  authority for the exact operation and target.
