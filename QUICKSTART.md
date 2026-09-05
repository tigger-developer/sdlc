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

1. resolves configured values and uses numbered prompts for unresolved choices;
2. asks whether the project is greenfield or brownfield;
3. discovers the available technology standards and asks which apply;
4. asks whether the project has no infrastructure contract, consumes one owned
   elsewhere, or provides and implements one;
5. asks whether it may run `specify init` and uses the specification harness as
   its agent integration;
6. installs the `sdlc-standards` preset;
7. records the project selections in an ignored `.env`;
8. renders `.specify/templates/overrides/constitution-template.md`;
9. commits only that generated scaffold;
10. launches the specification harness to create an unratified project constitution;
    and
11. verifies that the complete generated Engineering Standards, Specification
    and Evidence, and Mandatory Independent Audits covenants remain present.

Select `HUGO` together with `WEB` for a Hugo site. Select `NODE` whenever
Node.js is the application runtime or npm-managed dependencies participate in
the build; incidental use of an npm-installed developer tool still brings its
dependency graph into vulnerability checking when it can affect the artefact.
Select `JAVASCRIPT` whenever the project owns JavaScript or TypeScript source,
including a host-application plugin. Add `WEB` for browser-facing HTML, CSS, DOM,
or accessibility behaviour. An Obsidian plugin normally selects `JAVASCRIPT`
and `NODE`, plus `WEB` when it owns interface or stylesheet behaviour.

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
constitution on its own merits and correct or remove unsuitable project-specific
clauses. Only the human may authorize removal or weakening of the shared
Engineering Standards, Specification and Evidence, or Mandatory Independent
Audits covenant. Keep the current compact Sync Impact Report as the first line
of the constitution. Replace it on amendment rather than accumulating report
history.

Keep `.env` and project-local agent runtime directories such as `.agents/`,
`.claude/`, and `.codex/` untracked. Resolve every ratification TODO, explicitly
ratify the constitution as version 1.0.0, and checkpoint the resulting project
state before specifying the first feature.

For initial ratification, replace the Sync Impact transition with
`UNRATIFIED -> 1.0.0`, remove the `SDLC-GENERATED-SCAFFOLD` comment, confirm the
adopted SDLC revision, set the ratification and last-revised dates, remove
resolved blockers and draft-only qualifications, and do not describe edits to
the unratified draft as amendments.

## Initialize a brownfield project

Start from a recoverable Git checkpoint. Existing unrelated or human-authored
changes are not an initialization workspace; checkpoint or separate them
according to the project's normal source-control practice before proceeding.

First invoke `$migrate-legacy-acs-to-sdlc-v1`. It archives every issue and
comment under `docs/archive/migrated-tickets/`, then works only from that local
snapshot. Before appending criteria, it losslessly converts the normally
pre-existing `docs/ACs.md` to the canonical foldable `docs/ACs.org` structure and
removes the Markdown source. It may run the supported whole-suite command once
and performs one reverse checksum from every maintained RT to `docs/ACs.org`. Before classifying
any ticket, it reviews every maintained RT and builds an RT-to-ticket-to-AC
delivery-evidence map. It does not rerun historical tests or investigate old
code ticket by ticket. Ticket status, comments, timeline commit references, and
code-commit evidence are retained in the archive. If the suite fails, the skill
checkpoints only the verified archive, migration index, and format-only Org
conversion, then stops before classification, semantic reconciliation, or remote
mutation.

The Org conversion preserves every requirement, identifier, provenance field,
status, test relationship, evidence note, supersession, footnote, glossary
entry, and explanatory note. It standardizes those fields through nested
headings and keeps criteria in identifier sequence. An AC already present in the
source ledger is delivery evidence even when its archived ticket omitted test
results. Archived tickets and comments remain provenance only, never current AC
authority.

The skill creates `docs/ticket-migration.org` immediately after verifying the
archive, then updates it after every evidence phase and ticket classification.
It records progress, suite evidence, the RT map, open defects, abandoned or
undelivered feature scope, unresolved items, and a final delivered-ticket list.
A
near-complete ticket with only one or two unrecorded UT or OT results is
automatically treated as delivered. The same applies when a ticket has no
recorded passing evidence but any of its RTs remain in the passing harness.
ACs relying solely on either inference and their delivered-list entries are
footnoted. When no delivery evidence remains, closed scope is abandoned and
open scope is undelivered. Orphan RTs automatically receive AC provenance through one
baseline ticket. Once the archive, AC ledger, documentation
corrections, and unchanged implementation-plan archival are committed in a
batch, the skill closes every remaining legacy issue without comment and
records the result.

The archive is external memory: the agent reads one ticket at a time rather than
loading the whole corpus. If its harness compacts automatically, it rereads the
skill, manifest, AC ledger, and incremental Org record, then resumes from the
first unfinished ticket without rerunning completed work.

