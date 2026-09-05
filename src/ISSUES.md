# Specification Standards

This document governs requirements and acceptance criteria. It does not define
an issue lifecycle or a delivery workflow.

## Defined specifications

Code requires a durable specification agreed for the work. A specification is
defined when it identifies:

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
- **Brief:** give every Spec Kit section a distinct purpose; retain only the
  repetition needed to connect user value, behavioural examples, authoritative
  rules, and feature-level measurements.
- **Clear:** use direct, falsifiable language; separate independently failing
  conditions and name actors, states, boundaries, and outcomes precisely.

Prefer scan-friendly bullets to a paragraph containing several facts. Keep one
primary fact per bullet. Use ordinary Markdown headings without additional
emphasis. Within prose and bullets, bold the semantic spine: the smallest words
or phrases that carry the distinctive state, action, qualifier, quantity,
boundary, or outcome. Include a verb or modifier when it conveys the important
change; do not default mechanically to nouns. Read in order, the bold fragments
should give an accurate compressed summary of the statement. Do not bold
context already supplied by its heading or preceding line, whole sentences,
generic labels, identifiers, modal verbs, or every occurrence of a term.

In acceptance scenarios, put `GIVEN`, `WHEN`, `THEN`, and `AND` in unbolded
capitals on separate lines. Their position and capitalization provide the
structural emphasis; reserve bold text for the functional concepts.

Keep each user story to one or two sentences establishing actor, purpose, and
value. Acceptance scenarios carry concrete behavioural examples. Functional
requirements state the authoritative generalized rules demonstrated by those
scenarios. Success criteria measure overall feature success rather than
retelling individual requirements.

For example:

- A report **schedule** **declares** an **output format** and may declare a
  **recipient list**.
- An **empty recipient list** **disables delivery** without disabling report
  generation.
- Repeating a scheduled run with the **same execution identifier** creates **no
  duplicate report** or **notification**.

Prefer concise authoritative requirements and boundaries over narrative
restatements of the same behaviour. Length follows the complexity of the
change; brevity must not remove a material case, constraint, or source
relationship.

## Specification summary

Every Spec Kit specification begins with a `Specification Summary`, followed by
an exact `***` Markdown section break before metadata and detailed sections. It
uses these labels in order:

- **Outcome:** the user or system result in one sentence.
- **Before:** the affected current behaviour or relevant greenfield absence.
- **After:** the required observable behaviour.
- **Changes:** precise behaviour, interface, data, or constraint differences.
- **Unchanged:** preserved behaviour, compatibility boundaries, and exclusions.
- **Edge cases:** applicable empty, missing, invalid, limit, repetition,
  concurrency, partial-failure, security, privacy, accessibility, and
  compatibility behaviour.
- **Decisions:** resolved assumptions and any unresolved decision.
- **Evidence:** requirement and baseline sources plus a reference to the
  feature's external `audits.md`; never a copied mutable verdict.
- **Next step:** clarification still required or what operator sign-off permits.

Use short, keyword-anchored bullets with one principal fact per bullet. Draft
the detailed specification first and derive the summary from it. The summary is
a presentation layer within the same artefact, not a separate authority. It
must introduce no requirement or interpretation absent below the section break,
contradict none, and omit no material change, unchanged boundary, decision, or
applicable edge case. A mismatch is a blocking specification-audit finding.

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
- If no requirement covers the behaviour, amend the specification before
  implementation. Do not manufacture a parallel requirement that conflicts
  with the original feature specification.
- Treat the human's observation as evidence to investigate. A code-reading
  hypothesis does not disprove it.
- Reproduce or otherwise isolate the failure before stating its cause.

Do not contradict a reported observation merely because the current code or a
different environment suggests it should be impossible. Treat the discrepancy
as diagnostic evidence: both observations may be true under different paths,
configuration, state, or timing.

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
