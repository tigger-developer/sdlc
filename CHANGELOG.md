# Changelog

## Unreleased

Reworked the `spec-kit-prototype` branch as a standards-only companion to
GitHub Spec Kit. The deployed SDLC now retains universal, specification,
testing, coding, Git, documentation, language, and domain standards while
removing SDLC-owned modes, approval keywords, ticket lifecycle, and build or
review orchestration.

Added `sdlc-project-init`, a cross-platform deterministic initializer for Spec
Kit projects. It discovers technology standards, resolves CLI, project, and
user configuration, renders a fixed constitution baseline, supports an optional
external infrastructure contract, no-ops without prompting when current, and
then invokes Codex, Claude, or Hermes for project-specific constitution text.
Project configuration includes separate delivery and audit provider/model
values.

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
