# SDLC v3 Implementation Plan

Status: Accepted implementation handover. Implementation is active on the
`sdlc-v3` branch; release remains conditional on validation and the first paired
project migration.

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

- [X] Record the current branch, remotes, tags, and dirty paths.
- [X] Preserve the complete current SDLC v2 state on an archive branch before
  removing or replacing its unfinished work.
- [X] Keep the existing v3 design commits and this review pack reachable from
  the implementation branch.
- [X] Do not stage, overwrite, or discard unrelated working-tree changes.

### 2. Establish the v3 source layout

- [X] Keep universal standards under `src/` and technology standards under
  `src/technologies/`.
- [X] Add canonical installed templates for `spec.org`, `work.org`, audit
  evidence, and validation evidence.
- [X] Add a schema for `~/.agents/sdlc.yaml` and `.sdlc/project.yaml`.
- [X] Remove active Spec Kit presets, adapters, commands, templates, and
  references from the v3 source tree.
- [X] Archive any v2 material retained for historical explanation rather than
  leaving it in active routing.

### 3. Recover the authoritative SDLC v1 definitions

- [X] Treat the latest SDLC v1 release tag, currently `v1.0.3`, as the source
  evidence for concepts deliberately resurrected in v3.
- [X] Extract the acceptance-criterion definition and AC/test boundary from
  `v1.0.3:ISSUES.md`.
- [X] Extract RT, UT, and OT classification, automated TDD, real-user boundary,
  one-off lifecycle, and user-test ownership from `v1.0.3:TESTING.md`.
- [X] Review `v1.0.3:MAIN.md` and the relevant v1 skills for paired delivery,
  emergency delivery, identifier allocation, approval ownership, and useful
  handoff language.
- [X] Write a short provenance map naming each resurrected definition and its
  v1 source before rewriting the v3 instructions.
- [X] Do not restore v1 gate ceremony, ticket workflow, modes, or terminology
  merely because they occur beside a useful definition. Only the concepts
  explicitly selected for v3 return.

### 4. Rewrite routing and workflow documents

- [X] Make `MAIN.md` route only the standards and workflow needed for the
  current activity.
- [X] Define the standard two-phase workflow:
  definition, definition gate, operator sign-off, TDD delivery,
  implementation gate, validation, operator closure.
- [X] Retain the paired-development route with live user validation and
  change-scoped evidence.
- [X] Retain the exact `BYPASS-GATE-7` emergency route. The agent must never
  infer or invoke it.
- [X] Require every specification to pass the context-independent handoff
  test before definition sign-off.
- [X] Resurrect the SDLC v1 RT, UT, and OT test classifications and its
  automated TDD rule. Do not present these as inherited from Spec Kit or as new
  SDLC v3 terminology.
- [X] Preserve documentation, Git, security, and technology standards.
- [X] Remove v2-only terms such as constitution stages, Spec Kit commands,
  generated plans, tasks, and convergence.

### 5. Create or revise global skills

All v3 skills are installed globally under `~/.agents/skills`. Each skill must
bootstrap from the current global SDLC and must not depend on a project-local
copy of itself.

- [X] **Define change:** ask no more than three concise free-form questions per
  unresolved area, then create or revise the single `spec.org` artefact.
- [X] **Definition gate:** apply specification, design, and test-definition
  audits together. Run up to five local remediation rounds, then all three
  audits in one retained external context.
- [X] **Deliver change:** begin from a signed-off specification, write
  automated tests first, observe RED, implement, and observe GREEN.
- [X] **Implementation gate:** apply test and code audits together. Run up to
  five local remediation rounds, then both audits in one retained external
  context.
- [X] **Pair change:** retain the agreed paired-development behaviour and
  consolidate live decisions and validations into the durable artefacts.
- [X] **Emergency change:** implement only after the exact operator keyword,
  then backfill specification, design, documentation, and evidence.
- [X] **Migrate legacy tickets:** adapt the existing skill to prepare a v1
  project for v3 without requiring Spec Kit.
- [X] **Focused audits:** retain `audit-spec`, `audit-design`, `audit-tests`,
  and `audit-code` as criteria and optional diagnostic skills. Their direct
  invocation must not accidentally create additional auditor contexts.
