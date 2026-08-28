# Spec Kit Migration Pilot: Writeback

**Date:** 2026-08-28

**Status:** In progress

**SDLC branch:** `spec-kit-prototype`
**Pilot project:** Writeback, branch `spec-kit-pilot`

## Purpose

This record captures the first brownfield trial of the standards-only SDLC with
GitHub Spec Kit. It records operator prompts, observed behaviour, workarounds,
and process gaps. It is evidence for improving the integration after the trial,
not a normative workflow.

The operator runs the delivery agent in a separate shell. The agent maintaining
this record acts only as coach and scribe. The trial deliberately supplies
context step by step so the final review can distinguish:

- behaviour Spec Kit handled without assistance;
- context supplied by the SDLC preset;
- context that required operator prompting;
- incorrect assumptions or missing dependencies; and
- procedures worth extracting into optional, operator-invoked skills.

## Brownfield specification model

Writeback already has an implemented specification and regression evidence.
Spec Kit must understand both the existing system and the proposed change.

### Existing sources

The as-built specification may be distributed across:

- project documentation, which is the strongest current narrative of the
  system;
- `docs/ACs.md`, the centralized acceptance-criteria and test-traceability
  record;
- older closed tickets whose acceptance criteria predate migration into
  `docs/ACs.md`;
- open tickets whose implementation and automated tests exist but whose human
  approval remains outstanding; and
- implementation and tests, which provide evidence but do not themselves
  establish human approval.

Ticket state is therefore not sufficient to classify behaviour. An open ticket
may be implemented, and a closed early ticket may contain acceptance criteria
that were never centralized.

### Two classes of open legacy ticket

The migration must distinguish two operator-selected actions. The agent must
not infer the action from GitHub's open state, implementation evidence, test
results, or ticket content.

- **CLOSE:** The operator supplies a numbered list of implemented tickets that
  have now received the human validation required for closure. The closure
  procedure reconciles project documentation, centralizes AC and test
  traceability where applicable, comments on the tickets, and closes them.
- **MIGRATE:** The operator supplies still-live tickets written for the former
  SDLC. Their requirements and design must be translated into the appropriate
  Spec Kit artefacts for continued delivery. Migration does not itself mean
  approval, implementation, or closure.

An implemented but unapproved ticket may enter either class only through an
explicit operator decision: CLOSE after the operator validates it, or MIGRATE
when remaining requirements or validation work should continue under Spec Kit.

The current candidate is one operator-invoked `$migrate-legacy-sdlc` skill. It
asks the operator to divide the relevant tickets into disjoint CLOSE and MIGRATE
lists. Unlisted tickets remain untouched, and a ticket appearing in both lists
is reported for operator classification rather than inferred by the agent.

### Requirement states

| State | Meaning |
|---|---|
| Approved and implemented | Established as-built specification |
| Approved but stored only in a ticket | Established specification missing from `docs/ACs.md` |
| Implemented but awaiting approval | Provisional behaviour, not an approved requirement |
| Tested but awaiting human validation | Evidence exists, but approval remains outstanding |
| Proposed only | Future work, not part of the as-built baseline |
| Superseded | Historical requirement retained for traceability |

Passing automated tests does not imply human approval. Only the operator may
record approval or a human-validation result.

### Feature specification boundary

A new Spec Kit feature specification defines a delta against the as-built
specification. When the feature touches an existing acceptance criterion, the
feature context must include:

- the existing criterion, with identifier and descriptor;
- its regression-test traceability;
- the proposed new or changed behaviour; and
- one of these relationships: unchanged, extended, replaced, or potentially
  affected.

An intended traceability shape is:

| Existing specification | Existing evidence | Relationship | New requirement |
|---|---|---|---|
| `AC-12 - descriptor` | `RT-4 - descriptor`, `RT-7 - descriptor` | Extended | `FR-3 - descriptor` |

The trial will not bulk-convert the complete historical acceptance-criteria
record. It will bring forward only the baseline relevant to the feature.

## CLOSE: approved legacy ticket closure before cutover

