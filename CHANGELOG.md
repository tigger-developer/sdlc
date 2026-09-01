# Changelog

## Unreleased

Narrowed `migrate-legacy-acs-to-sdlc-v1` to a fast readiness reconciliation.
It now downloads every open and closed ticket with comments once, relies only on
that cache, uses at most one whole-suite run, and produces only a current
`docs/ACs.md`, refreshed stale project documentation, and unchanged archival of
the legacy implementation plan. It creates one batched local commit, never
comments on historical tickets, and may offer one operator-authorized silent
closure batch. Implementation-plan archival moved out of `sdlc-project-init`,
which retains only the mechanical legacy-ledger prefix.

Standardized the Make interface across adopting projects. Added canonical
contracts for `test`, `vulncheck`, `install`, `sync`, and `deploy`; defined
`make sync` as add, concise commit, pull, and push; and standardized
`COMMIT_MESSAGE`, `SKIP_TESTS=1`, and `VERBOSE=1` behaviour. Deployment may skip
an already-established regression run explicitly but never uses that control to
skip vulnerability checking.

Added a universal application-security standard requiring a read-only,
fail-closed `make vulncheck` interface for deployable applications, kept
separate from behavioural regression tests. Added selectable Hugo and Node.js
standards, including an explicit justification requirement for new Node.js
runtimes; tightened Python for internet-facing services; and added
stack-specific `govulncheck`, `pip-audit`, package-manager audit, `cpan-audit`,
Trivy, and OSV-Scanner guidance. Project initialization now includes the
security standard in every constitution, while deployment systems may consume
the common target without transferring host-security ownership to applications.

Prohibited metered external models, APIs, hosted services, tests, and probes
from automated regression suites, build targets, CI, scheduled jobs, and other
repeated automation. Validly invoked skills and workflows may perform their
required metered operations and retry loops without per-call approval. Bounded
one-off and user tests selected within authorized work have the same authority.

Added an exact-basename `.env` read prohibition to the universal standards and
provider configuration layer. Claude receives its recursive native read rule;
Claude, Codex, Copilot, and Hermes register the shared pre-tool guard for native
read/search inputs and direct shell references. `.env.example` and `.env.local`
remain readable. Provider-specific configuration merges preserve unrelated
settings and hooks, and repeated installation remains idempotent. The guard is
documented as a harness boundary rather than an operating-system sandbox.

Allowed operators to authorize named consecutive delivery phases in one
instruction. After audited design sign-off, "move on to test and build" now
covers task and test-traceability generation, test-audit convergence, analysis,
TDD implementation, and code-audit convergence without a redundant intermediate
handback. Required audits, RED/GREEN evidence, human-controlled decisions,
external authority, and autonomous-loop stop conditions remain unchanged.

Kept `migrate-legacy-acs-to-sdlc-v1` explicit-only, preserved descriptor-bearing
operator adjudication for genuine ambiguities and closure batches, and retained
temporary long-report presentation through `HTML_PREVIEW_TOOL` with a text-editor
fallback. Undelivered scope remains open and untouched.

## v2.0.2 - 2026-08-31

Added `sdlc-audit`, an isolated audit controller installed with the other SDLC
helpers. Audit skills now delegate to one-shot Hermes processes in empty
temporary working directories. The controller embeds the canonical audit
prompt and only explicitly selected file contents, applies project audit
configuration over user defaults, and passes provider and model explicitly. A
non-Hermes harness value warns and falls back to Hermes. Inputs are constrained
to exact named files in approved directories, child environment and runtime are
bounded, and malformed or misidentified reports are rejected. Hermes reasoning
or display prefixes are
discarded so callers receive only the validated report. Audit prompts now have
a single deployed source under `src/prompts/audits/` while skills remain small
invocation adapters. Exact external authorities use `--external-context FILE`
without authorizing a directory tree. Regression coverage uses a fake harness;
live hosted-model verification is a one-off test and is excluded from CI and
persistent regression targets.

