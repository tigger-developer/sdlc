# SDLC v3 Redesign Log

Status: Accepted design record implemented on the `sdlc-v3` branch. Validation
through isolated installer, new-project, v1, and v2 fixtures has passed. Live
installation, the first paired project migration, and one ordinary delivery
remain before release tagging.

This document preserves decisions and open questions while the live SDLC is being redesigned. Update it as decisions are made so work can resume after context compaction or in a new agent session.

## Problem being solved

- SDLC v2 is the live production SDLC, not a prototype.
- The standards library, progressive routing, audits, paired development, emergency delivery, and migration work remain valuable.
- Spec Kit orchestration introduced excessive artefacts, ceremony, context loading, audit loops, and blocking handbacks.
- Removing Spec Kit therefore requires a production migration that preserves active work and history.

## Agreed direction

- Disregard Spec Kit procedure while designing and implementing SDLC v3.
- The rules and standards maintained by this SDLC project remain authoritative.
- SDLC v3 begins with Codex as its only supported agent harness and OpenAI models as its only supported models.
- Keep the standards provider-neutral so additional harnesses can be supported later without rewriting engineering rules.
- Retain progressive loading through the canonical SDLC root and project-selected technology and domain standards.
- Install SDLC skills globally. Do not copy skill implementations into projects.
- Do not create `sdlc-project-update` for v3. Its only established purpose was
  refreshing project-local Spec Kit skills, and global skills remove that need.

## Standard delivery workflow

The normal workflow has two phases and two operator gates:

```text
Define
  -> one unified specification
  -> definition gate
  -> operator sign-off

Deliver
  -> write automated tests
  -> observe RED
  -> implement
  -> observe GREEN
  -> implementation gate
  -> execute one-off and user tests
  -> operator closure
```

The unified specification contains:

- context;
- acceptance criteria describing falsifiable system states rather than test actions;
- test definitions derived from and traced to the acceptance criteria;
- edge cases; and
- a solution design covering architecture context, high-level design, relevant architectural qualities, and mandatory security impact.

The specification is a context-independent delivery handoff. A new agent must
be able to deliver it using only the specification, the project profile, its
named authorities, and the repository. It must not depend on the drafting
conversation, phrases such as "as discussed", unnamed source material,
uncaptured operator decisions, or assumed implementation knowledge. Failure of
this handoff test blocks the definition gate.

Test definitions retain the SDLC v1 classifications:

- `RT`: persistent automated regression test;
- `UT`: human user test; and
- `OT`: bounded one-off test.

These terms and the automated TDD rule are deliberately resurrected from SDLC
v1. They are not Spec Kit concepts or newly invented SDLC v3 terminology.

## SDLC v1 source recovery

- Recover resurrected definitions from the latest v1 release rather than from
  memory or their later v2 restatements. The current latest v1 tag is `v1.0.3`.
- Use `v1.0.3:ISSUES.md` for the original falsifiable-system-state definition
  of an acceptance criterion and the AC/test boundary.
- Use `v1.0.3:TESTING.md` for RT, UT, and OT classification, automated TDD,
  observable user boundaries, one-off lifecycle, and human ownership of user
  tests.
- Review `v1.0.3:MAIN.md` and relevant v1 skills for any selected pairing,
  emergency, identifier, approval, and handoff semantics.
- Record a concise provenance map while implementing v3. Do not restore the v1
  process wholesale; recover only the definitions explicitly chosen for v3.

## Audit topology

Each gate uses no more than two contexts: the authoring context and one external auditor context.

### Definition gate

- Apply the `audit-spec`, `audit-design`, and `audit-tests` criteria together in the authoring context.
- Remediate and repeat for at most five local rounds.
- Start one external auditor context only after all three pass locally.
- Apply all three audits in that single external context and return one combined verdict.
- After remediation, rerun the local checks and resume the same external auditor context.

### Implementation gate

