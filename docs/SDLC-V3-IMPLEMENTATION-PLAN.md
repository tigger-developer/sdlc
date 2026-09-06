# SDLC v3 Implementation Plan

Status: Review candidate. This document translates the decisions in
`SDLC-V3-REDESIGN-LOG.md` into an implementation handover. It does not
authorize implementation.

## Intended outcome

Deliver a lean, standalone SDLC that:

- keeps the public engineering standards and progressive routing;
- uses one definition artefact and two gates for ordinary delivery;
- supports paired and emergency delivery without Spec Kit ceremony;
- installs global skills rather than copying skills into projects;
- migrates SDLC v1 and v2 projects without losing history or unfinished work;
- uses Org for foldable, linked specifications and work tracking; and
- initially supports Codex with OpenAI models only.

## Implementation checklist

### Source impact map

The implementation should begin from this explicit inventory rather than
discovering scope while editing:

- **Rewrite:** `cmd/sdlc-project-init/main.go` and
  `internal/projectinit/`. Replace constitution and Spec Kit setup with the
  once-only v3 initialization and migration controller.
- **Rewrite:** `src/config/project-init.schema.yaml`. Replace the environment
  key schema with the global and project YAML schema.
- **Adapt:** `src/libexec/load-sdlc-env.sh` and `internal/configenv/`. Restrict
  them to allowlisted one-time migration of historical SDLC keys from `.env`;
  they are not the v3 runtime configuration loader.
- **Rewrite:** `cmd/sdlc-install/main.go` and `internal/installer/`. Deploy the
  v3 common root, global skills, Codex support, CLI tools, and exact retirement
  plan for unsupported providers.
- **Review carefully:** `internal/installer/hermes.go` and the Claude, Copilot,
  and Codex configuration mutation functions in `internal/installer/installer.go`.
  Remove only public SDLC-owned hooks and rules from unsupported providers;
  preserve private agent configuration.
- **Retire or replace:** `cmd/sdlc-audit/main.go` and
  `internal/auditrunner/`. The current multi-harness external runner does not
  implement the v3 local-first, single-retained-Codex-context topology. Keep no
  compatibility CLI unless it has a defined v3 user.
- **Retain:** `internal/audit/verdict.go` only if its verdict parsing is reused
  by the v3 gate evidence contract.
- **Rewrite:** active files under `skills/`, including the two composite gate
  skills and the legacy migration skill. Remove Spec Kit terminology and
  project-local skill assumptions.
- **Rewrite:** `src/MAIN.md`, `src/AUDITS.md`, `src/ISSUES.md`,
  `src/TESTING.md`, `src/CODING.md`, `src/PAIRING.md`, and `src/EMERGENCY.md`
  around the two-phase v3 workflow while preserving the standards they own.
- **Review and retain:** `src/DOCUMENTATION.md`, `src/GIT.md`,
  `src/SECURITY.md`, and `src/technologies/`. Remove only obsolete v2 routing
  or configuration references.
- **Remove from active source:** `src/presets/sdlc-standards/` and all Spec Kit
  adapters, commands, and templates.
- **Add:** installed Org templates and the public `sdlc-preview` source.
- **Rewrite:** `Makefile`. Build and install only the v3 installer,
  initializer, previewer, and any CLI with a proven v3 purpose. Remove the
  `sdlc-project-update` installation and obsolete audit-runner wiring.
- **Rewrite:** README, quickstart material, architecture, changelog, and
  learnings after the implementation behaviour is stable.

### 1. Preserve the current state

- [ ] Record the current branch, remotes, tags, and dirty paths.
- [ ] Preserve the complete current SDLC v2 state on an archive branch before
  removing or replacing its unfinished work.
- [ ] Keep the existing v3 design commits and this review pack reachable from
  the implementation branch.
- [ ] Do not stage, overwrite, or discard unrelated working-tree changes.

### 2. Establish the v3 source layout

- [ ] Keep universal standards under `src/` and technology standards under
  `src/technologies/`.
- [ ] Add canonical installed templates for `spec.org`, `work.org`, audit
  evidence, and validation evidence.
- [ ] Add a schema for `~/.agents/sdlc.yaml` and `.sdlc/project.yaml`.
- [ ] Remove active Spec Kit presets, adapters, commands, templates, and
  references from the v3 source tree.
- [ ] Archive any v2 material retained for historical explanation rather than
  leaving it in active routing.