Before starting the first feature, the operator identified approved legacy
tickets that should complete the former SDLC closure procedure. Their project
documentation, acceptance criteria, and test traceability must be reconciled
before the new feature begins.

### Candidate operator prompt: close approved legacy tickets

The following prompt is a candidate for reuse as the CLOSE phase of
`$migrate-legacy-sdlc`. It does not govern MIGRATE tickets.

```text
BYPASS-GATE-7

This is authorized legacy SDLC closure maintenance for Writeback. The currently deployed Spec Kit SDLC no longer contains the former ticket-closure procedure, so follow the complete procedure stated here.

I have personally reviewed and approve the following tickets:

- #<ticket number>: <descriptor>
- #<ticket number>: <descriptor>
- #<ticket number>: <descriptor>

My approval is authoritative for these tickets. It includes authorization to:

- update the project documentation;
- migrate their acceptance criteria and test traceability;
- mark pending UT and RT tests as PASSING;
- close successfully processed tickets.

Do not reassess or second-guess this approval.

## Scope and safeguards

1. Work only on the listed tickets.
2. Do not apply Spec Kit feature-generation commands to this maintenance operation.
3. Do not modify implementation code or tests.
4. You may inspect implementation code and tests when necessary to establish the documented as-built behaviour, but do not infer unsupported requirements.
5. Inspect the Git working tree before editing. Preserve unrelated and human-authored changes.
6. The complete operation must be idempotent:
   - do not duplicate ACs, tests, provenance, comments, or documentation;
   - reconcile partial migrations in place;
   - do not create an empty commit when no repository changes are required.
7. Complete all unblocked tickets in the batch. Do not interrupt the batch to ask about one ticket.

## Establish the project baseline

1. Before processing any ticket, read the project documentation in full, including:
   - README.md;
   - every applicable document under docs/;
   - docs/ACs.md.

2. Treat the project documentation as the best available description of the as-built system, while recognizing that the listed open tickets may contain implemented changes not yet incorporated into it.

3. Treat docs/ACs.md as the current centralized requirements and test-traceability record, while recognizing that older tickets may contain ACs that predate the migration into that document.

4. Use implementation code and existing tests as supporting evidence when needed, but do not treat test existence or test success as a substitute for human approval.

## Processing order

1. Sort the authorized tickets into ascending numerical order.
2. Process them from the smallest ticket number to the largest.
3. This ascending ticket order is the defined chronological precedence for this migration.
4. A requirement from a higher-numbered ticket takes precedence over a materially conflicting requirement from a lower-numbered ticket.
5. Never regress documentation that already reflects a later ticket merely to reproduce an older ticket's wording or design.

## Process each ticket

For each ticket, in ascending numerical order:

1. Read the complete ticket body and every comment.
2. Determine the implemented behaviour and its effects on:
   - user-visible behaviour;
   - interfaces;
   - stored data;
   - constraints;
   - operations;
   - architecture;
   - testing;
   - other documented behaviour.

3. Compare the ticket's resulting behaviour with the current project documentation.

4. For every ticket, including a bug-fix ticket without an AC table:
   - update any project documentation made incomplete or stale by the ticket;
   - preserve documentation that already describes the correct current behaviour;
   - make no documentation change when the ticket is already accurately represented;
   - update documentation to describe the cumulative as-built system, not the historical intermediate state;
   - never remove historical requirements from docs/ACs.md.

5. Classify the ticket:
   - if it contains an AC table, follow the AC migration procedure below;
   - if it contains no AC table, treat it as a bug-fix ticket requiring no AC migration, but still complete the documentation review and any necessary documentation updates.

## AC migration procedure

For every ticket containing an AC table:

1. Transfer every acceptance criterion and its associated test specification, test reference, and traceability information into docs/ACs.md.

2. Preserve:
   - the original AC identifier and descriptor;
   - the original requirement wording;
   - the originating ticket number and descriptor;
   - test identifiers and descriptors;
   - AC-to-test relationships;
   - relevant status and provenance information.

3. Do not:
   - summarize;
   - renumber;
   - silently rewrite;
   - omit information;
   - duplicate an existing or partially migrated entry.

4. Reconcile partial existing entries in place.

5. Apply these test-status rules while transferring test information:
   - change every UT marked pending to PASSING;
   - change every RT marked pending to PASSING;
   - preserve the status of every OT exactly as recorded;
   - do not infer, promote, or otherwise change an OT status;
   - preserve the status of every other test type unless this prompt explicitly directs otherwise.

6. This prompt is the human authorization for the pending-to-PASSING transition for UT and RT tests. Do not require another approval or rerun tests solely to authorize that transition.

## Conflicting ACs

When comparing an AC being migrated with ACs already in docs/ACs.md or migrated from another listed ticket:

1. Determine whether the requirements materially contradict one another and cannot both be satisfied.

2. Do not treat requirements as conflicting merely because:
   - one is more detailed;
   - one applies to a narrower scope;
   - they use different wording;
   - one extends the other without invalidating it.

3. For a genuine contradiction between tickets:
   - the AC from the higher-numbered ticket takes precedence;
   - preserve the lower-ticket AC in docs/ACs.md;
   - mark the lower-ticket AC RETIRED or SUPERSEDED;
   - identify the superseding ticket using its number and descriptor;
   - identify the superseding AC using its identifier and descriptor;
   - preserve the retired AC's original wording, provenance, and test traceability;
   - ensure the general project documentation describes only the resulting current behaviour.

4. Apply ticket-number precedence regardless of the order in which an AC was previously copied into docs/ACs.md.

5. If an apparent contradiction cannot be resolved confidently using these rules:
   - preserve both ACs;
   - do not invent a resolution;
   - report the ambiguity using identifiers and descriptors;
   - leave the affected ticket open;
   - continue processing the remaining tickets.

## Incomplete information

If a listed ticket lacks information required for faithful documentation or AC migration:

1. Inspect the relevant existing implementation, tests, project documentation, and ticket discussion for evidence.
2. Do not invent missing requirements or behaviour.
3. Leave that ticket open if the missing information still prevents faithful processing.
4. Record exactly what is missing, using ticket and AC descriptors.
5. Continue processing every other unblocked listed ticket.

A bug-fix ticket without an AC table is not incomplete merely because it has no AC table. It may be closed once its documentation impact has been reviewed and addressed.

## Final consistency review

After processing all listed tickets:

1. Review the complete resulting documentation change.

2. Verify consistency across:
   - the project documentation;
   - docs/ACs.md;
   - the processed tickets;
   - migrated test traceability;
   - the cumulative as-built behaviour.

3. Verify specifically that:
   - every migrated AC retains its identifier, descriptor, provenance, and test traceability;
   - no AC or test entry was duplicated;
   - pending UT and RT statuses became PASSING;
   - OT statuses remained unchanged;
   - superseded ACs remain present and identify their replacements;
   - current project documentation reflects the latest applicable behaviour;
   - older tickets did not regress documentation that already reflected later behaviour;
   - every bug-fix ticket received a documentation-impact review.

4. Correct any inconsistency introduced by this migration before committing.

## Commit, push, and ticket closure

1. Commit all resulting documentation and AC migration changes in one coherent legacy-closure commit.

2. Use a concise commit message that describes the legacy ticket closure and documentation migration.

3. Push the commit.

4. Do not create an empty commit if the documentation and AC record were already current.

5. Add a concise comment to each successfully processed AC-bearing ticket stating:
   - that its ACs and test traceability were transferred or reconciled in docs/ACs.md;
   - that related project documentation was reviewed and updated where necessary;
   - the commit reference, when a commit was required.

6. Add a concise comment to each successfully processed bug-fix ticket stating:
   - that it was approved as a legacy bug fix;
   - that it contained no AC table and required no AC migration;
   - that its documentation impact was reviewed and updated where necessary;
   - the commit reference, when a commit was required.

7. Close every successfully processed ticket as completed.

8. Do not close a ticket whose required documentation update or AC migration remains incomplete.

## Final report

Report only:

- AC-bearing tickets migrated and closed, with ticket numbers and descriptors;
- bug-fix tickets closed without AC migration, with ticket numbers and descriptors;
- retired or superseded ACs, identifying both old and replacement ACs with descriptors;
- tickets left open, with descriptors and exact reasons;
- project documents changed;
- the commit reference, or that no commit was necessary;
- push status;
- anything authorized by this prompt that was not completed.

Do not apply Spec Kit feature-generation commands to this maintenance operation.
```

