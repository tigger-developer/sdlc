---
title: Specification Standards
version: 6
last-updated: 2026-10-31
---

# Specification Standards

This document governs requirements and acceptance criteria. It does not define
an issue lifecycle or a delivery workflow.

Follow `~/.agents/sdlc/ORG-SCHEMA.md` for machine-checkable AC/test headings and
their matching execution records. Keep test definitions here, results in
`validation.org`, and all audit results solely in `audits.yaml` under
`AUDIT-RECORD.md`. Label self and independent methods explicitly; do not maintain
parallel review documents.
Reviewing a specification does not authorize adopting or executing its workflow.

## Defined specifications

A specification is defined when it identifies:

- the user or system outcome;
- behaviour within and outside scope;
- inputs, outputs, state changes, and important interfaces;
- failure, boundary, security, privacy, and accessibility behaviour where
  relevant; and
- acceptance criteria that can distinguish compliance from non-compliance.

Open questions that can materially change those facts must be resolved in the
specification before implementation. Routine, reversible implementation details
may remain to the plan.

## Brownfield context pass

Before drafting a brownfield specification, examine the affected area rather
than treating one requirements document as an exhaustive system description.
Consult:

- the current and historical requirement authorities named by the project;
- the current design and architecture authorities;
- relevant tickets, comments, decisions, or equivalent work records found
  through direct traceability references and a targeted search;
- the maintained regression test pack and its requirement or ticket
  traceability; and
- the affected implementation as evidence of current facts and untested
  behaviour.

Record a concise baseline relationship in the active specification: the sources
consulted and the existing behaviour it preserves, changes, supersedes, or
leaves unaffected. Do not copy the baseline into the delta specification.
Tests and code are implementation evidence, not authority to approve or change
requirements. Make missing, conflicting, stale, or unresolved sources explicit
rather than silently choosing one.

## Requirement quality

Each requirement must be:

- **observable:** expressed as system behaviour or a durable constraint;
- **falsifiable:** evidence can show whether it holds;
- **atomic enough to assess:** unrelated outcomes are separate requirements;
- **technology-neutral where possible:** implementation belongs in the plan;
- **bounded:** actors, conditions, exclusions, and limits are explicit; and
- **traceable:** its origin and later amendment remain discoverable.

Avoid vague terms such as "fast", "secure", "user-friendly", or "appropriate"
unless the specification supplies a measurable or reviewable meaning.

## ABC presentation contract

A specification is an approval and working artefact. Every section, paragraph,
and bullet must be **Accurate, Brief, and Clear**. Failure of any one quality
means the specification is not ready for audit or approval.

- **Accurate:** agree with the request and cited authorities; distinguish
  requirements, evidence, assumptions, and unresolved decisions.
- **Brief:** give every section a distinct purpose; retain only the
  repetition needed to connect user value, behavioural examples, authoritative
  rules, and feature-level measurements.
- **Clear:** use direct, falsifiable language; separate independently failing
  conditions and name actors, states, boundaries, and outcomes precisely.

Prefer scan-friendly bullets to a paragraph containing several facts. Keep one
primary fact per bullet. In `spec.org`, use native Org headlines for document
hierarchy and native description lists for labelled facts. Do not simulate
either with bold text. Keep the fixed major sections at level one, use level two
for addressable subsections or records, and use deeper levels only for genuine
children.
Within Synopsis, use the eight level-two categories from the template with short
bullets beneath each. One category may need several bullets, but each bullet
has one principal fact. Do not pack paragraphs into description-list labels.

Within substantive prose and bullets, bold the semantic spine: the smallest
words or phrases that carry the distinctive state, action, qualifier, quantity,
boundary, or outcome. Include a verb or modifier when it conveys the important
change; do not default mechanically to nouns. Read in order, the bold fragments
alone must communicate the statement's functional meaning. Include enough of
the subject or context to make that bold-only reading understandable; a
headline or preceding line does not replace it. In Org, use native `*bold*`
syntax. Do not bold whole sentences, generic labels, identifiers, modal verbs,
or every occurrence of a term. Use the smallest sufficient set of fragments; if
most of a sentence is bold, reduce it until the emphasis is selective again.

In acceptance scenarios, put `GIVEN`, `WHEN`, `THEN`, and `AND` in unbolded
capitals on separate lines. Their position and capitalization provide the
structural emphasis; reserve bold text for the functional concepts.

Keep each user story to one or two sentences establishing actor, purpose, and
value. Acceptance scenarios carry concrete behavioural examples. Functional
requirements state the authoritative generalized rules demonstrated by those
scenarios. Success criteria measure overall feature success rather than
retelling individual requirements.

For example, an Org specification may contain:

```org
- A report *schedule* *declares* an *output format* and may declare a
  *recipient list*.
- An *empty recipient list* *disables delivery* without disabling report
  generation.
- Repeating a scheduled run with the *same execution identifier* creates *no
  duplicate report* or *notification*.
```

