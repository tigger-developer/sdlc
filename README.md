---
title: Lean SDLC for Coding Agents
version: 2
last-updated: 2026-10-01
---

# Lean SDLC for Coding Agents

## What it is

The Lean SDLC helps a person and a coding agent take a software change from an
idea to reviewed delivery. It supplies shared engineering standards, workflow
skills, and evidence the person can inspect before approving important
decisions. It supports Codex, Claude Code, GitHub Copilot CLI, and Hermes.

## Why it exists

A coding agent can write code quickly while missing a requirement, inventing a
design choice, or mistaking its own checks for independent review. This
framework keeps project decisions with the person, gives the agent the
standards relevant to its task, and checks a change where review matters. Its
balance is one definition document, focused independent audits, and a choice
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
began here as a quick Markdown preview script. As the SDLC adopted Org and
browser annotations for reviewing specifications and designs in context, the
viewer grew useful beyond this framework and became its own project.

## Quickstart

Install from this repository with Go 1.23 or later, Git, the `trash` command,
and at least one supported coding-agent harness available. GitHub CLI is needed
only for migrating an SDLC v1 project's GitHub tickets.

```sh
make install
```

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
- **Emergency:** the exact `BYPASS-GATE-7` token selects the bounded emergency
  patch process. It records the decision and verification afterwards without
  suspending safety or evidence requirements.

The [standard workflow](src/MAIN.md#normal-workflow),
[paired route](src/PAIRING.md), and [emergency route](src/EMERGENCY.md)
set out their authority and validation rules.

### Define the change

`$define-change` prepares one `spec.org` containing the outcome, boundaries,
acceptance criteria, test definitions, and enough solution design for another
agent to implement it. An independent definition review checks that package.
The person signs off before standard delivery begins.

### Deliver and verify

`$deliver-change` takes a signed-off definition through written tests,
implementation, recorded results, and review of the finished change. Automated
regression tests establish RED and GREEN where applicable; operator and one-off
tests remain separately identified.

### Understand the audit harness

The audit harness sends a defined evidence package to an independent reviewer,
records its verdict against that revision, and manages bounded retries and
recovery. It also performs local readiness checks before spending a review
round. An audit verdict is evidence for human judgement, not human approval.
The [audit harness guide](src/HARNESS.md) covers commands and diagnostics.

## Find the detail

| Need | Read |
|---|---|
| Set up an agent and walk through a project | [Quickstart](QUICKSTART.md) |
| Understand the design history and v2 experiment | [Design learnings](LEARNINGS.md) |
| Configure global defaults or project overrides | [Configuration reference](docs/archive/README-v3.5.2.md#configuration) and [Quickstart setup](QUICKSTART.md#2-set-optional-global-defaults) |
| Inspect installation behaviour and recovery | [Installation reference](docs/archive/README-v3.5.2.md#install) |
| Diagnose an audit or its readiness check | [Audit harness guide](src/HARNESS.md) and [Org schema](src/ORG-SCHEMA.md) |
| Understand project authority and standards routing | [Core SDLC rules](src/MAIN.md) |
| Locate templates, skills, and implementation | [Repository layout](docs/archive/README-v3.5.2.md#repository-layout) |
| Review legacy migration evidence | [Migration reference](docs/archive/README-v3.5.2.md#migration-evidence) |

The [previous detailed README](docs/archive/README-v3.5.2.md) is preserved as
historical reference. Current commands and rules are defined by the linked
standards, skills, and command help.

## Document history

- Version 2 (2026-10-01): Reorganized the front page around purpose, history,
  quickstart, and the main steps; preserved the previous detailed reference.