### Candidate unified migration skill

The working name is `$migrate-legacy-sdlc`. The skill would remain strictly
operator-invoked because it performs high-impact, transitional work and
contains substantial procedural detail that should not burden ordinary feature
sessions.

Its first operation should present open tickets using identifiers and
descriptors, then ask the operator for two lists:

```text
CLOSE
- #<ticket number> - <descriptor>

MIGRATE
- #<ticket number> - <descriptor>
```

It should then execute two distinct phases:

1. Reconcile, commit, and close every authorized CLOSE ticket.
2. Translate each authorized MIGRATE ticket into Spec Kit artefacts, obtaining
   operator direction when several tickets may form one coherent feature.

The skill should:

- require explicit, disjoint operator-supplied CLOSE and MIGRATE lists;
- never discover or close additional tickets autonomously;
- process tickets oldest to newest;
- reconcile documentation, ACs, and test traceability;
- preserve superseded requirements and evidence;
- apply the defined UT, RT, and OT status rules;
- commit, push, comment, and close only successfully migrated tickets;
- remain idempotent; and
- complete the unblocked batch before reporting unresolved tickets.

The skill will not be designed or implemented until the trial shows where the
prompt needs correction.

## MIGRATE: live legacy tickets into Spec Kit

MIGRATE applies to open tickets that still represent live work under the former
SDLC. These tickets are not being retrospectively approved and closed. Their
durable requirements and design must move into Spec Kit so delivery can
continue under the new orchestration model.

