# SDLC Architecture

This document explains how the public SDLC repository is organized, how its
parts relate, and what installation produces. It is descriptive guidance for
users and maintainers. The documents under [`src/`](../src/) are the normative
agent instructions.

## Framework Boundary

The SDLC is a standalone, provider-neutral framework for software delivery
with AI coding agents. It does not depend on the author's private agent
configuration or on a parent repository.

The framework begins when an agent is directed to read
`~/.agents/sdlc/MAIN.md` for applicable coding work. The user or their provider
instructions must establish when that happens. Personal preferences, general
conversation rules, and provider bootstrap instructions remain outside this
repository.

This separation has two consequences:

- Installing the SDLC does not replace a user's general agent instructions.
- MODE PAIR, MODE DELIVER, gates, and the coding lifecycle apply only after
  `MAIN.md` has been loaded.

The example in [`templates/AGENTS-or-CLAUDE.example.md`](../templates/AGENTS-or-CLAUDE.example.md)
shows one way to connect provider-level instructions to the public framework.
It is an example, not a required personal configuration.

## Repository Map

```text
sdlc/
|-- src/                 Normative framework entry point and routed standards
|-- commands/            Human-invoked workflow prompts
|-- skills/              Reusable drafting, audit, and advisory workflows
|-- hooks/               Optional provider-integrated command safeguards
|-- templates/           Provider configuration examples and installer inputs
|-- cmd/sdlc-install/    Installer command
|-- internal/installer/  Deployment planning and provider integration
|-- docs/                Explanations for users and maintainers
|-- README.md            Installation and first-use guide
|-- LEARNINGS.md         Design history and observed failure patterns
`-- CHANGELOG.md         Public release history
```

The directory boundary is deliberate. Only `src/`, `commands/`, `skills/`,
and `hooks/` contain agent-runtime material. Repository documentation,
installer source, tests, templates, and build metadata are not copied into the
canonical runtime merely because they exist in the repository.

## Source and Installed Layout

The repository is the staging source. The installed runtime has exactly one
canonical root:

```text
~/.agents/sdlc/
|-- MAIN.md
|-- ISSUES.md
|-- TESTING.md
|-- CODING.md
|-- GIT.md
|-- DOCUMENTATION.md
|-- language and domain standards
|-- commands/
|-- skills/
`-- hooks/
```

Files inside repository `src/` are copied directly to the root of
`~/.agents/sdlc`. The other runtime directories retain their names. Agents
must use this literal root and must never search the filesystem to discover an
alternative SDLC location.

Provider-native adapters are separate from the canonical runtime. For
example, a provider may require skills or commands in its own discovery
directory. The installer copies those adapters from the canonical public
sources, but it does not create a second authoritative SDLC tree in the
provider home.

## Loading Model

The framework uses progressive loading so irrelevant rules do not consume the
agent's context or dilute applicable instructions:

```text
Provider or user instruction
          |
          | applicable coding work
          v
~/.agents/sdlc/MAIN.md
          |
          | routing by task
          v
Selected process, craft, language, and domain standards
          |
          v
Project-specific documentation and instructions
```

`MAIN.md` is both the lifecycle entry point and the reference router. It
defines the code-specific process, then selects only the standards needed for
the current work. A Go implementation might load `CODING.md`, `TESTING.md`,
`GIT.md`, and `GO.md`; unrelated language standards stay out of context.

Project-specific instructions come after the public framework in this model.
They describe the repository being changed and must not redefine the public
SDLC or act as another installed copy of it.

## Normative and Explanatory Documents

| Material | Role |
|---|---|
| `src/MAIN.md` | Normative lifecycle, routing, modes, gates, and authority |
| Other files under `src/` | Normative standards when selected by `MAIN.md` |
| `commands/` | Explicit human entry points into named workflow stages |
| `skills/` | Reusable workflows whose invocation authority is defined elsewhere |
| `hooks/` | Optional runtime enforcement integrated by supported providers |
| `README.md` | Installation, connection, and first-use guidance |
| `docs/` | Explanations for users and maintainers |
| `LEARNINGS.md` | Historical rationale and lessons from observed failures |
| `CHANGELOG.md` | Summary of public changes over time |