- [X] Review `diagnose-issue`, `recommendations-please`, `summarize-issues`,
  and `useful-be` for v3 terminology and routing.
- [X] Retire `audit-definition` and `audit-implementation` implementations
  that encode the abandoned multi-harness or Spec Kit workflow before
  replacing them with the v3 gate skills.
- [X] Retain `convert-migrated-acs-to-org` as an explicit repair path for v1
  projects whose ticket migration is complete but whose ledger still uses
  `docs/ACs.md`; ordinary v1 initialization performs the conversion itself.

### 6. Implement the two gate runners

- [X] Keep local self-audit in the authoring context. It must not launch a
  script or another agent.
- [X] Launch one external Codex auditor only after every local component audit
  passes.
- [X] Pass all component audit instructions and the bounded artefact/context
  set to that one auditor.
- [X] Resume the same external auditor context after remediation.
- [X] Enforce a maximum of five local rounds and five external rounds per
  gate.
- [X] Record the artefact revision, each component verdict, combined verdict,
  findings, remediation, and superseding rerun in `audits.org`.
- [X] Treat malformed output, timeout, or harness failure as an audit-runner
  incident, never as PASS or as an artefact finding.
- [X] Use the configured timeout, defaulting to five minutes when absent.
- [X] Do not embed metered audit invocations in regression tests, CI, builds,
  or scheduled automation.

### 7. Replace project configuration

- [X] Use `~/.agents/sdlc.yaml` for non-secret global defaults.
- [X] Use `.sdlc/project.yaml` for tracked project facts and explicit local
  overrides.
- [X] Implement precedence as command line, process environment, project YAML,
  global YAML, then schema default.
- [X] Drive prompts, choices, validation, and persistence from one YAML schema.
- [X] Show a global default when present and allow a project-local alternative;
  an empty answer inherits without copying the global value.
- [X] Never permit a global default for whether a project is greenfield or
  brownfield.
- [X] Migrate only allowlisted SDLC keys from `.env` through a shell wrapper.
  Never read or expose unrelated `.env` values.
- [X] Preserve unrelated `.env` content and remove migrated SDLC keys only
  after the YAML write has succeeded.
- [X] Keep secrets out of both YAML files.

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
    - docs/work.org
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

- [X] Run exactly once. Refuse when `.sdlc/project.yaml` already exists.
- [X] Verify the project is a Git repository and inspect its working tree
  before changing tracked files.
- [X] Identify the current primary branch without assuming `master` or `main`.
- [X] Detect new, SDLC v1, or SDLC v2 project state deterministically.
- [X] Create a dated archive branch containing the exact pre-migration state
  and ask whether to push that branch to `origin`.
- [X] Create a migration branch from the primary branch.
- [X] For an eligible GitHub-backed v1 project, offer the ticket-migration
  skill before continuing.
- [X] Ask schema-driven non-authority project questions before migration.
- [X] After migration, derive README, vision, and architecture candidates from
  the bounded Git inventory. Let the operator toggle multiple paths and rescan
  after moving a misplaced document.
- [X] Record `docs/work.org` and the sole applicable AC ledger as deterministic
  requirement authorities. Reject simultaneous `docs/ACs.org` and
  `docs/ACs.md` ledgers.
- [X] Only for v2 migrations with archived specifications, invoke headless
  Codex with the configured audit model and an exact bounded path list. Accept
  structured disposition evidence only; do not ask the model to select
  authorities or edit the work ledger.
- [X] Keep the proposal in the project-owned `.sdlc/.init/` directory whose
  presence identifies an interrupted initialization; remove it before staging
  the completed migration. Track `.sdlc/.gitignore` with `.init/` excluded.
  Never use private `.git` paths as application storage.
- [X] Create `docs/work.org` from its canonical template before discovery. For
  v2 migration, validate complete classification coverage and render every
  archived specification into the appropriate existing work-ledger section.
  The definition skill creates each new v3 specification directory only when
  that work begins.
- [X] Perform the applicable v1 or v2 migration, or initialize a new project.
- [X] Validate and commit the migration as coherent checkpoints rather than a
  per-file or per-ticket commit loop.