The likely artefact mapping is:

| Legacy ticket material | Spec Kit destination |
|---|---|
| User outcome, scope, requirements, and ACs | Feature `spec.md` |
| Existing behaviour affected by the change | Baseline and traceability section in `spec.md` |
| Technical design and architecture delta | `plan.md` |
| Test specifications and outstanding validation | Requirements traceability and later `tasks.md` |
| Unresolved product questions | Clarification inputs before planning |

The migration must preserve the originating ticket number and descriptor, AC
and test identifiers, approval state, outstanding human validation, and links
between old and new artefacts. It must not silently convert implementation or
test evidence into approval.

This is the second phase of `$migrate-legacy-sdlc`, not a separate standing
skill. Its exact scope, whether each listed ticket becomes one specification or
several tickets form a coherent feature, and the ticket state after successful
migration remain decisions for the live trial. No MIGRATE prompt has yet been
validated.

## Spec Kit initialization

The operator initialized the existing Writeback checkout with Spec Kit 1.0.1:

```sh
specify init --here --force --non-interactive --integration codex --script sh
```

`sh` is Spec Kit's selector for its Bash implementation. The installed scripts
use `#!/usr/bin/env bash`; `bash` is not an accepted selector value. The Codex
integration installs project-local skills under `.agents/skills`. An explicit
`--integration-options="--skills"` was found to be redundant because skills are
the Codex integration's default.

The operator then installed the SDLC preset:

```sh
specify preset add --dev ~/.agents/sdlc/presets/sdlc-standards
```

Template resolution reported this composition chain:

1. Spec Kit's core `constitution-template`.
2. The SDLC preset's appended `constitution-addendum.md`.

The generated `.specify/` infrastructure and installed preset were committed on
the pilot branch before the first constitution run. Project-local `.agents/`
runtime state remained ignored.

## First constitution attempt

### Candidate operator prompt: initial brownfield constitution