### 3. Recover the authoritative SDLC v1 definitions

- [ ] Treat the latest SDLC v1 release tag, currently `v1.0.3`, as the source
  evidence for concepts deliberately resurrected in v3.
- [ ] Extract the acceptance-criterion definition and AC/test boundary from
  `v1.0.3:ISSUES.md`.
- [ ] Extract RT, UT, and OT classification, automated TDD, real-user boundary,
  one-off lifecycle, and user-test ownership from `v1.0.3:TESTING.md`.
- [ ] Review `v1.0.3:MAIN.md` and the relevant v1 skills for paired delivery,
  emergency delivery, identifier allocation, approval ownership, and useful
  handoff language.
- [ ] Write a short provenance map naming each resurrected definition and its
  v1 source before rewriting the v3 instructions.
- [ ] Do not restore v1 gate ceremony, ticket workflow, modes, or terminology
  merely because they occur beside a useful definition. Only the concepts
  explicitly selected for v3 return.

### 4. Rewrite routing and workflow documents

- [ ] Make `MAIN.md` route only the standards and workflow needed for the
  current activity.
- [ ] Define the standard two-phase workflow:
  definition, definition gate, operator sign-off, TDD delivery,
  implementation gate, validation, operator closure.
- [ ] Retain the paired-development route with live user validation and
  change-scoped evidence.
- [ ] Retain the exact `BYPASS-GATE-7` emergency route. The agent must never
  infer or invoke it.
- [ ] Require every specification to pass the context-independent handoff
  test before definition sign-off.
- [ ] Resurrect the SDLC v1 RT, UT, and OT test classifications and its
  automated TDD rule. Do not present these as inherited from Spec Kit or as new
  SDLC v3 terminology.
- [ ] Preserve documentation, Git, security, and technology standards.
- [ ] Remove v2-only terms such as constitution stages, Spec Kit commands,
  generated plans, tasks, and convergence.

### 5. Create or revise global skills

All v3 skills are installed globally under `~/.agents/skills`. Each skill must
bootstrap from the current global SDLC and must not depend on a project-local
copy of itself.

- [ ] **Define change:** ask no more than three concise free-form questions per
  unresolved area, then create or revise the single `spec.org` artefact.
- [ ] **Definition gate:** apply specification, design, and test-definition
  audits together. Run up to five local remediation rounds, then all three
  audits in one retained external context.
- [ ] **Deliver change:** begin from a signed-off specification, write
  automated tests first, observe RED, implement, and observe GREEN.
- [ ] **Implementation gate:** apply test and code audits together. Run up to
  five local remediation rounds, then both audits in one retained external
  context.
- [ ] **Pair change:** retain the agreed paired-development behaviour and
  consolidate live decisions and validations into the durable artefacts.
- [ ] **Emergency change:** implement only after the exact operator keyword,
  then backfill specification, design, documentation, and evidence.
- [ ] **Migrate legacy tickets:** adapt the existing skill to prepare a v1
  project for v3 without requiring Spec Kit.
- [ ] **Focused audits:** retain `audit-spec`, `audit-design`, `audit-tests`,
  and `audit-code` as criteria and optional diagnostic skills. Their direct
  invocation must not accidentally create additional auditor contexts.
- [ ] Review `diagnose-issue`, `recommendations-please`, `summarize-issues`,
  and `useful-be` for v3 terminology and routing.
- [ ] Retire `audit-definition` and `audit-implementation` implementations
  that encode the abandoned multi-harness or Spec Kit workflow before
  replacing them with the v3 gate skills.
- [ ] Retire `convert-migrated-acs-to-org` once the main migration path
  performs that conversion directly and safely.

### 6. Implement the two gate runners

- [ ] Keep local self-audit in the authoring context. It must not launch a
  script or another agent.
- [ ] Launch one external Codex auditor only after every local component audit
  passes.
- [ ] Pass all component audit instructions and the bounded artefact/context
  set to that one auditor.
- [ ] Resume the same external auditor context after remediation.
- [ ] Enforce a maximum of five local rounds and five external rounds per
  gate.
- [ ] Record the artefact revision, each component verdict, combined verdict,
  findings, remediation, and superseding rerun in `audits.org`.
- [ ] Treat malformed output, timeout, or harness failure as an audit-runner
  incident, never as PASS or as an artefact finding.
- [ ] Use the configured timeout, defaulting to five minutes when absent.
- [ ] Do not embed metered audit invocations in regression tests, CI, builds,
  or scheduled automation.