- Apply the `audit-tests` and `audit-code` criteria together in the authoring context.
- Remediate and repeat for at most five local rounds.
- Start one external auditor context only after both pass locally.
- Apply both audits in that single external context and return one combined verdict.
- After remediation, rerun the local checks and resume the same external auditor context.

For the first SDLC v3 release, use Codex-native agent contexts rather than multi-harness audit dispatch. The exact mechanism must be verified before it becomes implementation authority.

## Variant workflows

### Paired development

- Retain the existing paired-development contract.
- The bounded objective and explicit iteration instructions are the working specification.
- Use automated TDD where appropriate without replacing valid live user validation with artificial tests.
- Consolidate the accepted result into the unified specification and validation record at closure.
- Apply only the change-scoped audits warranted by the work.

### Emergency delivery

- Only the exact operator invocation `BYPASS-GATE-7` enables this route.
- The sufficiently bounded request is the temporary specification.
- Select test evidence, observe RED where an automated regression test applies, implement, and observe GREEN.
- Run the implementation gate against the resulting tests and code.
- Execute required one-off and user tests.
- Reconcile the durable specification, solution design, and project documentation after implementation.
- Do not impose a retrospective definition gate.

## Candidate project artefacts

The intended lean shape is:

```text
.sdlc/project.yaml
docs/work.org
specs/NNN-descriptor/spec.org
specs/NNN-descriptor/audits.org
specs/NNN-descriptor/validation.org
```

- `.sdlc/project.yaml` is the tracked, non-secret project SDLC profile.
- `work.org` is the proposed local work ledger and must use proper Org heading hierarchy for folding.
- `spec.org` is the single definition and approval artefact.
- `audits.org` preserves revision-specific audit evidence outside the audited specification.
- `validation.org` is required only when execution evidence must be recorded, especially for one-off and user tests.
- Separate Spec Kit plans, tasks, checklists, and generated metadata are not part of the proposed workflow.

Use a deliberately small, portable Org subset: headings, ordinary and
description lists, property drawers, `CUSTOM_ID`, internal and file links,
verbatim text, and source blocks. Do not use Babel execution, macros, dynamic
blocks, generated agenda machinery, or executable Emacs configuration. Keep
TODO states in `work.org`; the specification is an authority document rather
than a task list.

Give every acceptance criterion and test definition a stable `CUSTOM_ID`.
Acceptance criteria link to the tests that prove them, and tests link back to
the criteria they cover. This bidirectional traceability must remain readable as
plain text and navigable in Emacs or a rendered document.

Use `.sdlc/project.yaml` so the purpose remains explicit without colliding with
an application's own root configuration.

## Configuration direction

- Replace SDLC configuration currently persisted in project `.env` files with YAML that agents may safely read.
- Use `~/.agents/sdlc.yaml` for global non-secret defaults.
- Use tracked `.sdlc/project.yaml` for project facts and project-local overrides.
- When a global value exists, show it in the initializer and offer a
  project-local alternative. An empty response inherits the global value and
  does not copy it into the project file.
- Provide a deterministic migration from existing SDLC-managed `.env` keys to YAML without exposing or parsing unrelated project secrets.
- Do not require agents to read `.env`.
- Keep secrets out of both YAML files.
- Resolve command-line, process-environment, project, global, then schema-default
  values in that order. Persist project facts and explicit project overrides,
  not inherited global defaults.
- Drive questions, types, choices, defaults, validation, and persistence from a
  deterministic schema rather than hard-coded prompts.

## Migration requirements

### From SDLC v2 with Spec Kit

- Do not normalize every unfinished feature during project migration.
- Remove Spec Kit from the active project, including `.specify` infrastructure
  and project-local Spec Kit skills.
- Preserve incomplete Spec Kit work unchanged as archived migration evidence and
  index it in `work.org`; do not normalize it until the operator deliberately
  resumes that work.