```text
$speckit-constitution

Establish the constitution for this existing brownfield Writeback project.

Before drafting it:

1. Read README.md and every applicable project document under docs/ in full.
2. Read the existing build, test, CI, dependency, and repository configuration needed to identify established engineering constraints.
3. Treat the project documentation as the best available description of the as-built system.
4. Treat docs/ACs.md as the centralized specification and test-traceability record for previously implemented behaviour.
5. Recognize that future Spec Kit feature specifications will define new or changed behaviour as a delta against that as-built specification.
6. Record that when a new feature touches an existing AC, its feature specification must identify the existing AC with its descriptor, retain its regression-test traceability, and classify the relationship as unchanged, extended, replaced, or potentially affected.
7. Preserve superseded requirements and their provenance rather than silently rewriting history.
8. Apply the installed SDLC Engineering Standards Profile and select only the standards applicable to this project.
9. Derive enduring project principles and constraints. Do not copy the entire AC table, roadmap, architecture, or individual feature requirements into the constitution.
10. Do not create the word-count feature specification, implementation plan, tasks, or code in this operation.

Create only the project constitution and any constitution-dependent Spec Kit template synchronization required by the constitution workflow. Report what sources informed each material principle and identify any unresolved contradiction without inventing a resolution.
```

### Observed dependency failure

The constitution skill stopped before drafting because the generated Bash
resolver could not import PyYAML:

```text
Error: PyYAML is required to resolve preset template composition
```

The failure exposed an undeclared execution dependency:

- the `sh` implementation invokes ambient Python internally;
- preset template composition additionally imports PyYAML;
- `specify init` did not identify or provision that dependency;
- `specify preset resolve` succeeded through the Specify CLI environment, while
  the generated agent-time resolver used another Python environment; and
- launching Codex through a PyYAML-equipped `uv run` environment did not make
  that environment available to Codex's later command execution.

The direct-interpreter prohibition was not the cause. The agent submitted the
tracked resolver script, which is a project-owned entry point. The script then
invoked Python internally, which the standards explicitly permit.

### Verified pilot workaround

Wrapping the resolver itself supplied its undeclared dependency at the point of
use:

```sh
uv run --isolated --with PyYAML --python 3.13 .specify/scripts/bash/resolve-template.sh constitution-template --json
```

This returned the composed core template and SDLC Engineering Standards Profile.
The operator then authorized the delivery agent to use that exact command and
resume the constitution workflow without invoking `python` or `python3`
directly.

A permanent solution remains undecided. Options to assess after the trial
include an upstream fix, a preset integration change, or a portable
repository-owned launcher.

## Initial constitution result

The delivery agent created and committed Writeback constitution version 1.0.0.
The initial draft included project-specific privacy, authorization, retention,
accessibility, architectural, verification, governance, and standards-profile
rules.

The review found these omissions:

- `docs/ACs.md` was not mentioned;
- the as-built specification versus feature-delta model was absent;
- existing-AC regression traceability was absent;
- the unchanged, extended, replaced, or potentially affected classification was
  absent;
- `~/.agents/sdlc/ISSUES.md` was missing from the standards profile;
- the requested per-principle source report was not produced;
- `make test` was described as a gate, reintroducing unnecessary gate
  terminology; and
- the installed standards tree carried no source revision marker, so the agent
  could not populate the adopted revision without external information.

The source checkout used for the deployed prototype was verified at revision
`0689a399986f`. The ratification date correctly remained unresolved because the
operator had not yet approved the draft.

### Candidate operator prompt: correct the unratified constitution

This prompt is retained as a candidate for the pilot process and any later
constitution guidance:

