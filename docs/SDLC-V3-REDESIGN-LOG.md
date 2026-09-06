# SDLC v3 Redesign Log

Status: Working design record. This is not an approved specification or implementation authority.

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
- Ask schema-driven project questions, create `.sdlc/project.yaml`, validate and
  commit the migration, then ask whether to merge it into the original primary
  branch.

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

## Requested design artefacts

- The implementation handover checklist is drafted in
  `docs/SDLC-V3-IMPLEMENTATION-PLAN.md`.
- The Org unified-specification review candidate is drafted in
  `docs/SDLC-V3-SPEC-TEMPLATE.org`.
- The Org work-ledger review candidate is drafted in
  `docs/SDLC-V3-WORK-TEMPLATE.org`.
- The templates and the decisions explicitly left open in the implementation
  plan must be reviewed before repository implementation begins.

The specification template must make context-independent handoff explicit and
record the delivery objective, relevant starting state, exact authorities,
constraints, dependencies, resolved decisions, unresolved blockers, and enough
solution detail for a new agent to begin with the defined tests.

## Open decisions

- Exact local work-item and derived acceptance-criteria/test identifier formats.
- The treatment and archive location of v2 constitutions and generated Spec Kit infrastructure.
- The verified Codex mechanism for retaining one external auditor context across gate retries.