Brownfield specification and planning now begin with an evidenced context pass
across relevant requirement and design authorities, historical work records,
the maintained regression test pack and its traceability, and affected
implementation. Feature artefacts record what they preserve, change, supersede,
and leave unaffected. The constitution scaffold, project-initializer prompt,
Spec Kit command overlays, audit skills, standards, public guidance, and
acceptance criteria all carry the same source-coverage contract. Removed the
brittle unit assertions that searched the constitution prompt for prose
fragments; the retained launcher test covers harness, model, arguments, and
working-directory behaviour.

Made `validation.md` the mandatory Spec Kit record for selected one-off and user
tests. Test design now creates traceable `PENDING` entries, implementation
records evidence-backed results after code audit, and `audit-tests` verifies the
planned ledger and execution tasks. `audit-code` assesses implementation without
requiring final non-automated results; completion and convergence require a
current code-audit PASS and current passing validation. Later changes preserve
revision-specific history while refreshing only materially affected audits and
tests.

Defined `BYPASS-GATE-7` as an exact-keyword, operator-only emergency route whose
surrounding request must state the current behaviour, required observable
behaviour, precise scope, and material constraints. The route now explicitly
selects every applicable automated regression, one-off, and user test. Only
automated tests use TDD and require a pre-change failure; omitted automation
requires a specific justification. The route retains durable specification and
documentation reconciliation, change-scoped code audit, and remediation to
effective PASS. Normal Spec Kit implementation applies the same test
classification and ordering.

Added an explicitly selected paired-development track for live human-agent
iteration. Bounded operator instructions define each iteration, visual and
subjective approvals remain first-class user-test evidence, and one final
confirmation records the session's current validation ledger. Automation is
required only when it adds durable evidence. Paired closure uses a change-scoped
code audit for material implementation and invokes other audits only when the
corresponding durable artefacts exist.

Added a machine-checkable PROVISIONAL audit verdict for exact mandatory
corrections that require no further judgement. Each `[CONDITION]` includes a
deterministic `VERIFY` clause. A condition receipt records the audited and
corrected revisions, evidence, and effective PASS without another model audit;
any additional, unverifiable, or judgement-based change returns to the normal
fresh-audit loop.

Audits now distinguish `[BLOCKING]` findings from optional `[ADVISORY]`
improvements. A PASS may contain advisories, while a FAIL requires at least one
blocking finding. The fail-closed verdict parser enforces those classifications.

Added a shared `AUDITS.md` contract for independent review and autonomous phase
convergence. The main authoring context remediates current-phase blockers and
dispatches fresh audits without operator handback for at most five attempts.
It returns after PASS, the fifth FAIL, or an earlier human-controlled decision,
with the decisions, assumptions, advisories, attempt history, and blockers
needed for operator sign-off.

Constitution, specification, clarification, planning, checklist, and task
commands now present approval artefacts through the optional
`HTML_PREVIEW_TOOL`, falling back to an available text editor or the exact
artefact paths. Presentation occurs after the applicable audit or validation
and is explicitly non-blocking and non-approving.

The README now documents the complete Spec Kit feature lifecycle for new users
and projects migrating from the earlier SDLC. It distinguishes the reusable
main authoring context from mandatory fresh audit contexts and identifies when
context reset is otherwise appropriate.

Constitution scaffolds now include Spec Kit's Sync Impact Report as one compact
HTML-comment line at the top of the constitution. Ratification and amendments
replace that line instead of accumulating a separate constitution changelog.

Project-initializer prose now lives in deployed files under `src/prompts/` and
`src/templates/` rather than embedded Go strings. This includes the
constitution scaffold, constitution-generation prompt, and brownfield
acceptance-criteria prefix.

