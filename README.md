# Lean SDLC for Coding Agents

This public repository provides a standalone, provider-neutral engineering
standards library and a lean delivery workflow for coding agents. SDLC v3 uses
one definition artefact and two human gates. It does not require GitHub Spec Kit.

The first v3 release supports **Codex with OpenAI models** for delivery and
external audit contexts. The standards themselves are written so additional
harnesses can be supported later.

## Prerequisites

- Go 1.23 or later to build the installer and helper commands.
- Git for recoverability, migration branches, and delivery checkpoints.
- Codex for the initial v3 agent workflow.
- Pandoc for browser previews of Markdown and Org documents.
- GitHub CLI only when migrating an SDLC v1 project's GitHub tickets.

## Install

```sh
make install
```

The installer:

- builds and installs `sdlc-install`, `sdlc-project-init`, `sdlc-preview`, and
  `sdlc-merge-legacy-acs`;
- synchronizes the canonical standards to `~/.agents/sdlc`;
- installs SDLC skills globally under `~/.agents/skills`;
- retains Codex as the only supported v3 provider adapter; and
- retires known public SDLC v1 and v2 commands, skills, prompts, and hooks from
  Claude, Hermes, and Copilot without removing unrelated provider configuration.

It lists only missing or differing artefacts. An unchanged rerun writes nothing
and asks no question. Set `VERBOSE=1` to include matching artefacts. Interactive
confirmation accepts `y` or `yes`.

## Repository layout

| Path | Purpose |
|---|---|
| `src/MAIN.md` | Universal rules and progressive routing |
| `src/*.md` | Requirements, testing, auditing, coding, Git, documentation, security, paired, emergency, and Org standards |
| `src/technologies/` | Automatically discoverable technology standards |
| `src/templates/v3/` | Unified specification, work, audit, validation, and preview templates |
| `src/prompts/` | Saved prompts used by bounded headless initializer analysis |
| `skills/` | Globally installed workflow and focused audit skills |
| `cmd/` and `internal/` | Installer, initializer, ledger merger, and preview implementation |

## Initialize a project once

Run from a clean, named Git branch:

```sh
sdlc-project-init
```

The initializer:

1. refuses when `.sdlc/project.yaml` already exists;
2. detects a new project, an SDLC v1 project, or an SDLC v2 Spec Kit project;
3. uses the configured audit model for one bounded, read-only assessment of
   technologies materially used by the project, then presents those findings
   as preselected choices in the schema-driven technology question;
4. asks the remaining schema-driven questions needed before migration,
   including project role, infrastructure ownership, branch strategy, agent
   settings, and audit timeout;
5. creates a dated branch preserving the exact pre-migration state and offers
   to push it;
6. creates a dedicated v3 migration branch;
7. offers the optional legacy-ticket migration for eligible SDLC v1 projects
   and continues with the existing ledger when declined;
8. archives and removes active Spec Kit artefacts for v2 projects without
   normalizing unfinished work;
9. inventories README, vision, and architecture Markdown or Org documents and
   lets the operator select, deselect, move, and rescan them;
10. folds any canonical `docs/ACs.org` into a preserved or newly created
   `docs/work.org`, which becomes the sole requirement authority;
11. for v2 projects with archived specifications, runs one bounded read-only
    classification with the configured audit model, removes their obsolete
    top-level Status fields, and renders their lifecycle into `docs/work.org`;
12. creates `.sdlc/project.yaml` and commits the migration with a concise
    commit summary; and
13. asks whether to merge the migration branch into the original branch.

Set `VERBOSE=1` to show Git's complete changed-file inventory during the
migration commit.

The first real project migration should be performed with the operator watching.
See [QUICKSTART.md](QUICKSTART.md) for each migration path.

## Configuration

Global non-secret defaults and the deployed SDLC release live at
`~/.agents/sdlc.yaml`. Project facts and explicit overrides live in tracked
`.sdlc/project.yaml`. Resolution order is:

1. command line;
2. process environment;
3. project YAML;
4. global YAML; and
5. schema default.

When a global value exists, the initializer shows it and lets the operator press
Enter to inherit it or enter a project override. Inherited global values are not
copied into the project file. Project identity, technology selection, and
infrastructure role are always project decisions.

When no explicit technology selection exists, the initializer asks the
configured audit model for a read-only assessment before displaying the
schema-owned technology choices. Maintained product, test, build, packaging,
and deployment artefacts count as evidence. Archived, generated, vendored,
framework-owned, and incidental tooling does not. The recommendations begin
selected and include concise evidence; the operator may accept or replace them.
A genuinely blank project has no inferred stack and requires a manual choice.

Every interactive question begins on a separate line. Schema choices and
document authorities share the same `[x]` and `[ ]` presentation; configuration
questions select one value, while authority questions permit toggling and
rescanning. After validation, each schema response prints the resolved value;
multi-choice questions redraw their complete state after each numbered or named
toggle and require Enter to confirm it.