- When a work item is resumed, convert its relevant specification, plan, task, audit, and validation material into the lean unified specification without changing its identifier or losing history.
- Preserve existing effective audit and operator-approval evidence where it maps cleanly to the new gates.
- Treat the constitution as migration evidence, not mandatory continuing ceremony.
- Extract genuine project-wide invariants and ownership boundaries into the new project profile or existing authoritative documentation.
- Preserve the complete pre-migration repository state on its dated archive
  branch before removing superseded Spec Kit artefacts from the migrated branch.

### From SDLC v1

- Retain the existing pre-migration skill.
- Archive GitHub tickets and comments locally.
- Produce the complete historical acceptance-criteria ledger and ticket-migration record.
- Close and classify legacy tickets according to the established migration rules.
- Transfer unresolved defects and undelivered feature ideas into the local work ledger without re-specifying delivered work.
- Treat the migrated acceptance-criteria ledger as baseline history for later changes.

### New projects

- Create the project profile, work ledger, and lean specification facilities directly.
- Do not install or initialize Spec Kit.

## Global commands

### `sdlc-project-init`

- Run exactly once for a project. The presence of `.sdlc/project.yaml` means the
  project is initialized and prevents repeated migration.
- Identify the primary branch, create a dated branch preserving the exact SDLC
  v1 or v2 state, and ask whether to push that branch to `origin`.
- Create a migration branch from the primary branch and detect a new, SDLC v1,
  or SDLC v2 project.
- For eligible GitHub-backed legacy projects, offer the existing ticket pre-migration skill.
- Run the appropriate migration or new-project initialization path.
- Never process dormant incomplete v2 work merely because the project is being migrated.
- Ask schema-driven non-authority questions before migration. After migration,
  use a saved read-only headless Codex prompt to propose product, architecture,
  and requirement documents as structured YAML lists. Show descriptors and
  rationales, let the operator confirm or replace the paths, then create
  `.sdlc/project.yaml`.
- Keep the temporary proposal in the project-owned `.sdlc/.init/` directory so
  an interrupted initialization is recognizable, then remove it before staging
  the completed migration. Never use private `.git` paths as application
  storage.
- Validate and commit the migration, then ask whether to merge it into the
  original primary branch.

### `sdlc-project-update`

Park this command. Global skills and standards must update without rewriting
every initialized project. Add a project-update command only if a future
concrete migration requirement cannot be satisfied safely another way.

## Global skill installation and unsupported harness cleanup

- `make install` installs the supported SDLC skills once under the global
  canonical root `~/.agents/skills`. Project initialization records
  configuration and creates project artefacts, but it does not copy skills.
- The first v3 installation removes every public SDLC-managed v1 and v2
  artefact from Claude, Hermes, and Copilot because those harnesses are
  temporarily unsupported. Their provider directories must contain no SDLC
  commands, hooks, prompts, or skills after cleanup, and v3 installs nothing
  there until support for that harness is explicitly restored.
- Cleanup uses an explicit inventory of SDLC-owned paths. It must not remove
  provider configuration or personal artefacts owned by the separate agents
  project.
- Codex is the only v3 delivery harness in the first release. Later provider
  support requires an explicit compatibility design and validation.

## Document preview

- Org is the canonical v3 specification format because it provides foldable
  hierarchy, stable internal links, properties, and bidirectional AC-to-test
  navigation.
- Pandoc support makes Org specifications viewable as high-quality HTML without
  requiring Emacs.
- Include a public, cross-platform `sdlc-preview` utility derived from the useful
  behaviour of the operator's private `htmlpreview` tool without inheriting its
  private configuration or dependencies.
- The viewer accepts Markdown and Org, renders through Pandoc, opens the result
  in the user's browser, and reports a clear dependency error when Pandoc is
  unavailable. Previewing is presentation, not an approval gate.
- Render to a collision-resistant random HTML filename in the source document's
  directory so document-relative links continue to resolve. Never overwrite an
  existing file.
