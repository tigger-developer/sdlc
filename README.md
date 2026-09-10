# Lean SDLC for Coding Agents

This public repository provides a standalone, provider-neutral engineering
standards library and a lean delivery workflow for coding agents. SDLC v3 uses
one definition artefact and two human gates. It does not require GitHub Spec Kit.

SDLC v3 supports **Codex**, **Claude Code**, **GitHub Copilot CLI**, and
**Hermes**. One canonical skill source is exposed through each harness's native
global skill location; fixed adapters preserve their different model, provider,
output, and resumable-session contracts.

## Prerequisites

- Go 1.23 or later to build the installer and helper commands.
- Git for recoverability, migration branches, and delivery checkpoints.
- At least one supported coding-agent harness: Codex, Claude Code, GitHub
  Copilot CLI, or Hermes.
- GitHub CLI only when migrating an SDLC v1 project's GitHub tickets.

## Install

`make install` also initializes the pinned
[tigger-developer/HTML-Preview](https://github.com/tigger-developer/HTML-Preview)
submodule and runs its own `make install` if `htmlpreview` is absent from PATH.
Existing installations are retained. This previewer was
developed primarily for SDLC document review and renders **Org and Markdown**
as **high-fidelity HTML**, including navigable document structure.

```sh
make install
```

The repository-internal `sdlc-install` command is invoked by `make install`; it
is not installed on the global path. The installation:

- builds `sdlc-install` locally, deploys `sdlc-init` and
  `sdlc-merge-legacy-acs` under `~/.agents/sdlc/bin`, and links those operator
  commands onto the global path;
- deploys the internal `sdlc-harness` runner under `~/.agents/sdlc/bin` without
  adding it to the global path;
- synchronizes the canonical standards to `~/.agents/sdlc`;
- installs SDLC skills globally under `~/.agents/skills`;
- links canonical skills into the native Claude, Copilot, and Hermes skill
  locations; and
- registers the same canonical command guard through all four native hook
  mechanisms without replacing unrelated provider configuration.

It lists only missing or differing artefacts. An unchanged rerun writes nothing
and asks no question. Set `VERBOSE=1` to include matching artefacts. Interactive
confirmation accepts `y` or `yes`.

## Repository layout

| Path | Purpose |
|---|---|
| `src/MAIN.md` | Universal rules and progressive routing |
| `src/*.md` | Requirements, testing, auditing, coding, Git, documentation, security, paired, emergency, and Org standards |
| `src/technologies/` | Automatically discoverable technology standards |
| `src/templates/v3/` | Unified specification, work, audit, and validation templates |
| `src/prompts/` | Saved prompts used by bounded headless initializer analysis |
| `skills/` | Globally installed workflow and focused audit skills |
| `cmd/` and `internal/` | Installer, initializer, harness adapters, and ledger merger |

## Initialize a project once

Run from a clean, named Git branch:

```sh
sdlc-init
```

The initializer:

1. refuses when `.sdlc/project.yaml` already exists;
2. detects a new project, an SDLC v1 project, or an SDLC v2 Spec Kit project;
3. applies schema-defined heuristics to the bounded Git inventory and presents
   the detected technologies as preselected choices for operator confirmation;
4. asks the remaining schema-driven questions needed before migration,
   including project role, infrastructure ownership, implementation branch
   strategy, agent settings, and audit timeout;
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
    classification with the configured audit harness and model, removes their obsolete
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

When a global value exists, the initializer inherits it without asking and does
not copy it into the project file. Pass `--override-global-config` to display
those questions and record selected project overrides. Project identity,
technology selection, and infrastructure role are always project decisions.

When no explicit technology selection exists, the initializer applies the
deterministic file heuristics in `config/project-init.schema.yaml` to the bounded
Git inventory. Maintained manifests and source files provide recommendations;
archives, generated output, vendored dependencies, and provider runtime state
are excluded. Recommendations begin selected with concise matching evidence,
and the operator confirms or corrects them. A genuinely blank project has no
inferred stack and requires a manual choice. The heuristics optimize recall and
need not establish the stack without operator confirmation.

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
Only archived v2 specification disposition requires model judgement. Pass
`--no-agent-scan` to preserve those specifications as unresolved `REVIEW` work
instead. A temporary `.sdlc/.init/` working directory remains
if initialization is interrupted. Rerunning `sdlc-init` on the migration
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
    max_rounds: 5
infrastructure:
  owner: Example platform team
  contract: /absolute/path/to/PROJECT-INTEGRATION.md
```

`version` identifies the configuration schema. `release` identifies the latest
semantic-version Git tag deployed by `make install`; the installer maintains it
without replacing unrelated settings. Projects follow those global skills and
standards, so `.sdlc/project.yaml` does not duplicate an SDLC release pin.

Document inspection uses `htmlpreview` directly. No `HTML_PREVIEW_TOOL` variable
or manual viewer installation is required. See the
[HTML-Preview prerequisites](https://github.com/tigger-developer/HTML-Preview#prerequisites-and-installation)
for its supported Go and Pandoc versions. Preview is presentation, not a gate.

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
retained context on the configured external audit harness run all three audits.
On effective PASS,
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

Audit results live in `audits.yaml`. They are evidence for an exact revision, not
human approval.

Audit criteria are selected by the composite audit harness; individual audit
skills are not deployed. The harness-invoked prompt returns findings; only the
harness writes audit state. Duplicated audit state in a supplied artefact is a
gate failure.

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
specs/WNNN-descriptor/spec.org
specs/WNNN-descriptor/audits.yaml
specs/WNNN-descriptor/validation.org
```

Org provides foldable hierarchy and stable internal links without making Emacs
a dependency. Read `~/.agents/sdlc/ORGMODE.md` before editing Org artefacts.

## Migration evidence

SDLC v1 ticket migration remains an explicit operator-invoked workflow. It
creates a lossless local ticket archive, an intermediate canonical
`docs/ACs.org`, and `docs/ticket-migration.org`, updates stale project documents,
archives the old
implementation plan, and closes tickets only after durable evidence is
committed.

`sdlc-init` then validates that intermediate ledger against the
canonical Org structure and folds it into `docs/work.org` before importing
unresolved legacy work or archived Spec Kit specifications. If deterministic
validation cannot safely interpret the ledger, the initializer invokes one
headless Codex repair using the configured audit model and validates the result
before continuing. An unsuccessful repair stops initialization without merging
or deleting the source. The AC disposition becomes its headline state and the
redundant Status field is removed. After the embedded copy is verified, the
separate `docs/ACs.org` is removed.

For an already initialized v3 project, run `sdlc-merge-legacy-acs` once from the
project root. A successful rerun is a no-op.

SDLC v2 migration preserves `.specify` and existing feature directories under
`docs/archive/sdlc-v2/`, removes them from the active workflow, and indexes the
preserved work for later operator disposition. It does not force incomplete work
through a migration-time definition exercise.
