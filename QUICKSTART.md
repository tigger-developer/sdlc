# SDLC v2 Quickstart

> **Prerequisite:** Install
> [GitHub Spec Kit](https://github.com/github/spec-kit) 1.0 or later and ensure
> the `specify` CLI is available on `PATH`. SDLC v2 uses Spec Kit for project
> initialization, specification, planning, tasks, analysis, and implementation.

## Install SDLC v2

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
4. discovers the available technology standards and asks which apply;
5. asks whether another project owns deployment or runtime infrastructure;
6. records the project selections in an ignored `.env`;
7. renders `.specify/templates/overrides/constitution-template.md`; and
8. launches the selected agent to create an unratified project constitution.

The generated baseline records the exact SDLC commit or release when the
initializer was built from clean versioned source. A modified or unversioned
build leaves an explicit `TODO(SDLC_REVISION)` and cannot be ratified until that
traceability is resolved.

Review the generated constitution before ratification. It should contain the
fixed standards and audit baseline, up to four durable project-wide principles,
one concern-specific authority hierarchy, explicit governance, and no copied
feature requirements or detailed design.

Track the project configuration needed to reproduce it, including:

- `.specify/templates/overrides/constitution-template.md`;
- `.specify/memory/constitution.md`;
- the installed preset and shared Spec Kit infrastructure under `.specify/`;
  and
- any project documentation created for the constitution.

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
existing working tree with its `--here --force` initialization path. The SDLC
initializer does not replace application source or project documentation. It
selects standards and asks the constitution agent to read the existing
documentation as evidence.

For brownfield projects:

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

Review and track the same `.specify/` constitution and baseline files listed in
the greenfield procedure. Confirm that the generated authority hierarchy gives
approved specifications authority over observable behaviour, approved design
authority over technical choices within those requirements, and code and tests
the status of implementation evidence rather than requirement approval.

## Non-interactive selection

Project or automation inputs may be supplied explicitly:

```bash
sdlc-project-init \
    --harness codex \
    --technologies GO,WEB \
    --infra no
```

Use `--no-launch` to render and inspect the baseline without starting an agent:

```bash
sdlc-project-init --no-launch
```

The initializer reads user defaults from the platform user configuration
directory under `sdlc/.env` and project overrides from the ignored project
`.env`. Command-line values take precedence. Run `sdlc-project-init --help` for
the complete interface.

## Rerun and update

An unchanged rerun is silent: it writes nothing, asks nothing, and does not
launch an agent.

```bash
sdlc-project-init
```

To change a recorded selection, update the applicable `SDLC_*` value in the
project `.env` or supply a command-line override. A changed standards selection
regenerates the baseline and relaunches the constitution workflow. Review and
checkpoint both the generated baseline and constitution change.

After updating the SDLC staging clone, redeploy and rerun initialization so the
project records the new adopted revision:

```bash
make install
```

```bash
sdlc-project-init
```
