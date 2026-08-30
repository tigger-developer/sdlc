# SDLC v2 Quickstart

> **Prerequisite:** Install
> [GitHub Spec Kit](https://github.com/github/spec-kit) 1.0 or later and ensure
> the `specify` CLI is available on `PATH`. SDLC v2 uses Spec Kit for project
> initialization, specification, planning, tasks, analysis, and implementation.

## Install SDLC v2

Building or installing from source requires a Go toolchain and `rsync`. The
installed Go executables do not require a separate Go runtime.

The SDLC installer also installs the `sdlc-install` and `sdlc-project-init`
helpers under `~/.local/bin`.

```bash
git clone https://github.com/tigger-developer/sdlc.git ~/code/sdlc
```

```bash
cd ~/code/sdlc
```

```bash
make install
```

Verify both prerequisites:

```bash
specify --version
```

```bash
sdlc-project-init --help
```

The installer deploys the canonical standards to `~/.agents/sdlc`. Provider
directories contain discovery adapters only; they are not alternate standards
roots.

## Initialize a greenfield project

Create the repository according to local conventions and establish a Git
checkpoint before tracked work begins. From the project root, run:

```bash
sdlc-project-init
```

When `.specify/` is absent, the initializer:

1. asks whether it may run `specify init`;
2. asks for the agent integration when no default is configured;
3. installs the `sdlc-standards` preset;
4. asks whether the project is greenfield or brownfield;
5. discovers the available technology standards and asks which apply;
6. asks whether another project owns deployment or runtime infrastructure;
7. records the project selections in an ignored `.env`;
8. renders `.specify/templates/overrides/constitution-template.md`;
9. commits only that generated scaffold; and
10. launches the selected agent to create an unratified project constitution.

The generated scaffold records an exact SDLC release tag when the initializer
was built at that tag; other clean versioned builds record their source commit.
A modified or unversioned build leaves an explicit `TODO(SDLC_REVISION)` and
cannot be ratified until that traceability is resolved.

Review the generated constitution before ratification. It should contain the
fixed standards and audit baseline, up to four durable project-wide principles,
one concern-specific authority hierarchy, explicit governance, and no copied
feature requirements or detailed design.

The initializer checkpoints the generated scaffold without including unrelated
staged or working-tree changes. Track the remaining project configuration needed
to reproduce the constitution, including:

- `.specify/templates/overrides/constitution-template.md`;
- `.specify/memory/constitution.md`;
- the installed preset and shared Spec Kit infrastructure under `.specify/`;
  and
- any project documentation created for the constitution.

The scaffold has no authority before ratification. Review the complete
constitution on its own merits and correct or remove any unsuitable scaffold
clause. Ratification and later amendment reports belong in the append-only
`.specify/memory/constitution-changelog.md`, not inside the constitution.

Keep `.env` and project-local agent runtime directories such as `.agents/`,
`.claude/`, and `.codex/` untracked. Resolve every ratification TODO, explicitly
ratify the constitution as version 1.0.0, and checkpoint the resulting project
state before specifying the first feature.

## Initialize a brownfield project

Start from a recoverable Git checkpoint. Existing unrelated or human-authored
changes are not an initialization workspace; checkpoint or separate them
according to the project's normal source-control practice before proceeding.

Run the same initializer from the existing project root:

```bash
sdlc-project-init
```

After authorization, Spec Kit merges its project infrastructure into the
existing working tree with its `--here --force` initialization path. For the
established SDLC v1 document shape, the initializer then performs a bounded,
mechanical migration:

1. `docs/implementation_plan.md` moves unchanged to
   `docs/archive/implementation_plan.md`.
2. A fixed, marked legacy introduction is added to `docs/ACs.md`.
3. The initializer stages only those paths, shows their Git diff, and asks
   whether to commit.

Vision, architecture, and README documents remain active. The initializer does
not rewrite them. Adapting those documents to the adopted process requires
project-specific semantic review.

Declining leaves the reviewed migration staged and stops before constitution
generation. Rerunning presents the same diff for confirmation. An accepted or
already-current migration is silent on later runs. Brownfield repositories
without this document shape are not altered.

For brownfield projects:

- the constitution populates the fixed `Specification Baseline` with the exact
  requirement, design, historical-authority, and regression-lineage sources
  that later delta specifications must consult;
- project documentation is evidence for durable constitutional invariants, not
  text to copy into the constitution;
- existing requirements, acceptance criteria, code, and tests describe the
  as-built baseline with their existing approval and traceability status;
- the constitution records project-wide principles and authority, not the
  feature catalogue or detailed architecture;
- a new Spec Kit feature specification defines its change as a delta against
  the relevant existing behaviour; and
- superseded requirements and regression lineage remain discoverable.

Initialization does not bulk-migrate legacy tickets or acceptance criteria.
Reconcile that material separately and bring forward only the baseline relevant
to each new or migrated feature.

Review and track the same `.specify/` constitution and scaffold files listed in
the greenfield procedure. Confirm that the generated authority hierarchy gives
approved specifications authority over observable behaviour, approved design
authority over technical choices within those requirements, and code and tests
the status of implementation evidence rather than requirement approval.

## Non-interactive selection

Project or automation inputs may be supplied explicitly:

```bash
sdlc-project-init \
    --harness codex \
    --project-type brownfield \
    --technologies GO,WEB \
    --infra no
```

Use `--no-launch` to render and inspect the scaffold without starting an agent.
Any required mechanical brownfield migration still displays its diff and asks
for commit confirmation:

```bash
sdlc-project-init --no-launch
```

The initializer reads user defaults from `~/.agents/.env` and project overrides
from the ignored project `.env`. Every resolved global default is copied into a
new project's `.env`; changing the global file later does not silently change
existing projects. Project classification is project-only: supply
`--project-type`, record `SDLC_PROJECT_TYPE` in the project `.env`, or answer the
initializer prompt. A value in `~/.agents/.env` is ignored. Command-line values
take precedence. Run `sdlc-project-init --help` for the complete interface.

## Rerun and update

An unchanged rerun is silent: it writes nothing, asks nothing, and does not
launch an agent.

```bash
sdlc-project-init
```

To change a recorded selection, update the applicable `SDLC_*` value in the
project `.env` or supply a command-line override. A changed standards selection
regenerates and checkpoints the scaffold before relaunching the constitution
workflow. Review and checkpoint the resulting constitution change.

After updating the SDLC staging clone, redeploy and rerun initialization so the
project records the new adopted revision:

```bash
make install
```

```bash
sdlc-project-init
```
