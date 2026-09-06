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

- builds and installs `sdlc-install`, `sdlc-project-init`, and `sdlc-preview`;
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
| `cmd/` and `internal/` | Installer, initializer, and preview implementation |

## Initialize a project once

Run from a clean, named Git branch:

```sh
sdlc-project-init
```

The initializer:

1. refuses when `.sdlc/project.yaml` already exists;
2. detects a new project, an SDLC v1 project, or an SDLC v2 Spec Kit project;
3. asks the schema-driven questions needed before migration, including project
   role, technologies, infrastructure ownership, branch strategy, agent
   settings, and audit timeout;
4. creates a dated branch preserving the exact pre-migration state and offers
   to push it;
5. creates a dedicated v3 migration branch;
6. offers the optional legacy-ticket migration for eligible SDLC v1 projects
   and continues with the existing ledger when declined;
7. archives and removes active Spec Kit artefacts for v2 projects without
   normalizing unfinished work;
8. inventories README, vision, and architecture Markdown or Org documents and
   lets the operator select, deselect, move, and rescan them;
9. records `docs/work.org` and the applicable AC ledger as deterministic
   requirement authorities;
10. for v2 projects with archived specifications, runs one bounded read-only
    classification with the configured audit model and renders the results into
    `docs/work.org`;
11. creates `.sdlc/project.yaml`, commits the migration; and
12. asks whether to merge the migration branch into the original branch.

The first real project migration should be performed with the operator watching.
See [QUICKSTART.md](QUICKSTART.md) for each migration path.

## Configuration

Global non-secret defaults live at `~/.agents/sdlc.yaml`. Project facts and
explicit overrides live in tracked `.sdlc/project.yaml`. Resolution order is:

1. command line;
2. process environment;
3. project YAML;
4. global YAML; and
5. schema default.

When a global value exists, the initializer shows it and lets the operator press
Enter to inherit it or enter a project override. Inherited global values are not
copied into the project file. Project identity, technology selection, and
infrastructure role are always project decisions.

Every interactive question begins on a separate line. Schema choices and
document authorities share the same `[x]` and `[ ]` presentation; configuration
questions select one value, while authority questions permit toggling and
rescanning.

Product and architecture authorities are selected after migration from the
project's bounded Git file inventory. The chooser finds Markdown and Org files
whose stems contain README, VISION, or ARCHITECTURE, preserves exact path case,
and supports multiple selections and rescanning after a move. Requirement
authorities are automatic: `docs/work.org` always applies, followed by
`docs/ACs.org` or, when it is the only legacy ledger, `docs/ACs.md`. Two AC
ledgers are an error. Only archived v2 specification disposition requires a
model; its temporary YAML result contains status evidence rather than authority
recommendations. A temporary `.sdlc/.init/` working directory remains if
initialization is interrupted and is removed before a completed migration is
staged. A tracked `.sdlc/.gitignore` excludes that temporary directory from
commits. The initializer never uses private `.git` paths as application storage.

Example global configuration:

```yaml
version: 3
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

Invoke `$audit-definition`. The authoring context applies the specification,
design, and test audits together for at most five local rounds. Only after local
PASS does one external Codex context run all three audits. That same context is
resumed after remediation. The gate never spans more than two contexts.

The operator reviews and signs off the definition before implementation.

### Deliver

Invoke `$deliver-change`. Automated regression tests are written first and must
show the intended RED result. The agent implements the smallest coherent change
and returns the suite to GREEN.

Invoke `$audit-implementation`. The authoring context applies the implemented
test and code audits together for at most five local rounds. One retained
external Codex context then runs both. After PASS, required OT and UT evidence is
recorded in `validation.org`, affected documentation is reconciled, and the
operator decides closure.

Audit results live in `audits.org`. They are evidence for an exact revision, not
human approval.

## Variant workflows

- `$pair-change` supports explicitly selected live human-agent implementation.
  The bounded objective and each explicit iteration instruction form the
  working specification. User validations are first-class evidence and are
  consolidated into durable artefacts at closure.
- `$emergency-change` is available only when the operator includes the exact
  token `BYPASS-GATE-7` in the same request. It uses a bounded temporary
  specification, preserves TDD where an automated test is justified, runs the
  implementation gate, then backfills the durable specification, design,
  validation, and documentation.

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
creates a lossless local ticket archive, canonical `docs/ACs.org`, and
`docs/ticket-migration.org`, updates stale project documents, archives the old
implementation plan, and closes tickets only after durable evidence is
committed.

SDLC v2 migration preserves `.specify` and existing feature directories under
`docs/archive/sdlc-v2/`, removes them from the active workflow, and indexes the
preserved work for later operator disposition. It does not force incomplete work
through a migration-time definition exercise.