The brownfield acceptance-criteria ledger now leads with an explicit
`# LEGACY DOCUMENT` heading in a managed prefix ending at `***`. The initializer
treats the ledger body as opaque: it prepends an absent prefix, no-ops when the
prefix SHA-256 matches the deployed resource, and replaces only a stale prefix
through the delimiter. A marker outside the expected prefix fails without
changing the document.

Vision, architecture, and README documents remain active and outside the
mechanical legacy migration. The initializer never modifies them; their Spec Kit
adaptation remains project-specific semantic work.

## v2.0.1 - 2026-08-30

Constitution templates no longer embed the configured audit provider or model;
runtime selection remains in project configuration. The generated template is
now explicitly editable pre-ratification scaffolding rather than an immutable
authority. Constitution generation performs a final fitness review over the
assembled document and removes unsuitable scaffold or agent-authored material.
Ratified constitutions are amended directly without reapplying the initialization
template. Sync Impact Reports are append-only history in
`.specify/memory/constitution-changelog.md`, not embedded constitution content.

Brownfield templates now require an explicit authority boundary between
legacy-process requirement records and approved Spec Kit feature
specifications, including explicit supersession and lineage preservation. Before
constitution generation, `sdlc-project-init` recognizes the established SDLC v1
documentation shape, mechanically adds fixed authority introductions, moves
`docs/implementation_plan.md` unchanged under `docs/archive/`, updates its
README link, displays only the managed Git diff, and obtains operator approval
before an isolated commit. Declined migrations remain staged for review; current
or unrelated brownfield layouts are silent no-ops.

Added the explicit-only `migrate-legacy-acs-to-sdlc-v1` skill for brownfield
projects. It caches complete GitHub issue bodies, comments, and implementation
links; reconciles ticket-based SDLC v0.1 AC and test lineage into the centralized
SDLC v1 record required before SDLC v2 adoption; ignores bug-fix tickets without
AC tables; and reserves multiple tables or inconclusive test evidence for
operator adjudication. The skill recommends Luna without defining a fallback
model.

The project initializer now asks whether each adopting project is greenfield or
brownfield and renders a fixed `Specification Baseline` appropriate to that
classification. Brownfield constitutions name exact current and historical
requirement authorities, design authorities, regression lineage, and source
precedence without copying the underlying documents. Greenfield constitutions
record that approved feature specifications establish requirements
prospectively. Project classification cannot be defaulted from the user-level
`~/.agents/.env`.

The constitution-generation prompt now populates that fixed authority map, and
the specification, clarification, and specification-audit instructions consume
it. Brownfield work is defined as a bounded behavioural delta against cited
approved requirements; unchanged baseline behaviour is referenced rather than
duplicated, while code and tests remain evidence rather than requirement
authority.

The project initializer now commits the generated
`.specify/templates/overrides/constitution-template.md` in an isolated Git
checkpoint before launching the constitution agent. An already-current but
untracked or modified scaffold is checkpointed without relaunching the agent,
and unrelated staged or working-tree changes are excluded.

Moved initializer user defaults to the deterministic, user-owned
`~/.agents/.env` path. New projects now snapshot every resolved global default
into their ignored project `.env`, while CLI and existing project values retain
precedence.

Replaced the ambiguous delivery provider/model settings with separate
specification and build settings. Constitution generation uses the specification
runtime; independent audits retain their own runtime. Legacy delivery values are
accepted as specification defaults and rewritten under the new names when a
project is initialized again.

## v2.0.0 - 2026-08-29

Released the Spec Kit integrated edition of the SDLC. GitHub Spec Kit 1.0 or
later is now a prerequisite and owns specification and delivery orchestration.
The deployed SDLC retains universal, specification, testing, coding, Git,
documentation, language, and domain standards while removing SDLC-owned modes,
approval keywords, ticket lifecycle, and build or review orchestration.

