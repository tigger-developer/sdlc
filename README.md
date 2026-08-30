# Engineering Standards for AI Coding Agents

> **Prerequisite:** SDLC v2 requires
> [GitHub Spec Kit](https://github.com/github/spec-kit) 1.0 or later. Install
> the `specify` CLI and ensure it is available on `PATH` before installing or
> initializing this SDLC.

This public repository provides a provider-neutral engineering standards
library for coding agents. Spec Kit owns
specification and delivery orchestration; this repository supplies the coding,
testing, Git, documentation, language, and domain standards applied within it.

The library does not require private agent instructions. Its canonical installed
root is exactly `~/.agents/sdlc` for every supported provider.

## What is included

| Path | Purpose |
|---|---|
| `src/MAIN.md` | Compact universal rules and progressive-loading routes |
| `src/ISSUES.md` | Specification and acceptance-criteria standards |
| `src/TESTING.md` | Behavioural testing and evidence standards |
| `src/CODING.md` | Cross-language implementation standards |
| `src/GIT.md` | Source-control and recoverability standards |
| `src/DOCUMENTATION.md` | Public technical-documentation standards |
| `src/technologies/*.md` | Automatically discoverable technology standards |
| `src/presets/sdlc-standards/` | Spec Kit preset that selects standards progressively |
| `src/prompts/project-init/` | Constitution-generation prompt resource |
| `src/templates/project-init/` | Constitution scaffold and managed brownfield document block |
| `skills/` | Findings-only audits and advisory tools |
| `hooks/` | Optional provider-integrated command safeguard |
| `cmd/` and `internal/` | Installer, project initializer, and audit-verdict implementation |

Only runtime material from `src/`, `skills/`, and `hooks/` is
deployed. Repository documentation, installer source, tests, and build metadata
do not enter the live instruction tree.

## Install the standards

Building or installing from source requires Spec Kit 1.0 or later, a Go
toolchain, `rsync`, and the tools required by the repository's build targets.
The installed Go executables do not require a separate Go runtime. The complete
greenfield and brownfield setup is documented in
[`QUICKSTART.md`](QUICKSTART.md).

```bash
git clone https://github.com/tigger-developer/sdlc.git ~/code/sdlc
```

```bash
cd ~/code/sdlc
```

```bash
make install
```

The installer:

- installs `sdlc-install` and `sdlc-project-init` under `~/.local/bin`;
- synchronizes `src/` into `~/.agents/sdlc` and retains the other runtime
  directory names;
- installs common and provider-native skill copies for detected agents where
  required for discovery;
- compares source and destination before prompting;
- lists only missing or differing destinations by default;
- backs up owned drift before replacement;
- backs up and retires the known command and drafting-skill paths removed by
  the standards-only model; and
- performs no deployment prompt or write when every detected destination is
  current.

Set `VERBOSE=1` to include matching destinations in the plan:

```bash
VERBOSE=1 make install
```

To inspect installer options:

```bash
make build
```

```bash
bin/sdlc-install --help
```

Provider-specific homes are adapters, not alternate SDLC roots. Agents must
load standards from `~/.agents/sdlc` and must never search the filesystem for a
different copy.

## Initialize a Spec Kit project

Run the initializer from the adopting project:

```bash
sdlc-project-init
```

The initializer separates deterministic selection from semantic drafting. It:

- initializes Spec Kit when authorized and absent;
- installs the deployed SDLC preset;
- asks whether the adopting project is greenfield or brownfield;
- discovers available standards from `~/.agents/sdlc/technologies/`;
- asks once which technologies and infrastructure ownership apply;
- renders the editable constitution scaffold at
  `.specify/templates/overrides/constitution-template.md`;
- pins the exact SDLC release tag used to build the initializer, or the source
  commit for another clean versioned build, and leaves an explicit ratification
  TODO for an unversioned or modified build;
- commits only the generated constitution scaffold before invoking an agent,
  without including unrelated staged or working-tree changes;
- reports only a changed template, with additional variance detail under
  `VERBOSE=1`; and
- invokes the selected agent harness to complete only the project-specific
  parts of the constitution.

The initializer writes only its named SDLC selections into the project `.env`
and adds that file to `.gitignore`; unrelated existing values are preserved.
User defaults come from `~/.agents/.env`. Every resolved default is copied into
a new project's `.env`, producing a stable project snapshot. Later changes to
the user defaults affect new projects, not existing snapshots. Command-line
values override project values, which override user defaults. Project
classification is deliberately project-only and is never read as a user
default. Supported keys are:

```text
SDLC_AGENT_HARNESS
SDLC_SPEC_PROVIDER
SDLC_SPEC_MODEL
SDLC_BUILD_PROVIDER
SDLC_BUILD_MODEL
SDLC_AUDIT_PROVIDER
SDLC_AUDIT_MODEL
SDLC_PROJECT_TYPE
SDLC_TECHNOLOGIES
SDLC_INFRA_ENABLED
SDLC_INFRA_OWNER
SDLC_INFRA_CONTRACT
```

`SDLC_PROJECT_TYPE` accepts `greenfield` or `brownfield`. Set it in the project
`.env` or with `--project-type`; a value in `~/.agents/.env` is ignored.

Specification settings apply to constitution, specification, clarification,
design, planning, and task-definition agent invocations. Build settings apply
to implementation and convergence. Audit settings apply to independent audit
invocations. The initializer accepts legacy `SDLC_DELIVERY_PROVIDER` and
`SDLC_DELIVERY_MODEL` values as specification defaults and rewrites project
snapshots using the new names.

The external owner and contract are optional project inputs. The public SDLC
does not assume any particular infrastructure project. Run
`sdlc-project-init --help` for non-interactive overrides and `--no-launch`.

When the rendered scaffold, brownfield migration, and selections are already current, the initializer
writes nothing, asks nothing, and does not relaunch an agent. If the current
generated scaffold is untracked or modified, it checkpoints that file without
relaunching the agent.

For a greenfield repository, the initializer offers to run `specify init`, then
creates the standards profile and unratified constitution. For a brownfield
repository, it preserves the existing project, installs Spec Kit into the
working tree, applies the current managed legacy prefix to the
acceptance-criteria ledger, archives
`docs/implementation_plan.md`, and shows the bounded Git diff before asking
whether to commit the migration. Vision, architecture, and README documents
remain active and require project-specific semantic review. The initializer then
generates a proposed `Specification Baseline` authority map for the constitution
agent to populate without copying feature or design detail. Projects without
that legacy document shape are left unchanged.

The managed `docs/ACs.md` prefix comes from the deployed SDLC resource rather
than Go source. A rerun compares the prefix through its `***` delimiter by
SHA-256. It prepends the block when absent, writes nothing when current, and
replaces only a stale managed prefix while preserving the ledger body byte for
byte. A marker outside the managed prefix is reported as an error.

The generated template has no authority before ratification. The candidate
constitution may correct, remove, or replace any scaffold clause. Ratification
makes `.specify/memory/constitution.md` the sole governance authority; impact
reports are append-only history in
`.specify/memory/constitution-changelog.md`, not embedded constitution content.
Follow the complete procedures in [`QUICKSTART.md`](QUICKSTART.md).

Thereafter:

- `speckit.specify` and `speckit.clarify` load specification standards, after
  which `audit-spec` must pass in a fresh context;
- `speckit.plan` loads cross-language and selected technology standards, after
  which `audit-design` must pass in a fresh context;
- `speckit.tasks` loads testing and documentation standards, after which
  `audit-tests` must pass in a fresh context;
- `speckit.analyze` checks consistency and coverage; and
- `speckit.implement` loads the selected standards, after which `audit-code`
  must pass in a fresh context before completion or convergence.

Each verdict is recorded in the active feature's `audits.md`. A finding,
missing or malformed verdict, or subsequent artefact change invalidates PASS.
`speckit.analyze` does not replace an independent audit.

`speckit.taskstoissues` retains Spec Kit's task artefacts as the source of truth
and applies the rule that human-facing identifiers always include descriptors.

The preset adds standards to these commands without replacing Spec Kit's
orchestration. The shared documents remain the single source of truth.

## Emergency exception

Spec Kit is the supported v2 project workflow. Provider or project instructions
may still direct an agent to read `~/.agents/sdlc/MAIN.md` for an isolated coding
task, but an equivalent durable project specification is required before code
is written.

For a small emergency change before project Spec Kit artefacts exist, the human
may include `BYPASS-GATE-7` in the request. The request then serves as a
temporary specification for its stated outcome and scope. The exception does
not restore the retired SDLC modes, gates, or ticket lifecycle and does not
override safety or engineering standards.

The public template at
[`templates/AGENTS-or-CLAUDE.example.md`](templates/AGENTS-or-CLAUDE.example.md)
shows the minimum standalone integration. Adapt provider discovery filenames and
locations to current provider documentation.

## Skills and safeguards

The audit skills are findings-only and have a common machine-checkable verdict:

- `audit-spec` challenges requirements and acceptance criteria;
- `audit-design` challenges boundaries, trade-offs, failure handling, migration,
  operability, security, and testability;
- `audit-tests` challenges evidence and coverage; and
- `audit-code` reviews implementation against the selected standards.

Advisory skills diagnose problems, recommend technical decisions, summarize
open work, or load minimal project context. Audits run in a fresh context, never
modify the judged artefact, and identify both provider and model in their
verdict. A skill's frontmatter may recommend an inexpensive audit provider and
model; runtime configuration has precedence.

The explicit-only `migrate-legacy-acs-to-sdlc-v1` skill snapshots every GitHub
issue, including comments and recorded implementation links, then reconciles
ticket-based SDLC v0.1 acceptance criteria and test lineage into the centralized
SDLC v1 record required as a brownfield baseline for SDLC v2. Tickets without
AC tables are ignored as bug fixes; multiple tables and unresolved test
evidence are reserved for operator adjudication.

The optional Hermes hook and provider rules reinforce the common prohibitions
on `rm`, `sed`, `awk`, and direct `python` or `python3` interpreter commands.
They do not prevent a documented project target from invoking its own runtime.

Hermes must create `~/.hermes/config.yaml` through its startup configuration
before the installer can safely amend it. If the Hermes home exists without that
file, installation stops with a prerequisite diagnostic.

## Updating and rollback

Update the staging clone, inspect the branch, and redeploy:

```bash
git pull --ff-only
```

```bash
make install
```

Release tags provide stable rollback points. Check out the required release and
rerun `make install` to redeploy its managed runtime files. Arbitrary
destination-only files are preserved. The installer recognizes only the known
commands, skills, root-level technology files, and obsolete constitution
addendum retired by the v2 migration. It renames them to adjacent
`<path>.<epoch>.bak` backups and leaves all other destination-only material
untouched.

## Further reading

- [`QUICKSTART.md`](QUICKSTART.md) covers installation and greenfield and
  brownfield initialization.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) describes source, deployment,
  loading, and Spec Kit composition.
- [`LEARNINGS.md`](LEARNINGS.md) records the design lessons behind the current
  standards model.
- [`CHANGELOG.md`](CHANGELOG.md) records repository changes.

## Licence

Apache License 2.0. See [`LICENSE`](LICENSE).
