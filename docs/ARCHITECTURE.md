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

The installer exposes the canonical global skills through the native skill
locations used by Claude, Copilot, and Hermes; Codex consumes the canonical
skill root directly. Each supported harness receives its native registration
for the shared command guard and uses a thin capability-aware adapter for
bounded external audits. Private provider configuration remains outside this
repository's ownership.

## Project architecture

An initialized project contains:

```text
.sdlc/project.yaml
docs/work.org
specs/NNN-descriptor/spec.org
specs/NNN-descriptor/audits.yaml
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
not execute code. Rendering uses the separate
[tigger-developer/HTML-Preview](https://github.com/tigger-developer/HTML-Preview)
project, pinned as a Git submodule. The SDLC installation initializes that
submodule and invokes its own installer only when `htmlpreview` is missing.
HTML-Preview owns rendering, assets, cleanup, and its installation; the SDLC
does not implement a second preview command.

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

Each gate retains one external session per work item. Composite prompts live in
`src/prompts/audits.yaml`; separate audit skills are no longer deployed. The
harness owns session identity, timeout, round limits, and persistence. Authoring
agents remediate all findings before resuming the same gate session.

Lifetime audit numbering is separate from the current bounded allowance.
Operator-authorized `resume --reset-session` appends a timestamped budget
boundary and exits without invoking a provider. It preserves native context,
findings and cached results; subsequent launches consume the fresh allowance.
Spec changes and provider recovery do not reset it automatically.

Codex and Claude receive original absolute paths and a metadata-only evidence
manifest. Hermes runs without filesystem tools and receives the verified UTF-8
contents through stdin on every invocation, without retained copies. The harness
hashes regular files before and after execution, rejects
mutation, and checkpoints each invocation's manifest before execution in
`audits.yaml` under `history[].evidence`. Resume compares against that gate's
last recorded manifest. There is no document-copy cache; hashes do not imply
that an auditor remembers unchanged content. Provider read-only controls remain
necessary, and additional reads outside the input list are not hash-verified.

Validated verdicts are cached per work item/gate using a versioned SHA-256 key
over the full supplied manifest, prompt registry, caller context and effective
primary/fallback model settings. The newest matching incident-free response is
returned before session recovery and round admission, without any record mutation.
Unkeyed legacy records and incomplete attempts cannot produce cache hits.

For Claude, the captured manifest also supplies exact invocation-local Read
permissions, including resolved path aliases. JSON settings preserve literal
filename characters without granting directory access. The adapter retains
Read-only built-in tools and plan mode; existing permission denials still apply.
This is evidence-access plumbing, not a complete filesystem sandbox.

Native session identity is checkpointed as soon as the adapter observes it.
Timeouts retain the attempt manifest and consume the same recorded round budget
without producing a verdict. YAML supplies the caller recovery diagnostic and
the resumed auditor's continuation instruction. Missing identity or exhausted
bounds stop timeout recovery. Explicit missing-session diagnostics and configuration
changes permit a harness-owned replacement, with previous identities/configuration
retained and all historical findings supplied to the new context. An optional
phase-local fallback tuple handles explicit authentication failures once per call.
Missing-session restart is likewise limited to once per call. Every launch consumes
the existing round budget and retains the configured per-attempt deadline.

The work ledger owns lifecycle state, the specification owns requirements and
operator sign-off, and `audits.yaml` alone owns audit state. Legacy `audits.org`
records are preserved in YAML through the existing migration. Delivery does not
repeat the definition gate when entering a new context.

Provider stdout is decoded incrementally for bounded progress metadata and error
diagnostics on stderr. Claude uses native streaming JSON; its final result alone
becomes the returned response. Failed exits and timeouts flush the final partial
record instead of discarding it. Capture is limited to 16 MiB, with an explicit
overflow incident; provider stderr remains live. Event metadata is not a verdict
or proof that the provider has finished. Tool payloads and prompts are not
replayed as progress, and existing process deadlines still govern termination.

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

Writing migrations use direct native CLI invocations, separate from the
read-only audit runner. Ticket migration, AC conversion and repair select the
`delivery.definition` harness/provider/model.
Codex runs with workspace-write, Hermes with normal tools, and Claude with
acceptEdits; native command checks and hooks remain active. Provider applies
only to Hermes. Other harnesses retain the manual handoff. Native failures stop
the initializer, which offers ticket migration again on resume when its index
is absent. Canonical AC ledgers require no agent; a repair must pass the same
deterministic validator before consolidation. No audit session, round budget or
audit timeout is attached to a writing migration.

Legacy AC output uses `VALID` for applicability; it does not synthesize a passing
test result. `HOLD`/`HOLDING` are accepted only as historical legacy-parser
aliases. Consolidation upgrades the known legacy type declaration while leaving
normal work-item lifecycle, qualifiers and recorded test results unchanged.

After migration, the initializer derives stable authorities from the bounded
Git file inventory. It presents Markdown and Org files whose stems contain
README, VISION, or ARCHITECTURE as multi-select product or architecture
candidates. Exact path case is preserved, canonical locations begin selected,
and the operator can rescan after moving a file. The initializer preserves an
existing `docs/work.org`, adds only missing migration structure and records,
folds any canonical `docs/ACs.org` into it, and records only `docs/work.org` as
the requirement authority. A remaining separate AC ledger is an error.

The other bounded headless harness operation applies only to an SDLC v2 migration
with archived specifications. The configured audit harness and model receive the exact
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

Recovery validates archive ancestry for the migration and primary branches;
their tips need not still equal the archived baseline. On `master` or `main`,
the sole migration branch requires a warning and explicit confirmation before
an ordinary Git switch. Ambiguous branches, unrelated history and switch
conflicts stop recovery without deleting state, stashing edits or rewriting refs.

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