Prefer concise authoritative requirements and boundaries over narrative
restatements of the same behaviour. Length follows the complexity of the
change; brevity must not remove a material case, constraint, or source
relationship.

## Unified specification

Normal delivery uses `tickets/<ticket>-descriptor/spec.org`. A new work item
uses W, a defect D and a spike S; a defect's later fix specification retains
its D identity and lives beside `defect.org`. A spike brief is `spike.org`.
The specification contains, in order:

- a scan-friendly summary covering outcome, before, after, changes, unchanged
  boundaries, edge cases, decisions, and next step;
- context, scope, authorities, constraints, decisions, and blockers;
- acceptance criteria;
- traced AT, RT, UT, and OT test definitions, including at least one meaningful
  UT under TESTING.md;
- edge cases;
- solution design and architecture impact, including security; and
- a context-independent delivery handoff.

Use the canonical `~/.agents/sdlc/templates/v3/spec.org`. Draft the detail first
and derive the opening summary from it. The summary introduces nothing absent
below its section break, contradicts nothing, and omits no material change,
unchanged boundary, decision, or edge case. A mismatch blocks the definition
gate.

Set `#+TITLE:` to the work identifier followed by a concise, change-specific
title. Generic titles such as `Spec sheet`, `Specification`, `Change`, or
`Untitled` fail the definition gate.
Name the opening level-one section `Synopsis`, not a repetition of the document
title. Retain its `CUSTOM_ID: summary` so existing summary links remain stable.

### Template version and ownership

The canonical template's `#+SCHEMA:` value is its **positive-integer version**.
New specifications copy that value. Maintainers increment it once per published
revision that changes the template or its authoring contract, independently of
the SDLC release. Multiple edits in one pending release share that increment.
It identifies a documented format, not an executable validator or standards pin.
Do not silently stamp or upgrade existing specs; an authorized conversion must
actually conform to the chosen template revision and preserve history.

- **Audit state:** authorized reviewers or callers update `audits.yaml` under
  `AUDIT-RECORD.md`; no
  `DEFINITION_GATE` or other copied audit status belongs in `spec.org`.
- **Author review:** separate internal records; independent results retain their
  actual method and separation. Neither writes harness state.
- **Human approval:** only explicit sign-off sets `OPERATOR_SIGNOFF` to
  `APPROVED YYYY-MM-DD`; add scope only where limited, not personal attribution
  or conversation narration. It is not an audit verdict or permission to lift a hold.
- **Lifecycle:** the work item's TODO keyword in `work.org`.
- **Validation:** actual test results in `validation.org`.
- **Review feedback:** pending annotations in `spec.org`; processed annotations
  and their dispositions in the work item's `feedback.yaml`, not audit state.

Resolve older metadata through authorized edits, not an automatic rewrite of
signed-off specifications. Do not treat an old copied gate value as current
audit evidence.

### Annotation review

Annotations are explicitly identified review feedback, not specification text,
approval or automatically executable instructions. Their exact encoding belongs
to the annotation tool's documented format; do not invent markers or classify
ordinary Org comments, quoted examples or mentions of annotations as feedback.
If the encoding cannot be interpreted safely, preserve it and ask for clarification.

Before audit or sign-off, `spec.org` must contain no annotation records or blocks,
including resolved or empty ones. Removing feedback without preserving its
disposition is not resolution. Supply `feedback.yaml` as audit evidence when
present; it must not duplicate findings, verdicts or session state from `audits.yaml`.
Audit the revised definition as one package, not once per annotation. Material
changes to an approved specification require renewed operator approval.

Processed feedback evidence retains each annotation's stable identity, author,
original text, referenced passage and reviewed revision, plus its disposition
(applied, answered or rejected), reason, date, affected sections and resulting
specification hash. Earlier entries remain available. Missing provenance must
be disclosed, not fabricated; unresolved or conflicting dispositions block
acceptance. These are evidence criteria, not authority for a reviewer to edit.

### One detailed home

Keep the core sections; adapt subordinate structure to the change:

- **ACs** own required system outcomes and durable product constraints.
- **Design** owns exact after-state interfaces and mechanisms, preferably as
  addressable examples or contract tables. Link stable architecture, do not copy it.
- **Tests** own actions and explicit observable expectations, traced to ACs.
  For dense case groups, retain shared setup and one test ID; use named child
  cases or a compact case/outcome table with AC mappings. Formatting does not
  require more test IDs or another harness. "See design" alone is insufficient.
- **Edge cases** add non-obvious boundary outcomes and link existing ACs/tests,
  rather than repeating whole contracts.
- **Handoff** owns entry conditions and required reading, linking prerequisites
  under Context / Constraints and dependencies. Administrative sequencing is
  not itself a product AC or a reason to invent another product test.

