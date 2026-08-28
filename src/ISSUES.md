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

Identifiers are optional unless the project or tooling requires them. Once a
published identifier is used for traceability, do not silently reuse or
renumber it.

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
