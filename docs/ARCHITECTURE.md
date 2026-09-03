# SDLC Standards Architecture

This document explains how the public standards library is composed, installed,
and integrated with Spec Kit. Documents under [`src/`](../src/) are normative;
this document is explanatory.

## Boundary

The repository owns engineering standards:

- universal coding constraints;
- specification and acceptance-criteria quality;
- testing and evidence;
- application vulnerability checking and exception governance;
- explicitly selected paired-development evidence;
- cross-language and stack-specific implementation standards;
- Git and documentation standards;
- findings-only audit and advisory skills; and
- provider-integrated command and sensitive-file safeguards.

It does not own agent-wide delivery modes, ticket states, or a parallel
autonomous lifecycle. Spec Kit or the adopting project owns those concerns. The
preset defines audit evidence, bounded autonomous remediation, and operator
handback requirements within Spec Kit's existing stages. The one additional
path is explicit paired development, where the operator remains present and
validates bounded iterations.

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
|   |-- SECURITY.md
|   |-- AUDITS.md
|   |-- PAIRING.md
|   |-- CODING.md
|   |-- GIT.md
|   |-- DOCUMENTATION.md
|   |-- technologies/
|   |-- presets/sdlc-standards/
|   |-- prompts/project-init/
|   |-- prompts/audits/
|   |-- templates/project-init/
|   `-- templates/migration/
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
|-- SECURITY.md
|-- AUDITS.md
|-- PAIRING.md
|-- CODING.md
|-- GIT.md
|-- DOCUMENTATION.md
|-- technologies/
|-- presets/sdlc-standards/
|-- prompts/project-init/
|-- prompts/audits/
|-- templates/project-init/
|-- templates/migration/
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
    +-- application security --> SECURITY.md
    +-- audited phase ---------> AUDITS.md
    +-- paired development ----> PAIRING.md
    +-- implementation --------> CODING.md
    +-- source control --------> GIT.md
    +-- technical docs --------> DOCUMENTATION.md
    `-- selected stack --------> technologies/GO.md, technologies/HUGO.md, and so on
