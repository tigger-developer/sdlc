---
title: Lean SDLC for Coding Agents
version: 13
last-updated: 2026-10-29
---

# Lean SDLC for Coding Agents

## Review without the retired audit harness

**The old automatic audit harness is retired.** Workflow audit steps use `definition-audit`, `test-code-audit` and
`code-audit` internally in the current session, without attempting an external
call first. Readiness checks, review criteria and human approval remain.

The three audit skills perform **self-audit only**. For an explicitly requested
independent review, start a fresh session or an authorized isolated native
sub-agent at [INDEPENDENT-AUDITOR.md](src/INDEPENDENT-AUDITOR.md), with a bounded
brief and original candidate evidence. Do not pass the author's conversation
or use a self-audit skill. [AUDITS.md](src/AUDITS.md) defines the caller's
isolation and evidence responsibilities. The reviewer loads engineering
criteria, without MAIN or delivery orchestration. Both methods use the same
[review criteria](src/REVIEW.md); self-review is labelled non-independent.
Both routes record results in `audits.yaml` under [the canonical audit record
contract](src/AUDIT-RECORD.md). Do not maintain parallel review documents.

## What it is

The Lean SDLC helps a person and a coding agent take a software change from an
idea to reviewed delivery. It supplies shared engineering standards, workflow
skills, and evidence the person can inspect before approving important
decisions. It supports Codex, Claude Code, GitHub Copilot CLI, Hermes, and Goose.

## Why it exists

A coding agent can write code quickly while missing a requirement, inventing a
design choice, or mistaking its own checks for independent review. This
framework keeps project decisions with the person, gives the agent the
standards relevant to its task, and checks a change where review matters. Its
balance is one definition document, focused reviews, and a choice
of delivery modes, proportionate to the change.

## How it got here

- **v1** established the engineering discipline. Its instruction set grew and
  consumed an increasing share of the agent's context window.