- Remove the generated HTML one second after opening it, using delayed cleanup
  that cannot race the browser's initial read. Do not leave stale previews after
  an open or render failure.
- Begin with a basic bundled HTML and CSS template. A community template may be
  adopted later only after its quality, licence, and provenance are reviewed.
- Add a short Org primer under `src/` and route it only for Org work. It must
  explain the supported portable subset and require Org =verbatim= or ~code~
  syntax rather than Markdown backticks, even though Pandoc accepts backticks.

## Org authoring and folding

- Use Org as an outliner, not as Markdown with different punctuation. Headings
  are meaningful addressable nodes; lists contain subordinate material within
  a node.
- Keep the hierarchy shallow enough to navigate. Three or four heading levels
  are normally sufficient; a deeper structure is a signal to restructure or
  split the document.
- Distinguish checkboxes for finite list-item completion from TODO keywords for
  the lifecycle of a heading.
- Use tags for cross-cutting classification and properties for structured node
  attributes. Do not reproduce the heading hierarchy as tags.
- Use typed blocks according to meaning: source blocks for code, example blocks
  for literal output, and quote blocks for quoted prose.
- Use drawers for metadata or secondary supporting material. Do not hide
  first-class requirements, design decisions, or evidence in a drawer merely to
  shorten the visible document.
- Use `#+STARTUP: overview` or `content` for file-wide initial folding. A dense
  supporting subtree not intended for casual inspection may set
  `:VISIBILITY: folded` in its property drawer.
- Folding is a view over the explicit tree, not a substitute for structure.
  Documents must remain understandable as plain text and rendered HTML.

## Requested design artefacts

- The implementation handover checklist is drafted in
  `docs/SDLC-V3-IMPLEMENTATION-PLAN.md`.
- The accepted unified-specification template is installed from
  `src/templates/v3/spec.org`.
- The accepted work-ledger template is installed from
  `src/templates/v3/work.org`.
- The operator accepted both Org templates on 2026-09-06.
- The implementation plan records the resolved decisions and remaining release
  checks.

The specification template must make context-independent handoff explicit and
record the delivery objective, relevant starting state, exact authorities,
constraints, dependencies, resolved decisions, unresolved blockers, and enough
solution detail for a new agent to begin with the defined tests.

## Resolved implementation decisions

- Use project-wide never-reused `WNNN`, `ACNNN.n`, `RTNNN.n`, `UTNNN.n`, and
  `OTNNN.n` identifiers.
- Preserve complete v2 state on the dated archive branch and move active
  `.specify` and `specs` material unchanged under `docs/archive/sdlc-v2/`.
- Classify every archived v2 specification through the temporary discovery
  YAML and render it under the existing canonical `docs/work.org` sections.
  Keep individual feature specifications out of the stable authority lists;
  the work ledger supplies their state, provenance, and links.
- Composite audit skills use one Codex subagent and retain its identity for
  follow-up reviews; focused audits never spawn separate contexts themselves.

## Implementation and pilot chronology

This section records significant actions as well as decisions. It exists so an
interrupted or compacted agent context can resume from repository evidence
without reconstructing the work from conversation history.

### 2026-09-06

- Created the `sdlc-v3` development branch and recorded the initial redesign
  decisions (`9ceb42c`).
- Made context-independent delivery specifications an explicit requirement
  (`214542c`).
- Settled the v1, v2, and new-project migration architecture (`da5592f`) and
  prepared the implementation handover (`33bc1cc`, `6722121`).
- Defined the portable Org outlining contract and canonical v3 templates
  (`72847f8`).
- Implemented the lean v3 routing, globally installed skills, two-phase
  delivery workflow, paired and emergency variants, and Org artefacts
  (`ea60974`).
- Implemented the first v3 installer and migration path, including removal of
  superseded Spec Kit runtime artefacts (`aee79cd`).
