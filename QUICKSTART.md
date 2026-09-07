# SDLC v3 Quickstart

## 1. Install the framework

Prerequisites are Go, Git, Codex, and Pandoc. GitHub CLI is required only for
legacy GitHub-ticket migration.

```sh
make install
```

This builds the repository-internal installer, deploys the public commands under
`~/.agents/sdlc/bin`, and links those deployed copies onto the global path. The
links do not point back into the source checkout.

Use non-interactive application only when the listed variances have already been
reviewed:

```sh
INSTALL_FLAGS=--apply make install
```

## 2. Set optional global defaults

Create `~/.agents/sdlc.yaml` with non-secret defaults. Omit anything that should
be decided per project. A project may inherit a global value without copying it
or record an explicit override in `.sdlc/project.yaml`.

## 3. Initialize one project

Start on the branch that represents the project state to preserve. Ensure the
worktree is clean, then run:

```sh
sdlc-init
```

The initializer prints a concise migration commit summary. Use
`VERBOSE=1 sdlc-init` only when the complete Git changed-file inventory
is useful.

Before the technology question, the configured audit model performs one
bounded, read-only stack assessment. Its evidence-backed recommendations appear
as `[x]` selections. Press Enter to accept them or enter the complete corrected
selection. The assessment ignores archives, generated output, dependencies,
framework-owned files, and incidental tooling. A truly blank project receives
no inferred selection.

The command creates the archive branch before making changes. Review any push
and final merge question by branch descriptor, not by an unexplained name.

After migration, the initializer lists README, vision, and architecture
Markdown or Org files from the bounded Git inventory. Selected entries use
`[x]`; enter numbers to toggle them, `r` to rescan after moving a file, or press
Enter to accept. Exact path case is preserved. `docs/work.org` is recorded
automatically as the sole requirement authority rather than offered as a
choice. Any canonical legacy AC ledger is folded into it first.

An SDLC v2 project with archived Spec Kit specifications starts one additional
headless Codex context. The configured audit model classifies those
specifications in one bounded read-only pass. The initializer removes their
obsolete top-level Status fields and renders their lifecycle into
`docs/work.org`. Projects without archived specifications skip this second
model operation.

If initialization is interrupted during or after authority selection, its
temporary working directory remains at `.sdlc/.init/`. Rerun
`sdlc-init` on the migration branch. It resumes when the matching dated
archive and original primary branch are unambiguous; otherwise it stops without
creating another branch. The initializer records `.init/` in
`.sdlc/.gitignore` so ordinary Git operations cannot commit the temporary state.

### New project

- Select the project role and applicable technology standards.
- Select any existing product and architecture authorities. Empty lists are
  valid for a genuinely blank project.
- Record product and architecture documents as they are created.
- Use `docs/work.org` for defects, ideas, active work, review, and closed work.
- Start the first change with `$define-change`.

### SDLC v1 project

- When any historical GitHub tickets and `docs/ACs.md` are found, the
  initializer offers `$migrate-legacy-acs-to-sdlc-v1` before the general
  project-profile questions. Declining skips that optional workflow and
  continues initialization from the existing AC ledger.
- The migration runs the supported regression suite once and stops if it fails.
- It archives every ticket and comment, creates `docs/ACs.org` and
  `docs/ticket-migration.org`, refreshes stale documentation, archives the old
  implementation plan, commits the evidence, then closes the legacy tickets.
- If the resulting `docs/ACs.org` cannot be interpreted safely by the canonical
  importer, initialization invokes one headless Codex repair using the
  configured audit model. It continues only when the repaired ledger validates.
- The initializer folds `docs/ACs.org` beneath the initially collapsed `Legacy
  Acceptance Criteria (SDLC v1)` section of `docs/work.org`, removes the
  redundant file, then records unresolved defects and undelivered ideas without
  re-specifying them.

### SDLC v2 project

- The archive branch preserves the complete pre-migration repository.
- Active `.specify` and `specs` material moves unchanged under
  `docs/archive/sdlc-v2/`.
- Project-local Spec Kit integrations move into the same archive; unrelated
  project-local integrations remain active.
- Preserved work is indexed for operator disposition. It is converted to the v3
  unified specification only when the operator resumes it. New v3 work IDs are
  allocated above every historical work, AC, test, and ticket number; the v2 ID
  remains recorded as provenance.

### Already initialized project with a separate legacy ledger

From the project root, run:

```sh
sdlc-merge-legacy-acs
```

The command preserves an existing `docs/work.org`, folds the complete canonical
`docs/ACs.org` hierarchy into it, updates `.sdlc/project.yaml` when present, and
removes the redundant ledger. Rerunning after success changes nothing. If an
interrupted run leaves two differing copies, the command stops rather than
selecting one.

## 4. Define a change

Invoke:

```text
$define-change
```

Review the generated `spec.org` in the browser. It must be sufficient for a new
agent with no drafting-conversation context. The same skill runs the combined
specification, design, and test-definition gate locally, then in one retained
external Codex context. Sign off only after `spec.org` records the resulting
PASS and `docs/work.org` places the item in `REVIEW`.

## 5. Deliver the signed-off change

Invoke:

```text
$deliver-change
```

Delivery requires an `ACTIVE` work item whose specification records a current
definition PASS and operator sign-off. It does not rerun definition audits. The
sequence is automated test, RED, implementation, GREEN, then the combined
implemented-test and code gate. Test definitions classified as OT or UT are
recorded but do not pretend to be TDD. After PASS, execute them, record results
in `validation.org`, reconcile documentation, and decide closure.

Use `$audit-definition`, `$audit-implementation`, or any focused audit
independently when a separate review is wanted.

## 6. Use a variant only when selected

For live paired implementation:

```text
$pair-change
```

For an emergency, include the exact token in the same instruction:

```text
BYPASS-GATE-7 <bounded emergency change>
```

The exact token invokes `$emergency-change`. The agent cannot infer, suggest, or
self-authorize the route.

## 7. Preview an artefact

```sh
sdlc-preview specs/001-example/spec.org
```

The generated HTML uses the source directory so relative links resolve, and is
removed one second after the browser opens it.