Product and architecture authorities are selected after migration from the
project's bounded Git file inventory. The chooser finds Markdown and Org files
whose stems contain README, VISION, or ARCHITECTURE, preserves exact path case,
and supports multiple selections and rescanning after a move. The initializer
preserves an existing `docs/work.org`, adds missing migration structure, and
folds a canonical legacy AC ledger beneath `Legacy Acceptance Criteria (SDLC
v1)`. It then records only `docs/work.org` as the requirement authority.
Technology assessment and archived v2 specification disposition require a
model. Their temporary YAML contains recommendations or status evidence rather
than authority decisions. A temporary `.sdlc/.init/` working directory remains
if initialization is interrupted. Rerunning `sdlc-project-init` on the migration
branch resumes when the dated archive and original primary branch are
unambiguous. A tracked `.sdlc/.gitignore` excludes that temporary directory from
commits. The initializer never uses private `.git` paths as application storage.

Example global configuration:

```yaml
version: 3
release: v3.0.0
delivery:
  branch_strategy: current
  definition:
    harness: codex
    provider: openai
    model: gpt-5.6-sol
  build:
    harness: codex
    provider: openai
    model: gpt-5.6-terra
  audit:
    harness: codex
    provider: openai
    model: gpt-5.6-luna
    timeout: 5m
infrastructure:
  owner: Example platform team
  contract: /absolute/path/to/PROJECT-INTEGRATION.md
```

`version` identifies the configuration schema. `release` identifies the latest
semantic-version Git tag deployed by `make install`; the installer maintains it
without replacing unrelated settings. Projects follow those global skills and
standards, so `.sdlc/project.yaml` does not duplicate an SDLC release pin.

The initializer can import only schema-allowlisted historical `SDLC_*` values
from a project `.env` through its shell wrapper. After the YAML profile is
written, it removes those exact historical keys while preserving all unrelated
lines. Agents never read `.env`, and v3 runtime configuration never depends on
it.

## Lean delivery workflow

### Define

Invoke `$define-change`. It asks only the clarification needed and creates one
`spec.org` containing:

- context and scope;
- falsifiable acceptance criteria;
- RT, UT, and OT test definitions linked to those criteria;
- edge cases;
- solution design and architecture impact, with mandatory security analysis;
  and
- a context-independent delivery handoff.

The same skill applies the specification, design, and test-definition audits,
remediating for at most five local rounds. Only after local PASS does one
retained external Codex context run all three audits. On effective PASS,
`spec.org` records the gate PASS and `docs/work.org` moves the item to `REVIEW`;
explicit operator sign-off records its authority in the specification and moves
the work item to `ACTIVE`. `$audit-definition` remains available for a
separately requested combined review.

The operator reviews and signs off the definition before implementation.

### Deliver

Invoke `$deliver-change`. It admits only an `ACTIVE` work item whose
specification records a definition-gate PASS and operator sign-off; it does not
rerun the definition audits. Automated regression tests are written first and
must show the intended RED result. The agent implements the smallest coherent
change, returns the suite to GREEN, and applies the implemented-test and
production-code audits locally before one retained external audit context.
After PASS, required OT and UT evidence is recorded in `validation.org`,
affected documentation is reconciled, and the operator decides closure.
`$audit-implementation` remains available for a separately requested combined
review.

Audit results live in `audits.org`. They are evidence for an exact revision, not
human approval.

The independently invocable focused audits are `$audit-spec`, `$audit-design`,
`$audit-test-definitions`, `$audit-test-code`, and `$audit-code`.

## Variant workflows

- `$pair-change` supports explicitly selected live human-agent implementation.
  The bounded objective and each explicit iteration instruction form the
  working specification. User validations are first-class evidence and are
  consolidated into durable artefacts at closure.
- The exact operator token `BYPASS-GATE-7` invokes `$emergency-change`. It uses
  a bounded temporary specification, preserves TDD where an automated test is
  justified, runs the implementation gate, then backfills the durable
  specification, design, validation, and documentation.

## Project artefacts

```text
.sdlc/project.yaml
docs/work.org
specs/NNN-descriptor/spec.org
specs/NNN-descriptor/audits.org
specs/NNN-descriptor/validation.org
```

Org provides foldable hierarchy and stable internal links without making Emacs
a dependency. `sdlc-preview FILE` renders Markdown or Org with Pandoc beside the
source, opens it in the browser, then removes the temporary HTML after one
second. Read `~/.agents/sdlc/ORGMODE.md` before editing Org artefacts.

## Migration evidence

SDLC v1 ticket migration remains an explicit operator-invoked workflow. It
creates a lossless local ticket archive, an intermediate canonical
`docs/ACs.org`, and `docs/ticket-migration.org`, updates stale project documents,
archives the old
implementation plan, and closes tickets only after durable evidence is
committed.

`sdlc-project-init` then folds that intermediate ledger into `docs/work.org`
before importing unresolved legacy work or archived Spec Kit specifications.
The AC disposition becomes its headline state and the redundant Status field is
removed. After the embedded copy is verified, the separate `docs/ACs.org` is
removed.
For an already initialized v3 project, run `sdlc-merge-legacy-acs` once from the
project root. A successful rerun is a no-op.

SDLC v2 migration preserves `.specify` and existing feature directories under
`docs/archive/sdlc-v2/`, removes them from the active workflow, and indexes the
preserved work for later operator disposition. It does not force incomplete work
through a migration-time definition exercise.