### 7. Replace project configuration

- [ ] Use `~/.agents/sdlc.yaml` for non-secret global defaults.
- [ ] Use `.sdlc/project.yaml` for tracked project facts and explicit local
  overrides.
- [ ] Implement precedence as command line, process environment, project YAML,
  global YAML, then schema default.
- [ ] Drive prompts, choices, validation, and persistence from one YAML schema.
- [ ] Show a global default when present and allow a project-local alternative;
  an empty answer inherits without copying the global value.
- [ ] Never permit a global default for whether a project is greenfield or
  brownfield.
- [ ] Migrate only allowlisted SDLC keys from `.env` through a shell wrapper.
  Never read or expose unrelated `.env` values.
- [ ] Preserve unrelated `.env` content and remove migrated SDLC keys only
  after the YAML write has succeeded.
- [ ] Keep secrets out of both YAML files.

Candidate global configuration:

```yaml
schema: 1

models:
  definition: gpt-5.6-sol
  build: gpt-5.6-terra
  audit: gpt-5.6-luna

audit:
  timeout: 5m

git:
  branch_strategy: current
```

Candidate project configuration:

```yaml
schema: 1

project:
  type: brownfield
  initialized_from: sdlc-v2
  initialized_with: v3.0.0

technologies:
  - GO
  - SHELL

authorities:
  product:
    - docs/VISION.md
  architecture:
    - docs/architecture.md
  requirements:
    - docs/ACs.org

infrastructure:
  role: consumer
  owner: Example infrastructure project
  contract: /path/to/integration-contract.md

git:
  primary_branch: master

migration:
  archive_branch: sdlc_v2_state_YYYY-MM-DD
  migrated_at: YYYY-MM-DD
```

Inherited defaults and empty sections are omitted from project YAML.

### 8. Rewrite `sdlc-project-init`

- [ ] Run exactly once. Refuse when `.sdlc/project.yaml` already exists.
- [ ] Verify the project is a Git repository and inspect its working tree
  before changing tracked files.
- [ ] Identify the current primary branch without assuming `master` or `main`.
- [ ] Detect new, SDLC v1, or SDLC v2 project state deterministically.
- [ ] Create a dated archive branch containing the exact pre-migration state
  and ask whether to push that branch to `origin`.
- [ ] Create a migration branch from the primary branch.
- [ ] For an eligible GitHub-backed v1 project, offer the ticket-migration
  skill before continuing.
- [ ] Ask schema-driven project questions and write `.sdlc/project.yaml`.
- [ ] Create `docs/work.org` and the specification directory structure.
- [ ] Perform the applicable v1 or v2 migration, or initialize a new project.
- [ ] Validate and commit the migration as coherent checkpoints rather than a
  per-file or per-ticket commit loop.
- [ ] Show the resulting variance and ask whether to merge the migration branch
  into the original primary branch.
- [ ] Do not copy skills into the project.
- [ ] Do not launch an agent merely to initialize an empty project.
- [ ] Do not implement `sdlc-project-update` for v3.

### 9. Migrate SDLC v1 projects

- [ ] Run the existing fast pre-migration workflow when the operator accepts
  it.
- [ ] Run `make test` once at the beginning and stop if it fails.
- [ ] Archive every GitHub ticket, its comments, status, and commit references
  under `docs/archive/migrated-tickets/`.
- [ ] Reconcile the historical AC ledger as `docs/ACs.org` using the canonical
  Org hierarchy and without losing information.
- [ ] Use live regression tests and existing AC entries as delivery evidence.
- [ ] Apply the agreed delivery heuristics and mark heuristic-only conclusions
  visibly.
- [ ] Write unresolved defects and undelivered features to `docs/work.org`.
- [ ] Close all migrated GitHub tickets only after the local archive and
  migration artefacts have been committed and the operator has authorized the
  closure batch.
- [ ] Update stale project documentation in batches. Inspect code only when
  documentation and evidence cannot resolve a genuine contradiction.
- [ ] Archive the obsolete implementation plan.

### 10. Migrate SDLC v2 projects

- [ ] Preserve `.specify/`, constitutions, specifications, plans, tasks,
  audits, and validation records on the archive branch.
- [ ] Copy incomplete v2 work unchanged into a documented archive location on
  the migration branch and index each item in `docs/work.org`.
