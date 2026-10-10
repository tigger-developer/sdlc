---
title: SDLC Configuration and Installation Reference
version: 2
last-updated: 2026-10-11
---

# SDLC configuration and installation reference

This reference describes installation, configuration, and repository layout
for SDLC v3. The [README](../README.md) introduces the framework and the
[Quickstart](../QUICKSTART.md) walks through first use.

## Install

For the recommended **driver/auditor setup**, including model capability,
authentication and billing, see the [quickstart configuration guidance](../QUICKSTART.md#recommended-driver-and-auditor-setup).

`make install` also initializes the pinned
[tigger-developer/HTML-Preview](https://github.com/tigger-developer/HTML-Preview)
submodule and runs its own `make install` if `htmlpreview` is absent from PATH.
Existing installations are retained. This previewer renders **Org and
Markdown** as **high-fidelity HTML**, including navigable document structure.

```sh
make install
```

The installer lists the existing `~/.codex`, `~/.claude`, `~/.copilot`, and
`~/.hermes` directories and asks which detected agents should receive native
adapters. A top-level `active-harness` list in `~/.agents/sdlc.yaml` supplies
that selection without asking; listed harnesses whose provider home does not
exist are skipped. An explicit `--agent` selection takes precedence.

The repository-internal `sdlc-install` command is invoked by `make install`; it
is not installed on the global path. The installation:

- builds `sdlc-install` locally, deploys `sdlc-init` and
  `sdlc-merge-legacy-acs` under `~/.agents/sdlc/bin`, and links those operator
  commands onto the global path;
- deploys `sdlc-harness` under `~/.agents/sdlc/bin` and exposes it as
  `~/.local/bin/sdlc-audit`; `sdlc-audit --help` describes the audit interface;
- moves recognized retired command remnants to Trash rather than leaving
  executable `.sdlc-*-retired` names in command directories;
- synchronizes the canonical standards to `~/.agents/sdlc`;
- installs SDLC skills globally under `~/.agents/skills`;
- links canonical skills into the native Claude, Copilot, and Hermes skill
  locations of the selected harnesses; and
- registers the same canonical command guard through each selected harness's
  native hook mechanism without replacing unrelated provider configuration.

It lists only missing or differing artefacts. An unchanged rerun writes nothing
and asks no question. Set `VERBOSE=1` to include matching artefacts. A selected
plan is applied without a second aggregate confirmation. Replacing an unknown
same-path provider configuration conflict still requires an explicit `y` or
`yes`.

Replaced managed files, links, retired SDLC artefacts, and merged
**configuration files** go to **Trash**, not adjacent `.bak` files. This includes
provider settings and global SDLC defaults. Applying changes requires `trash`;
a failed operation stops replacement. Already trashed items remain recoverable
through Trash. Existing backup files are not cleaned up. `sdlc-init` preserves
its Git and document migration archives; those are historical evidence, not
disposable deployment copies.

`make sync` refreshes HTML-Preview from its upstream default branch and runs its
own `make install` to update the installed command. It then stages and commits
local changes including the updated submodule pin, and pulls and pushes the
SDLC branch. `COMMIT_MESSAGE` defaults to `chore: sync`. Each step stops on
failure. Ordinary installation uses the recorded pin; synchronization updates
both the pin and the installed previewer.

HTML-Preview's installer preserves conflicting files and links. If a previous
checkout owns `~/.local/bin/htmlpreview`, move that link aside once before
`make sync` installs the submodule-managed command.

## Configuration

Global non-secret defaults and the deployed SDLC release live at
`~/.agents/sdlc.yaml`. Project facts and explicit overrides live in tracked
`.sdlc/project.yaml`. Resolution order is:

1. command line;
2. process environment;
3. project YAML;
4. global YAML; and
5. schema default.

When a global value exists, the initializer inherits it without asking and does
not copy it into the project file. Pass `--override-global-config` to display
those questions and record selected project overrides. Project identity,
technology selection, and infrastructure role are always project decisions.

Example global configuration:

```yaml
version: 3
release: vX.Y.Z # maintained by make install
active-harness:
  - codex
  - claude
delivery:
  branch_strategy: current
  definition:
    harness: codex
    provider: openai
    model: gpt-5.6-sol
  build:
    harness: codex
    provider: openai
    model: gpt-5.6-terra
  goal:
    max_turns: 20             # Hermes and Copilot only.
                             # Claude Code caps consecutive Stop-hook continuations at 8.
    max_token_budget: "100,000" # Codex only; commas separate thousands.
  audit:
    harness: codex
    provider: openai
    model: gpt-5.6-luna
    timeout: 10m
    total_timeout: 15m
    response_start_timeout: 3m # A personal override can use 90s.
    response_idle_timeout: 2m
    max_rounds: 5
    max_failures: 3
    # Optional: use when an audit attempt ends without a usable verdict.
    # fallback:
    #   cool_off_period: 1h
    #   harness: hermes
    #   provider: nous
    #   model: z-ai/glm-5.3
infrastructure:
  owner: Example platform team
  contract: /absolute/path/to/PROJECT-INTEGRATION.md
```

`version` identifies the configuration schema. `release` identifies the latest
semantic-version Git tag deployed by `make install`; the installer maintains it
without replacing unrelated settings. Projects follow those global skills and
standards, so `.sdlc/project.yaml` does not duplicate an SDLC release pin.

Each of `delivery.definition`, `delivery.build` and `delivery.audit` accepts an
optional `fallback` mapping containing `harness`, `provider` where supported, and
`model`. Project YAML replaces the global tuple as a whole; `fallback: {}` disables
it. Omission inherits it. Audit fallback handles attempts ending without a usable
verdict, including usage limits, launch or authentication errors, timeout and
malformed output. PASS, FAIL and PROVISIONAL PASS never trigger it; running
attempts must finish or time out. Local configuration, evidence and record
errors remain blockers. Non-audit definition and build retain
authentication-only fallback. See the [audit harness guide](../src/HARNESS.md)
for recovery, cooldown and diagnostics.

Audit durations require units, for example `90s` or `3m`, and a minimum of
one second. The three additional deadline keys also accept
`SDLC_AUDIT_TOTAL_TIMEOUT`, `SDLC_AUDIT_RESPONSE_START_TIMEOUT` and
`SDLC_AUDIT_RESPONSE_IDLE_TIMEOUT` process overrides. The earliest deadline wins;
`total_timeout` spans provider recovery and fallback. A recoverable response
timeout is not an instruction to intervene. See the
[audit deadline contract](../src/HARNESS.md#audit-execution-deadlines).
Hermes auditing additionally requires `uv` and a supported installed Hermes
runtime; it uses existing packages offline. Development checks use the locked
project-local Python environment through `make test-hermes` and `make lint`.

Goal values above match the schema defaults. Plain positive integers and
correctly comma-grouped thousands are accepted; periods, decimals, malformed
grouping, scientific notation, signs, zero and leading zeroes are rejected before
numeric conversion. Project overrides use the same keys under
`.sdlc/project.yaml`. Native goal activation is capability-dependent; see the
[native goal instructions](../skills/deliver-change/references/native-goal.md).

Document inspection defaults to `htmlpreview`. Set `HTML_PREVIEW_TOOL` to an
executable name or path to use another viewer; an unset or empty value uses the
default. Preview is presentation, not a gate.

The initializer can import only schema-allowlisted historical `SDLC_*` values
from a project `.env` through its shell wrapper. After the YAML profile is
written, it removes those exact historical keys while preserving all unrelated
lines. Agents never read `.env`, and v3 runtime configuration never depends on
it.

## Repository layout

| Path | Purpose |
|---|---|
| `src/MAIN.md` | Universal rules and progressive routing |
| `src/*.md` | Requirements, testing, auditing, coding, Git, documentation, security, paired, emergency, Org, and Markdown standards |
| `src/technologies/` | Automatically discoverable technology standards |
| `src/templates/v3/` | Unified specification, work, audit, and validation templates |
| `src/prompts/` | Audit gate prompts and saved prompts for bounded initializer analysis |
| `skills/` | Globally installed workflow skills |
| `hooks/` | Shared command guard registered with each selected harness |
| `cmd/` and `internal/` | Installer, initializer, validator, harness adapters, and ledger merger |

## Document history

- Version 2 (2026-10-11): Document audit deadlines, units, overrides and Hermes
  runtime requirements.

- Version 1 (2026-10-05): Moved the current installation, configuration, and
  repository-layout reference out of the archived v3.5.2 README, updated it for
  harness selection, and removed retired focused audit skills from the layout.
