# Engineering Standards for AI Coding Agents

> **Prerequisite:** SDLC v2 requires
> [GitHub Spec Kit](https://github.com/github/spec-kit) 1.0 or later. Install
> the `specify` CLI and ensure it is available on `PATH` before installing or
> initializing this SDLC.

This public repository provides a provider-neutral engineering standards
library for coding agents. Spec Kit owns
specification and delivery orchestration; this repository supplies the coding,
testing, Git, documentation, language, and domain standards applied within it.

The library does not require private agent instructions. Its canonical installed
root is exactly `~/.agents/sdlc` for every supported provider.

## What is included

| Path | Purpose |
|---|---|
| `src/MAIN.md` | Compact universal rules and progressive-loading routes |
| `src/ISSUES.md` | Specification and acceptance-criteria standards |
| `src/TESTING.md` | Behavioural testing and evidence standards |
| `src/CODING.md` | Cross-language implementation standards |
| `src/GIT.md` | Source-control and recoverability standards |
| `src/DOCUMENTATION.md` | Public technical-documentation standards |
| `src/technologies/*.md` | Automatically discoverable technology standards |
| `src/presets/sdlc-standards/` | Spec Kit preset that selects standards progressively |
| `src/prompts/project-init/` | Constitution-generation prompt resource |
| `src/templates/project-init/` | Constitution scaffold and managed brownfield document block |
| `skills/` | Findings-only audits and advisory tools |
| `hooks/` | Optional provider-integrated command safeguard |
| `cmd/` and `internal/` | Installer, project initializer, and audit-verdict implementation |

Only runtime material from `src/`, `skills/`, and `hooks/` is
deployed. Repository documentation, installer source, tests, and build metadata
do not enter the live instruction tree.

## Install the standards

Building or installing from source requires Spec Kit 1.0 or later, a Go
toolchain, `rsync`, and the tools required by the repository's build targets.
The installed Go executables do not require a separate Go runtime. The complete
greenfield and brownfield setup is documented in
[`QUICKSTART.md`](QUICKSTART.md).

```bash
git clone https://github.com/tigger-developer/sdlc.git ~/code/sdlc
```

```bash
cd ~/code/sdlc
```

```bash
make install
```

The installer:

- installs `sdlc-install` and `sdlc-project-init` under `~/.local/bin`;
- synchronizes `src/` into `~/.agents/sdlc` and retains the other runtime
  directory names;
- installs common and provider-native skill copies for detected agents where
  required for discovery;
- compares source and destination before prompting;
- lists only missing or differing destinations by default;
- backs up owned drift before replacement;
- backs up and retires the known command and drafting-skill paths removed by
  the standards-only model; and
- performs no deployment prompt or write when every detected destination is
  current.

Set `VERBOSE=1` to include matching destinations in the plan:

```bash
VERBOSE=1 make install
```

To inspect installer options:

```bash
make build
```

```bash
bin/sdlc-install --help
```

Provider-specific homes are adapters, not alternate SDLC roots. Agents must
load standards from `~/.agents/sdlc` and must never search the filesystem for a
different copy.

## Initialize a Spec Kit project

Run the initializer from the adopting project:

```bash
sdlc-project-init
```

The initializer separates deterministic selection from semantic drafting. It:

- initializes Spec Kit when authorized and absent;
- installs the deployed SDLC preset;
- asks whether the adopting project is greenfield or brownfield;
- discovers available standards from `~/.agents/sdlc/technologies/`;
- asks once which technologies and infrastructure ownership apply;
- renders the editable constitution scaffold at
  `.specify/templates/overrides/constitution-template.md`;
- pins the exact SDLC release tag used to build the initializer, or the source
  commit for another clean versioned build, and leaves an explicit ratification
  TODO for an unversioned or modified build;
- commits only the generated constitution scaffold before invoking an agent,
  without including unrelated staged or working-tree changes;
- reports only a changed template, with additional variance detail under
  `VERBOSE=1`; and
- invokes the selected agent harness to complete only the project-specific
  parts of the constitution.

The initializer writes only its named SDLC selections into the project `.env`
and adds that file to `.gitignore`; unrelated existing values are preserved.
User defaults come from `~/.agents/.env`. Every resolved default is copied into
a new project's `.env`, producing a stable project snapshot. Later changes to
the user defaults affect new projects, not existing snapshots. Command-line
values override project values, which override user defaults. Project
classification is deliberately project-only and is never read as a user
default. Supported keys are:

```text
SDLC_AGENT_HARNESS
SDLC_SPEC_PROVIDER
SDLC_SPEC_MODEL
SDLC_BUILD_PROVIDER
SDLC_BUILD_MODEL
SDLC_AUDIT_PROVIDER
SDLC_AUDIT_MODEL
SDLC_PROJECT_TYPE
SDLC_TECHNOLOGIES
SDLC_INFRA_ENABLED
SDLC_INFRA_OWNER
SDLC_INFRA_CONTRACT
```

`SDLC_PROJECT_TYPE` accepts `greenfield` or `brownfield`. Set it in the project
`.env` or with `--project-type`; a value in `~/.agents/.env` is ignored.

Specification settings apply to constitution, specification, clarification,
design, planning, and task-definition agent invocations. Build settings apply
to implementation and convergence. Audit settings apply to independent audit
invocations. The initializer accepts legacy `SDLC_DELIVERY_PROVIDER` and
`SDLC_DELIVERY_MODEL` values as specification defaults and rewrites project
snapshots using the new names.

The external owner and contract are optional project inputs. The public SDLC
does not assume any particular infrastructure project. Run
`sdlc-project-init --help` for non-interactive overrides and `--no-launch`.

When the rendered scaffold, brownfield migration, and selections are already current, the initializer
writes nothing, asks nothing, and does not relaunch an agent. If the current
generated scaffold is untracked or modified, it checkpoints that file without
relaunching the agent.

For a greenfield repository, the initializer offers to run `specify init`, then
creates the standards profile and unratified constitution. For a brownfield
repository, it preserves the existing project, installs Spec Kit into the
working tree, applies the current managed legacy prefix to the
acceptance-criteria ledger, archives
`docs/implementation_plan.md`, and shows the bounded Git diff before asking
whether to commit the migration. Vision, architecture, and README documents
remain active and require project-specific semantic review. The initializer then
generates a proposed `Specification Baseline` authority map for the constitution
agent to populate without copying feature or design detail. Projects without
that legacy document shape are left unchanged.

The managed `docs/ACs.md` prefix comes from the deployed SDLC resource rather
than Go source. A rerun compares the prefix through its `***` delimiter by
SHA-256. It prepends the block when absent, writes nothing when current, and
replaces only a stale managed prefix while preserving the ledger body byte for
byte. A marker outside the managed prefix is reported as an error.

The generated template has no authority before ratification. The candidate
constitution may correct, remove, or replace any scaffold clause. Ratification
makes `.specify/memory/constitution.md` the sole governance authority. Its first
line is the current compact Sync Impact Report; amendments replace that line
rather than accumulating a separate changelog.
Follow the complete procedures in [`QUICKSTART.md`](QUICKSTART.md).

## Deliver a feature with Spec Kit

Spec Kit owns the feature workflow. Its
[Agentic SDD guide](https://github.com/github/spec-kit/blob/main/docs/reference/agentic-sdd.md)
defines the upstream command sequence and artefacts. The SDLC preset keeps that
orchestration and adds progressively loaded engineering standards plus four
independent audit gates.

Spec Kit documentation writes commands as `/speckit.*`. Codex skills mode uses
the equivalent `$speckit-*` form. The examples below use Codex syntax; use the
form exposed by the selected agent integration.

### Context model

Use one main authoring context for the feature. It may run specification,
clarification, planning, task generation, analysis, implementation, and
convergence. A context hand-off is not an approval gate; the files under the
active `specs/` feature directory carry the durable state.

An independent audit is the only stage transition that mandates a different
context. The auditor must not be the context that authored the artefact. The
main context may dispatch a fresh subagent and resume after its verdict, or the
operator may run the audit in a separate agent session.

Apply these rules to the main context:

- **Start fresh for each feature** so unrelated feature history is not carried
  in.
- **Reset after unsafe compaction** when the context can no longer account for
  the current specification, plan, tasks, and audit record.
- **Consider resetting before a large implementation** when a build-focused
  context loaded from the approved artefacts would be clearer.

Do not reset the main context merely because the workflow advances from
`specify` to `clarify`, `plan`, or `tasks`. If a new or compacted context does
continue the feature, it must read the constitution and the active feature's
current artefacts and audit record before acting.

### Step-by-step feature workflow

1. **Specify the required behaviour.**

   Invoke `$speckit-specify` with the requested change. Describe observable
   behaviour, purpose, boundaries, and important failure cases, not the
   implementation. It creates the feature directory and `spec.md`. In a
   brownfield project, the SDLC overlay requires a bounded delta against the
   authorities named in the constitution rather than a restatement of the
   existing system.

2. **Clarify material ambiguity.**

   Invoke `$speckit-clarify` before planning. It asks up to five targeted
   questions and writes the answers into `spec.md`. It may report that no
   critical ambiguity remains. Repeat it with a focus area when necessary.
   Clarification is for decisions that affect observable behaviour, scope,
   security, access, persisted data, or validation. Technical choices belong
   in planning.

3. **Audit the specification in a fresh context.**

   Run `audit-spec`. It must return `PASS` before planning. Record the provider,
   model, artefact revision, verdict, and findings in the feature's `audits.md`.
   Any later specification or clarification change invalidates that PASS and
   requires another fresh audit.

4. **Plan the implementation.**

   Resume the main context and invoke `$speckit-plan`. This is where
   architecture, technology, interfaces, data structures, migration,
   compatibility, security, operability, and test architecture are decided.
   Depending on the feature, Spec Kit may create `plan.md`, `research.md`,
   `data-model.md`, `contracts/`, and `quickstart.md`.

5. **Audit the design in a fresh context.**

   Run `audit-design`. It must return `PASS` before test design or task
   generation. A later plan or design change invalidates that PASS.

6. **Optionally generate focused checklists.**

   `$speckit-checklist` creates domain-specific quality checks for the written
   requirements. It is useful for security, accessibility, privacy, or other
   areas needing an explicit completeness review. It does not replace an SDLC
   audit.

7. **Generate tasks and test traceability.**

   Invoke `$speckit-tasks`. It converts the approved specification and design
   into an ordered `tasks.md`, including the required verification,
   documentation, migration, and human-validation work.

8. **Audit the tests and tasks in a fresh context.**

   Run `audit-tests`. It must return `PASS` before implementation. A later
   change to the test design, traceability, or affected tasks invalidates that
   PASS.

9. **Analyse cross-artefact consistency.**

   Invoke `$speckit-analyze` after tasks and before implementation. It checks
   consistency and coverage across the specification, plan, and tasks. Resolve
   material findings in the authoritative artefact, then rerun any audit that
   the change invalidated. Analyse does not replace an independent audit.

10. **Implement the approved tasks.**

    Invoke `$speckit-implement`. A small feature may remain in the main context.
    A large feature may start a fresh build-focused context or be implemented in
    bounded phases; each context must load the approved artefacts before
    changing code. Implementation must not change required behaviour silently.

11. **Audit the implementation in a fresh context.**

    Run `audit-code` after implementation and verification. It must return
    `PASS` before completion or convergence. Any subsequent implementation
    change invalidates that PASS.

12. **Converge and repeat when necessary.**

    Invoke `$speckit-converge` to assess the implementation against the
    specification, plan, and tasks. If it appends missing work to `tasks.md`,
    implement that work, rerun affected verification and `audit-code` in a
    fresh context, then converge again. Stop only when no work remains and all
    required audit records are current.

The resulting control flow is:

```text
main:  specify -> clarify
fresh: audit-spec PASS
main:  plan
fresh: audit-design PASS
main:  checklist (optional) -> tasks
fresh: audit-tests PASS
main:  analyze -> implement
fresh: audit-code PASS
main:  converge -> repeat affected stages when work is appended
```

`$speckit-taskstoissues` is optional. It projects tasks into GitHub issues while
retaining Spec Kit's task artefacts as the source of truth and applying the
rule that human-facing identifiers always include descriptors.

For very large features, follow Spec Kit's
[spec-of-specs guidance](https://github.com/github/spec-kit/blob/main/docs/concepts/spec-of-specs.md)
instead of forcing an agent to retain an oversized feature in one context.

### Moving from the earlier SDLC

SDLC v2 does not use the former modes, gate ceremonies, approval keywords, or
ticket lifecycle. Their useful guarantees now live in durable artefacts and
explicit transitions:

- the constitution replaces preloaded process instructions;
- `spec.md`, `plan.md`, and `tasks.md` replace conversational scope hand-offs;
- independent audit PASS records replace approval gates; and
- Spec Kit commands own progression between stages.

The operator may review, correct, or pause at any stage. No legacy keyword is
required to continue. `BYPASS-GATE-7` remains only the narrowly defined
emergency exception described below, not an alternative feature workflow.

The preset adds standards to Spec Kit without replacing its orchestration. The
shared SDLC documents remain the single source of truth for those standards.

## Emergency exception

Spec Kit is the supported v2 project workflow. Provider or project instructions
may still direct an agent to read `~/.agents/sdlc/MAIN.md` for an isolated coding
task, but an equivalent durable project specification is required before code
is written.

For a small emergency change before project Spec Kit artefacts exist, the human
may include `BYPASS-GATE-7` in the request. The request then serves as a
temporary specification for its stated outcome and scope. The exception does
not restore the retired SDLC modes, gates, or ticket lifecycle and does not
override safety or engineering standards.

The public template at
[`templates/AGENTS-or-CLAUDE.example.md`](templates/AGENTS-or-CLAUDE.example.md)
shows the minimum standalone integration. Adapt provider discovery filenames and
locations to current provider documentation.

## Skills and safeguards

The audit skills are findings-only and have a common machine-checkable verdict:

- `audit-spec` challenges requirements and acceptance criteria;
- `audit-design` challenges boundaries, trade-offs, failure handling, migration,
  operability, security, and testability;
- `audit-tests` challenges evidence and coverage; and
- `audit-code` reviews implementation against the selected standards.

Advisory skills diagnose problems, recommend technical decisions, summarize
open work, or load minimal project context. Audits run in a fresh context, never
modify the judged artefact, and identify both provider and model in their
verdict. A skill's frontmatter may recommend an inexpensive audit provider and
model; runtime configuration has precedence.

The explicit-only `migrate-legacy-acs-to-sdlc-v1` skill snapshots every GitHub
issue, including comments and recorded implementation links, then reconciles
ticket-based SDLC v0.1 acceptance criteria and test lineage into the centralized
SDLC v1 record required as a brownfield baseline for SDLC v2. Tickets without
AC tables are ignored as bug fixes; multiple tables and unresolved test
evidence are reserved for operator adjudication.

The optional Hermes hook and provider rules reinforce the common prohibitions
on `rm`, `sed`, `awk`, and direct `python` or `python3` interpreter commands.
They do not prevent a documented project target from invoking its own runtime.

Hermes must create `~/.hermes/config.yaml` through its startup configuration
before the installer can safely amend it. If the Hermes home exists without that
file, installation stops with a prerequisite diagnostic.

## Updating and rollback

Update the staging clone, inspect the branch, and redeploy:

```bash
git pull --ff-only
```

```bash
make install
```

Release tags provide stable rollback points. Check out the required release and
rerun `make install` to redeploy its managed runtime files. Arbitrary
destination-only files are preserved. The installer recognizes only the known
commands, skills, root-level technology files, and obsolete constitution
addendum retired by the v2 migration. It renames them to adjacent
`<path>.<epoch>.bak` backups and leaves all other destination-only material
untouched.

## Further reading

- [`QUICKSTART.md`](QUICKSTART.md) covers installation and greenfield and
  brownfield initialization.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) describes source, deployment,
  loading, and Spec Kit composition.
- [`LEARNINGS.md`](LEARNINGS.md) records the design lessons behind the current
  standards model.
- [`CHANGELOG.md`](CHANGELOG.md) records repository changes.

## Licence

Apache License 2.0. See [`LICENSE`](LICENSE).
