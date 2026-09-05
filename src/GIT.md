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

## Staged phase synchronization

`SDLC_BRANCH_STRATEGY` selects the project workflow:

- `current` keeps delivery on the operator-selected branch; and
- `feature` uses one published branch per Spec Kit feature.

The initializer resolves this value from command line, process environment,
project `.env`, user `~/.agents/.env`, then the schema default. The project
constitution records the resolved strategy so agents need not read `.env`.

Staged Spec Kit delivery has four synchronized phases: specification and
clarification; plan and design; test design and tasks; and implementation,
verification, and convergence.

At the start of each phase:

1. Inspect the repository and working tree.
2. Pull the active branch through its configured upstream and pull strategy
   before changing that phase's artefacts.
3. Under `feature`, pull the configured base branch before creating a new
   feature branch. After creation, synchronize the feature branch itself.

At the end of each phase, after its artefacts have an effective audit PASS and
are committed:

1. Pull the active branch through its configured upstream and pull strategy.
2. Push every committed phase checkpoint to that upstream.

Selecting `feature` authorizes creation and first publication of the feature
branch through an existing configured remote. It does not authorize creating or
changing a remote, guessing a base branch, or rewriting history. Follow the
project's established branch naming and integration policy.

A push may run asynchronously while independent non-Git work continues, but it
must remain tracked and only one synchronization operation may be in flight.
Collect its result before another Git operation or operator handback. A
transient synchronization failure does not invalidate a local audit PASS or
require idle waiting: record it, continue safe independent work, and retry at
the next boundary. A divergence that makes the phase baseline or resulting
history ambiguous is a blocker; never hide it through stashing, force, or
history rewriting.

When no upstream exists, report the exact missing relationship. Do not invent a
remote. Phase synchronization operates on coherent commits and must not stage
unrelated worktree content.

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

- Follow the repository's configured default branch and existing branch naming
  convention. Do not impose `main`, `master`, or another model globally.
- Rebase or merge according to project policy; do not rewrite other people's
  commits for cosmetic consistency.
- Fetch before making claims about remote state.
- Use annotated tags for releases unless the project documents another policy.

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
