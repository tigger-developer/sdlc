## SDLC standards

If `~/.agents/sdlc/MAIN.md` is absent or unreadable, report that exact path;
never search for another copy.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`, and
`~/.agents/sdlc/ISSUES.md` in full. Apply their defined-specification,
requirement-quality, acceptance-criteria, traceability, and identifier rules
throughout this command. Do not put implementation or test procedures into
requirements.

Read `.specify/memory/constitution.md` and its `Specification Baseline` before
drafting. When the project classification is brownfield:

- complete the brownfield context pass in `ISSUES.md` before drafting;
- read the named requirement and design authorities relevant to the requested
  change;
- consult relevant historical work records, including ticket comments or their
  equivalent, by following direct traceability references and performing a
  targeted search of the affected area;
- inspect the maintained regression tests for the affected behaviour, follow
  their requirement and ticket references, and read their assertions as
  evidence of actively protected functional behaviour;
- inspect the affected implementation for current facts and untested behaviour;
- treat approved existing requirements as the requirement baseline and use
  tests and code to establish the current implemented state;
- specify only the requested behavioural delta and its compatibility boundaries;
- record the sources consulted and what existing behaviour the delta preserves,
  changes, supersedes, or leaves unaffected;
- cite unchanged baseline requirements with identifiers and descriptors instead
  of copying them into the feature specification;
- use design authorities only for technical context, not to manufacture product
  requirements; and
- use tests and code as evidence, not as independent requirement authority.

Do not expand a brownfield delta into a standalone specification of the existing
system. Do not invent normalization, migration, persistence, workflow, fixture,
or test-oracle requirements merely to fill the core template. If a material
baseline source is missing, conflicting, or unresolved, make that clarification
explicit rather than guessing.

The resolved SDLC `spec-template` is authoritative for the specification's
structure. Preserve its opening Specification Summary and following `***`
section break, Spec Kit-compatible User Scenarios and Testing, Acceptance
Scenarios, Requirements, Success Criteria, Assumptions, and optional Key
Entities sections. Preserve Scope and, for brownfield work, Existing Baseline.

Write the detailed specification first, then derive the opening Specification
Summary from that current text. Preserve the exact Outcome, Before, After,
Changes, Unchanged, Edge cases, Decisions, Evidence, and Next step labels in
their template order. Use the summary to expose the behavioural delta and its
important boundaries for rapid operator review. It is not a second authority:
it MUST introduce no requirement or interpretation absent below the section
break, contradict none, and omit no material change, unchanged boundary,
decision, or applicable edge case. Evidence names the sources supporting the
summary and refers to `audits.md`; it never copies a mutable audit verdict into
the audited specification.

Give those sections distinct jobs:

- a user story uses no more than one or two sentences to establish actor,
  purpose, and value;
- its priority rationale is one sentence;
- its independent test names one independently observable outcome rather than a
  test procedure;
- acceptance scenarios carry the concrete behavioural examples and boundary
  conditions;
- functional requirements state the authoritative generalized rules
  demonstrated by those scenarios; and
- success criteria measure overall feature success without retelling each
  requirement.

Retain only the repetition needed to connect those views. Do not repeat detailed
behaviour in narrative merely to fill every section.

Apply the ABC presentation and semantic-spine contract in
`~/.agents/sdlc/ISSUES.md`, including its scan-friendly bullets and unbolded
acceptance-scenario signposts.

Use the fictional example at
`~/.agents/sdlc/presets/sdlc-standards/examples/spec-example.md` only to
understand structure and presentation. Never copy its requirements,
terminology, or behaviour into the project specification.

Before audit, replace the stock specification checklist expectations with these
checks:

- every segment is Accurate, Brief, and Clear;
- the Specification Summary is the opening section, is followed by an exact
  `***` section break, retains every required label in order, and accurately
  represents the current detailed specification without material omission;
- each required section performs its distinct job without unnecessary narrative
  repetition;
- user stories remain brief while acceptance scenarios carry concrete
  behavioural examples;
- every requirement is observable, falsifiable, bounded, and has a descriptor;
- the bold semantic spine captures the distinctive state, action, qualifier,
  quantity, boundary, and outcome without emphasizing surrounding context;
- multi-fact prose is split into scan-friendly bullets, and acceptance-scenario
  signposts use unbolded capitals;
- scope, baseline relationships, failure behaviour, assumptions, and unresolved
  decisions are explicit where relevant; and
- implementation details and test procedures are absent unless they are part of
  the public contract.

After specification and clarification are complete, the specification MUST
receive `audit-spec` PASS or satisfy a PROVISIONAL receipt under `AUDITS.md`.
Record the verdict and any receipt in the active feature's `audits.md`.
Planning MUST NOT begin without a current effective PASS.
Converge under the autonomous contract in `AUDITS.md` before requesting
operator sign-off.

After the applicable audit PASS, or validation when none applies, present
approval artefacts with `HTML_PREVIEW_TOOL`; otherwise use an available
non-blocking text editor or report the exact paths. Previewing is not approval
and must not stop the workflow.
