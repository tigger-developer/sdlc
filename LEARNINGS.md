# SDLC Standards: Design Learnings

## Annotations are a review inbox, not a second specification

Keep pending feedback beside the passage it concerns, but separate its semantics
from authoritative requirements and approval. Process it in batches: record the
original feedback and disposition, reconcile the specification, then remove the
resolved annotation. A clean document alone is not evidence of resolution; retain
revision-linked history separately from harness-owned audit state. Define this
workflow before fixing the tool's encoding, without inventing a parallel format.

## Verify each boundary without repeating the whole workflow

Local provider doubles prove orchestration and cache behaviour, not native
authentication, session continuity or model quality. Bounded native one-off
tests establish those adapter behaviours; an independent audit assesses the
change; deployed binary and prompt checks establish which version is installed.
Keep these claims separate. Do not repeat metered reviews merely to satisfy a
mechanical documentary condition, or turn native probes into regression-pack
entries. Record the tested revision and carry evidence forward only when the
relevant behaviour is demonstrably unchanged.

## Cache the audit contract, not just the primary file

An unchanged spec can receive a different audit when its standards, supporting
evidence, instructions or selected model change. Reuse the existing SHA-256
manifest and fingerprint the complete supplied request. Cache all valid verdicts,
including FAIL, but never confuse a provider incident with an audit result.
Returning a cached response must not consume a round or rewrite its provenance.
Unlisted ambient inputs remain outside this guarantee.

## Session continuity is an optimization, not a dependency

Provider contexts can disappear and configuration can change. Retain native
ownership and history, then recover within the same budget on an explicit
missing-session diagnostic. Authentication fallback needs a complete alternate
tuple, not a mixture of primary and fallback settings. Neither silence nor a
generic failure justifies another provider or invented evidence.

## Qualify native adapters through their real interface

A command accepting an option does not prove that its selected execution path
honours it. Native start/resume probes exposed a Hermes shortcut that bypassed
session options. Use its chat interface and native stderr identity. Where an
adapter cannot expose read-only file tools, verified text over stdin preserves
review capability without granting writes. This resends evidence on resume;
path-capable adapters retain the more economical path-based approach.

## Version the template, not a second copy of the release

A specification needs to identify the format it follows, independently of the
installed SDLC. Use `#+SCHEMA:` as the canonical template's version and copy it
into new specs. Increment once per published template or authoring-contract
revision; do not relabel older specifications without an authorized conversion.
This identifies a documented contract, not an executable schema validator.

## Readability depends on structure, not a word limit

The reviewed HTML-Preview specifications retained useful detail but exposed
dense summary bullets, long test-case paragraphs and repeated contracts.
Use outlined Synopsis categories, short bullets and selective functional
emphasis. Keep detailed outcomes in ACs, contracts in design and reproducible
expectations in tests, linking between them. Group test cases without inventing
more IDs or harnesses. Administrative prerequisites belong in Context and
Handoff, not artificial product ACs. Keep optional sections proportionate while
retaining security analysis and enough detail for a fresh delivery context.

## Templates must respect state ownership

A copied definition-gate field contradicted the harness-only audit record and
could make template-conforming specs fail audit. Keep verdicts, findings and
sessions solely in `audits.yaml`; human approval belongs in the specification,
lifecycle in `work.org` and test results in `validation.org`. Design approval
does not lift a delivery hold. Place optional decision logs last and keep dummy
records inside authoring comments so they cannot resemble real history.

## Document titles and section headings have different purposes

Repeating a specification's change title as its first section heading duplicates
the title in rendered HTML and the Org outline. Keep the change name in
`#+TITLE:` and name the opening summary section `Synopsis`. Retain its stable
anchor and summary-to-detail contract; presentation changes need not change
requirements or traceability.

## An evidence manifest is not a native permission grant

Naming an external file in the audit prompt does not authorize the provider's
file tool to read it. Translate the explicit manifest into exact read permissions
on both start and resume; retain write restrictions and native deny policy.
Permission patterns must preserve literal filenames and resolved aliases without
opening their parent directories. Local adapter tests prove the emitted contract;
native access still requires bounded live evidence.

## Auditor capability and billing are setup decisions

An inexpensive auditor is not necessarily a cheaper review: difficulty assessing
the author's work can produce unnecessary scope expansion and repeated rounds.
Select for the complexity being reviewed, then confirm the harness can run that
model and retain its session. A separate model offers a different perspective;
a different harness name alone establishes neither capability nor independence.
Authentication support and included subscription usage are separate questions.

## Reading an authority does not expand project scope