- [X] Show the resulting variance and ask whether to merge the migration branch
  into the original primary branch.
- [X] Do not copy skills into the project.
- [X] Do not launch an agent merely to initialize an empty project.
- [X] Do not implement `sdlc-project-update` for v3.

### 9. Migrate SDLC v1 projects

- [X] Run the existing fast pre-migration workflow when the operator accepts
  it. Declining skips that optional workflow without stopping initialization.
- [X] Run `make test` once at the beginning and stop if it fails.
- [X] Archive every GitHub ticket, its comments, status, and commit references
  under `docs/archive/migrated-tickets/`.
- [X] Reconcile the historical AC ledger as `docs/ACs.org` using the canonical
  Org hierarchy and without losing information.
- [X] Use live regression tests and existing AC entries as delivery evidence.
- [X] Apply the agreed delivery heuristics and mark heuristic-only conclusions
  visibly.
- [X] Write unresolved defects and undelivered features to `docs/work.org`.
- [X] Close all migrated GitHub tickets only after the local archive and
  migration artefacts have been committed and the operator has authorized the
  closure batch.
- [X] Update stale project documentation in batches. Inspect code only when
  documentation and evidence cannot resolve a genuine contradiction.
- [X] Archive the obsolete implementation plan.

### 10. Migrate SDLC v2 projects

- [X] Preserve `.specify/`, constitutions, specifications, plans, tasks,
  audits, and validation records on the archive branch.
- [X] Move v2 work unchanged into a documented archive location on the
  migration branch and index each item in `docs/work.org` as delivered,
  approved but undelivered, abandoned, or unresolved.
- [X] Do not convert or elaborate dormant incomplete work during migration.
- [X] When the operator later resumes one item, convert only that item to the
  unified `spec.org` without reusing or renumbering its identifier.
- [X] Preserve reliable audit and operator-approval evidence unchanged in the
  v2 archive. Use explicit operator approval or operator-authorized delivery
  for migration classification; an audit PASS alone is not approval.
- [X] Prompt the operator to record genuine project-wide authorities and
  infrastructure boundaries in `.sdlc/project.yaml`; preserve the constitution
  itself in the v2 archive.
- [X] Remove active `.specify/` infrastructure and project-local Spec Kit
  skills from the migrated branch.

### 11. Rewrite installation and deployment

- [X] Keep `~/.agents/sdlc` as the canonical installed standards and workflow
  root.
- [X] Install v3 skills globally under `~/.agents/skills`.
- [X] Install the supported Codex adapters and v3 CLI tools.
- [X] Stop installing project-local skills or provider-specific copies of the
  SDLC corpus.
- [X] On the first v3 install, back up and retire every public SDLC-managed v1
  and v2 command, prompt, skill, and hook from Claude, Hermes, and Copilot.
- [X] Remove the SDLC-managed restrictions or hooks previously injected into
  unsupported harness configuration only where ownership markers make that
  safe. Do not alter private configuration owned by the separate agents repo.
- [X] Retire Spec Kit preset material and obsolete v2 command-line tools.
- [X] Preserve concise variance-only output, `VERBOSE=1` inventory output,
  `y`/`yes` confirmation, and no prompt on an idempotent rerun.
- [X] Build and install the public `sdlc-preview` utility.

### 12. Provide document preview

- [X] Accept one Markdown or Org file.
- [X] Render through Pandoc to a collision-resistant random HTML filename in
  the source document's directory so document-relative links continue to work.
- [X] Open the result in the platform browser without making preview an
  approval gate.
- [X] Remove the generated HTML one second after handing it to the browser.
  Schedule cleanup asynchronously or otherwise ensure deletion cannot happen
  before the browser has read the file.
- [X] Remove stale generated output after failures as well as successful opens.
- [X] Never overwrite a pre-existing file when allocating the preview name.
- [X] Work on macOS and Linux, and emit a clear unsupported-platform message
  elsewhere.
- [X] Report a clear dependency error when Pandoc is unavailable.
- [X] Do not depend on the operator's private scripts, YAML, shell functions,
  or templates.
