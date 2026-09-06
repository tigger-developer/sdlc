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
  CLI tools  -> ~/.local/bin/
```

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
ledger carries work state and provenance without duplicating requirements,
tests, or design. Each unified specification is a context-independent delivery
handoff. Audit and validation records remain separate so evidence cannot
silently rewrite the authority it assesses.

Org provides addressable hierarchy, stable internal links, properties, folding,
and readable plain text. The supported subset is deliberately portable and does
not execute code. `sdlc-preview` uses Pandoc to provide browser rendering without
requiring Emacs.

## Gate architecture

The definition gate applies specification, design, and test-definition criteria.
The implementation gate applies implemented-test and code criteria.

Each gate uses at most two contexts:

1. the authoring context performs and remediates up to five local rounds; and
2. one external Codex context applies all component audits and is resumed after
   any remediation.

Focused audit skills define reusable review criteria. They never spawn their own
contexts. Composite gate skills own context creation, retention, retry limits,
evidence recording, and the human handback.

## Initialization and migration

`sdlc-project-init` is a once-only migration controller. It requires a clean Git
worktree, creates a dated archive branch before mutation, and performs all work
on a dedicated migration branch.

- New projects receive the project profile and empty work ledger.
- SDLC v1 projects may first run the lossless GitHub-ticket migration. Historical
  requirements remain in `docs/ACs.org`; unresolved work moves to `work.org`.
- SDLC v2 projects preserve active Spec Kit material unchanged under
  `docs/archive/sdlc-v2/`, remove it from active workflow, and index it for later
  operator disposition.

Only after that migration does the initializer invoke headless Codex with the
deployed read-only authority-discovery prompt. The model returns a temporary
structured YAML proposal containing lists of product, architecture, and
requirement documents, with descriptors and rationales. The initializer
validates every proposed or amended path against repository files, requires
operator confirmation, and writes only the confirmed path lists to the project
profile.

The proposal lives in a temporary working directory in Git metadata. Its
presence tells a later invocation that initialization did not complete. The
directory remains available when initialization fails and is removed after a
successful initialization and optional merge decision.

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

The entire unfinished v2 implementation is preserved on
`sdlc_v2_state_2026-09-06`. V3 is developed on `sdlc-v3`. It is not tagged until
the focused unit suite, installer fixtures, Org rendering, and the first paired
project migration have been reviewed.