An infrastructure contract, consumed API, or SDLC instruction may govern the
current project without making its owning project part of the task. Report
external blockers; do not take over their diagnosis or design without explicit
operator scope. A filesystem sandbox can constrain writes, but cannot replace
this responsibility boundary. Instruction compliance and enforced isolation are
separate claims.

## Distinguish provider buffering from runner buffering

A final-only provider format cannot emit progress before completion; a runner
cannot flush bytes it has not received. Conversely, retaining an already received
error until process exit, then discarding it on failure, hides the cause behind
an exit status. Use a provider's streaming format where supported, decode events
incrementally, and preserve bounded failure diagnostics separately from the final
response. Heartbeats establish runner liveness, not provider progress or success.
Test unsuccessful exits and unterminated records as well as successful output.

## Audit timing must match the delivery route

Test definitions and implemented tests answer different questions. Definitions
belong with requirements and design; implemented tests belong with production
code in a shared implementation review. A separate pre-code test audit adds a
gate without replacing meaningful RED evidence.

Paired work needs a skeleton ticket before coding; emergency work needs the fix
before ticket ceremony. Both need a durable authorized scope before implementation
audit and a retrospective definition review afterwards. That review must detect
misrepresentation and material design defects without pretending to grant prior
approval or expanding the change. Documentary corrections can be remediated;
behaviour and design decisions return to the operator.

## Goal configuration is not native activation

Hands-off instructions alone do not ensure that a provider continues after a
response boundary. A delivery skill must activate an exposed native control and
verify its state, or disclose the operator action needed. Configuration parsing,
goal activation, native limit enforcement and real-task completion are separate
claims. Goal continuation never replenishes the audit allowance or replaces
operator approval. Native availability is qualified per harness, not inferred
from a shared skill file.

Budget parsing must inspect original YAML scalars before number conversion.
Permit decimal digits or comma-separated groups of three; reject periods rather
than guessing locale. Stripping a period can inflate an intended 100-token
budget to 100,000. Validate grouping before stripping commas, and bound integers
to exact JSON transport. Human-readable examples and schema defaults must agree.

## Qualify audit isolation independently of normal harness configuration

W007 begins with an opt-in, tool-free Claude safe-mode experiment. Instruction
suppression, file-access enforcement and retained-session behaviour are separate
claims; none follows merely from changing the provider or removing skills.
Keep normal personal configuration intact while qualifying an isolated invocation.

The first live probe with Claude Code 2.1.236 stopped at authentication after
redirecting runtime state to a scratch configuration directory. It returned no
audit result. Safe mode retaining authentication in an existing configuration
does not establish that a separate configuration can reuse that authentication.
Do not extract credentials or silently change billing routes to make a probe run.

A later login discrepancy was traced to a literal newline inside the quoted
configuration path, creating a different profile identity. After login to the
intended path, the synthetic audit caught its deliberate defect in 2.970 seconds;
the same-session correction passed in 2.524 seconds without repeating the
requirement. Neither synthetic instruction marker appeared in the transcript.
Only the internal structured-output tool was observed. These results qualify
the small invocation/resume experiment, not original-file isolation or the
duration and scope discipline of a realistic audit.

A subsequent native permission probe returned in 6.183 seconds: an exact-path
grant allowed the original synthetic file, while an unlisted sibling and a
symlink with an allowed name but an unlisted target were denied. The transcript
contained actual error tool results, not merely a model assertion of restraint.
This used safe mode, only Read, dontAsk, an empty read-denied working directory,
and explicit runtime-state denies. It required no custom reader. The result is
specific to the tested paths and Claude Code 2.1.236, not a universal filesystem
sandbox claim.

The subsequent resume kept the same native session identity, read changed
fixture content from disk and again denied the unlisted file and symlink in
6.087 seconds. The prompt supplied neither the replacement contents nor the
paths again. New tool results establish fresh reads and permission checks;
remembered model statements alone would not. Permission revocation and realistic
audit scope still require separate qualification before production integration.

The first real-code qualification reviewed the retained-timeout recovery delta
with original-path grants and no operational tool except Read. Its five-minute
external timeout terminated it without a structured result. All 13 observed
reads stayed in the allowlist, with no permission failures; all 22 supplied
original files retained their hashes. Isolation and a small synthetic response
do not establish practical audit latency. Retain the native session and report
the missing result rather than declaring failure of the reviewed code or
assuming the response was buffered. Finding quality remains unassessed.

## Recovery instructions need recoverable state

Telling an agent to resume after timeout is insufficient when native identity
and evidence are saved only after a final response. The harness now checkpoints
attempts before invocation and native IDs when observed. Codex and Hermes IDs
must come from their output, not the runner's generated placeholder.

