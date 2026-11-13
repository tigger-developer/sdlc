---
title: SDLC Configuration and Installation Reference
version: 10
last-updated: 2026-11-13
---

# SDLC configuration and installation reference

This reference describes installation, configuration, and repository layout
for SDLC v3. The [README](../README.md) introduces the framework and the
[Quickstart](../QUICKSTART.md) walks through first use.

## Install

For canonical deployment only, run `make install INSTALL_FLAGS="--bare"`, or
`bin/sdlc-install --bare` after building. It installs the framework, skills,
commands, hooks and global release metadata under `/opt/sdlc`. It does not
discover agents, read legacy home configuration, create home links, retire
home artefacts, or configure harnesses. Nix or another declarative manager owns
those integrations. Existing canonical defaults remain intact except for the
release metadata. With no canonical configuration, a minimal one is created.

Bare mode never requests sudo. Pre-create a writable `/opt/sdlc` when ordinary
creation is unavailable. Use `--bare --dry-run` for a non-mutating preview.
Provider flags and `--configure` cannot be combined with `--bare`. The normal
installation behaviour described below remains the default.

For the recommended **driver/auditor setup**, including model capability,
authentication and billing, see the [quickstart configuration guidance](../QUICKSTART.md#recommended-driver-and-auditor-setup).

The separate `make install-preview` target initializes the pinned
[tigger-developer/HTML-Preview](https://github.com/tigger-developer/HTML-Preview)
submodule and runs its own `make install` if `htmlpreview` is absent from PATH.
Existing installations are retained. This previewer renders **Org and
Markdown** as **high-fidelity HTML**, including navigable document structure.

```sh
make install
```

The installer lists the existing `~/.codex`, `~/.claude`, `~/.copilot`,
`~/.hermes`, and `~/.config/goose` directories and asks which agents should
receive native adapters. A top-level `active-harness` list in
`/opt/sdlc/sdlc.yaml`, or the legacy `~/.agents/sdlc.yaml` during first
installation, supplies that
selection without asking; listed harnesses whose provider home does not
exist are skipped. An explicit `--agent` selection takes precedence.

The repository-internal `sdlc-install` command is invoked by `make install`; it
is not installed on the global path. The installation:

- builds `sdlc-install` locally, deploys `sdlc-init` and
  `sdlc-merge-legacy-acs` under `/opt/sdlc/bin`, and links those operator
  commands onto the global path;
- deploys `sdlc-harness` under `/opt/sdlc/bin` and exposes it as
  `~/.local/bin/sdlc-audit`; `sdlc-audit --help` describes the audit interface;
- moves recognized retired command remnants to Trash rather than leaving
  executable `.sdlc-*-retired` names in command directories;
- synchronizes the canonical standards to `/opt/sdlc/docs`;
- installs canonical SDLC skill directories under `/opt/sdlc/skills` and links
  each one into `~/.agents/skills`;
- links canonical skills into the native Claude, Copilot, and Hermes skill
  locations of the selected harnesses; and
- registers the same canonical command guard through each selected harness's
  native hook mechanism without replacing unrelated provider configuration.

Hook configuration also migrates recognized legacy SDLC guard paths and removes
duplicate SDLC registrations. Codex, Claude, Hermes and Copilot retain unrelated
hooks; Codex and Claude keep their original matchers. Known Goose plugin
generations migrate automatically. Canonical plugin files are replaced with the
complete staged tree and the previous contents remain in its rollback backup;
unknown plugin content in a harness home still requires review.
This reconciliation occurs during ordinary selected-harness installation, not
bare deployment of user-home configuration. Native hook execution still needs
host verification after installation.

It lists only missing or differing artefacts. An unchanged rerun writes nothing
and asks no question. Set `VERBOSE=1` to include matching artefacts. A selected
plan is applied without a second aggregate confirmation. Replacing an unknown
same-path provider configuration conflict still requires an explicit `y` or
`yes`.

The installer replaces an existing skill directory with a canonical link only
when its complete contents match the current source or the recorded installed
release. It replaces an existing Goose plugin directory only when its files
match the former SDLC plugin projection. Existing links must point to a known
SDLC location. Unknown skill files, altered skills, and unrelated links stop
the installation for manual review. Recognized Finder and Explorer metadata
files are ignored when comparing skill and Goose plugin content; they remain
recoverable with the old directory in Trash. Ordinary extra files and
directories still prevent replacement. Naming a path like an SDLC skill is
not proof of ownership.

The canonical tree is staged, validated and replaced as a whole. Previous
contents and a recovery journal remain under `/opt/sdlc/.installer/`; the root
itself is never renamed. `sdlc.yaml` retains operator settings through release
migration. Stale canonical artefacts, including extra files inside skills, leave
the active tree without a file-retirement inventory. An interrupted operation
is restored on the next applying invocation; dry-run reports pending recovery.

Replaced managed home files, links, and merged provider **configuration files**
go to **Trash**, not adjacent `.bak` files. The transaction also snapshots these
adapters so an application failure restores their prior state. Applying home adapter changes requires `trash`;
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
`/opt/sdlc/sdlc.yaml`. Project facts and explicit overrides live in tracked
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
    # reasoning_effort: high # optional
  test_code:
    harness: codex
    provider: openai
    model: gpt-5.6-sol # writes tests from the signed-off specification
    # reasoning_effort: high
  build:
    harness: codex
    provider: openai
    model: gpt-5.6-terra
    # reasoning_effort: low
  goal:
    max_turns: 20             # Hermes and Copilot only.
                             # Claude Code caps consecutive Stop-hook continuations at 8.
    max_token_budget: "100,000" # Codex only; commas separate thousands.
  audit:
    harness: codex
    provider: openai
    model: gpt-5.6-luna
    # reasoning_effort: high
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

Each phase (`definition`, `test_code`, `build`, `audit`) accepts optional
`reasoning_effort` beside its harness, provider and model. Omission requests no
effort override. Supported levels vary by harness, provider and model. The public
installer copies these global preferences into advisory metadata on the relevant
installed skills. `deliver-change` carries the build preference and a separate
`preferred_test_code_*` set for its test-code phase. The skill source remains
unchanged. A harness may ignore the metadata; a skill invocation alone does not
switch a running session's model. After editing `/opt/sdlc/sdlc.yaml`, run
`make install` again to refresh installed skill metadata. Project overrides and
scheduler launch settings are resolved at invocation time, not by reinstalling
the shared skill tree.

An authorized independent audit follows `AUDITS.md`: its caller resolves the
effective audit tuple, requests it when launching a fresh reviewer, and records
the actual invocation and verdict in `audits.yaml`. The former `sdlc-harness`
is not the active audit route. Its fallback and deadline settings below describe
the retained legacy executable and do not launch a reviewer.

`delivery.test_code` selects the harness, provider and model for writing
executable tests from the signed-off specification. Test definitions remain
part of `delivery.definition`. The test-code tuple does not split
`deliver-change` into separate jobs; a scheduler may use it when dispatching a
distinct test-code phase.

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
| `src/MAIN.md` | Delivery-agent rules and progressive workflow routing |
| `src/READING.md` | Full-read reuse and shared safeguards |
| `src/ADVICE.md` | Advisory project routing, excluded from auditor reads |
| `src/DELIVERY.md` | Author execution, record and closure responsibilities |
| `src/INDEPENDENT-AUDITOR.md` | Independent fresh-context auditor entry; no MAIN |
| `src/REVIEW.md` | Shared substantive review criteria and profiles |
| `src/SELF-AUDIT.md` | Non-independent author review and evidence |
| `src/AUDITS.md` | Delivery gate orchestration |
| `src/EXTERNAL-AUDIT.md` | Preserved suspended transport procedure |
| `src/*.md` | Requirements, testing, auditing, coding, Git, documentation, security, paired, emergency, Org, and Markdown standards |
| `src/technologies/` | Automatically discoverable technology standards |
| `src/templates/v3/` | Unified specification, work, audit, and validation templates |
| `src/prompts/` | Audit gate prompts and saved prompts for bounded initializer analysis |
| `skills/` | Globally installed workflow skills |
| `hooks/` | Shared command guard registered with each selected harness |
| `cmd/` and `internal/` | Installer, initializer, validator, harness adapters, and ledger merger |

## Document history

Version 6 completes the test-code model setting with its own harness and
provider, allowing a distinct test-code job to select its full agent tuple.

Version 5 adds `delivery.test_code.model` without changing the delivery
workflow or the definition and build model roles.

Version 4 separates independent review, author self-audit and delivery-only
responsibilities. Earlier rules remain in their canonical owners and Git history.

- Version 3 (2026-10-24): Separate delivery, independent-audit and self-audit entry points.

- Version 2 (2026-10-11): Document audit deadlines, units, overrides and Hermes
  runtime requirements.

- Version 1 (2026-10-05): Moved the current installation, configuration, and
  repository-layout reference out of the archived v3.5.2 README, updated it for
  harness selection, and removed retired focused audit skills from the layout.
