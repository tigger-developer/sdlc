# SDLC v3 Architecture

This document explains how the public framework is composed. Normative
engineering requirements live under [`src/`](../src/); workflow entry points
live under [`skills/`](../skills/).

## Design objective

SDLC v3 minimizes orchestration while preserving explicit requirements,
automated TDD, independent challenge, human authority, and durable evidence.
It uses one definition artefact and two gates rather than separate specification,
plan, task, and implementation lifecycles.

## Installed architecture

```text
staging repository
  src/       -> ~/.agents/sdlc/
  skills/    -> ~/.agents/skills/
  hooks/     -> ~/.agents/sdlc/hooks/
  bin/       -> ~/.agents/sdlc/bin/
                 ^
                 |
  ~/.local/bin/sdlc-* symlinks
```

`sdlc-install` remains a repository-internal build artefact invoked by
`make install`; it is not deployed to the global path. The installed CLI tools
operate on individual projects or documents. Their global links target the
deployed release under `~/.agents/sdlc/bin`, never the mutable source checkout.

`~/.agents/sdlc/MAIN.md` is the only standards bootstrap. It routes an agent to
the minimum relevant universal, domain, technology, and workflow documents.
Skills are global and refer to that canonical root, so a standards update does
not require every initialized project to be refreshed.

The first v3 release installs no SDLC workflow into Claude, Hermes, or Copilot.
The installer removes only known public SDLC-owned v1/v2 artefacts from those
provider homes. Private provider configuration remains outside this repository's
ownership.

## Project architecture

An initialized project contains:

```text
.sdlc/project.yaml
docs/work.org
specs/NNN-descriptor/spec.org
specs/NNN-descriptor/audits.org
specs/NNN-descriptor/validation.org
```

The project YAML contains non-secret project facts, named product, architecture,
and requirement authorities, migration branch metadata, and explicit overrides.
It does not copy inherited global defaults from `~/.agents/sdlc.yaml`. The work
profile does not pin an SDLC release: global skills and standards follow the
release recorded in `~/.agents/sdlc.yaml`. Its schema `version` remains distinct
from that installer-managed `release`. The work ledger carries work state,
provenance, and an initially folded legacy
acceptance-criteria subtree. Current v3 requirements, tests, and design remain
in each unified specification rather than being duplicated into the ledger.
Each specification is a context-independent delivery handoff. Audit and
validation records remain separate so evidence cannot
silently rewrite the authority it assesses.

Org provides addressable hierarchy, stable internal links, properties, folding,
and readable plain text. The supported subset is deliberately portable and does
not execute code. `sdlc-preview` uses Pandoc to provide browser rendering without
requiring Emacs.

### Storage and parser boundary

The framework uses YAML for machine configuration, Org for work and
specification authorities, and Markdown for explanatory documentation. Markdown
with YAML frontmatter is not an SDLC storage format unless an integrated tool
requires it.

Within `docs/work.org`, the TODO keyword is the sole current lifecycle state.
Tags are non-exclusive classifications, properties are stable metadata, and
links identify other authorities without copying them. A linked `spec.org`
defines requirements, traced tests, edge cases, and solution design, but does
not repeat the work item's lifecycle state.

Initialization writes Org from canonical templates and performs narrow,
structure-aware insertions for migrated work and legacy acceptance criteria. It
does not parse and re-render an existing operator-authored ledger. A Go Org
parser is a future validation and query boundary, not a whole-document writer,
until round-trip preservation of the supported subset is demonstrated.

## Gate architecture

The definition gate applies specification, design, and test-definition criteria.
The implementation gate applies implemented-test and code criteria.

Each gate uses at most two contexts:

1. the authoring context performs and remediates up to five local rounds; and
2. one external Codex context applies all component audits and is resumed after
   any remediation.

Focused audit skills define reusable review criteria. They never spawn their own
contexts. `define-change` and `deliver-change` own their normal composite gates;
`audit-definition` and `audit-implementation` expose those gates for separate
operator-requested review. The composite workflow owns context creation,
retention, retry limits, evidence recording, and the human handback.

The work ledger records the work item's sole lifecycle state. The specification
records its current definition-gate status and operator sign-off. Detailed
revision-specific findings remain in `audits.org`. Delivery reads these
separate authorities and does not repeat the definition gate when entering a
new context.