Timeout is an incident, not an audit finding. It consumes a bounded attempt but
permits unchanged evidence to be resubmitted. The caller receives an actionable
recovery diagnostic; the retained auditor receives a distinct continuation
instruction. Both texts live in the YAML prompt registry. Missing identity or
exhausted bounds still stop automatic continuation. These changes address lost
continuation, not why a provider took longer than its timeout. Path manifests
neither prove remembered context nor guarantee lower audit duration.

## Retained audit context needs stable evidence identities

Copying every input into a new temporary directory on every audit added file
handling and unstable paths without establishing what the auditor remembered.
The harness now supplies original paths and compares SHA-256 manifests retained
once per recorded round in `audits.yaml`. This replaces the copied-bundle design;
it does not introduce a document cache or a purge policy.

Unchanged bytes are not proof of a previous read or retained understanding.
Resume instructions prioritize changed evidence while permitting rereads after
compaction or where a change affects an unchanged authority. Pre/post hashing
detects mutation, not prevents it, and is not an immutable snapshot. Existing
provider read-only controls remain in place. Token savings and live provider
file-access behaviour require separate bounded observation, not assertions from
the manifest alone.

## Delegate document rendering to its maintained tool

`tigger-developer/HTML-Preview` owns Markdown and Org rendering, presentation
assets, and cleanup. The SDLC pins it as a submodule and delegates installation
to its own Makefile only when `htmlpreview` is missing from PATH. This avoids
maintaining a second renderer or requiring manual preview configuration.
`HTML_PREVIEW_TOOL` remains an optional executable override so that the supplied
default does not prevent operators from choosing another viewer.
The checkout lives under `tools/`, because Go reserves `vendor/` for dependency
vendoring.

## The harness is the audit boundary

Individual audit skills encouraged accidental context proliferation and made
session ownership ambiguous. Composite gate prompts belong in a versioned YAML
registry, while the harness owns timeout, round limits, start/resume selection,
and the durable `audits.yaml` session mapping. This keeps audit policy
inspectable without deploying a separate skill for every criterion.

The external prompt must return findings only; it must not edit the project or
the audit record. The harness persists the response, and one-way migration
preserves any legacy `audits.org` content in YAML before removing the duplicate.
This keeps audit state canonical without discarding historical evidence.

Work-item numbering is similarly a project convention: new v3 work uses
`WNNN`, while migration preserves an existing project's established prefix and
never reuses a number.

This document records why the public SDLC is structured as a standards library
and how its harder rules were derived. Normative requirements live under
[`src/`](src/).

## Lean orchestration is part of correctness

SDLC v2 demonstrated that a standards library can integrate successfully with
an external specification framework, but the combined process was not usable in
routine delivery. Small changes accumulated separate specification, plan, task,
audit, constitution, and migration artefacts. Independent one-shot audits then
reloaded the same project context repeatedly. High-quality review arrived, but
delivery was delayed by ceremony, context cost, and remediation loops.

This was an observed production cost, not merely a preference for fewer
documents. More than a dozen refinements still left ordinary work running five
to ten fresh-context audits and taking roughly four times as long for the
operator to define and deliver. The final case was a small script change,
comparable to a one-line operator edit, that generated about 1,400 lines of
artefacts and ran nine audits. Five audits belonged to the test-and-build cycle,
which ultimately failed its own gate without delivering the change.

The direct financial signal was equally clear. The operator is an independent
developer who rarely exhausted the weekly allowance of a Codex subscription
costing €90 ($105) per month. After adopting Spec Kit, that allowance was
consumed in one afternoon. Continuing required an upgrade to €180 ($210) per
month, without a corresponding improvement in delivery value.

SDLC v3 retains the valuable constraints and removes the duplicate lifecycle:

- one unified specification defines requirements, tests, and solution design;
- automated tests still precede implementation;
- focused audits remain distinct criteria, but each human gate combines them;
- authors review locally before paying for an independent context; and
- one external auditor context is retained across retries rather than recreated.

The lesson is not that audits or specifications are excessive. A poor first
definition makes every downstream review expensive, while separate artefacts
make contradictions easier to introduce. A concise context-independent handoff
and bounded audit topology improve both safety and throughput.

## Delivery mode must be explicit

The former MODE DELIVER experience exposed a useful distinction that should not
be lost in a lean workflow. Definition requires an attended operator because
unresolved product and scope decisions cannot be guessed safely. Delivery has
the opposite cost profile: once a specification is signed off, a premature
agent handback wastes the operator's unattended delivery window. V3 therefore
uses one preflight question set, then continues hands-off unless the operator
selects attended. Routine reversible assumptions belong in the specification's
decision log; human decisions, genuine blockers, and exhausted audit limits
remain valid stopping conditions.

