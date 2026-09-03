# Acceptance Criteria

## Public SDLC deployment

The copy-based deployment supersedes the former common-live-tree/provider-link
adapter model.

### Detection and confirmation

- `make install` creates the common `~/.agents` home when it is absent and
  detects existing `~/.claude`, `~/.codex`, `~/.copilot`, and `~/.hermes`
  provider homes.
- It preflights every SDLC-owned destination before writing.
- One `y` or `yes`, matched case-insensitively, applies the complete detected
  batch; one refusal writes nothing.
- Explicit `--agent` mode limits provider deployment and retains deliberate
  `--apply` and `--configure` operations.

### Canonical runtime and provider adapters

- The canonical runtime exists only at `~/.agents/sdlc`; provider homes do not
  receive alternate SDLC trees.
- Files discovered recursively beneath `src/` deploy directly to the canonical
  root. Files beneath `skills/` and `hooks/` retain those directory names.
- README, changelog, learnings, project documentation, templates, installer
  source, tests, and build metadata are not runtime deployment candidates.
- Every top-level SDLC skill is discovered from `skills/` and recursively
  copied to each supported skill home.
- A branch that contains the optional legacy `commands/` source copies its
  commands to Claude `commands/` and Codex/Copilot `prompts-commands/`.
- No SDLC-owned deployed destination is a symlink.
- The seven retired workflow commands, retired drafting or design skills,
  `audit-acs`, root-level technology documents, and obsolete constitution
  addendum are backed up and removed from active discovery when their source
  paths are absent.
- All other destination-only content survives deployment.

### Shared ownership

- Every differing existing artefact, including stale symlinks, wrong types,
  and provider configuration, is renamed to `<path>.<epoch>.bak` before the
  canonical replacement is copied.
- A failed backup stops replacement of that artefact.
- The installer never claims an entire agent, skill, or command directory.
- Deployment decisions are made per managed file. Repository-level changes
  outside the runtime roots do not create deployment variances.
- Legacy retirement decisions are limited to the explicit migration list and
  never infer ownership from an arbitrary destination-only path.
- Hermes configuration changes own only the SDLC command-guard hook and
  preserve private instructions and unrelated configuration.
- Claude, Codex, Copilot, and Hermes provider adapters install a pre-tool guard
  that denies direct reads of any path whose exact basename is `.env`.
- The guard denies native file-read and content-search requests and direct shell
  references to `.env`; names such as `.env.example` and `.env.local` remain
  outside the rule.
- Provider configuration changes preserve unrelated settings and hooks. A
  repeated installation reports no variance and requires no confirmation.

## Spec Kit project initialization

- The project initializer discovers Markdown technology standards
  alphabetically from `~/.agents/sdlc/technologies/` without a hard-coded
  technology list.
- Universal specification, coding, testing, security, independent-audit,
  paired-development, documentation, Git, and entry-point standards are
  included exactly once.
- Configuration precedence is CLI, process environment, project `.env`, user
  `.env`, then a schema-declared fallback.
- A YAML schema defines configuration fields, CLI flags, types, choices, prompt
  order, user-default eligibility, persistence, conditional requirements, and
  phase provider applicability. Fixed unresolved choices use numbered prompts;
  resolved fields are not prompted.
- Bash evaluates user and project `.env` files. Only schema-listed SDLC values
  cross the wrapper boundary; Go does not interpret shell expressions.
- Specification, build, and audit harnesses independently fall back to the
  default agent harness. The specification harness controls Spec Kit integration
  and constitution generation.
- Infrastructure relationship is recorded as `none`, `consumer`, or `provider`.
  Consumer projects comply with an externally owned contract; provider projects
  define, implement, evolve, and honour the contract they publish. Selected
  relationships record an owner descriptor and integration-contract path
  without assuming a private infrastructure project.
- Rendering is deterministic. When the rendered baseline and selections are
  current, the initializer asks nothing, writes nothing, and launches no agent.
- The baseline is written as the project `constitution-template` override, not
  over Spec Kit's core fallback template, and resolves without parsing preset
  manifests.
- A changed baseline is written atomically before the selected Codex, Claude,
  or Hermes harness receives the constitution-only semantic prompt.
