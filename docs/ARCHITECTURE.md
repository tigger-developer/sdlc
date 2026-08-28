# SDLC Standards Architecture

This document explains how the public standards library is composed, installed,
and integrated with Spec Kit. Documents under [`src/`](../src/) are normative;
this document is explanatory.

## Boundary

The repository owns engineering standards:

- universal coding constraints;
- specification and acceptance-criteria quality;
- testing and evidence;
- cross-language and stack-specific implementation standards;
- Git and documentation standards;
- findings-only audit and advisory skills; and
- optional provider-integrated command safeguards.

It does not own delivery modes, approval gates, ticket states, planning phases,
or implementation orchestration. Spec Kit or the adopting project owns those
concerns.

This separation makes the standards usable with more than one agent provider
and prevents a large lifecycle prompt from being loaded before the task needs
it.

## Repository map

```text
sdlc/
|-- src/
|   |-- MAIN.md
|   |-- ISSUES.md
|   |-- TESTING.md
|   |-- CODING.md
|   |-- GIT.md
|   |-- DOCUMENTATION.md
|   |-- language and domain standards
|   `-- presets/sdlc-standards/
|-- skills/
|-- commands/
|-- hooks/
|-- templates/
|-- cmd/ and internal/
`-- public documentation and build metadata
```

Runtime discovery is bounded to `src/`, `skills/`, `commands/`, and `hooks/`.
This prevents the README, changelog, installer tests, and repository internals
from becoming agent instructions.

## Source and installed layout

The repository is a staging source. Installation produces one authoritative
runtime:

```text
~/.agents/sdlc/
|-- MAIN.md
|-- ISSUES.md
|-- TESTING.md
|-- CODING.md
|-- GIT.md
|-- DOCUMENTATION.md
|-- language and domain standards
|-- presets/sdlc-standards/
|-- skills/
|-- commands/
`-- hooks/
```

Repository `src/` maps directly to the root of `~/.agents/sdlc`. Other runtime
directories retain their names.

Provider-native skill and command locations contain adapters or copies required
for discovery. They are not authoritative standards roots. Every independently
invoked entry point uses the literal `~/.agents/sdlc` path and prohibits
filesystem discovery if it is unavailable.

## Progressive loading

`MAIN.md` is the compact universal entry point. It contains rules that must be
present for any coding task and a routing table for additional documents.

```text
coding task
    |
    v
~/.agents/sdlc/MAIN.md
    |
    +-- requirements ----------> ISSUES.md
    +-- verification ----------> TESTING.md
    +-- implementation --------> CODING.md
    +-- source control --------> GIT.md
    +-- technical docs --------> DOCUMENTATION.md
    `-- selected stack --------> GO.md, WEB.md, and so on
```

The project constitution records a standards profile. An agent loads only the
profile entries relevant to the current activity. The shared files remain the
single source of truth; their contents are not copied into every feature
artefact.

## Spec Kit composition

The deployed preset lives at
`~/.agents/sdlc/presets/sdlc-standards`. It uses Spec Kit's composition
strategies rather than replacing core commands:

- a `constitution-template` addendum records the project's selected standards;
- command preambles load the relevant canonical standards; and
- Spec Kit's lower-priority core command remains responsible for its normal
  operation.

The composition is intentionally selective:

| Spec Kit command | Standards loaded |
|---|---|
| `speckit.constitution` | Universal rules, then documents selected from verified project evidence |
| `speckit.specify`, `speckit.clarify` | Universal and specification standards |
| `speckit.plan` | Universal, coding, Git, documentation, and selected stack standards |
| `speckit.tasks` | Universal, testing, documentation, and profile-specific standards |
| `speckit.analyze` | Universal, specification, and testing standards |
| `speckit.implement` | Universal and only the profile entries needed by current tasks |
| `speckit.converge` | Universal, coding, testing, and selected profile entries |
| `speckit.taskstoissues` | Universal identifier and source-of-truth rules |

Spec Kit copies preset material into project state and materializes composed
commands for the active integration. Installing or changing a preset does not
silently rewrite an existing live constitution; the operator invokes the
constitution command to adopt the new composed template.

The `--integration` selected during `specify init` controls the project-local
agent adapter. It does not launch an agent and does not change the provider-
neutral specification artefacts.

## Skills and compatibility paths

Audit skills are independent, read-only challenges of requirements, tests, or
code. They report evidence and remediation without changing the subject or
declaring approval. Advisory skills load context, diagnose, recommend, or
summarize.

The previous SDLC commands and drafting skill paths are retained as concise
compatibility notices. This is required because installation preserves
destination-only files. Keeping the paths with inert guidance makes branch
switching and redeployment reversible without destructive cleanup.

## Installer responsibilities

The installer:

1. discovers runtime files from the bounded source roots;
2. compares each owned source and destination file;
3. shows only variances unless `VERBOSE=1` is set;
4. performs no prompt or write when destinations are current;
5. backs up drift before synchronizing;
6. installs provider-native adapters; and
7. validates provider prerequisites before mutation.

It never treats repository documentation or arbitrary files as deployable merely
because they exist. It does not create provider-specific authoritative SDLC
trees.

## Extending the standards

| Change | Location |
|---|---|
| Universal engineering rule | `src/MAIN.md` |
| Cross-language implementation standard | `src/CODING.md` |
| Requirement or acceptance-criteria standard | `src/ISSUES.md` |
| Testing standard | `src/TESTING.md` |
| Language or domain standard | A focused file under `src/`, plus a route in `MAIN.md` |
| Spec Kit command selection | `src/presets/sdlc-standards/` |
| Findings-only reusable review | `skills/<name>/SKILL.md` |
| Provider command safeguard | `hooks/` plus the relevant installer adapter |

Keep `MAIN.md` concise. A rule belongs there only when every coding task needs
it before progressive routing. Put detailed craft guidance in a selected
standards document and reference that document from the preset rather than
duplicating the rule.