## Project-local workflow copies do not scale

Spec Kit installed workflow skills into each project. Any framework correction
then required every initialized repository to be refreshed. Stable global skills
under `~/.agents/skills`, with standards at the canonical SDLC root, eliminate
that fleet-wide update problem. Project repositories retain only durable facts,
work, specifications, audits, and validation evidence.

## Configuration must be safe for agents to read

An application's `.env` may contain secrets and shell expressions. It was the
wrong home for SDLC settings. V3 uses readable YAML: global defaults in
`~/.agents/sdlc.yaml` and explicit project facts or overrides in
`.sdlc/project.yaml`. A shell wrapper imports only allowlisted historical keys
once, then removes those exact keys while preserving unrelated content; agents
never read `.env`.

## Match the storage format to the authority

No single text format serves configuration, narrative documentation, and a
human-operated work ledger equally well. SDLC v3 assigns each format one
semantic role:

- YAML stores machine configuration in `.sdlc/project.yaml` and
  `~/.agents/sdlc.yaml`;
- Org stores the canonical work ledger and context-independent specifications;
- Markdown stores explanatory project documentation such as README, vision,
  and architecture documents; and
- Markdown with YAML frontmatter is used only when an external tool requires
  that interface.

This division keeps each fact canonical. The TODO keyword in `docs/work.org` is
the sole current lifecycle state. Org tags classify a work item, properties
hold stable metadata, and links point to requirements, design, evidence, and
history without copying their authority. A specification defines the change; it
does not repeat the work item's lifecycle state.

The Go Org ecosystem includes a parser and writer that expose headings, TODO
states, tags, properties, links, and an outline tree. That is useful for
validation and queries, but broad parse-and-render mutation would place
operator formatting and unsupported Org syntax at the mercy of a partial
implementation. Initial ledger creation therefore renders the canonical
template, and migration performs narrow structure-aware insertions. A Go Org
parser may validate the resulting supported subset after the format stabilizes;
it should not rewrite the whole document merely to insert known nodes.

Migration must honour the same single-source rule. Archived Spec Kit
specifications retain their definitions and evidence, but their obsolete
top-level lifecycle Status field is removed after the untouched source has been
preserved on the dated archive branch. A completed `docs/ACs.org` is different:
its complete requirement hierarchy is folded into `docs/work.org`, its Status
field becomes the legacy AC headline state, and the redundant source file is
removed only after the embedded copy is verified.

## Deterministic migration needs a bounded semantic fallback

Historical ledgers contain presentation and vocabulary variations that a
deterministic importer must not guess at. Expanding the parser for each observed
variant would silently turn uncertain interpretation into policy. Stopping for
manual repair, however, makes a once-only migration unnecessarily operator
intensive.

The safe middle ground is one bounded semantic repair. The initializer keeps
the deterministic importer as the acceptance boundary. Only when that importer
rejects `docs/ACs.org` does a fresh agent normalize the ledger against the
canonical template, without touching other project files. The same importer
then validates the result. A failed or still-invalid repair stops migration and
preserves the source ledger. This keeps inference visible and constrained while
making known legacy variance recoverable without hand editing.

## Semantic discovery can remain human-controlled

Configuration code can validate paths and persist lists, but it cannot reliably
decide whether a README owns product policy or merely introduces a repository.
Asking the operator to recall every authority path is equally unreliable,
especially after migration has moved or created documents.

V3 therefore separates semantic proposal from deterministic control. Authority
paths come from a bounded Git inventory and remain explicit operator choices.
Where classification needs semantic judgement, a bounded read-only agent
returns strict structured YAML that the initializer validates before use. The
model proposes or classifies; it never silently grants authority.

Technology selection has the same recall problem but a narrower, mechanically
detectable answer space. The schema now owns basename, extension, exclusion,
and implication heuristics over the bounded Git inventory. Their result is a
preselected recommendation, not hidden configuration or proof. This avoids
loading an agent and the SDLC merely to recognize familiar project files while
retaining operator confirmation for ambiguity and omissions.

Likewise, a populated global default needs no repetitive project question.
Initialization inherits it without copying it; an explicit command option
reopens those questions when a project needs local overrides.

Technology standards should distinguish an established ecosystem reach-for from
an unconditional dependency endorsement. Rust makes that boundary especially
important: mature crates can provide strong defaults, but features, transitive
dependencies, minimum compiler versions, security history, and operational
cost remain project-specific evidence. The standard therefore names common
choices by concern while retaining the same dependency review required for any
new library.

## Migration preserves history without processing dormant work

