# SDLC v3 Quickstart

## 1. Install the framework

Prerequisites are Go, Git, Codex, and Pandoc. GitHub CLI is required only for
legacy GitHub-ticket migration.

```sh
make install
```

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
sdlc-project-init
```

The initializer prints a concise migration commit summary. Use
`VERBOSE=1 sdlc-project-init` only when the complete Git changed-file inventory
is useful.

The command creates the archive branch before making changes. Review any push
and final merge question by branch descriptor, not by an unexplained name.

After migration, the initializer lists README, vision, and architecture
Markdown or Org files from the bounded Git inventory. Selected entries use
`[x]`; enter numbers to toggle them, `r` to rescan after moving a file, or press
Enter to accept. Exact path case is preserved. `docs/work.org` and the single
applicable AC ledger are recorded automatically rather than offered as choices.

Only an SDLC v2 project with archived Spec Kit specifications starts headless
Codex. The configured audit model classifies those specifications in one
bounded read-only pass, and the initializer renders their statuses into
`docs/work.org`. Projects without archived specifications skip this call.

If initialization is interrupted during or after authority selection, its
temporary working directory remains at `.sdlc/.init/`. A later invocation
reports that path instead of treating the project as a fresh start. The
initializer records `.init/` in `.sdlc/.gitignore` so ordinary Git operations
cannot commit the temporary state.

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
- The initializer then records unresolved defects and undelivered ideas in the
  v3 work ledger without re-specifying them.

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

## 4. Define a change

Invoke:

```text
$define-change
```

Review the generated `spec.org` in the browser. It must be sufficient for a new
agent with no drafting-conversation context.

Run the combined gate:

```text
$audit-definition
```

The gate applies specification, design, and test-definition audits locally,
then in one retained external Codex context. Sign off only after PASS.

## 5. Deliver the signed-off change

Invoke:

```text
$deliver-change
```

The delivery sequence is automated test, RED, implementation, GREEN. Test
definitions classified as OT or UT are recorded but do not pretend to be TDD.

Run the combined gate:

```text
$audit-implementation
```

After PASS, execute the selected OT and UT checks, record results in
`validation.org`, reconcile documentation, and decide closure.

## 6. Use a variant only when selected

For live paired implementation:

```text
$pair-change
```

For an emergency, include the exact token in the same instruction:

```text
BYPASS-GATE-7
```

The agent cannot infer or self-authorize the emergency route.

## 7. Preview an artefact

```sh
sdlc-preview specs/001-example/spec.org
```

The generated HTML uses the source directory so relative links resolve, and is
removed one second after the browser opens it.
