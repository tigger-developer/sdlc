# Changelog

## Unreleased

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