A production process cannot be replaced by pretending its existing work never
existed. V3 creates an archive branch before mutation. SDLC v1 projects retain
their lossless ticket corpus and acceptance-criteria ledger. SDLC v2 projects
retain their complete Spec Kit state under the archive branch and move active
artefacts unchanged into a project archive, but incomplete work is not
normalized until
the operator deliberately resumes it. This separates lossless preservation from
unnecessary migration-time ceremony.

Initialization and delivery branching solve different problems. Initialization
always protects the prior state on a dated archive branch and performs the
migration on a temporary branch. The configured delivery branch strategy
governs later changes only; it must not weaken or complicate the once-only
migration boundary.

## Migration readiness is authority convergence

Moving files is not enough to establish a usable migrated project. Readiness
requires the live repository to have one work and requirement authority, valid
project authority paths, no active superseded workflow, and an explicit
classification for preserved unfinished work. Archived material remains
evidence and provenance; it does not silently retain current authority.

Framework-specific authority text needs the same care as files. Exact known
obsolete clauses may be removed deterministically during migration. An unknown
reference must be preserved and surfaced for operator review rather than
deleted or reinterpreted by pattern matching.

## Historical SDLC v2 learnings

The remaining sections record decisions and observations from SDLC v2. They
are retained as design history. Where they conflict with the v3 sections above
or the current normative documents under `src/`, the current v3 material
governs.

## Separate standards from orchestration

The original framework combined engineering standards with delivery modes,
approval gates, ticket structure, bespoke audit sequencing, and closure
ceremonies.
That supplied discipline, but it also made every coding session load a large
instruction set before the task's actual context.

Different agent harnesses already have their own orchestration. Spec Kit, in
particular, owns a coherent sequence and durable artefacts:

- `spec.md` defines behaviour;
- `plan.md` records implementation design;
- `tasks.md` decomposes the work; and
- the project constitution records enduring project rules.

Reimplementing those concerns in this SDLC would create competing authorities.
The standards-only model keeps the hard-won engineering rules and lets Spec Kit
own its workflow. The four independent audits are evidence preconditions on
Spec Kit's existing stages, not a second delivery lifecycle.

## Load less to follow more

Instruction volume is not free. Rules that do not apply to the current work
compete with the rules that do. Some models respond to that competition by
following the visible process vocabulary while missing a craft constraint
buried hundreds of lines later.

The resulting loading model has three layers:

1. `MAIN.md` contains the small set required before any code work.
2. The project constitution selects applicable standards documents once.
3. Each Spec Kit command loads only the selected documents needed for that
   activity.

This is progressive disclosure without duplication. The shared document remains
authoritative; project artefacts record the selection and local decisions.

## Copy stable adapters, not evolving instructions

Spec Kit materializes project-local skills, so placing full SDLC command
instructions in a preset creates stale copies in every initialized project.
Maintaining a project inventory or provider-dependent symlinks would add another
deployment problem.

The preset now contributes only a small, stable adapter for each command. The
adapter reads the corresponding full instruction file from the exact canonical
SDLC root at invocation time and refuses filesystem discovery when it is absent.
Ordinary instruction and specification-template changes then propagate through
`make install`; project initialization is needed only to install a changed
adapter shape or other structural project material.

The same boundary removes runtime specification-template composition. The
specification command reads the canonical template directly, so an optional
preset parser dependency cannot block ordinary specification work.

## A specification is the coding boundary

"A question is not an instruction" prevents an agent from turning discussion
into mutation. "Never write code without a defined specification" supplies the
positive boundary that a process gate previously approximated.

A useful specification states observable behaviour, scope, boundaries,
constraints, and important failure cases. It does not need to dictate routine
implementation details. It must, however, settle choices that change product
behaviour, architecture, security, persisted data, access, compatibility, or
irreversible outcomes.

An emergency exception is still necessary while a project is adopting Spec
Kit, and for rare small changes where constructing the normal artefact chain
would obscure rather than clarify the request. `BYPASS-GATE-7` therefore makes
the human's same-message request a temporary specification. It is deliberately
narrow: it removes missing workflow artefacts as a blocker but does not remove
safety, evidence, scope, or engineering constraints.

Spec Kit makes this boundary durable. A plan cannot silently change a
requirement, and an implementation cannot use ambiguity as authority to invent
one.

## Paired work preserves live validation

Some outcomes cannot be specified efficiently before the operator sees them.
Visual hierarchy, typography, interaction feel, editorial presentation, and
similar work improve through short implementation and review cycles. Requiring
the complete staged Spec Kit chain for every adjustment replaces useful human
judgement with ceremony.

The paired-development track keeps the specification boundary without that
overhead. The bounded objective and each explicit instruction define the current
iteration. The operator's observed approval is first-class user-test evidence,
not a prompt to manufacture an automated browser imitation afterwards.