Run the same initializer from the existing project root:

```bash
sdlc-project-init
```

After authorization, Spec Kit merges its project infrastructure into the
existing working tree with its `--here --force` initialization path. The
initializer then verifies the migrated ledger boundary before constitution
generation:

1. A completed migration must provide `docs/ACs.org`.
2. `docs/ACs.md` must no longer coexist with the Org ledger.
3. Migration artefacts without `docs/ACs.org` stop with an instruction to run
   `$convert-migrated-acs-to-org`.

Vision, architecture, and README documents remain active. The initializer does
not rewrite them or archive the implementation plan; those are pre-migration
skill responsibilities.

For projects migrated before the Org format was introduced, invoke
`$convert-migrated-acs-to-org`. It refuses unless the archived ticket corpus and
completed migration record exist and GitHub reports zero open issues. It
performs no ticket, test, code, or requirement changes.

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
The pre-migration skill leaves undelivered scope untouched. Bring forward only
the baseline relevant to each new or migrated feature.

Review and track the same `.specify/` constitution and scaffold files listed in
the greenfield procedure. Confirm that the generated authority hierarchy gives
approved specifications authority over observable behaviour, approved design
authority over technical choices within those requirements, and code and tests
the status of implementation evidence rather than requirement approval.

## Add the deployment vulnerability gate

Every deployable application must expose:

```bash
make vulncheck
```

The project implements that target with the scanner selected by its technology
standards. It scans committed locks, resolved dependencies, or the built
artefact; makes no changes; and fails when scanning is unavailable, incomplete,
or reports a non-exempt vulnerability. Keep it separate from `make test`.

An external infrastructure owner may invoke the same target before deployment.
That application check complements rather than replaces host, operating-system,
service, container, and package-closure controls. See
`~/.agents/sdlc/SECURITY.md` for scanner selection and exception requirements.

Projects using Make use the canonical `test`, `vulncheck`, `install`, `sync`,
and `deploy` target names where those operations apply. Deployment runs tests
and vulnerability checking before the project-specific action. To omit only an
already-established regression run:

```bash
SKIP_TESTS=1 make deploy
```

`make sync` stages, commits, pulls, and pushes in that order. Its optional
`COMMIT_MESSAGE` is a short subject and defaults to `chore: sync`; the target
does not try to manufacture a detailed commit message.

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
    --branch-strategy current \
    --project-type brownfield \
    --technologies GO,WEB \
    --infra-role none
```

Use `--no-launch` to render and inspect the scaffold without starting an agent.
Any required mechanical brownfield migration still displays its diff and asks
for commit confirmation:

```bash
sdlc-project-init --no-launch
```

The initializer evaluates user defaults from `~/.agents/.env` and project
overrides from the ignored project `.env` through its deployed Bash wrapper.
Only schema-listed SDLC values are returned to the Go initializer. User
defaults remain global and apply wherever a project has no explicit override.
Precedence is command line, process environment, project `.env`, user `.env`,
then a declared fallback. Project classification is project-only: supply
`--project-type`, record `SDLC_PROJECT_TYPE` in the project `.env`, or answer the
initializer prompt. A value in `~/.agents/.env` is ignored. Command-line values
take precedence. Run `sdlc-project-init --help` for the complete interface.

`SDLC_BRANCH_STRATEGY` accepts `current` or `feature` and defaults to `current`.
Set it in `~/.agents/.env` as the global default or in a project `.env` to
override that default. Staged specification, design, test, and implementation
phases pull at entry, then pull and push their committed artefacts after
effective audit PASS. Feature strategy publishes the feature branch through the
existing configured remote; current strategy does not change branches.

The field definitions live in
`~/.agents/sdlc/config/project-init.schema.yaml`. Fixed choices are displayed as
numbered options and are not prompted when already resolved. `SDLC_INFRA_ROLE`
accepts `none`, `consumer`, or `provider`. Consumer projects
comply with an externally owned integration contract. Provider projects define,
implement, evolve, and honour the infrastructure side of the contract they
publish. Use `--infra-role` to override the configured default for an
infrastructure repository.

Set the phase-specific harnesses `SDLC_SPEC_HARNESS`, `SDLC_BUILD_HARNESS`, and
`SDLC_AUDIT_HARNESS` when different tools should perform those phases. Each
falls back to `SDLC_AGENT_HARNESS`. Set `SDLC_AUDIT_PROVIDER` and
`SDLC_AUDIT_MODEL` to select the independent auditor for new projects.
Set `SDLC_AUDIT_TIMEOUT` to a whole-second duration of at least one second,
such as `4m`; it defaults to `5m`, and `sdlc-audit --timeout` overrides it for
one invocation.
Set a project override only when that project must differ from the current
global value. Provider fields apply only to harnesses that accept explicit
provider selection. The audit
runner passes provider and model to Hermes, but passes model only to Codex or
Claude.

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