- Exercised isolated new-project, v1, and v2 migrations and recorded the
  rehearsal evidence (`246d832`).
- Added post-migration Codex discovery of project authority documents with
  structured YAML output and operator confirmation (`2827c3e`).
- Corrected installer handling of managed prompt files so a successful install
  verifies as current (`2a1bd58`).
- Moved model-question wording, choices, defaults, and persistence into the
  deterministic project-initialization schema, then corrected the generic
  renderer so every question has the same order and presentation
  (`9c23e5d`, `909fd4f`, `6bc8375`, `a039b8c`).
- Changed v2 migration discovery to identify approved Spec Kit specifications,
  migrate their state into `docs/work.org`, and keep individual archived specs
  out of stable requirement-authority configuration (`e36ccb7`, `b9113f8`).
- Began the first paired live migration in First Folio. The interrupted run
  exposed that initializer recovery state had been stored under `.git`, which
  violated the repository boundary and made recovery opaque.
- Replaced the private `.git` workspace with the explicit project-owned
  `.sdlc/.init/` workspace and added the absolute rule that agents must never
  directly manipulate `.git` internals (`6a54e5c`).
- The operator preserved the interrupted First Folio migration in a Git stash,
  returned to `master`, and removed the obsolete `.git/sdlc-project-init`
  directory. A residual untracked archived `feature.json` was identified before
  restarting. The live migration has not yet been rerun with the corrected
  initializer.
- Clarified the deployment command: from the SDLC repository use
  `make install`; from another directory,
  `make -C /Users/tigger/code/agents/sdlc install` invokes the same target
  without first changing directory.
- The resumed First Folio pilot showed that requirement-authority discovery
  proposes `docs/ACs.org` but omits `docs/work.org`. This is deterministic
  current behaviour: both the discovery prompt and proposal validator exclude
  the work ledger. It conflicts with the operator's expectation that the
  canonical ledger locate current v3 requirements and remains to be corrected.
- Changed the read-only authority-discovery invocation to use
  `SDLC_AUDIT_MODEL` rather than `SDLC_SPEC_MODEL`. This bounded classification
  work should use the configured fast audit model; the default is
  `gpt-5.6-luna`.
- Refined authority resolution after the First Folio pilot. The initializer
  should determine stable authority paths without a model: always include
  `docs/work.org`; include `docs/ACs.org` when present, otherwise include
  `docs/ACs.md`; and fail when both AC ledgers exist. Detect `README.md`, the
  established upper- or lower-case vision path under `docs/`, and established
  root or `docs/` architecture paths from Git's exact path inventory. Preserve
  the tracked case, validate every selected file, and ask the operator only
  when a required authority is missing or genuinely ambiguous. The headless
  audit model remains useful only for classifying archived Spec Kit work by
  disposition so those entries can populate `docs/work.org`; it should not
  select the stable authority documents.
- Refined deterministic product and architecture discovery to inventory the
  current project through Git. Match Markdown or Org files whose filename stem
  contains `README`, `VISION`, or `ARCHITECTURE`, case-insensitively, so names
  such as `ARCHITECTURE-v2.md` and `V2_VISION_DRAFT.md` remain visible. Preserve
  the exact tracked or untracked path and case.
- Present every matching product or architecture candidate with an explicit
  selected or unselected state. Canonical current locations may begin selected;
  versioned, draft, proposal, archived, and legacy locations remain visible but
  begin unselected. Let the operator toggle multiple candidates before
  accepting the category.
- The authority prompt must offer a refresh action that repeats the Git
  inventory for the current category without discarding earlier migration work,
  confirmed answers, or the current selection for paths that still exist. This
  allows the operator to relocate a misplaced authority such as a root-level
  `ARCHITECTURE.md` into `docs/`, then rescan and continue without restarting.
  Do not use the `find` command or inspect `.git` internals.