```

The project constitution references the selected standards. `SECURITY.md` is a
universal reference; technology documents select its concrete scanners. An
agent loads only the entries relevant to the current activity. The shared files
remain the single source of truth; their contents are not copied into every
feature artefact.

## Deterministic project initialization

`sdlc-project-init` discovers technology standards from the installed
`technologies/` directory, resolves command-line, project, and user defaults,
and renders
`.specify/templates/overrides/constitution-template.md`. This documented Spec
Kit override is resolved before preset composition, so constitution generation
does not require a preset manifest parser. Adding a technology document makes
it available without changing the initializer.

Before initialization, the explicit
`migrate-legacy-acs-to-sdlc-v1` skill performs the semantic readiness migration.
It archives every open and closed issue with comments once under
`docs/archive/migrated-tickets/`, uses only that local snapshot thereafter, and
creates `docs/ticket-migration.org` immediately as the incremental progress,
evidence, and disposition map from the deployed canonical Org template. Its
heading hierarchy makes report sections, tickets, ticket details, and individual
ACs independently foldable; bullets remain limited to flat summaries. Its
embedded native-syntax guide and Pandoc parse check keep the durable state
machine-readable. The
maintained regression harness and at most one whole-suite run reconcile
`docs/ACs.md`. Every maintained RT is reviewed and mapped to tickets and ACs
before ticket classification begins. Ticket-linked commits provide further
delivery evidence, while marked near-complete and live-RT heuristics preserve
the distinction between observed and inferred evidence. Documentation
reconciliation groups ACs by product area and consults code only when the
documented authorities remain ambiguous.
The normally pre-existing AC ledger is augmented in identifier sequence;
presence in that ledger is itself delivery evidence when the originating ticket
omits test results.
If the supported whole-suite run fails, the skill checkpoints only the verified
archive and incremental Org record, then stops before classification or GitHub
mutation. One-ticket-at-a-time processing treats the archive as external memory;
automatic compaction resumes from the durable Org record without repeating
completed work.
Targeted corrections and implementation-plan archival join one pre-closure
commit. The skill then closes every issue that was open in the snapshot and
commits the closure record.

Before rendering a brownfield constitution, the initializer performs only the
remaining mechanical authority update. It:

- applies the file-backed, marked legacy prefix to `docs/ACs.md`, replacing
  only an older managed prefix when the resource changes;
- shows that path's Git diff and obtains operator confirmation before an
  isolated commit; and
- performs no write, prompt, agent launch, or commit when the migration is
  already current.

Vision, architecture, and README documents remain active and outside the
mechanical legacy migration. The initializer never modifies them; adaptation
and implementation-plan archival belong to the pre-migration skill.

The initializer's mechanical update never invokes an agent. Unexpected
overlapping changes or content that cannot be transformed by the fixed rules
are reported rather than guessed. Declining the commit leaves the reviewed
working-tree change in place and stops constitution generation.

### Project-initializer resources

User-facing blocks, templates, and prompts are source files under `src/`; they
are not assembled from prose embedded in Go. The installer deploys these files
under the canonical SDLC root, and `sdlc-project-init` performs only explicit
template-field substitution.

The legacy acceptance-criteria notice is a managed prefix ending at a line that
contains only `***`. Its stable marker identifies the prefix independently of
the document's own headings or content. The initializer treats the remainder of
`docs/ACs.md` as opaque bytes:

- if no managed prefix exists, prepend the current block;
- if the prefix hash matches the current template, write nothing; and
- if the marker exists at the start but the hash differs, replace only the
  prefix through its delimiter.

A managed marker outside the prefix is an error rather than permission to
duplicate or reorganize document content.

The generated template is editable pre-ratification scaffolding. It contains
the fixed standards proposal, universal standard references, selected
technology references, an optional external
infrastructure contract, mandatory independent audits, a fixed specification
baseline selected as greenfield or brownfield, and bounded placeholders for
project-specific principles. The brownfield structure separates current and
historical requirement authority, historical work context, design authority,
regression evidence, and source precedence. A build made at an exact SDLC
release tag records that tag;
other clean versioned builds record their source commit. Modified or unversioned
builds leave an explicit ratification TODO. The initializer commits only the
generated scaffold, leaving unrelated staged and working-tree changes
untouched. The selected agent harness receives that template only after
deterministic rendering and the isolated Git checkpoint. After the harness
exits, the initializer verifies that the candidate retains the complete
Engineering Standards, Specification and Evidence, and Mandatory Independent
Audits sections from that exact rendered scaffold. Project-specific proposed
clauses remain editable; only the human may authorize changing shared governance
before ratification.
Ratification makes `.specify/memory/constitution.md` the sole governance
authority; later amendments use that document directly rather than reapplying
the initialization template.

The current Sync Impact Report is one compact HTML-comment line at the top of
the constitution. Ratification and amendments replace that line rather than
accumulating report history in a separate file. The project constitution
template is an initialization artefact and may be removed after first
ratification; Git retains its provenance.

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
| `speckit.specify`, `speckit.clarify` | Universal, audit, and specification standards |
| `speckit.plan` | Universal, audit, coding, testing, Git, documentation, and selected stack standards |
| `speckit.tasks` | Universal, audit, testing, documentation, and profile-specific standards |
| `speckit.analyze` | Universal, audit, specification, and testing standards |
| `speckit.implement` | Universal, audit, and only the profile entries needed by current tasks |
| `speckit.converge` | Universal, audit, coding, testing, and selected profile entries |
| `speckit.taskstoissues` | Universal identifier and source-of-truth rules |

Spec Kit copies preset material into project state and materializes composed
commands for the active integration. The initializer invokes the chosen harness
for the constitution operation only when the rendered scaffold or project
selection changes after any required brownfield documentation migration.

The `--integration` selected during `specify init` controls the project-local
agent adapter. It does not launch an agent and does not change the provider-
neutral specification artefacts.

For brownfield authoring, the command preambles require a bounded context pass
before drafting. Specification and planning consult the relevant requirement
and design authorities, historical work records, maintained regression tests
and their traceability, and affected implementation. The authored artefact
records what it preserves, changes, supersedes, and leaves unaffected. Audit
skills independently check that source coverage without requiring the baseline
to be copied into each delta artefact.

## Paired development path

Paired development is selected explicitly for a bounded change whose outcome is
refined through live operator review. It is not another autonomous lifecycle.
The session objective and each explicit iteration instruction supply the
specification boundary for that slice, while the constitution and selected
standards continue to apply.

The path separates three concerns:

```text
explicit instruction -> reviewable change -> operator validation
                                      |              |
                                      v              v
                             objective checks   user-test ledger
                                      \              /
                                       closure review