- [X] Begin with a basic bundled HTML and CSS template. A higher-quality
  redistributable community template may replace it later after its licence and
  provenance have been reviewed.
- [X] Add a concise `src/ORGMODE.md` primer and route it only when an agent
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
- [X] Require meaningful outline hierarchy and folding. Prefer no more than
  three or four heading levels; restructure or split a document rather than
  creating a deep star ladder. Use `#+STARTUP: overview` or `content` for the
  file view. A dense supporting subtree not intended for casual inspection may
  use `:VISIBILITY: folded`; drawers may hold subordinate detail that should
  remain available but normally collapsed.
- [X] Warn that folding is a view over explicit structure. Required information
  must remain understandable in plain text and rendered HTML, and must not be
  hidden merely because it is visually inconvenient.
- [X] Cover internal and file links, bold text, =verbatim=, ~code~,
  description lists, property drawers, named source blocks, example blocks,
  and quote blocks. Explicitly prohibit Markdown backticks in Org even though
  Pandoc accepts them.

### 13. Update public documentation

- [X] Rewrite the README and quickstart around v3 rather than Spec Kit.
- [X] Document normal, paired, and emergency delivery with concrete command
  examples.
- [X] Document new, v1, and v2 project initialization and migration.
- [X] Explain Org viewing through Emacs and `sdlc-preview`.
- [X] Explain global versus project configuration and precedence.
- [X] Update architecture, changelog, and learnings.
- [X] State clearly that v3 initially supports Codex and OpenAI models only.
- [X] Remove the Spec Kit prerequisite and obsolete v2 operational guidance.

### 14. Verify proportionately

- [X] Unit-test deterministic schema parsing, precedence, rendering, migration
  classification, identifier allocation, installer planning, and retirement
  ownership.
- [X] Do not add documentation-grep tests or metered agent calls to the
  regression suite.
- [X] Exercise installer idempotence with isolated homes.
- [X] Exercise initialization for a blank project.
- [X] Exercise migration fixtures representing SDLC v1 and v2 states.
- [X] Verify Org templates parse with Pandoc and retain stable internal links.
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

## Resolved decisions

- Project-wide work, AC, and test identifiers are monotonically increasing and
  never reused. Migration allocates above every historical identifier and keeps
  original v2 identifiers as provenance.
- The exact pre-migration state remains on the dated archive branch. Active v2
  `.specify`, `specs`, and exact Spec Kit integrations move unchanged under
  `docs/archive/sdlc-v2/`; unrelated project integrations remain active.
- Initialization requires a clean worktree and reports the blocker rather than
  stashing or committing unknown work.
- The global skills are `define-change`, `audit-definition`, `deliver-change`,
  `audit-implementation`, `pair-change`, and `emergency-change`. The focused
  audits remain available as criteria and diagnostics.
- `work.org` carries status and links, not duplicated requirements or design.
- Historical `.env` migration is schema-allowlisted and removes only migrated
  SDLC keys after the YAML profile is safely written.
- Unsupported-provider cleanup uses exact owned paths and hook signatures.

## Remaining release checks

- Perform the first representative project migration with the operator
  watching.
- Deliver one small ordinary change through both v3 gates.
- Reconcile any discrepancy found by those trials before tagging v3.0.0.

## Validation evidence before live installation

- The complete Go test suite, `go vet`, `golangci-lint`, ShellCheck, and shfmt
  pass against the implementation commit.
- Every v3 and migration Org template parses successfully with Pandoc, and the
  specification template's internal links resolve to declared custom IDs.
- The built installer was exercised against an isolated user home. Its first
  interactive run installed the v3 corpus and Codex adapter; its unchanged
  second run reported only that all detected copies were current.
- The built initializer was exercised against an isolated blank Git project.
  It created the archive branch, migration branch, project profile, work ledger,
  and coherent migration commit.
- The built initializer was exercised against an isolated v2 project. It moved
  active Spec Kit state unchanged into the project archive, removed the active
  paths, retained the original ID as provenance, and allocated the v3 work ID
  above the historical AC number.
- An isolated v1 fixture verifies the GitHub and Codex boundaries without
  making a metered call. The first real v1 migration remains the paired release
  trial.