- **v2** integrated those standards with Spec Kit. The experiment taught useful
  lessons but did not remain the delivery model. Read [why the Spec Kit
  integration ended](LEARNINGS.md#why-the-spec-kit-integration-ended).
- **v3** uses those lessons to route relevant standards, retain review at
  meaningful boundaries, and balance assurance with the pace of routine work.

The [design journey and detailed learnings](LEARNINGS.md#the-journey-from-v1-to-v3)
explain those choices. [HTML-Preview](https://github.com/tigger-developer/HTML-Preview)
began here as a quick Markdown preview script. As the SDLC adopted Org, it
became a Markdown and Org viewer with browser annotations: an operator could
review a specification for approval and place feedback beside the passage an
agent needed to revise. Its usefulness beyond this framework and growing
complexity led to a separate project.

## Quickstart

Building and installing the SDLC commands requires Go 1.23 or later[^go],
Git, a `trash` command on `PATH`[^trash], and at least one supported
coding-agent harness. GitHub CLI is needed only for migrating an SDLC v1
project's GitHub tickets. Hermes auditing also requires `uv` and a supported
installed Hermes runtime; see the [harness guide](src/HARNESS.md).

```sh
make install
```

The installer asks which of the existing `~/.codex`, `~/.claude`,
`~/.copilot`, `~/.hermes`, and `~/.config/goose` homes should receive native adapters. A top-level
`active-harness` list in `~/.agents/sdlc.yaml` supplies that selection without
asking. It does not create a missing provider home.

From a clean, named branch in the project you want to use, run:

```sh
sdlc-init
```

The initializer sets up a new project or guides an existing SDLC v1 or v2
project through migration. It preserves the starting state on a dated archive
branch and asks before merging the migration branch. See the
[step-by-step Quickstart](QUICKSTART.md#3-initialize-one-project) for its choices
and recovery path.

For the first change, invoke `$define-change` in a supported agent. Review and
sign off its definition before invoking `$deliver-change`. The agent then
writes tests, implements the approved change, records validation, and requests
final review. The [full walkthrough](QUICKSTART.md#4-define-a-change) covers
the commands and human decisions.

## The main steps

### Initialize a project

`sdlc-init` creates the project profile and work ledger. For SDLC v1, it can
migrate legacy tickets and acceptance criteria. For SDLC v2, it preserves Spec
Kit material in a project archive and indexes unfinished work.
The process does not silently turn unfinished work into approved v3 changes.

### Choose the delivery mode

- **Standard:** define a change, obtain human sign-off, then deliver and verify
  it. This is the normal route for work that can be specified before coding.
- **Paired:** `$pair-change` supports live human review and short iterations,
  such as visual or editorial changes.
- **Hands-off spike:** `$hands-off-spike` answers a signed-off exploratory
  question unattended on a never-merged branch. Its code always receives an
  internal test-code and implementation reviews, and adoption returns to the standard route.
- **Emergency:** the exact `BYPASS-GATE-7` token selects the bounded emergency
  patch process. It records the decision and verification afterwards without
  suspending safety or evidence requirements.

The [standard workflow](src/MAIN.md#normal-workflow),
[paired route](src/PAIRING.md), [hands-off spike](skills/hands-off-spike/SKILL.md),
and [emergency route](src/EMERGENCY.md) set out their authority and validation
rules.

### Define the change

`$define-change` prepares one `spec.org` containing the outcome, boundaries,
acceptance criteria, test definitions, and enough solution design for another
agent to implement it. The definition-audit skill reviews that package.
The person signs off before standard delivery begins.

### Deliver and verify

`$deliver-change` takes a signed-off definition through written tests,
implementation, recorded results, and review of the finished change. Automated
regression tests establish RED and GREEN where applicable; operator and one-off
tests remain separately identified.

### Historical audit harness

The implementation described below is **retired historical material**. Its
commands, defaults and ownership rules do not govern current reviews. Use
AUDITS.md and AUDIT-RECORD.md; a future CI implementation must follow those
contracts rather than revive this interface.

The retired audit harness sent a defined evidence package to an independent reviewer,
records its verdict against that revision, and manages bounded retries and
recovery. It also performs local readiness checks before spending a review
round. An audit verdict is evidence for human judgement, not human approval.
The [audit harness guide](src/HARNESS.md) covers commands and diagnostics,
including per-turn, overall, response-start and response-idle deadlines.
Audit defaults are `10m`, `15m`, `3m` and `2m` respectively; duration strings
such as `90s` support personal overrides.
The harness supplies mandatory standards as verified text to every provider:
each profile receives its review criteria and applicable technologies, selected
from project declarations and evidence. Optional references contain only review
standards; MAIN and delivery/transport instructions are excluded. Spike review
adds no workflow, gate or product-documentation requirements. Native file tools are
disabled; unchanged standards are supplied once per retained context.
For a project without SDLC initialization, `sdlc-audit --input FILE` instead
reviews the supplied original project files against the installed coding and
documentation standards. It warns that the review is standalone and does not
create SDLC process records. Every audit is anchored to the project from which
the command is invoked and receives a bounded file inventory so omitted
project paths are visible, though their contents still require review.

## Find the detail

| Need | Read |
|---|---|
| Set up an agent and walk through a project | [Quickstart](QUICKSTART.md) |
| Understand the design history and v2 experiment | [Design learnings](LEARNINGS.md) |
| Configure global defaults or project overrides | [Configuration reference](docs/CONFIGURATION.md#configuration) and [Quickstart setup](QUICKSTART.md#2-set-optional-global-defaults) |
| Inspect installation behaviour and recovery | [Installation reference](docs/CONFIGURATION.md#install) |
| Understand audit evidence and readiness | [Audit record](src/AUDIT-RECORD.md) and [Org schema](src/ORG-SCHEMA.md) |
| Understand project authority and standards routing | [Core SDLC rules](src/MAIN.md) |
| Locate templates, skills, and implementation | [Repository layout](docs/CONFIGURATION.md#repository-layout) |
| Review legacy migration evidence | [Migration reference](docs/archive/README-v3.5.2.md#migration-evidence) |

The [previous detailed README](docs/archive/README-v3.5.2.md) is preserved as
historical reference. Current commands and rules are defined by the linked
standards, skills, and command help.

## Document history

This revision separates independent review, author self-audit and delivery-only
responsibilities. Earlier rules remain in their canonical owners and Git history.

- Version 11 (2026-10-24): Separate independent and self-audit routes, with shared
  criteria and retained-context reuse.

Version 10 records the suspended automatic auditor, internal workflow reviews
and optional manually invoked fresh-session audits through the review skills.

- Version 9 (2026-10-11): Document audit execution deadlines and Hermes runtime
  requirements.

- Version 7 (2026-10-10): Listed the hands-off spike route.
- Version 6 (2026-10-06): Document invocation-project anchoring and audit
  file inventory.
- Version 5 (2026-10-06): Introduced standalone code and documentation
  standards review through `sdlc-audit`.
- Version 4 (2026-10-05): Linked the current configuration, installation, and
  repository-layout reference instead of the archived README.
- Version 3 (2026-10-04): Documented interactive provider-home selection and
  the shared `active-harness` default.
- Version 2 (2026-10-01): Reorganized the front page around purpose, history,
  quickstart, and the main steps; preserved the previous detailed reference.

[^go]: `make install` also builds the pinned HTML-Preview source when
    `htmlpreview` is absent from `PATH`; that source currently requires Go
    1.26.8. The installed commands are compiled binaries and do not need Go
    just to run.

[^trash]: The installer needs a `trash` executable on `PATH` when it applies
    changes. This is a command requirement, not a requirement for one package
    manager. On macOS, check for `trash` first; [Homebrew's `trash`
    formula](https://formulae.brew.sh/formula/trash) is one way to obtain it if
    absent. On Linux,
    [trash-cli](https://github.com/andreafrancia/trash-cli) provides the
    command; other compatible implementations can be used. Homebrew itself
    is optional.

### Goose deployment

Goose receives SDLC skills under `~/.config/goose/skills`; shared rules remain
at `~/.agents/sdlc`. Its command guard is a native `PreToolUse` plugin under
`~/.agents/plugins/sdlc-command-guard-goose`. The installer preserves Goose
configuration and other plugins. See [Goose integration](docs/goose-support.md)
for requirements and the validation boundary.

### Branch defaults

Full-SDLC definition and ticket allocation stay on primary, whose `docs/work.org`
is the canonical register. Publish a ticket reservation before using its number.
Create a feature branch only for admitted code delivery, before writing tests or
production code; see [branch and allocation hygiene](src/DELIVERY.md).
This is not a personal or standalone coding-standard
rule, and this maintainer repository follows its own local branch policy.
Explicit implementation branch instructions and `delivery.branch_strategy: current` settings
remain overrides, including existing global or project settings. An omitted
strategy now resolves to `feature`; existing explicit settings are not migrated.
Emergency `BYPASS-GATE-7` work defaults to the recorded primary branch unless
the operator specifies another branch. See [Git standards](src/GIT.md).
