# Git and Source Control

## Fundamental Rules

- **Never use `--no-verify`** when committing. Ever.
- Refuse to work on code not in a git repository.
- Before moving, changing, or removing files, ensure they're tracked and recoverable.
- Every project `.gitignore` must ignore `.agent/`, `.agents/`, `.claude/`, and `.codex/` in their entirety. These are runtime-state directories, not project content.
- Files beneath those directories must never be tracked. If any are already tracked, alert the user immediately; do not silently untrack or delete them.
- Before committing, inspect staged and tracked files for unexpectedly large generated artefacts. A file larger than 100 MiB is a hard block unless its project path and retention purpose have explicit human approval.

## Forbidden Flags

These are prohibited:
- `--no-verify`
- `--no-hooks`
- `--no-pre-commit-hook`
- NEVER include AI attribution in commits or attribution lines (e.g. "Co-authored-by Claude" or equivalent)

Before using any unfamiliar git flag: state it, explain why, confirm it's not forbidden, get permission for any bypass.

## Commit Standards

Messages must be:
- Concise and descriptive
- Imperative mood, present tense
- Conventional commit format

```
feat: add user authentication
fix: resolve memory leak in data processor
docs: update API documentation
refactor: simplify validation logic
```

### Issue-linked commits

When a commit implements work tracked by a GitHub issue, use:

```
Implement #N: short description
```

Checkpoint commits within an issue should be useful version-history markers, not declarations of closure. Use concise messages that describe the checkpoint, for example:

```
Implement #N: add failing regression tests
Implement #N: pass issue tests
Implement #N: update review documentation
```

**Never use GitHub auto-close keywords** (`Fixes`, `Closes`, `Resolves`, or their variants) in commit messages. Issues are closed manually after human review. Auto-closing bypasses that review step.

## Workflow

Before starting work:

```bash
git pull
git status
```

Ensure local repo is current with remote.

## Pre-Commit Hook Failures

When hooks fail, follow this sequence exactly:

1. Read complete error output (explain what you see)
2. Identify which tool failed and why
3. Explain the fix and why it addresses root cause
4. Apply fix, re-run hooks
5. Commit only after all hooks pass

Cannot fix? Ask for help. Never bypass.

## Runtime-State and Size Check

Before every commit and during review:

1. Confirm the four project-local agent runtime directories are ignored.
2. Enumerate tracked paths beneath those directories. Any result is a hard block and must be reported with the offending path.
3. Inspect the size of staged files and generated files touched by the work. Any file above the documented project limit, or 100 MiB when no project limit exists, is a hard block unless the human explicitly approved it as a persistent project artefact.

An ignore rule prevents new untracked files from being added; it does not make an already-tracked file safe. Disk-use checks are required independently because ignored temporary data can still exhaust local storage.

## Project Hooks Must Chain to Global Hooks

When a project needs its own git hooks, the project hook must call the global hook first - never silently replace it. A project-local `.git/hooks/pre-commit` or a project-level `core.hooksPath` overrides the global hook entirely unless chaining is explicit.

Use this pattern:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Chain to global hook first — project hooks must not override global hooks
if global_hooks_path="$(git config --global core.hooksPath 2>/dev/null)"; then
    if [[ -x "$global_hooks_path/${0##*/}" ]]; then
        "$global_hooks_path/${0##*/}"
    fi
fi

# Project-specific checks below
```

Notes:
- `${0##*/}` resolves to the hook's own name (`pre-commit`, `pre-push`, etc.), so the pattern works for any hook type without modification.
- `set -e` ensures the global hook's exit code propagates - if the global hook fails, the project hook fails too.
- The `if` around `git config` handles the case where no global `core.hooksPath` is configured.

### Setup

Hooks in `.git/hooks/` are not version-controlled. To track project hooks in the repo:

1. Store hooks in a `hooks/` directory at the project root.
2. Copy them into `.git/hooks/` via a Makefile target:

```makefile
init:
	cp hooks/* .git/hooks/
```

This preserves the default hooks path, so any hook types the project doesn't define (e.g. `pre-push`, `commit-msg`) continue to work from the global configuration. Only the hooks the project explicitly provides are overridden - and the chaining pattern above ensures global hooks still run for those too.

## Pressure Response

When asked to commit/push with failing hooks:

- Do not rush to bypass quality checks
- State: "Pre-commit hooks are failing, I need to fix those first"
- Work through systematically
- Quality over speed, even when waiting

Pressure is never justification for bypassing checks.

## Accountability

Before any git command, ask:

- Am I bypassing a safety mechanism?
- Would this violate the provider-level instructions?
- Am I choosing convenience over quality?

If any answer is "yes" or "maybe", explain the concern first.

## Branch Strategy

Default: work on master.

The main branch is always `master`. If a repository's primary branch is named anything else, alert the user immediately before doing branch-sensitive work. Do not silently normalize, rename, or proceed as if `main` were acceptable.

If the issue specifies a feature branch:
- Create the branch
- Work only on that branch
- Push to origin

Alternatives available: worktrees, feature branches.

## Exception Process

When a rule genuinely cannot be followed (not "is inconvenient"):

1. **Document the situation** - what rule, why it can't be followed
2. **Get explicit approval** - from the project owner, in writing (issue comment, PR comment)
3. **Record the exception** - in the commit message or PR description
4. **Set an expiry** - exceptions should be temporary; create a follow-up issue to resolve properly

Format for commit messages with exceptions:
```
fix: emergency patch for production outage

EXCEPTION: Skipping integration tests - test DB unavailable
APPROVED BY: @owner in #123
FOLLOW-UP: #124 to restore test coverage
```

Exceptions are not get-out-of-jail-free cards. They create technical debt that must be repaid.

# Canary
Suffix the canary string with "GIT "