- Never ask the operator or a model whether `docs/work.org` or the applicable
  legacy AC ledger is an authority. The initializer always records
  `docs/work.org` under requirement authorities. It also records `docs/ACs.org`
  when present, otherwise `docs/ACs.md` when present. Both ledgers existing is
  an error; a missing ledger is an error only for a migration source that
  requires one. A genuinely new project has no legacy ledger to record.
- The only semantic model work during project initialization is an SDLC v2
  migration with archived Spec Kit specifications. Invoke the configured audit
  model with a short bounded prompt to classify each archived specification as
  delivered, approved-undelivered, abandoned, or unresolved. The model returns
  structured status evidence only; it does not select authorities or write the
  ledger.
- Render `docs/work.org` deterministically from the canonical template. Populate
  it from the bounded v2 classification for an SDLC v2 migration and from the
  ticket-migration record for an SDLC v1 migration. For a new project, or a v2
  project with no archived specifications, install the clean template skeleton
  rather than creating a zero-byte file. Skip the model invocation when there
  are no archived Spec Kit specifications to classify.
- Implemented the refined authority boundary. Product and architecture choices
  now come from the exact Git inventory with stem matching, deterministic
  defaults, multi-selection, manual replacement, and in-place refresh. The
  project profile always receives `docs/work.org` plus the sole applicable AC
  ledger. The obsolete requirement-authority configuration field is retained
  only as a retired `.env` key for migration cleanup.
- Narrowed the headless prompt and YAML contract to archived Spec Kit work
  classification. It runs only when archived specifications exist, uses
  `SDLC_AUDIT_MODEL`, receives the exact specification paths, and cannot return
  authority selections. The script validates complete coverage and renders the
  work ledger itself.
- Verified the implementation with the complete repository-owned `make test`
  target, including Go vetting, linting, shell checks, formatting checks, and
  all Go unit tests.
- The second First Folio rehearsal exposed two initializer defects. Declining
  the advertised optional ticket pre-migration aborted v1 initialization, and
  the project workspace lacked the agreed nested ignore file, allowing a
  manual synchronization to commit `.sdlc/.init/` on the interrupted branch.
- Changed a declined ticket migration to continue with the existing ledger.
  Added a canonical project ignore template and merge-safe creation of
  `.sdlc/.gitignore`, ensuring `.init/` is ignored without replacing existing
  project entries.
- Added focused unit coverage for both defects and reran the complete
  repository-owned `make test` target successfully.
- Unified initializer question presentation. Every question now starts after a
  blank line, and schema choices and authority candidates use one shared `[x]`
  and `[ ]` renderer. The schema continues to own questions, options, defaults,
  and validation; scalar and multi-select input behaviour remains separate.
- Replaced the initializer's default per-file Git migration inventory with one
  concise commit summary. `VERBOSE=1` preserves the complete Git output, and a
  failed commit replays its captured stdout and stderr before returning the
  error.
- Completed the first end-to-end SDLC v2 to v3 project migration in First
  Folio. The initializer created `.sdlc/project.yaml` and `docs/work.org`, moved
  the active Spec Kit installation and three feature directories unchanged to
  `docs/archive/sdlc-v2/`, and committed the result on
  `sdlc-v3-migration-2026-09-06-3`.
- The operator merged that migration branch into First Folio's `master` as
  commit `13001c6`. The active `.specify/` and `specs/` directories are absent;
  the archived copies, v3 project profile, work ledger, and named authority
  documents are present. No project-local `SKILL.md` remains under
  `.agents/skills/`; the v3 workflow uses the global skills deployed under
  `~/.agents/skills/`.
- The migrated First Folio ledger classifies unified font configuration and
  source frontmatter configuration for human review, and manuscript block
  layout as closed work. Migration preserved these dispositions rather than
  forcing unfinished v2 work through a definition exercise.
- First Folio's local `master` was three commits ahead of `origin/master` after
  the merge. Publishing that branch remains an operator synchronization step;
  it does not affect the local migration structure.
