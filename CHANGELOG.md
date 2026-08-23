# Changelog

## Unreleased

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