- [ ] Do not convert or elaborate dormant incomplete work during migration.
- [ ] When the operator later resumes one item, convert only that item to the
  unified `spec.org` without reusing or renumbering its identifier.
- [ ] Carry forward reliable audit and operator-approval evidence where its
  meaning maps cleanly to a v3 gate.
- [ ] Extract genuine project-wide facts from a constitution into
  `.sdlc/project.yaml` or an existing authoritative project document.
- [ ] Remove active `.specify/` infrastructure and project-local Spec Kit
  skills from the migrated branch.

### 11. Rewrite installation and deployment

- [ ] Keep `~/.agents/sdlc` as the canonical installed standards and workflow
  root.
- [ ] Install v3 skills globally under `~/.agents/skills`.
- [ ] Install the supported Codex adapters and v3 CLI tools.
- [ ] Stop installing project-local skills or provider-specific copies of the
  SDLC corpus.
- [ ] On the first v3 install, back up and retire every public SDLC-managed v1
  and v2 command, prompt, skill, and hook from Claude, Hermes, and Copilot.
- [ ] Remove the SDLC-managed restrictions or hooks previously injected into
  unsupported harness configuration only where ownership markers make that
  safe. Do not alter private configuration owned by the separate agents repo.
- [ ] Retire Spec Kit preset material and obsolete v2 command-line tools.
- [ ] Preserve concise variance-only output, `VERBOSE=1` inventory output,
  `y`/`yes` confirmation, and no prompt on an idempotent rerun.
- [ ] Build and install the public `sdlc-preview` utility.

### 12. Provide document preview

- [ ] Accept one Markdown or Org file.
- [ ] Render through Pandoc to a collision-resistant random HTML filename in
  the source document's directory so document-relative links continue to work.
- [ ] Open the result in the platform browser without making preview an
  approval gate.
- [ ] Remove the generated HTML one second after handing it to the browser.
  Schedule cleanup asynchronously or otherwise ensure deletion cannot happen
  before the browser has read the file.
- [ ] Remove stale generated output after failures as well as successful opens.
- [ ] Never overwrite a pre-existing file when allocating the preview name.
- [ ] Work on macOS and Linux, and emit a clear unsupported-platform message
  elsewhere.
- [ ] Report a clear dependency error when Pandoc is unavailable.
- [ ] Do not depend on the operator's private scripts, YAML, shell functions,
  or templates.
- [ ] Begin with a basic bundled HTML and CSS template. A higher-quality
  redistributable community template may replace it later after its licence and
  provenance have been reviewed.
- [ ] Add a concise `src/ORGMODE.md` primer and route it only when an agent
  reads or writes Org artefacts. It must teach the semantic model rather than
  merely list punctuation:
  - headings are addressable nodes in the document hierarchy;
  - lists are subordinate content within a node;
  - checkboxes track finite list items, while TODO keywords track a node's
    lifecycle;
  - tags express cross-cutting classifications;
  - properties express structured attributes;
  - typed blocks distinguish source, literal output, and quotations; and
  - drawers hold secondary or metadata material, not first-class requirements.
- [ ] Require meaningful outline hierarchy and folding. Prefer no more than
  three or four heading levels; restructure or split a document rather than
  creating a deep star ladder. Use `#+STARTUP: overview` or `content` for the
  file view. A dense supporting subtree not intended for casual inspection may
  use `:VISIBILITY: folded`; drawers may hold subordinate detail that should
  remain available but normally collapsed.
- [ ] Warn that folding is a view over explicit structure. Required information
  must remain understandable in plain text and rendered HTML, and must not be
  hidden merely because it is visually inconvenient.
- [ ] Cover internal and file links, bold text, =verbatim=, ~code~,
  description lists, property drawers, named source blocks, example blocks,
  and quote blocks. Explicitly prohibit Markdown backticks in Org even though
  Pandoc accepts them.

### 13. Update public documentation

- [ ] Rewrite the README and quickstart around v3 rather than Spec Kit.
- [ ] Document normal, paired, and emergency delivery with concrete command
  examples.
- [ ] Document new, v1, and v2 project initialization and migration.
- [ ] Explain Org viewing through Emacs and `sdlc-preview`.
- [ ] Explain global versus project configuration and precedence.
- [ ] Update architecture, changelog, and learnings.
- [ ] State clearly that v3 initially supports Codex and OpenAI models only.
- [ ] Remove the Spec Kit prerequisite and obsolete v2 operational guidance.