## Initialization and migration

`sdlc-init` is a once-only migration controller. It requires a clean Git
worktree, creates a dated archive branch before mutation, and performs all work
on a dedicated migration branch.

Before configuration, the initializer applies technology-detection rules from
`config/project-init.schema.yaml` to the bounded Git inventory. Rules identify
strong basenames and extensions, exclude archived or generated paths, and may
express implications such as Hugo requiring the web standard. The results are
only preselected recommendations in the existing schema-driven multi-select;
the operator confirms or corrects them. This applies equally to maintained
brownfield projects and greenfield projects with initial scaffolding. A blank
project has no evidence from which to infer a stack.

Global configuration values are inherited without a question and without being
copied into the project profile. `--override-global-config` deliberately
re-enables those questions.

- New projects receive the project profile and empty work ledger.
- SDLC v1 projects may first run the lossless GitHub-ticket migration. Historical
  requirements are normalized through `docs/ACs.org`, then folded into
  `docs/work.org`; unresolved work enters the same ledger.
- SDLC v2 projects preserve active Spec Kit material under
  `docs/archive/sdlc-v2/`, remove it from active workflow, and classify every
  archived specification in `docs/work.org` without adding individual feature
  specifications to the project authority lists. The dated archive branch
  retains the exact source state; the archived working copy omits the obsolete
  top-level Status field because `docs/work.org` is now its sole lifecycle
  authority.

After migration, the initializer derives stable authorities from the bounded
Git file inventory. It presents Markdown and Org files whose stems contain
README, VISION, or ARCHITECTURE as multi-select product or architecture
candidates. Exact path case is preserved, canonical locations begin selected,
and the operator can rescan after moving a file. The initializer preserves an
existing `docs/work.org`, adds only missing migration structure and records,
folds any canonical `docs/ACs.org` into it, and records only `docs/work.org` as
the requirement authority. A remaining separate AC ledger is an error.

The other bounded headless Codex operation applies only to an SDLC v2 migration
with archived specifications. The configured audit model receives the exact
specification paths and may read only their feature directories. Its temporary
structured YAML classifies each specification as delivered, approved but
undelivered, abandoned, or unresolved, with priority, creation date, and
supporting evidence. Audit PASS alone is not operator approval. The initializer
validates exact coverage, then renders each classification beneath `Work items`
in the canonical `docs/work.org` hierarchy. `--no-agent-scan` instead records
each archived specification as unresolved `REVIEW` work. Model classification
never selects authority documents or edits project artefacts.

The proposal lives in the project-owned `.sdlc/.init/` directory. Its presence
tells a later invocation that initialization did not complete. The directory
remains available when initialization fails and is removed before a completed
migration is staged. The initializer creates or extends `.sdlc/.gitignore` from
the canonical template so `.init/` cannot be committed. It never uses private
`.git` paths as application storage.

The initializer commits the coherent migration and asks whether to merge it into
the original branch. It never normalizes dormant unfinished work merely because
the framework changed.

## Configuration architecture

`src/config/project-init.schema.yaml` defines field names, YAML paths, flags,
types, choices, defaults, global-default eligibility, prompts, and historical
`.env` aliases. It also identifies post-migration authority fields and their
discovery categories. Questions, authority-list persistence, and configuration
ordering are derived from this schema.

Resolution order is command line, process environment, project YAML, global
YAML, then schema default. During first initialization no project YAML exists;
the generated file records only project facts and deliberate overrides.

The `.env` compatibility wrapper evaluates only the existing shell file and
returns only schema-allowlisted keys. This is a one-time migration boundary, not
a v3 runtime configuration mechanism. After a successful YAML write, it removes
only the migrated SDLC keys and preserves unrelated `.env` content.

## Recovery and release boundary

Annotated semantic-version Git tags are the release authority. The build embeds
the latest release tag reachable from its current commit. `make install` records
that value as the top-level `release` in `~/.agents/sdlc.yaml`, preserving the
operator's other global defaults. Project profiles do not duplicate it.

The entire unfinished v2 implementation is preserved on
`sdlc_v2_state_2026-09-06`. V3 was developed on `sdlc-v3`; that branch is merged
into `master` before the annotated release tag. The release candidate requires
the local regression suite, installer fixtures, Org rendering, and a watched
project migration to pass.
