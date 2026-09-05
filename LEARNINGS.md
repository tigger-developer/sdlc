# SDLC Standards: Design Learnings

This document records why the public SDLC is structured as a standards library
and how its harder rules were derived. Normative requirements live under
[`src/`](src/).

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
and [OWASP threat-modelling guidance](https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modeling_Cheat_Sheet.html),
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

## Licence

Apache License 2.0. See [`LICENSE`](LICENSE).