```text
$speckit-constitution

Revise the unratified Writeback constitution draft to correct omissions from the original instruction.

Use the project-approved resolver command:

uv run --isolated --with PyYAML --python 3.13 .specify/scripts/bash/resolve-template.sh constitution-template --json

Do not invoke python or python3 directly.

Make these specific corrections:

1. Keep the constitution at version 1.0.0 because this is correction of the initial unratified draft, not an amendment to an adopted constitution.

2. Keep RATIFICATION_DATE as a TODO. Do not claim that the constitution has been ratified.

3. Add a concise brownfield specification and traceability rule establishing that:
   - the project documentation is the best available narrative of the as-built system;
   - docs/ACs.md is the centralized specification and test-traceability record for previously implemented and approved behaviour;
   - a new Spec Kit feature specification defines new or changed behaviour as a delta against that baseline;
   - when a feature touches an existing AC, the specification must cite the AC using its identifier and descriptor, preserve its regression-test traceability, and classify the relationship as unchanged, extended, replaced, or potentially affected;
   - superseded requirements, provenance, and test lineage must be preserved rather than silently rewritten.

4. Add ~/.agents/sdlc/ISSUES.md to the Engineering Standards Profile because requirements and acceptance criteria are intrinsic to this project's Spec Kit workflow.

5. Set the adopted SDLC source revision to 0689a399986f, identified from the source checkout used for the deployed prototype.

6. Replace "make test gate" with "supported regression command" or equally neutral wording. Do not reintroduce SDLC approval-gate terminology.

7. Preserve the existing project-specific privacy, access, retention, accessibility, architecture, and evidence principles unless project evidence contradicts them.

8. In the final response, provide a concise mapping from each material constitutional principle to the project documents or repository evidence that informed it. Do not merely say that documentation was reviewed.

9. Do not create a feature specification, plan, tasks, tests, or implementation code.

Create a follow-up commit rather than rewriting the existing commit. Report the new commit identifier with its descriptor.
```

### Corrected constitution result

The delivery agent applied the candidate correction prompt in follow-up commit
`0817c0a - docs: correct unratified constitution traceability`.

Direct review of the committed diff verified that it:

- retained version 1.0.0 as an unratified draft;
- retained the unresolved ratification-date TODO;
- defined project documentation and `docs/ACs.md` as the brownfield baseline;
- defined new Spec Kit specifications as deltas against that baseline;
- required existing ACs to retain identifiers, descriptors, regression-test
  traceability, and unchanged, extended, replaced, or potentially affected
  classification;
- preserved superseded requirements, provenance, and test lineage;
- added `~/.agents/sdlc/ISSUES.md` to the standards profile;
- recorded SDLC source revision `0689a399986f`; and
- replaced `make test` gate wording with `supported regression command`.

The delivery agent also supplied the requested evidence mapping. Examples
verified during coaching included:

- privacy and private infrastructure in `docs/VISION.md` and data minimization
  in `docs/architecture.md`;
- the Go, SQLite, and minimal-infrastructure model in `docs/VISION.md`;
- feedback isolation, moderation, consolidated feedback, and keyboard access in
  `docs/VISION.md` and `docs/architecture.md`;
- the supported `make test` entry point in `Makefile`; and
- preserved superseded AC lineage in `docs/ACs.md`, including
  `AC68.1 - former purge_at column requirement`.

The constitution remains a draft until the operator explicitly ratifies it.

## Trial observations to carry forward

| Stage | Observation | Candidate improvement |
|---|---|---|
| CLOSE | Historical requirements may remain in old tickets, including closed tickets. | Provide a distinct CLOSE phase in the operator-invoked migration skill. |
| CLOSE | Ticket state does not establish implementation or approval state. | Require an explicit operator list and preserve human-test status. |
| MIGRATE | Still-live old-SDLC tickets require translation into Spec Kit artefacts. | Develop a distinct MIGRATE phase in the same operator-invoked skill. |
| Initialization | Codex uses skills by default; the explicit skills option is redundant. | Keep documented initialization concise. |
| Initialization | The `sh` selector installs Bash scripts. | Explain the selector accurately. |
| Preset resolution | Agent-time composition requires ambient Python and PyYAML. | Remove or provision the undeclared dependency. |
| Constitution | Brownfield traceability instructions were omitted from the first draft and added after an explicit correction. | Strengthen the preset or constitution prompt using trial evidence. |
| Standards profile | Installed copies do not expose their source revision. | Deploy a release or revision marker. |

## Remaining pilot work

- Review the corrected constitution and its source mapping.
- Obtain explicit operator ratification and record its date.
- Trial migration of a still-live old-SDLC ticket into Spec Kit without
  conflating migration, approval, implementation, or closure.
- Start a fresh feature context for the word-count calculation change.
- Create the feature specification as a delta against relevant existing ACs and
  regression tests.
- Record every additional prompt, omission, correction, and dependency through
  clarification, planning, tasks, analysis, implementation, and convergence.
- At the end of the trial, decide which guidance belongs in the lean preset,
  which belongs in optional skills, and which should be proposed upstream.