Added `sdlc-project-init`, a cross-platform deterministic initializer for Spec
Kit projects. It discovers technology standards, resolves CLI, project, and
user configuration, renders a fixed constitution baseline, supports an optional
external infrastructure contract, no-ops without prompting when current, and
then invokes Codex, Claude, or Hermes for project-specific constitution text.
Project configuration includes separate delivery and audit provider/model
values.

The generated constitution scaffold now records an exact SDLC release tag when
the initializer is built at that tag, or the source commit for another clean
versioned build. Modified or unversioned builds leave an explicit ratification
TODO rather than inventing traceability. Initial
drafts use `Last Revised` rather than amendment terminology.

`make install` now also installs both SDLC CLI helpers previously available only
through `make install-cli`.

Added a standalone quickstart for greenfield and brownfield adoption. The
document covers Spec Kit initialization, constitution review and ratification,
tracked project artefacts, brownfield evidence and delta specifications,
non-interactive selection, and idempotent reruns.

Corrected the generated constitution location to Spec Kit's project override
path. The resolver now returns the generated baseline before parsing preset
manifests, avoiding an otherwise unnecessary ambient PyYAML dependency during
constitution creation.

Tightened the project initializer's constitution invocation after a clean
brownfield trial still promoted feature requirements and detailed design into
project governance. The delivery prompt now applies explicit inclusion and
exclusion tests, limits project-specific principles, and requires one concise
authority hierarchy. A second trial clarified that authoritative project
documentation may support a constitutional invariant without its detailed
requirements being copied into the constitution. A third trial added
concern-specific source authority, explicit human governance, versioning,
deviation, compliance-review, and exhaustive-blocker requirements.

Moved technology standards under `src/technologies/` for automatic discovery.
The installer now backs up and retires the former root-level copies and obsolete
constitution addendum while preserving unrelated destination-only material.

Renamed `audit-acs` to `audit-spec`, added a scenario- and trade-off-based
`audit-design`, and standardized all four audits on fresh-context,
findings-only, machine-checkable PASS or FAIL verdicts that identify the auditor
provider and model. Added a fail-closed Go verdict parser. Spec Kit command
fragments now require a current independent PASS before the next delivery stage.

Added the deployable `sdlc-standards` Spec Kit preset. It composes progressive
standards loading into Spec Kit's constitution, specification, clarification,
planning, task, analysis, checklist, implementation, convergence, and
task-to-issue commands without duplicating the standards or replacing Spec
Kit's core workflow. The generated constitution references a project-specific
standards selection.

Reframed `ISSUES.md` as provider-neutral specification standards, simplified
testing and Git terminology, retained findings-only audit and advisory skills,
and removed the legacy command and drafting paths. Rewrote the README,
architecture guide, learnings, and provider example so the public project stands
alone without private agent configuration.

Retained `BYPASS-GATE-7` as an operator-only emergency exception for small,
clearly scoped work before Spec Kit or equivalent project artefacts exist. The
same-message request becomes a temporary specification without restoring SDLC
modes, approval gates, ticketing, or audit orchestration.

Added bounded, recoverable installer cleanup for the seven retired SDLC command
files and retired drafting, design, and audit skills. Active legacy paths in the
canonical tree, common skill directory, and supported provider adapters are
renamed to adjacent `<path>.<epoch>.bak` backups. All other destination-only
files remain untouched, and repeated installation returns to the no-prompt
current state.

Added a self-contained public architecture guide covering the framework
boundary, repository and installed layouts, progressive loading, delivery
lifecycle, commands and skills, installer ownership, and extension points.
Updated the README and design learnings so users need no knowledge of the
author's private agent configuration to understand or adopt the SDLC.

Separated agent-runtime instructions from repository content under `src/` and
made installation discover deployable files recursively from convention-based
runtime directories. Deployment decisions and default output are now per file;
changes to README, changelog, installer code, tests, build metadata, templates,
or project records cannot become file-deployment variances. Installer templates
remain available to the provider-configuration analysers that intentionally use
them.

Made direct `python` and `python3` interpreter commands operator-only while
retaining Python development through project-owned entry points.

