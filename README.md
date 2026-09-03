# Engineering Standards for AI Coding Agents

> **Prerequisite:** SDLC v2 requires
> [GitHub Spec Kit](https://github.com/github/spec-kit) 1.0 or later. Install
> the `specify` CLI and ensure it is available on `PATH` before installing or
> initializing this SDLC.

This public repository provides a provider-neutral engineering standards
library for coding agents. Spec Kit owns staged specification and delivery
orchestration. This repository supplies the coding, testing, Git,
documentation, language, and domain standards applied within it, plus a bounded
paired-development track for work refined through live operator review.

The library does not require private agent instructions. Its canonical installed
root is exactly `~/.agents/sdlc` for every supported provider.

## What is included

| Path | Purpose |
|---|---|
| `src/MAIN.md` | Compact universal rules and progressive-loading routes |
| `src/ISSUES.md` | Specification and acceptance-criteria standards |
| `src/TESTING.md` | Behavioural testing and evidence standards |
| `src/SECURITY.md` | Application vulnerability-checking and exception contract |
| `src/AUDITS.md` | Independent verdict and autonomous phase-convergence contract |
| `src/PAIRING.md` | Explicit paired-development and live user-validation contract |
| `src/CODING.md` | Cross-language implementation standards |
| `src/GIT.md` | Source-control and recoverability standards |
| `src/DOCUMENTATION.md` | Public technical-documentation standards |
| `src/technologies/*.md` | Automatically discoverable technology standards |
| `src/presets/sdlc-standards/` | Spec Kit preset that selects standards progressively |
| `src/prompts/project-init/` | Constitution-generation prompt resource |
| `src/prompts/audits/` | Canonical prompts for isolated audit processes |
| `src/templates/project-init/` | Constitution scaffold and managed brownfield document block |
| `skills/` | Findings-only audits and advisory tools |
| `hooks/` | Provider-integrated command and sensitive-file safeguard |
| `cmd/` and `internal/` | Installer, project initializer, isolated audit runner, and verdict implementation |

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

- installs `sdlc-install`, `sdlc-project-init`, and `sdlc-audit` under
  `~/.local/bin`;
- synchronizes `src/` into `~/.agents/sdlc` and retains the other runtime
  directory names;
- installs common and provider-native skill copies for detected agents where
  required for discovery;
- compares source and destination before prompting;
- lists only missing or differing destinations by default;
- backs up owned drift before replacement;
- configures Claude, Codex, Copilot, and Hermes pre-tool safeguards after the
  same variance review and confirmation;
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
SDLC_AUDIT_HARNESS
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
to implementation and convergence. Audit settings select the independent
audit harness, provider, and model. This release always uses Hermes: an unset
or `hermes` harness value is silent, while another value warns and falls back
to Hermes. Project `.env` values override `~/.agents/.env`, and provider and
model are passed explicitly to Hermes. The initializer accepts legacy
`SDLC_DELIVERY_PROVIDER` and
`SDLC_DELIVERY_MODEL` values as specification defaults and rewrites project
snapshots using the new names.

The external owner and contract are optional project inputs. The public SDLC
does not assume any particular infrastructure project. Run
`sdlc-project-init --help` for non-interactive overrides and `--no-launch`.

When the rendered scaffold, brownfield authority marker, and selections are
already current, the initializer writes nothing, asks nothing, and does not
relaunch an agent. If the current generated scaffold is untracked or modified,
it checkpoints that file without relaunching the agent.

For a greenfield repository, the initializer offers to run `specify init`, then
creates the standards profile and unratified constitution. For a brownfield
repository, first invoke `$migrate-legacy-acs-to-sdlc-v1`. That bounded skill
archives every ticket and comment in the repository, uses at most one
whole-suite run, reconciles `docs/ACs.md`, builds `docs/ticket-migration.org`
incrementally as durable working state, refreshes stale project documentation,
and archives
`docs/implementation_plan.md`. It performs no ticket-by-ticket code archaeology,
test reruns, migration comments, or per-ticket commits. After the archive commit
succeeds, it closes every legacy issue and records the closure result.

The initializer preserves the existing project, installs Spec Kit into the
working tree, and applies the current managed legacy prefix to the
acceptance-criteria ledger. It shows only that bounded Git diff before asking
whether to commit. It does not archive or semantically revise project
documentation. It then generates a proposed `Specification Baseline` authority
map for the constitution agent to populate without copying feature or design
detail. Projects without the legacy AC ledger are left unchanged.

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
independent audits with bounded autonomous phase convergence.

Spec Kit documentation writes commands as `/speckit.*`. Codex skills mode uses
the equivalent `$speckit-*` form. The examples below use Codex syntax; use the
form exposed by the selected agent integration.

### Context model

Use one main authoring context for the feature. It may run specification,
clarification, planning, task generation, analysis, implementation, and
convergence. A context hand-off is not an approval gate; the files under the
active `specs/` feature directory carry the durable state.

An independent audit is the only stage transition that mandates a different
context. The audit skill invokes `sdlc-audit`, which starts a fresh one-shot
Hermes process in an empty temporary directory. It embeds the
canonical audit contract, the audit-specific prompt, the candidate, and only
the exact context files named by the authoring agent. No authoring conversation
or child-agent context is inherited.

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

### Presenting approval artefacts

Document-producing commands present their current artefacts once they have
passed the required audit, or after validation when no audit applies. If the
optional `HTML_PREVIEW_TOOL` environment variable names an executable, the
command invokes it with the artefact paths. Otherwise it opens the artefacts in
an available text editor or reports their exact paths when no editor is
available.

Previewing is an operator convenience. It is not approval, a workflow gate, or
a reason for the agent to stop. The SDLC does not prescribe or install a
particular preview implementation.

### Autonomous phase convergence

The main authoring context owns convergence within the current phase. A failed
audit is not an operator handback when its blocking findings can be remedied in
the current artefact. The author remediates them, records its decisions and
assumptions, and dispatches a fresh independent audit. The initial audit counts
as attempt one; each phase permits at most five attempts.

The author stops earlier when remediation would change a signed-off upstream
artefact or requires a human-controlled product, scope, security, privacy,
access, persisted-data, external-contract, or irreversible decision. It must
not change upstream authority or switch auditors merely to obtain a PASS.

On PASS, or after satisfying a PROVISIONAL verdict's exact conditions, the
author presents the candidate and requests operator sign-off. The handback
states the attempt count, decisions and rationale, assumptions or upstream
changes requiring validation, retained advisories, and any unresolved blockers.
A fifth FAIL produces the same consolidated handback without phase sign-off.

The operator may sign off the current phase and authorize named consecutive
phases in one instruction. For example, after design sign-off, "move on to test
and build" authorizes tasks and test traceability, `audit-tests`, analysis, TDD
implementation, and `audit-code`. Each audit remains mandatory, but an effective
PASS advances automatically within the authorized sequence. The agent hands
back at the end, or earlier for a material decision, scope or upstream change,
safety boundary, separate external authority, or exhausted audit limit.

A PROVISIONAL verdict avoids another model audit only for exact, mechanical,
deterministically verifiable corrections that do not affect behaviour,
architecture, security, privacy, access, persisted data, external contracts, or
irreversible outcomes. The author records the audited and corrected revisions,
condition evidence, and effective PASS in `audits.md`. Any additional or
judgement-based change requires a fresh audit.

### Step-by-step feature workflow

1. **Specify the required behaviour.**

   Invoke `$speckit-specify` with the requested change. Describe observable
   behaviour, purpose, boundaries, and important failure cases, not the
   implementation. It creates the feature directory and `spec.md`. In a
   brownfield project, the SDLC overlay first requires a context pass across the
   relevant requirement and design authorities, historical work records,
   maintained regression tests and traceability, and affected implementation.
   The specification records what the bounded delta preserves, changes,
   supersedes, and leaves unaffected rather than restating the existing system.

2. **Clarify material ambiguity.**

   Invoke `$speckit-clarify` before planning. It asks up to five targeted
   questions and writes the answers into `spec.md`. It may report that no
   critical ambiguity remains. Repeat it with a focus area when necessary.
   Clarification is for decisions that affect observable behaviour, scope,
   security, access, persisted data, or validation. Technical choices belong
   in planning.

3. **Audit the specification in a fresh context.**

   The main context dispatches `audit-spec`. It remediates blocking findings
   within the specification and re-audits under the five-attempt limit. A
   specification or clarification change preserves the earlier PASS as history
   but makes it non-current unless it exactly satisfies a PROVISIONAL condition.
   Record every attempt and any condition receipt in the feature's `audits.md`;
   planning requires the current effective PASS and operator sign-off.

4. **Plan the implementation.**

   Resume the main context and invoke `$speckit-plan`. This is where
   architecture, technology, interfaces, data structures, migration,
   compatibility, security, operability, and test architecture are decided.
   Brownfield planning consults the current design authority, relevant tickets
   and comments or equivalent work history, regression tests and traceability,
   and affected code before recording inherited constraints, deliberate
   changes, superseded decisions, and unaffected boundaries.
   Depending on the feature, Spec Kit may create `plan.md`, `research.md`,
   `data-model.md`, `contracts/`, and `quickstart.md`.

5. **Audit the design in a fresh context.**

   The main context dispatches `audit-design`, remediates current-design
   blockers, and re-audits under the five-attempt limit. It may make reasonable,
   reversible technical decisions consistent with the signed-off specification
   and must report them at handback. Test design and task generation require the
   current effective PASS and operator sign-off.

6. **Optionally generate focused checklists.**

   `$speckit-checklist` creates domain-specific quality checks for the written
   requirements. It is useful for security, accessibility, privacy, or other
   areas needing an explicit completeness review. It does not replace an SDLC
   audit.

7. **Generate tasks and test traceability.**

   Invoke `$speckit-tasks`. It converts the approved specification and design
   into an ordered `tasks.md`, including the required verification,
   documentation, migration, and human-validation work. When one-off or user
   tests are selected, it also creates `validation.md` in the active feature
   directory with one `PENDING` entry per test.

8. **Audit the tests and tasks in a fresh context.**

   The main context dispatches `audit-tests`, remediates current test-design or
   traceability blockers, and re-audits under the five-attempt limit.
   Implementation requires the current effective PASS and operator sign-off.

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
    Automated tests follow RED/GREEN; one-off and user tests do not.
    Deployable applications also run their repository-owned `make vulncheck`
    security gate after the dependency graph and build artefact are current.
    This gate remains separate from `make test` and behavioural traceability.

11. **Audit the implementation in a fresh context.**

    The main context dispatches `audit-code` after implementation and
    verification. It remediates implementation blockers and re-audits under the
    five-attempt limit. The verdict assesses the implementation, not whether
    final non-automated results have already been recorded.

12. **Validate the audited candidate.**

    After `audit-code` has an effective PASS, execute every selected one-off and
    user test and record its result in the active feature's `validation.md`. If a
    result exposes a defect and remediation changes code, rerun affected
    automated tests, `audit-code`, and affected validation. The earlier audit
    remains evidence for its revision but is no longer current for completion.

13. **Converge and repeat when necessary.**

    Invoke `$speckit-converge` to assess the implementation against the
    specification, plan, and tasks. If it appends missing work to `tasks.md`,
    implement that work, rerun affected verification and `audit-code` in a
    fresh context, then converge again. Stop only when no work remains and all
    required audit and validation records are current.

The resulting control flow is:

```text
main:  specify -> clarify
fresh: audit-spec -> main remediation -> fresh re-audit, at most 5 attempts
main:  plan
fresh: audit-design -> main remediation -> fresh re-audit, at most 5 attempts
main:  checklist (optional) -> tasks
fresh: audit-tests -> main remediation -> fresh re-audit, at most 5 attempts
main:  analyze -> implement -> vulncheck for deployable applications
fresh: audit-code -> main remediation -> fresh re-audit, at most 5 attempts
main:  one-off and user tests -> validation.md -> converge
main:  repeat affected audit and validation when remediation changes code
```

### Check application vulnerabilities

Every deployable application exposes the stable `make vulncheck` interface
defined by `~/.agents/sdlc/SECURITY.md`. It uses the scanner selected by the
applicable technology standards, changes nothing, fails closed when scanning is
unavailable or incomplete, and blocks on non-exempt findings. Mixed-stack
projects aggregate every applicable dependency set.

The target is a security gate, not a behavioural regression test, and therefore
does not belong in `make test`. Deployment systems may invoke it before release;
projects should also use proportionate periodic scanning because a vulnerability
can be disclosed after an unchanged version has been deployed. Infrastructure
owners retain responsibility for host, operating-system, service, container, or
package-closure controls beyond the application boundary.

### Use the standard Make interface

Projects using Make use the same target names for the same operations:
`build`, `lint`, `test`, `vulncheck`, `install`, `sync`, and `deploy` where each
operation applies. `make sync` is deliberately simple: it stages the worktree,
commits changed content with the short `COMMIT_MESSAGE` subject, then pulls and
pushes. It does not invent a detailed commit message or add hidden Git policy.
The default subject is `chore: sync`.

Deployment runs `make test` and `make vulncheck` before the project-specific
deployment action. An operator may use `SKIP_TESTS=1 make deploy` when the suite
has already been established, but this does not skip vulnerability checking.
Use `VERBOSE=1` for full output; normal output remains concise. The complete
contracts live in `~/.agents/sdlc/CODING.md`, `GIT.md`, `TESTING.md`, and
`SECURITY.md`.

### Develop interactively with an operator

Paired development is available when the required result must be refined
through live human review, such as visual and experience-led web work. The
operator selects it explicitly and provides a bounded session objective. Each
explicit iteration instruction is the specification for that slice; a question
is not implementation authority.

The agent implements and verifies one reviewable slice, presents the user-visible
result, and retains explicit operator validation in a provisional ledger. At
closure it presents the current validations once and asks whether they may be
recorded as the user tests for the change. In a Spec Kit feature, the mandatory
record is the active feature's `validation.md`. It includes the behaviour,
reviewed revision or state, material viewing conditions, result, human
authority, and any superseded validation.

Automation is added only when it protects objective, stable behaviour and adds
evidence beyond the paired validation. The agent must not manufacture source
greps or synthetic browser tests to imitate a visual judgement the operator has
already made.

Paired work does not repeat every staged Spec Kit audit. A change-scoped
`audit-code` runs at closure for new or materially modified code, templates,
scripts, or non-trivial CSS. Other audits apply only when the change creates the
corresponding durable specification, design, or test architecture. The complete
contract is `~/.agents/sdlc/PAIRING.md`.

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

The paired-development track is not a restored agent-wide mode or alternative
autonomous workflow. It is an explicitly selected, operator-present path for a
bounded change, with its own closure evidence.

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
temporary specification only when it states the current behaviour, required
observable behaviour, precise scope, and important constraints or exclusions.
The agent may choose routine implementation details but may not invent the
intended outcome. If those facts are not clear enough to define distinguishing
evidence, the emergency change is not ready for implementation.

Select every applicable test type: automated regression tests, one-off tests,
and user tests may be combined. Only automated tests follow TDD and require a
pre-change failure. Define one-off and user-test evidence before implementation
where practical, then execute their final verification after `audit-code` has
an effective PASS. Omitting an automated regression test requires a specific
justification; urgency, difficulty, or inconvenience is insufficient.

After implementing and verifying the smallest coherent fix, reconcile the
durable specification, design, and affected documentation, then obtain a
change-scoped `audit-code` effective PASS and run the selected one-off and user
tests. A defect-driven code change makes the earlier PASS historical rather than
current and requires a fresh audit plus repetition of materially affected tests.
The route skips pre-implementation Spec Kit artefacts and stage audits, not
applicable testing, code remediation, documentation, or verification. The exact
keyword must appear in the human request and must never be inferred or invoked
by an agent.

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

Audits classify material phase blockers as `[BLOCKING]`, exact mechanical
corrections as `[CONDITION]`, and optional improvements as `[ADVISORY]`. PASS
permits advisories, PROVISIONAL requires at least one condition, and FAIL
requires at least one blocking finding. Audits run through `sdlc-audit`, never
modify the judged artefact, and identify both provider and model in their
verdict. The runner rejects malformed reports or an identity different from the
requested configuration. It bounds each invocation to 15 minutes. The shared
contract is `~/.agents/sdlc/AUDITS.md`.

Use `--external-context FILE` when an audit needs one exact authority outside
the project and canonical SDLC directories. This supplies only the named file;
it does not authorize a directory tree. Regression tests of the runner use a
fake harness. Live hosted-model invocations are metered one-off tests and must
not be added to `make test`, CI, or another persistent regression target.

The explicit-only `migrate-legacy-acs-to-sdlc-v1` skill retires a brownfield
project's legacy GitHub issue system before Spec Kit initialization. It archives
every issue and comment once under `docs/archive/migrated-tickets/`; subsequent
classification uses only that local source. Before classifying any ticket, it
reviews every maintained regression test, runs or accepts one current passing
whole-suite result, and builds an RT-to-ticket-to-AC delivery-evidence map. A
reverse checksum then reconciles implemented ACs into `docs/ACs.md` without
historical implementation research. Ticket-linked code commits provide further
delivery evidence. The normally pre-existing `docs/ACs.md` is augmented in
identifier sequence, and an AC already present there is itself delivery evidence
even when its ticket lacks recorded test results. When no delivery evidence
remains after the complete pass, ticket state determines only the disposition:
closed scope is abandoned and open scope is undelivered. A failed suite stops
the migration after a recoverable archive checkpoint and before classification
or remote mutation.

`docs/ticket-migration.org` then records open defects, abandoned or undelivered
feature scope, unresolved classifications, and delivered tickets.
It is created immediately after archive verification and updated after each
evidence phase and ticket classification. The agent reads one archived ticket at
a time and resumes from this durable record after any automatic compaction.
Reviving abandoned scope requires a new Spec Kit specification. The near-complete
heuristic automatically treats one or two otherwise-unrecorded UT or OT results
as assumed passing when every other ticket test passed and no contrary evidence
exists. A ticket without recorded passing evidence but with any maintained
passing RT is likewise assumed delivered in full. ACs supported only by either
inference and their delivered-list entries carry a footnote distinguishing
migration inference from observed test evidence. Orphan RTs automatically
receive AC provenance through one purpose-created baseline ticket.
After the durable archive commit, every issue that was open at the snapshot is
closed without comment and the closure outcome is committed. Later agents can
understand the legacy baseline without depending on GitHub.

Provider-native permissions and the shared pre-tool guard reinforce the common
prohibitions on `rm`, `sed`, `awk`, and direct `python` or `python3` interpreter
commands. They also deny native read/search requests and direct shell references
that name a file whose exact basename is `.env`; `.env.example` and `.env.local`
remain outside that rule. This is a harness-level guard, not an operating-system
sandbox: it does not inspect files opened internally by an otherwise permitted
project command.

The installer merges its Claude and Codex entries with existing settings,
creates one SDLC-owned Copilot hook file, and updates only the matching Hermes
hook. Unrelated provider configuration and hooks remain in place. Restart the
harness after installation and approve or trust the installed hook when the
provider requires it; an untrusted hook is not an active guard.

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