Automation remains valuable when it protects stable, objective behaviour. Audit
scope follows the same proportionality: review the changed implementation and
necessary context, not the untouched legacy codebase. Durable architectural,
specification, or test-design decisions still use their corresponding audits.

This is deliberately not an agent-wide mode. It exists only while the operator
is present and explicitly directing the bounded change.

## Constitutions should be generated, not hand-written repeatedly

A generic global constitution either becomes another large prompt or fails to
capture the project stack. A fully bespoke constitution per repository creates
manual work and drift.

The initializer therefore renders the stable constitutional base before an
agent sees it. The generated template lists:

- the universal standards;
- the selected technology documents;
- an optional external infrastructure owner and contract;
- the independent audit transitions;
- project-specific additions and explicit deviations; and
- bounded placeholders for evidence-derived project principles.

The full standards are referenced, not copied. The agent may add supported
project facts but cannot remove or weaken the generated base. A current rerun
asks nothing, writes nothing, and does not relaunch an agent.

## Audit skills remain independent

Spec Kit analysis checks consistency among its artefacts. That is useful but is
not the same as an adversarial review of specification, design, tests, or code.

The audit skills remain findings-only. They:

- inspect an existing artefact or implementation;
- cite evidence and the applicable requirement or standard;
- recommend a concrete correction;
- do not modify the thing they judge; and
- emit a machine-checkable PASS, PROVISIONAL, or FAIL verdict identifying
  provider and model, with exact deterministic conditions for PROVISIONAL.

Each audit runs in a fresh context. A blocking finding or malformed result
prevents PASS. A later relevant change preserves the earlier verdict as evidence
for its revision but makes it non-current for completion. Exact PROVISIONAL
conditions may be corrected and verified mechanically without another model
audit. This avoids self-review while keeping detailed standards in one place.

## Brownfield authors need the auditor's source discipline

An independent auditor may find established constraints that the author missed
when drafting from a single requirements ledger. That is an authoring-process
gap, not a reason to weaken the audit. Brownfield specification and design must
first reconcile the relevant requirement and design authorities, historical
work records, maintained regression tests and traceability, and affected
implementation. Recording what the change preserves, changes, supersedes, and
leaves unaffected makes that investigation reviewable without copying the
existing system into every feature artefact.

## Better input prevents ceremonial audit loops

Spec Kit accepts the text supplied with its specification command as the
feature description and encourages informed assumptions where detail is
missing. That works for a mature brief, but a short operator request can become
an overgrown specification containing plausible rather than authorized
requirements. A rigorous audit then spends several iterations correcting
precision the author invented instead of omissions in the requested outcome.

The answer is not a weaker audit. Specification creation now treats command
text as a brief and asks one bounded batch of ordinary-prose questions when the
outcome, boundary, validation path, or material risk is unresolved. Every
requirement must derive from operator input, an approved requirement, or a
necessary boundary directly implied by one of them. Industry convention may
prompt a question; it cannot silently become product scope.

One template supports two feature profiles. Compact specifications define one
bounded outcome without manufacturing sections or edge cases. Full
specifications carry the detail justified by multiple outcomes or material
data, security, external-contract, compatibility, or irreversible-operation
risk. The profile belongs to the feature rather than the size or age of the
project.

Test evidence follows the same credibility rule. Human validation is stronger
than an automated imitation for visual, ergonomic, editorial, and operational
judgement. A third-party double proves only the project's handling of a known
contract; it does not prove undocumented or live provider behaviour. Live
checks are bounded one-off tests with explicit authority, effects, call and
retry limits, timeout, cost, cleanup, and stop conditions. This retains strict
audits while preventing false confidence and uncontrolled external activity.

## Behaviour is not source text

Persistent tests must exercise an observable boundary. A grep proving that a
phrase exists in a prompt, document, template, or source file says nothing about
whether the product behaves correctly and couples the test to wording that may
legitimately evolve.

The useful distinction is:

- **source:** implementation material, unsuitable as proof of external
  behaviour;
- **rendered or built output:** the artefact consumed by another program,
  suitable for format-aware assertions; and
- **presented result:** post-execution or visual behaviour, requiring a browser
  or human judgement.

One-off technical review remains legitimate. It simply is not mislabelled as a
regression test.

## Test strategy, code audit, and validation are separate concerns

`audit-tests` approves the combination of automated regression, one-off, and
user tests before implementation. Only automated tests use RED/GREEN. One-off
and user tests may be the only justified evidence and are recorded in the active
feature's `validation.md` rather than converted into artificial automation.

