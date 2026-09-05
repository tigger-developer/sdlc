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
specs/NNN-descriptor/spec.md
specs/NNN-descriptor/audits.md
specs/NNN-descriptor/validation.md
```

- `project.yaml` is a proposed tracked, non-secret SDLC profile, not an existing convention.
- `work.org` is the proposed local work ledger and must use proper Org heading hierarchy for folding.
- `spec.md` is the single definition and approval artefact.
- `audits.md` preserves revision-specific audit evidence outside the audited specification.
- `validation.md` is required only when execution evidence must be recorded, especially for one-off and user tests.
- Separate Spec Kit plans, tasks, checklists, and generated metadata are not part of the proposed workflow.

The exact project path for `project.yaml` remains to be confirmed. `.sdlc/project.yaml` is the current candidate because it names both the purpose and filename without colliding with an application's own root configuration.

## Configuration direction

- Replace SDLC configuration currently persisted in project `.env` files with YAML that agents may safely read.
- Use a global non-secret configuration file, currently proposed as `~/.agents/sdlc.yaml`.
- Allow a tracked project `project.yaml` to override global defaults.
- Provide a deterministic migration from existing SDLC-managed `.env` keys to YAML without exposing or parsing unrelated project secrets.
- Do not require agents to read `.env`.
- Keep secrets out of both YAML files.
- Define the schema and precedence before implementation.

## Migration requirements

### From SDLC v2 with Spec Kit

- Do not normalize every unfinished feature during project migration.
- Preserve incomplete Spec Kit work unchanged until the operator deliberately resumes that work.
- When a work item is resumed, convert its relevant specification, plan, task, audit, and validation material into the lean unified specification without changing its identifier or losing history.
- Preserve existing effective audit and operator-approval evidence where it maps cleanly to the new gates.
- Treat the constitution as migration evidence, not mandatory continuing ceremony.
- Extract genuine project-wide invariants and ownership boundaries into the new project profile or existing authoritative documentation.
- Archive superseded Spec Kit machinery only through an explicit, recoverable migration step.

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

- Detect a new, SDLC v1, or SDLC v2 project.
- For eligible GitHub-backed legacy projects, offer the existing ticket pre-migration skill.
- Run the appropriate migration or new-project initialization path.
- Never process dormant incomplete v2 work merely because the project is being migrated.

### Project update command

The second global command and its exact contract remain to be confirmed. It is expected to update SDLC project configuration and bootstrap material without re-running migration.

## Requested design artefacts

- A Markdown unified-specification template using the operator's ADHD presentation contract, to be opened with `$HTML_PREVIEW_TOOL`.
- An Org work-ledger template using meaningful nested headings, to be opened in Emacs.
- Both templates are to be reviewed before repository implementation begins.

The specification template must make context-independent handoff explicit and
record the delivery objective, relevant starting state, exact authorities,
constraints, dependencies, resolved decisions, unresolved blockers, and enough
solution detail for a new agent to begin with the defined tests.

## Open decisions

- Final project path and schema for `project.yaml`.
- Exact global-to-project YAML precedence, including command-line and process-environment overrides.
- Exact local work-item and derived acceptance-criteria/test identifier formats.
- The treatment and archive location of v2 constitutions and generated Spec Kit infrastructure.
- The second global command's name and scope.
- The verified Codex mechanism for retaining one external auditor context across gate retries.