- After the harness exits, the initializer verifies against the generated
  scaffold that the complete Engineering Standards, Specification and Evidence,
  and Mandatory Independent Audits sections remain present. Line wrapping may
  change; omitted, summarized, or weakened shared governance fails
  initialization.
- Initial ratification uses `UNRATIFIED -> 1.0.0`, removes the generated-scaffold
  marker and resolved blockers, confirms the adopted SDLC revision, records the
  ratification and last-revised dates, and removes draft-only qualifications.
- A brownfield baseline distinguishes the centralized legacy requirement
  authority, the migration disposition and historical context, current design
  authority, and the maintained regression test pack with its supported command
  and requirement traceability. A completed migration makes `docs/ACs.org` the
  sole legacy-process requirement authority; live tickets, comments, the
  migration index, and archived snapshots remain disposition, provenance, and
  rationale rather than requirement or AC authority.
- Brownfield pre-migration archives every open and closed ticket body and all
  comments under `docs/archive/migrated-tickets/`, with a manifest. It reads
  only that tracked snapshot after download and does not re-fetch individual
  tickets.
- Its repository deliverables are a current `docs/ACs.org`, a concise
  `docs/ticket-migration.org` index, the complete ticket archive, targeted stale
  documentation corrections, and unchanged archival of the single file matching
  `docs/implementation?plan.md` when present, preserving its basename.
- It creates `docs/ticket-migration.org` immediately after archive verification
  from the canonical deployed Org template and updates it after each evidence
  phase and ticket classification. Report sections, ticket records, detail
  categories, and individual ACs use nested Org headings for operator folding;
  bullets are reserved for genuinely flat summaries. The template also defines
  native Org syntax and Pandoc parse validation. This durable state records the
  suite result, RT map, completed tickets, next ticket, assumptions,
  dispositions, and closure result.
- It runs the supported whole-suite command at most once and never reruns or
  diagnoses historical individual tests. A maintained regression test plus a
  current passing suite may correct a legacy pending or unverified status.
- After ticket-led reconciliation, it performs one reverse checksum over every
  RT still present in the maintained harness. This is the first analytical pass:
  every maintained RT is reviewed and mapped to its ticket and AC before ticket
  classification begins. Each live RT must resolve to at least one AC in
  `docs/ACs.org`, and every cited AC and RT relationship must be present in the
  ledger. Missing AC wording normally comes from the archived ticket record.
  When a ticket has no recorded passing evidence but retains at least one RT in
  the passing harness, the whole ticket is assumed delivered; inferred AC and
  test statuses are footnoted. Orphan live RTs automatically receive minimal
  evidence-bounded ACs under one purpose-created pre-migration baseline ticket.
- A failed supported whole-suite run stops migration after a recoverable
  checkpoint of the verified archive, migration index, and format-only Org
  conversion. No ticket classification, semantic ledger reconciliation, other
  documentation change, baseline-ticket creation, or ticket closure follows.
- When all ticket tests pass except one or two unrecorded UTs or OTs, and no
  contrary evidence exists, the ticket is automatically classified as
  delivered. An AC supported only by such an assumed pass, and its delivered
  ticket entry, carry a migration-heuristic footnote distinguishing inference
  from contemporaneous evidence.
- When no delivery evidence remains after the evidence pass, ticket state
  determines only the disposition: closed scope is abandoned and open scope is
  undelivered. Ticket-linked code commits are delivery evidence unless the
  archived record limits or contradicts them.
- The ledger normally predates migration as `docs/ACs.md`. Before any additions,
  the skill losslessly converts it to `docs/ACs.org` using the canonical deployed
  template and removes the Markdown source. Every identifier, requirement,
  provenance field, status, test relationship, evidence note, supersession,
  footnote, glossary entry, and explanatory note is preserved. An AC already
  present is delivery evidence even when its ticket has no recorded test result.
  Missing ACs are inserted at their correct identifier positions; existing
  identifiers are never renumbered.
- Every written `docs/ACs.org` is cross-checked against the canonical outline.
  Each AC uses a level-three heading containing both its identifier and a useful
  descriptor; identifier-only AC headings fail structural validation.