`audit-code` assesses the implementation after automated verification. Final
one-off and user tests then exercise the audited candidate. Completion requires
both a current code-audit PASS and current passing validation results. If final
validation exposes a defect, the earlier audit and results remain historical;
only the changed implementation and materially affected tests are reassessed.

## Audit decision boundaries, not every intermediate document

The initial Spec Kit integration treated specification, clarification, design,
test definition, and implementation as separate ceremonial transitions. Even a
small change could accumulate several overlapping artefacts, independent
contexts, audit loops, and operator handbacks before producing evidence.

The useful decision boundaries are simpler. One unified specification can carry
the requirements, traced test definitions, and concise solution design needed
for an implementation decision. Its author should first challenge the whole
document using the existing specialist criteria, then pay for one independent
combined definition audit. After operator approval, TDD implementation and its
tests receive one combined implementation audit. The specialist audits remain
available, but their criteria need not force separate workflow stages.

Paired and emergency work converge on the same durable authority without
pretending they followed the normal chronology. Paired work has explicit human
direction and live validation before each slice; emergency work has an exact
temporary specification before the fix. Both reconcile the accepted or verified
result afterwards. Retrospective documentation records authority and evidence;
it must not invent advance ceremony.

## Structured data needs structured tools

Text processors are attractive because they are available everywhere and can
produce a quick-looking result. They are also easy to apply across a syntactic
boundary they do not understand.

The standards prohibit `sed`, `awk`, and Perl one-liners for agent shell work
and direct agents toward parsers and explicit editors. The same reasoning
forbids embedding SQL, YAML, HTML, shell, or another language inside ad hoc
string construction when a serializer, parameterized interface, or template
engine exists.

Direct `python` and `python3` commands are also operator-controlled. This does
not prohibit Python projects. It requires an agent to use the project's
documented environment and task entry points instead of running arbitrary code
to answer a question or manipulate files.

## Error suppression compounds failures

Patterns such as `|| true`, broad catches, discarded stderr, disabled hooks, and
unchecked return values make a transcript look successful while removing the
evidence needed to understand it. The same problem appears across languages,
so the prohibition belongs in the cross-language standards.

Expected failure needs a specific branch and a defined recovery. Unexpected
failure remains visible.

## Shell is glue

Shell works well for short coordination across stable command interfaces. It
becomes fragile when it owns application state, nested control flow, several
external systems, structured transformations, or complex cleanup.

Complexity tripwires do not automatically demand a rewrite. They require the
architecture choice to be visible in the plan before more application logic is
added. The current shell standards also reject `IFS=$'\n\t'` as a universal
"strict mode" rule because it changes correct parsing semantics; quoted
expansions and arrays provide safer command construction.

Fish demonstrates why a broad "shell" category is not always a useful context
boundary. It is intentionally not POSIX-compatible, and its variables, lists,
expansion, functions, startup files, formatter, and parser differ materially
from Bash. Loading a Bash-focused standard for Fish-only work adds irrelevant
instructions and risks invalid mechanical translation.

Technology routing should therefore follow the maintained source language, not
an umbrella label. Fish has its own automatically discoverable standard;
`SHELL.md` remains the Bash and POSIX-style authority. A mixed project selects
both explicitly.

## Reproducibility files are source inputs

Lockfiles and dependency checksum files make builds repeatable and dependencies
auditable. They are not disposable local output. In particular, `go.sum`
belongs in source control.

## Public means standalone

A public SDLC cannot rely on personal provider instructions, shell functions,
private tools, private filesystem topology, or knowledge of the author's other
repositories. Installation therefore creates one literal root:
`~/.agents/sdlc`.

Every independently invoked skill or composed command either uses that literal
path or does not require the standards. It never searches the filesystem to
infer an installation. Provider-native copies exist only for discovery.

## Installation is a comparison, not an event

An idempotent installer decides from owned source and destination files. It
prompts only for actual variance, shows only variance by default, and writes
nothing when all detected destinations match. A verbose mode may expose the
full comparison without changing the decision.

Destination-only files are deliberately preserved because provider and user
material shares some adapter directories. Omission alone therefore cannot
retire a previously deployed path safely. The v2 installer uses a finite list of
the commands and drafting skills removed by this migration. Each active legacy
path is renamed to an adjacent backup; no generalized destination deletion or
ownership inference is introduced.

## Schema identity is not release identity

`version: 3` in SDLC YAML identifies the configuration schema, not the installed
framework release. The annotated semantic-version Git tag is the sole release
authority. The installer derives a separate global `release` value from the
latest tag reachable from its build commit and updates it without replacing the
operator's other defaults.