### 14. Verify proportionately

- [ ] Unit-test deterministic schema parsing, precedence, rendering, migration
  classification, identifier allocation, installer planning, and retirement
  ownership.
- [ ] Do not add documentation-grep tests or metered agent calls to the
  regression suite.
- [ ] Exercise installer idempotence with isolated homes.
- [ ] Exercise initialization for a blank project.
- [ ] Exercise migration fixtures representing SDLC v1 and v2 states.
- [ ] Verify Org templates parse with Pandoc and retain stable internal links.
- [ ] Run the first real project migration paired with the operator.
- [ ] Deliver one small ordinary change through both v3 gates before tagging
  v3.0.0.

## Candidate artefact contracts

### Audit evidence

`specs/NNN-descriptor/audits.org` is append-only gate evidence. Each run records:

- gate name and attempt number;
- authoring or external context;
- exact artefact revision or digest;
- each component audit verdict;
- combined verdict;
- findings and required remediation;
- runner incidents separately from findings; and
- the later run that supersedes a failed result.

### Validation evidence

`specs/NNN-descriptor/validation.org` exists only when evidence needs a durable
record. It records:

- linked test identifier and acceptance criterion;
- RT, UT, or OT classification;
- exact action or command where appropriate;
- expected and observed result;
- date and operator for human validation;
- PASS, FAIL, or BLOCKED status; and
- limitations, retained artefacts, and any metered-operation boundary.

## Decisions still required before implementation

### Identifier namespace

**Recommendation:** use one monotonically increasing project-wide work number.
Represent work as `W001`, acceptance criteria as `AC001.1`, and tests as
`RT001.1`, `UT001.1`, or `OT001.1`. Never reuse a number after work is
abandoned or archived. A migrated project starts above its highest historic
ticket, specification, or work number.

**Reason:** one namespace avoids collisions while retaining familiar v1
traceability. This must be approved before the templates become canonical.

### SDLC v2 archive location

**Recommendation:** retain the untouched full state on the archive branch, and
copy only incomplete work needed for local discoverability under
`docs/archive/sdlc-v2/`. Do not duplicate completed Spec Kit machinery on the
migrated branch.

**Decision needed:** whether completed v2 feature artefacts should also be
copied into that directory or remain available only through Git history and
the archive branch.

### Dirty project migration

**Recommendation:** initialization may inspect a dirty tree but must stop before
branching if changes cannot be attributed to an existing operator checkpoint.
It must never auto-commit or stash unknown work.

**Decision needed:** whether the initializer may offer to let the operator
create that checkpoint and rerun, or should only report the blocker.

### External auditor retention

**Unverified:** the exact Codex CLI mechanism for creating and resuming one
bounded auditor context has not yet been established. Verify its supported
interface before implementing either gate runner. Do not emulate persistence by
reloading a fresh agent under the same label.

### Gate command surface

**Recommendation:** expose four human-facing global skills: define, definition
gate, deliver, and implementation gate. Keep the four focused audits as
internal criteria and optional diagnostics.

**Decision needed:** final concise skill names and whether ordinary delivery is
one continued invocation or two explicit operator invocations separated by
definition sign-off.

### `work.org` ownership

**Recommendation:** `work.org` records work identity, status, priority, source,
and links. It must not duplicate acceptance criteria, test definitions, or
solution design from `spec.org`.

**Decision needed:** whether closed items remain in `work.org` indefinitely or
move periodically to a dated archive.

### Configuration migration

**Decision needed:** identify the exact allowlist mapping from every historical
`SDLC_*` `.env` key to the v3 YAML schema, including keys that should be retired
rather than migrated.

### Unsupported-provider cleanup

**Risk:** public SDLC code and the private agents installer have both modified
provider configuration over time. Cleanup must be based on exact ownership
markers and deployed-path inventories. A broad provider-directory cleanup could
remove private configuration.

### First paired migration

**Decision needed:** choose the first representative project after the v3 code
and isolated fixtures pass. The migration itself will be performed with the
operator watching, and its discrepancies will update the migration rules before
v3.0.0 is tagged.

## Definition of handover readiness

Implementation may begin when:

- the two Org templates are accepted, as recorded on 2026-09-06;
- the identifier namespace is accepted;
- the v2 archive scope is accepted;
- the gate skill names and invocation boundary are accepted;
- the Codex auditor-resumption mechanism has been verified; and
- the current v2 working state has a named recoverable branch strategy.