The Synopsis and reproducible test expectations are intentional repetitions.
The summary need not repeat every protocol constant or fixture, but must retain
every material outcome and boundary. Links supplement, not replace it.
Do not impose a word limit, create a second summary artefact, or split work only
to shorten a document.

Keep security impact mandatory. Include other subordinate design/quality
sections only when relevant, without repetitive N/A entries. Keep the optional
hands-off log at the end, initially folded; its example belongs in an authoring
comment, not a live-looking dated record. Preserve actual decision history.

## Acceptance criteria

Acceptance criteria describe required system states, not test procedures.

| Acceptance criterion | Verification |
|---|---|
| The CLI exits non-zero and identifies the invalid field when configuration is malformed. | Run it with malformed configuration and inspect status and stderr. |
| Repeating installation with unchanged sources performs no writes and asks no deployment question. | Install twice and compare the second run's output and filesystem effects. |

The left column remains true regardless of test framework. The right column may
change as tooling evolves.

Do not put commands, test functions, fixtures, clicks, source-code searches, or
implementation details in an acceptance criterion unless that mechanism is
itself part of the public contract.

## Coverage of compound requirements

A requirement containing several independent conditions needs evidence for
each condition. Split it when the conditions can fail independently or need
different verification. Do not infer complete coverage from one happy-path
example.

For each requirement, consider:

- normal and alternate paths;
- empty, minimum, maximum, malformed, and missing input;
- permission and authentication boundaries;
- partial failure and interrupted operations;
- repetition, concurrency, and idempotence;
- compatibility and migration behaviour; and
- accessibility and human judgement where automation is insufficient.

## Bugs and regressions

A bug is a mismatch between observed behaviour and a requirement.

- If an existing requirement covers the behaviour, cite it with a descriptor
  and add regression evidence against that requirement.
- If no requirement covers the behaviour, amend the normal specification or
  establish the paired/emergency working specification before implementation.
  Do not manufacture a parallel requirement that conflicts
  with the original feature specification.
- Treat the human's observation as evidence to investigate. A code-reading
  hypothesis does not disprove it.
- Reproduce or otherwise isolate the failure before stating its cause.

Do not contradict a reported observation merely because the current code or a
different environment suggests it should be impossible. Treat the discrepancy
as diagnostic evidence: both observations may be true under different paths,
configuration, state, or timing.

## Ticket identities and storage

All new tickets share one numeric namespace, regardless of prefix. Use W for
work/specifications, D for defects and S for spikes. W123 followed by a defect,
a spike and a work item produces D124, S125 and W126. Never restart or maintain
a separate counter for each prefix. Choose one more than the highest number
ever reserved across all types, including archived and abandoned entries, not
merely the last displayed line in `docs/work.org`.

Store tickets in one flat root: `tickets/W123-description/spec.org`,
`tickets/D124-description/defect.org`, `tickets/S125-description/spike.org`.
Do not create a defects or spikes subdirectory. All types remain in the same
primary-branch `docs/work.org` register with their existing lifecycle and type
tags. Reserve IDs under DELIVERY's publication protocol. Ticket identity stays
stable when a defect gains a fix specification; do not allocate a second number.

Existing IDs are historical identities, including unprefixed or older W-prefixed
spikes/defects. Directory migration does not change them. New allocations follow
W/D/S and include all historical numbers when determining the next number.
A project still containing root `specs/` must run `sdlc-init` before normal work.

## Identifiers and human communication

Identifiers are optional unless the project or tooling requires them. Once
assigned, an identifier is permanently reserved within its established
namespace. Never reuse or renumber it after deletion, archiving, abandonment,
retirement, or supersession.

For a sequential namespace, inspect current artefacts, archives, and version
history before allocating an identifier. Use a value greater than the highest
value ever assigned, not the lowest currently available value. Preserve gaps.
Where practical, retain a retired or superseded entry as durable lineage rather
than deleting it.

For full-SDLC work, primary's `docs/work.org` is the canonical ticket register.
Reserve ticket numbers through `DELIVERY.md` before treating them as assigned;
a feature-branch snapshot is never an allocation authority. This workflow rule
does not apply to an independent auditor's read-only standards review.

Never give a human an identifier without its descriptor. Write, for example,
`FR-012 - repeated installation is a no-op`, not merely `FR-012`. The same rule
applies to test IDs, ticket numbers, findings, and commit hashes.

## Change control

When implementation reveals a missing or conflicting requirement, update the
specification before relying on the new interpretation. Preserve the reason for
material changes and identify superseded behaviour rather than silently
rewriting history.

A plan or task list may refine how a requirement is delivered. It may not alter
the required outcome without a corresponding specification change.
# Canary

Append " ISSUES" (with a leading space) to the canary.