Global skills and standards move together, so copying that release into every
project would create stale duplicate state. `.sdlc/project.yaml` records only
project facts and explicit overrides; projects follow the release installed in
`~/.agents/sdlc.yaml`.

## Deterministic retirement must tolerate presentation wrapping

A known obsolete sentence may share a paragraph with retained text or wrap
across different lines. Matching only a standalone paragraph turns harmless
presentation differences into migration blockers. Remove the exact known text
with whitespace-tolerant matching, preserve adjacent content, then surface only
unknown references for human interpretation.

## Interrupted migrations must retain enough evidence to resume

An interruption marker must not permanently block the command that created it.
When the current migration branch, dated archive branch, source generation, and
original primary branch remain unambiguous, rerun the idempotent migration steps
in place. Ambiguous state must stop rather than create replacement branches.

## Compare Org sections by identity, not presentation spacing

Alignment spaces before Org tags are presentation. A generator that compares
the complete heading line can mistake its normalized output for a missing
section on rerun. Use the stable section title or `CUSTOM_ID` as identity and
test the render-normalize-render sequence.

## Generate the invariant and ask an agent for the evidence

A generic constitution scaffold asks an agent to invent both universal
engineering discipline and project-specific principles in one pass. The result
can be lengthy while still omitting coding standards, ownership boundaries, or
independent review.

The initializer therefore renders the invariant parts deterministically:
universal standards, selected technology standards, external infrastructure
ownership, and the four audit transitions. The agent receives that baseline
only to add durable facts supported by project evidence. Detailed shared
standards remain references rather than copied text, so there is one source of
truth and less prompt dilution.

The design audit uses a lightweight scenario and trade-off review. It challenges
the design's traceability, boundaries, failure and recovery behaviour, quality
trade-offs, security, migration, operability, and testability without imposing
one architecture method or technology stack. Its portable core draws on
[ISO/IEC/IEEE 42010 architecture-description concepts](https://www.iso.org/standard/74393.html),
the [SEI Architecture Tradeoff Analysis Method](https://www.sei.cmu.edu/library/the-architecture-tradeoff-analysis-method/),
and [OWASP threat-modelling guidance](https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modelling_Cheat_Sheet.html),
then applies the selected SDLC standards as project-specific constraints.

## Why the migration used a prototype branch

Moving from an SDLC-owned lifecycle to Spec Kit was a material change in working
practice. The `spec-kit-prototype` branch made the experiment reversible while
the v2 model was validated:

- deploy it to exercise the standards-only model;
- observe context load, agent compliance, and artefact quality; and
- switch branches and redeploy if the model is not useful.

The pilot tested the boundary without pretending that documentation design alone
proved better agent behaviour. The validated model was merged for the v2.0.0
release; release tags now provide the stable rollback points.

## Presentation is part of specification correctness

A specification is both an authority document and a human approval surface. A
generic title, dense multi-claim prose, or emphasis consumed by labels makes the
change harder to identify and review even when all required sections are
present.

Org already provides structural semantics for these jobs. Headlines express
the document hierarchy and description lists label facts. Bold emphasis can
therefore be reserved for the smallest functional phrases that identify the
distinctive state, action, qualifier, quantity, boundary, or outcome. Read in
sequence, those phrases provide a compressed view without creating a second
summary that can drift.

Presentation requirements must be enforced at three points: the canonical
template demonstrates them, the authoring skill applies them, and the
specification audit rejects material violations. Treating presentation as a
late cosmetic edit is too unreliable for an artefact that requires operator
sign-off.

## Executables belong to the deployed release

A global command that links into a mutable source checkout can silently change
when that checkout switches branch or advances beyond the installed standards.
Its executable and governing documents then describe different releases.

The installer must copy public commands into the canonical deployed SDLC tree
and make global command links target those copies. The installer itself remains
a repository-internal build artefact invoked by `make install`. Command files,
links, retirements, standards, skills, and release metadata belong to one
installation plan and one confirmation. Declining that confirmation makes no
changes. The installer is recoverable and idempotent, but does not claim atomic
rollback after application begins.

## Preferred linters still require reproducible CI provisioning

Developer convenience is not a CI contract. For web projects, the preferred
tools are `tidy-html5`/`tidy` for HTML, `oxlint` for JavaScript and TypeScript,
and Biome for CSS. Their availability through Homebrew on one workstation does
not establish that a build or deployment environment can run the same checks.
CI must provision and pin the selected versions through its declared toolchain,
using a pinned Nixpkgs input or a documented Debian/Ubuntu package or source
where available. This preserves the no-casual-Node/npm rule while making the
chosen checks reproducible.

## Licence

Apache License 2.0. See [`LICENSE`](LICENSE).