```

The provisional ledger becomes durable user-test evidence only after one final
operator confirmation. Automation remains proportional and must add evidence
rather than imitate human visual judgement. Auditing is likewise proportional:
one change-scoped code audit for material implementation, with specification,
design, or test audits only when corresponding durable artefacts exist.

This path permits artefact-native web work without treating CSS selectors or
HTML text as the specification. Explicit instructions govern the requested
outcome; Markdown, YAML, CSS, templates, and rendered pages carry the content,
configuration, implementation, and reviewed presentation without requiring a
duplicative prose design.

## Skills and retired paths

Audit skills are thin adapters to `sdlc-audit`. The runner composes the
audit-specific prompt, the common audit contract, judged artefact contents, and
only the exact context files selected by the authoring agent. It starts one
non-resumed Hermes process in an empty temporary working
directory, so authoring conversation and project-local instruction discovery
are not inherited. Every audit emits its name, provider, model, and exact PASS,
PROVISIONAL, or FAIL verdict. Blocking findings require
FAIL, exact mechanical conditions permit PROVISIONAL, and advisories may
accompany any verdict. A later relevant change preserves the earlier verdict as
revision-specific history but requires a fresh audit of the delta and necessary
context unless it exactly satisfies a PROVISIONAL condition and carries its
evidence receipt.

Ordinary context files are restricted to the project, canonical SDLC, and
operating-system temporary directories. `--external-context FILE` admits only
the exact named external authority; it grants no directory tree. Automated
regression tests inject a fake harness. A live Hermes and hosted-model call is a
metered one-off test and is excluded from `make test`, CI, scheduled automation,
and persistent regression targets.

The main authoring context owns convergence within each phase. It remediates
current-phase blockers and dispatches a fresh auditor for at most five attempts
without intermediate operator handback. It stops earlier for a signed-off
upstream change or human-controlled decision. PASS, the fifth FAIL, or that
earlier blocker produces one consolidated handback for operator sign-off. A
satisfied PROVISIONAL receipt is effective PASS for this purpose.

```text
specification -- audit-spec PASS --> plan and design
plan/design  -- audit-design PASS -> tests and tasks
test design  -- audit-tests PASS --> implementation
implementation -- audit-code PASS -> one-off and user tests
current validation PASS -----------> completion or convergence
```

Test design creates `PENDING` entries in the active feature's `validation.md`
for selected one-off and user tests, and `audit-tests` approves that strategy.
`audit-code` assesses the implementation after automated verification. Final
one-off and user tests then validate the audited candidate. Completion requires
both a current code-audit PASS and current passing validation results.

If final validation exposes a defect, a corrective code change preserves the
earlier audit and results as revision-specific history but makes affected
evidence non-current. A fresh code audit may review the corrective delta plus
necessary context, and only materially affected tests are repeated.

In this diagram, PASS includes effective PASS produced by satisfying every exact
condition in a PROVISIONAL verdict. The author records the audited revision,
corrected revision, deterministic evidence, and absence of additional changes
instead of dispatching another model audit.

Verdicts are retained in the active feature's `audits.md`. The shared parser
fails closed on missing or malformed verdicts, inconsistent finding
classifications, or an audit, provider, or model identity different from the
requested configuration. A valid FAIL passes through as a report. Project
`.env` values override user defaults from `~/.agents/.env`. Hermes is the sole
audit harness in this release; another configured harness warns and falls back
to Hermes, which receives provider and model explicitly. Each child has a
minimal environment and a 15-minute runtime budget and hard timeout. Advisory
skills load context, diagnose, recommend, or summarize.

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

The Claude, Codex, Copilot, and Hermes adapters register the shared pre-tool
guard without claiming unrelated provider configuration. Claude also receives
its native recursive `.env` read denial. The shared guard rejects native
read/search inputs and direct shell references whose exact path basename is
`.env`, while leaving `.env.example` and `.env.local` outside the rule. This is
a harness boundary; an operating-system sandbox is required to govern files
opened internally by an otherwise permitted program.

## Extending the standards

| Change | Location |
|---|---|
| Universal engineering rule | `src/MAIN.md` |
| Cross-language implementation standard | `src/CODING.md` |
| Requirement or acceptance-criteria standard | `src/ISSUES.md` |
| Testing standard | `src/TESTING.md` |
| Application-security and vulnerability-gate standard | `src/SECURITY.md` |
| Independent audit and phase-convergence contract | `src/AUDITS.md` |
| Paired-development contract | `src/PAIRING.md` |
| Technology standard | A focused file under `src/technologies/`, plus a route in `MAIN.md` |
| Spec Kit command selection | `src/presets/sdlc-standards/` |
| Project-initializer prompt | `src/prompts/project-init/` |
| Project-initializer template or managed block | `src/templates/project-init/` |
| Findings-only reusable review | `skills/<name>/SKILL.md` |
| Provider command or file-read safeguard | `hooks/` plus the relevant installer adapter |

Keep `MAIN.md` concise. A rule belongs there only when every coding task needs
it before progressive routing. Put detailed craft guidance in a selected
standards document and reference that document from the preset rather than
duplicating the rule.