Included supported provider-configuration variances in the interactive
installer's single preflight and confirmation batch. A detected Hermes home
without its first-run `config.yaml` now stops before any write with a visible,
actionable diagnostic. The Hermes command guard now also blocks `rm`, `sed`,
and `awk` to match the shared command policy. Hermes registration now targets
the provider-neutral `~/.agents/sdlc` hook and removes recognized obsolete
provider-local registrations during migration.

Made Hermes configuration analysis compare only the managed command-guard
semantics. Compliant Hermes-generated YAML is now retained byte-for-byte, while
required hook migrations preserve comments, key order, and unrelated values.

Required every human-facing ID to include a short adjacent descriptor.

Made operating modes explicitly code-only and self-contained within the SDLC.
Provider bootstrap instructions now decide only whether to load the SDLC and
no longer duplicate mode defaults, transitions, skill authority, or canaries.

Prohibited final handbacks while MODE DELIVER remains active. Agent-initiated
delivery exit now requires an auditable declaration that accounts for all
completed and incomplete scope, justifies every unfinished item, and records
either `DELIVERY READY` or `DELIVERY BLOCKED` before atomically returning to
MODE PAIR.

Made review-report presentation explicitly mode-dependent. MODE PAIR uses the
optional `HTML_PREVIEW_TOOL` and falls back to an available text editor, while
MODE DELIVER ignores preview tooling, records the Markdown evidence, and
continues without opening the report or treating it as a checkpoint.

Made MODE DELIVER a durable continuation contract. Delivery masters now carry
a completion matrix, decision records, and quality-check evidence; progress
reports are explicitly non-terminal while in-scope work remains executable.

Added ambiguity classification, consolidated blocker handbacks, and the
human-only `RESUME DELIVER n` directive for reconstructing an open delivery
after context loss or in a new session.

Established `~/.agents/sdlc` as the only canonical live SDLC tree and removed
provider-local SDLC copies from installer plans. Installation creates the
common root for a fresh home, deploys provider-native skill and command
adapters, preflights one complete batch, asks once, and uses rsync without
`--delete` so unrelated and agent-created content survives.

Replaced every SDLC-root placeholder with the literal canonical path and added
strict no-discovery wording to independently invoked skills, commands, and the
provider template. Clarified that documentation and prompt wording contracts
require human UT sign-off rather than source-inspection RTs.

Added adjacent `<path>.<epoch>.bak` backups before any drifted deployment or
configuration artefact is replaced.

Restricted SDLC ownership of Hermes configuration to the command-guard hook.
Private operations instructions and their bootstrap now remain entirely
outside this public project.

## v1.0.2 - 2026-08-20

Changed `make install` to run an interactive multi-agent deployment. It detects
installed provider homes, asks once for the shared live tree, and asks
separately for each provider adapter that differs. The reusable CLI link moved
to `make install-cli`, and provider configuration remains an explicit workflow.

Simplified deployment comparison by using rsync dry-run itemization for the
same repository-owned tree that rsync applies.

Changed `sdlc-install` to synchronize the complete staging tree into the common
live deployment at `~/.agents/sdlc`, exclude Git metadata, deploy every skill,
and point provider adapters only to live common paths.

Added cross-provider prohibitions for agent-submitted `python` and `python3` interpreter commands, including idempotent Claude and Codex configuration migration and a Hermes-compatible command guard.

Added Hermes as a first-class `sdlc-install` target. The SDLC installer now owns the Hermes operations bootstrap, terminal command-guard registration, configuration backup, and idempotent YAML merge alongside the Claude and Codex provider adapters.

## v1.0.1 - 2026-08-17

Restored the public repository URL in the README clone instructions.

## v1.0.0 - 2026-08-16

The repository was flattened for scrubbing for personal identifying information. This public release brings forward a pair-programming SDLC evolved over the last year or so and adds an Apache License 2.0 licence.
