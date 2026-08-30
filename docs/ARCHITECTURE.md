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
concerns. It defines evidence preconditions for advancing between Spec Kit's
existing stages, but does not create a parallel lifecycle.

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
|   |-- technologies/
|   `-- presets/sdlc-standards/
|-- skills/
|-- hooks/
|-- templates/
|-- cmd/ and internal/
`-- public documentation and build metadata
```

Runtime discovery is bounded to `src/`, `skills/`, and `hooks/`.
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
|-- technologies/
|-- presets/sdlc-standards/
|-- skills/
`-- hooks/
```

Repository `src/` maps directly to the root of `~/.agents/sdlc`. Other runtime
directories retain their names.

Provider-native skill locations contain adapters or copies required for
discovery. They are not authoritative standards roots. Every independently
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
    `-- selected stack --------> technologies/GO.md, technologies/WEB.md, and so on
```

The project constitution references the selected standards. An agent loads only
the entries relevant to the current activity. The shared files remain the
single source of truth; their contents are not copied into every feature
artefact.

## Deterministic project initialization

`sdlc-project-init` discovers technology standards from the installed
`technologies/` directory, resolves command-line, project, and user defaults,
and renders
`.specify/templates/overrides/constitution-template.md`. This documented Spec
Kit override is resolved before preset composition, so constitution generation
does not require a preset manifest parser. Adding a technology document makes
it available without changing the initializer.

Before rendering a brownfield constitution, the initializer mechanically
migrates the established project documentation shape. It:

- adds fixed, marked authority introductions to `docs/VISION.md`,
  `docs/architecture.md`, and `docs/ACs.md` without duplicating an introduction
  already present;
- moves `docs/implementation_plan.md` unchanged to
  `docs/archive/implementation_plan.md`;
- shows the managed-path Git diff and obtains operator confirmation before an
  isolated commit; and
- performs no write, prompt, agent launch, or commit when the migration is
  already current.

`README.md` is outside the mechanical migration. Obsolete links and process
descriptions require project-specific semantic review rather than path
substitution.

The migration never invokes an agent. Missing expected source documents,
unexpected overlapping changes, or content that cannot be transformed by the
fixed rules are reported rather than guessed. Declining the commit leaves the
reviewed working-tree changes in place and stops constitution generation.

The generated template is editable pre-ratification scaffolding. It contains
the fixed standards proposal, universal standard references, selected
technology references, an optional external
infrastructure contract, mandatory independent audits, a fixed specification
baseline selected as greenfield or brownfield, and bounded placeholders for
project-specific principles. The brownfield structure separates current and
historical requirement authority, design authority, regression evidence, and
source precedence. A build made at an exact SDLC release tag records that tag;
other clean versioned builds record their source commit. Modified or unversioned
builds leave an explicit ratification TODO. The initializer commits only the
generated scaffold, leaving unrelated
staged and working-tree changes untouched. The selected agent harness receives
that template only after deterministic rendering and the isolated Git
checkpoint. Before ratification, any proposed clause may be corrected, removed,
or replaced. Ratification makes `.specify/memory/constitution.md` the sole
governance authority; later amendments use that document directly rather than
reapplying the initialization template.

Ratification and amendment impact reports are append-only history in
`.specify/memory/constitution-changelog.md`. They are not embedded normative
constitution content. The project constitution template is an initialization
artefact and may be removed after first ratification; Git retains its provenance.

User defaults live in the user-owned `~/.agents/.env`, outside the synchronized
standards tree. Project `.env` values override those defaults; CLI values
override both. The initializer reads only its named `SDLC_*` keys and copies
every resolved value into the ignored project `.env`. The project therefore
retains its initialization snapshot when the user defaults later change, and a
current rerun requires no questions or writes. `SDLC_PROJECT_TYPE` is the sole
project-only selection: the initializer ignores it in `~/.agents/.env` and
accepts it only from the project, the command line, or the project prompt.

## Spec Kit composition

The deployed preset lives at
`~/.agents/sdlc/presets/sdlc-standards`. It augments Spec Kit commands rather
than replacing them:

- `sdlc-project-init` supplies the editable constitution scaffold;
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
commands for the active integration. The initializer invokes the chosen harness
for the constitution operation only when the rendered scaffold or project
selection changes after any required brownfield documentation migration.

The `--integration` selected during `specify init` controls the project-local
agent adapter. It does not launch an agent and does not change the provider-
neutral specification artefacts.

## Skills and retired paths

Audit skills are independent, read-only challenges of specification, design,
tests, or code. Every audit runs in a fresh context and emits its name, provider,
model, and exact PASS or FAIL verdict. Any finding requires FAIL; changed
artefacts require a fresh audit.

```text
specification -- audit-spec PASS --> plan and design
plan/design  -- audit-design PASS -> tests and tasks
test design  -- audit-tests PASS --> implementation
implementation -- audit-code PASS -> completion or convergence
```

Verdicts are retained in the active feature's `audits.md`. The shared parser
fails closed on missing or malformed verdicts. Advisory skills load context,
diagnose, recommend, or summarize.

The standards-only model removes the previous SDLC commands and its drafting
and design workflow skills. The installer has a bounded retirement list for
those known paths. When one exists without a current source, it is renamed to an
adjacent `<path>.<epoch>.bak` backup in the canonical tree, the common skill
directory, and every supported provider adapter. Arbitrary destination-only
files are not inferred to be SDLC owned and remain untouched.

## Installer responsibilities

The installer:

1. discovers runtime files from the bounded source roots;
2. compares each owned source and destination file;
3. shows only variances unless `VERBOSE=1` is set;
4. performs no prompt or write when destinations are current;
5. backs up drift before synchronizing;
6. recoverably retires the finite legacy paths removed by this model, including
   old root-level technology documents and `audit-acs`;
7. installs provider-native adapters; and
8. validates provider prerequisites before mutation.

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
| Technology standard | A focused file under `src/technologies/`, plus a route in `MAIN.md` |
| Spec Kit command selection | `src/presets/sdlc-standards/` |
| Findings-only reusable review | `skills/<name>/SKILL.md` |
| Provider command safeguard | `hooks/` plus the relevant installer adapter |

Keep `MAIN.md` concise. A rule belongs there only when every coding task needs
it before progressive routing. Put detailed craft guidance in a selected
standards document and reference that document from the preset rather than
duplicating the rule.