- The final documentation reconciliation corrects unambiguous stale migration
  and authority statements in active project instructions and documentation,
  including a project-root `AGENTS.md`, while leaving archived material intact.
- The explicit `convert-migrated-acs-to-org` repair path runs only when the
  ticket archive and completed migration index exist, GitHub reports zero open
  issues, `docs/ACs.md` exists, and `docs/ACs.org` does not. It preserves every
  ledger field, updates current references, leaves archived tickets unchanged,
  and commits the tracked rename only after Org parsing and field reconciliation.
- The Org index records open defects, defined but undelivered features, any
  unresolved classifications, and a final simple list of delivered tickets.
  Local archive links make this history usable without GitHub.
- It commits the archive and reconciled project state in a batch before closing
  every issue that was open in the snapshot. It closes without comments,
  records closure failures, minimizes operator questions, and never enters a
  per-ticket commit loop.
- `sdlc-project-init` does not archive or semantically revise project
  documentation. A migrated brownfield project must contain `docs/ACs.org` and
  no `docs/ACs.md`; migration artefacts without the Org ledger, or both formats
  together, stop initialization with an actionable diagnostic.

## Application vulnerability checking

- Every deployable application exposes a repository-owned `make vulncheck`
  security gate separate from `make test`.
- The target scans all application-managed dependency graphs that affect the
  built or deployed artefact and aggregates every selected stack.
- A missing scanner, stale or uncovered manifest, unavailable advisory source,
  or incomplete scan fails the gate.
- The gate never installs tools, resolves or updates dependencies, modifies
  locks, applies fixes, suppresses exit codes, or creates exceptions.
- Findings block by default. Every exception is exact, durable,
  human-authorized, evidenced, and time-bounded.
- Stack standards prefer `govulncheck` for Go, `pip-audit` for Python, the
  selected package manager's audit for Node.js, `cpan-audit` for Perl, and Trivy
  for Swift, with Trivy or OSV-Scanner available as documented fallbacks.
- External infrastructure may invoke the application target but retains its own
  host, operating-system, service, container, and package-closure controls.

## Standard Make interface

- Projects using Make expose canonical target names for each applicable
  operation: `build`, `lint`, `test`, `vulncheck`, `install`, `sync`, and
  `deploy`.
- Every Make-based Git software project provides `test` and `sync`; applicability
  determines the other targets without permitting successful no-ops.
- `make sync` stages the complete worktree, creates a commit when staged changes
  exist, pulls, and pushes, in that order. It uses a short `COMMIT_MESSAGE`
  value, defaulting to `chore: sync`, and does not infer or demand a long
  message.
- `make deploy` runs the complete regression suite and the application
  vulnerability gate before deployment.
- `SKIP_TESTS=1 make deploy` skips only the regression suite, reports the skip,
  and does not skip vulnerability checking.
- `VERBOSE=1` consistently requests full output. Normal output remains concise.
- Targets stop on failure and do not silently add rebasing, stashing, branch
  selection, force, hook bypass, or other inferred policy.

## Brownfield specification and design

- Before drafting, the author examines the relevant requirement and design
  authorities, historical work records and comments, maintained regression
  tests and traceability, and affected implementation.
- The specification or design records the sources consulted and what existing
  behaviour or decisions it preserves, changes, supersedes, or leaves
  unaffected without copying the baseline.
- Regression-test assertions are used as evidence of actively protected
  functional behaviour and to trace that behaviour to originating
  requirements. Tests and code do not approve requirements.
- `audit-spec` and `audit-design` report a source-coverage finding when a
  material authority, historical decision, protected regression behaviour,
  traceability link, or implementation fact was omitted from the context pass.

## Independent audit contract

- `audit-spec`, `audit-design`, `audit-tests`, and `audit-code` invoke the
  `sdlc-audit` runner rather than auditing in the authoring context.
- Each invocation starts a non-resumed harness process in an empty temporary
  directory and embeds only the canonical prompt, judged artefacts, and exact
  context files supplied by the caller.
- The runner resolves audit harness, provider, and model from command-line,
  process-environment, project, user, then fallback sources. Audit harness falls
  back to the default harness.
- The configured audit harness is invoked in a fresh context. Hermes receives
  provider and model explicitly; Codex and Claude receive model only.