Explanatory documentation does not override normative instructions. Conversely,
normative documents should not carry repository-maintainer guidance that an
agent does not need while delivering software.

## Delivery Lifecycle

The SDLC is issue-driven. Requirements, acceptance criteria, test evidence,
decisions, reviews, and closure remain attributable to the work they govern.
The detailed rules live in `MAIN.md`, `ISSUES.md`, and `TESTING.md`; the
high-level lifecycle is:

```text
Problem and scope
      |
      v
Issue, acceptance criteria, tests, and design
      |
      v
Authorized implementation and verification
      |
      v
Review evidence and human-only judgements
      |
      v
Approval and closure
```

### MODE PAIR

MODE PAIR is the interactive coding mode. The human and agent work through the
lifecycle with two explicit human gates:

| Gate | Purpose |
|---|---|
| `PROCEED n [n ...]` | Accept the issue definition and authorize implementation |
| `APPROVED n [n ...]` | Accept reviewed results and authorize closure |

The gates are checkpoints, not a request for the agent to stop between every
workflow step.

### MODE DELIVER

MODE DELIVER is the autonomous coding mode. It starts only after the agent has
stated a bounded delivery contract and the human has confirmed it. Work is
tracked through a durable master issue, completion matrix, child issues,
decisions, audits, and verification evidence.

The agent continues while safe, authorized work remains. Before returning
control, it must account for completed and incomplete scope, justify anything
left undone, record either `DELIVERY READY` or `DELIVERY BLOCKED`, and return
atomically to MODE PAIR. The normative exit conditions remain in `MAIN.md`.

## Commands, Skills, and Authority

Commands and skills package workflows; they do not grant themselves authority.

- Commands represent deliberate human actions, including gate-adjacent
  workflow stages.
- Drafting and design skills prepare auditable issue material.
- Audit skills challenge acceptance criteria, tests, code, or recommendations.
- Context skills load and summarize existing project state.

Installation makes these tools discoverable where a provider supports them.
Invocation remains governed by `MAIN.md`, the user's provider instructions,
and the current operating mode.

## Installer Responsibilities

The installer has three ownership layers:

1. **Canonical runtime:** compare and deploy files discovered beneath `src/`,
   `commands/`, `skills/`, and `hooks/`.
2. **Provider adapters:** place supported commands and skills in each
   provider's native discovery locations.
3. **Managed configuration fragments:** analyse and change only the narrow
   settings that integrate the SDLC with a supported provider.

Planning is per file. An absent or differing managed destination is a
variance; a matching destination is not. The default output lists variances
only, while `VERBOSE=1` also shows matching files.

Before replacing a managed artefact, the installer preserves the existing
version according to the backup behaviour documented in the README. It does
not delete destination-only material or claim unrelated provider
configuration. Repeating installation against compliant destinations produces
no deployment prompt.

## Extending the Framework

Choose the extension point by responsibility:

| Need | Location | Additional work |
|---|---|---|
| Universal coding-process rule | `src/MAIN.md` | Keep the entry point concise and route detail when possible |
| Cross-language craft standard | Existing routed file under `src/` | Update `MAIN.md` routing if its load conditions change |
| Language or domain standard | New file under `src/` | Add an explicit route and canary component in `MAIN.md` |
| Human-invoked workflow | `commands/` | Document its authority and supported provider adapters |
| Reusable agent workflow | `skills/<name>/` | Define its scope and invocation authority |
| Runtime safeguard | `hooks/` | Add provider integration and behavioural verification |
| Provider integration | `internal/installer/` and `templates/` | Own only the smallest necessary configuration fragment |
| User or maintainer explanation | `docs/` | Link it from the README when it is a primary guide |

New runtime files should fit one of the convention-based runtime directories.
That keeps deployment deterministic without maintaining a hard-coded manifest
and prevents repository-only material from entering agent context.

## Further Reading

- [`README.md`](../README.md) covers installation and connection to supported
  providers.
- [`LEARNINGS.md`](../LEARNINGS.md) records why the framework evolved this way.
- [`src/MAIN.md`](../src/MAIN.md) is the normative lifecycle and routing entry
  point.
- [`docs/ACs.md`](ACs.md) records the installer acceptance criteria retained
  with the project.
