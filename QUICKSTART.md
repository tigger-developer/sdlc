# SDLC v2 Quickstart

> **Prerequisite:** Install
> [GitHub Spec Kit](https://github.com/github/spec-kit) 1.0 or later and ensure
> the `specify` CLI is available on `PATH`. SDLC v2 uses Spec Kit for project
> initialization and staged specification, planning, tasks, analysis, and
> implementation. Explicitly paired development remains available for bounded
> work refined through live operator review.

## Install SDLC v2

Building or installing from source requires a Go toolchain and `rsync`. The
installed Go executables do not require a separate Go runtime.

The SDLC installer also installs the `sdlc-install`, `sdlc-project-init`, and
`sdlc-audit` helpers under `~/.local/bin`.

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

```bash
sdlc-audit audit-spec --help
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
clause. Keep the current compact Sync Impact Report as the first line of the
constitution. Replace it on amendment rather than accumulating report history.

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
2. The file-backed, marked legacy prefix is applied to `docs/ACs.md`.
3. The initializer stages only those paths, shows their Git diff, and asks
   whether to commit.

Vision, architecture, and README documents remain active. The initializer does
not rewrite them. Adapting those documents to the adopted process requires
project-specific semantic review.

Declining leaves the reviewed migration staged and stops before constitution
generation. Rerunning presents the same diff for confirmation. An accepted or
already-current migration is silent on later runs. Brownfield repositories
without this document shape are not altered.

The managed prefix ends at a line containing only `***`. The initializer hashes
that prefix independently of the existing ledger: an absent marker causes a
prepend, a matching prefix is a no-op, and an older managed prefix is replaced
through the delimiter. The remainder of `docs/ACs.md` is preserved byte for
byte. A marker outside the expected prefix is an error.

For brownfield projects:

- the constitution populates the fixed `Specification Baseline` with the exact
  requirement, design, historical-authority, historical-work, and
  regression-lineage sources that later delta specifications and plans must
  consult;
- project documentation is evidence for durable constitutional invariants, not
  text to copy into the constitution;
- existing requirements and acceptance criteria define intended behaviour,
  while code and tests provide evidence of the as-built baseline with their
  existing traceability status;
- the constitution records project-wide principles and authority, not the
  feature catalogue or detailed architecture;
- a new Spec Kit feature specification defines its change as a delta against
  the relevant existing behaviour; and
- superseded requirements and regression lineage remain discoverable.

Before drafting a brownfield specification or plan, the agent performs a
bounded context pass for the affected area. It reads the relevant requirement
and design authorities, follows traceability into tickets and comments or
equivalent work records, inspects the maintained regression tests and their
assertions, and checks the affected implementation. The resulting artefact
states what it preserves, changes, supersedes, and leaves unaffected. Tests and
code remain implementation evidence rather than requirement approval.

Initialization does not bulk-migrate legacy tickets or acceptance criteria.
Reconcile that material separately and bring forward only the baseline relevant
to each new or migrated feature.

Review and track the same `.specify/` constitution and scaffold files listed in
the greenfield procedure. Confirm that the generated authority hierarchy gives
approved specifications authority over observable behaviour, approved design
authority over technical choices within those requirements, and code and tests
the status of implementation evidence rather than requirement approval.

## Run an interactive paired change

Use paired development when the outcome must be refined through live operator
review, such as a Hugo layout, typography, navigation, or other visual change.
Select it explicitly and state the bounded objective:

```text
Use the SDLC paired-development track for this Hugo change. Preserve the
existing deployment contracts. Start by making the requested navigation change
and present the rendered result for review.
```

The agent then works in reviewable slices. Each explicit iteration instruction
authorizes that slice; questions and requests for advice do not. Existing
standards and project constraints remain mandatory.

During the session, the agent retains explicit user validations without asking
for a separate sign-off after each one. At closure it reports:

- what changed;
- current and superseded user validations;
- objective checks and their results;
- unvalidated behaviour and justified automation gaps; and
- applicable change-scoped audit results.

The operator then confirms whether the listed validations may be recorded in
the active feature's mandatory `validation.md`. Outside Spec Kit, use the
project's equivalent evidence record. Do not add an automated test merely to
imitate a visual validation.

Run `audit-code` once at closure when the change includes new or materially
modified code, templates, scripts, or non-trivial CSS. Run specification,
design, or test audits only when the paired change produces material artefacts
of those kinds. If code-audit remediation changes behaviour already validated
by the operator, repeat only the materially affected user tests and preserve the
earlier entries as superseded. See `~/.agents/sdlc/PAIRING.md` for the complete
contract.

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

Set `SDLC_AUDIT_HARNESS`, `SDLC_AUDIT_PROVIDER`, and `SDLC_AUDIT_MODEL` in
`~/.agents/.env` to select the default independent auditor for new projects.
The initializer snapshots those values into the project `.env`, whose values
take precedence when the runner executes. This release always invokes Hermes
and passes provider and model explicitly. An unset or `hermes` harness value is
silent; any other value warns and falls back to Hermes.

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