- Every phase-specific `*_PROVIDER` value is ignored when its selected harness
  does not accept explicit provider selection.
- Every report names the audit, auditor provider, auditor model, and exact PASS,
  PROVISIONAL, or FAIL verdict. Findings are classified as `[BLOCKING]`,
  `[CONDITION]`, or `[ADVISORY]`.
- PASS permits advisory findings only. PROVISIONAL requires at least one exact
  condition with a deterministic `VERIFY` clause and permits advisories. FAIL
  requires at least one blocking finding and permits advisories.
- Missing or malformed headers, unclassified findings, and verdicts inconsistent
  with their finding classifications are rejected.
- A report whose audit, provider, or model identity differs from the effective
  values is rejected. A structurally valid FAIL is returned to the authoring
  context as evidence rather than treated as runner failure.
- Harness display or reasoning text preceding the last exact requested audit
  header is discarded; only the validated report is emitted.
- Inputs must be regular non-symlink files beneath the project, canonical SDLC,
  or operating-system temporary directory. An exact external authority is
  supplied separately with `--external-context FILE`; it does not authorize a
  directory tree. The child receives a bounded environment and a 15-minute
  runtime budget and hard timeout.
- Automated regression coverage of `sdlc-audit` injects a fake harness and
  never invokes an agent harness or hosted model. A live end-to-end audit
  invocation is a metered one-off test only and must not be included in
  `make test`, CI, scheduled automation, or another persistent regression target.
- Effective specification PASS precedes planning; effective design PASS precedes
  tests and tasks; effective test PASS precedes implementation; effective code
  PASS precedes completion or convergence.
- A later relevant change preserves the earlier PASS as revision-specific
  history but makes it non-current for phase completion. A fresh audit assesses
  the delta and necessary context unless the change exactly satisfies a
  PROVISIONAL condition.
- A PROVISIONAL verdict matures to effective PASS without another model audit
  only when the author records both revisions, verifies every exact condition,
  proves no additional change entered the correction, and appends the evidence
  receipt to `audits.md`.
- The main authoring context remediates current-phase blocking findings and
  dispatches fresh independent audits without operator handback for at most five
  attempts, counting the first audit.
- Autonomous remediation stops earlier when it would change signed-off upstream
  authority or requires a human-controlled product, scope, security, privacy,
  access, persisted-data, external-contract, or irreversible decision.
- Phase handback occurs after PASS, effective PASS from satisfied PROVISIONAL
  conditions, the fifth FAIL, or an earlier human-controlled blocker and reports
  attempt history, decisions, assumptions requiring validation, advisories, and
  unresolved blockers.
- A Spec Kit feature that selects one-off or user tests contains
  `validation.md` with one traceable result entry per selected test.
- `audit-tests` verifies the planned `PENDING` entries and their execution tasks
  before implementation.
- `audit-code` assesses implementation after automated verification without
  requiring final one-off or user-test results.
- Completion and convergence require a current `audit-code` PASS and current
  passing `validation.md` results for every required non-automated test.
- A later relevant code change preserves the earlier audit and validation as
  revision-specific history but requires a fresh code audit and repetition of
  only the materially affected tests.

## Paired development contract

- The operator explicitly selects paired development for a bounded change; the
  agent never infers it from ordinary conversation.
- The session objective and each explicit iteration instruction define the
  specification for that iteration. A question does not authorize mutation.
- Human visual or subjective validation is first-class user-test evidence and
  is not replaced by an automated imitation.
- The agent maintains a provisional validation ledger and obtains one operator
  confirmation before recording the current entries as durable user tests.
- Each validation identifies the behaviour, reviewed state, material viewing
  conditions, and whether a later iteration superseded it.
- Automated regression coverage is added only when it protects objective,
  stable behaviour at proportionate cost and adds evidence beyond the user
  validation. Source-text grep is never behavioural evidence.
- Paired work receives one change-scoped `audit-code` at closure when code,
  templates, scripts, or non-trivial CSS changed. Unrelated legacy defects are
  non-blocking unless the change relies on or worsens them.
- Specification, design, and test audits apply only when the paired change
  creates material artefacts in their respective scopes.
