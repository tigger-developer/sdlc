# Lean Delivery Redesign

Status: Agreed process direction; composite-audit mechanics remain under review.

## Objective

Restore a proportionate delivery path resembling the useful core of SDLC v1:
one approval artefact defines the requirements, evidence, and solution outline;
one independent audit reviews that definition; implementation then follows TDD
and receives one independent implementation audit.

The redesign retains the four specialist audits. It combines them at two useful
decision boundaries instead of requiring four separate audit loops and their
associated handbacks and artefacts.

## Unified specification

One specification contains the approval material for the change. It defines
tests but does not contain or create test code.

```markdown
# Specification Summary

- **Outcome:** ...
- **Before:** ...
- **After:** ...
- **Changes:** ...
- **Unchanged:** ...
- **Edge cases:** ...
- **Decisions:** ...
- **Next step:** ...

Everything above this section break MUST accurately and completely represent
the detailed specification below it.

***

## Requirements

### R-001 - Descriptive requirement title

...

## Tests

### RT-001 - Descriptive regression test

- **Requirements:** R-001 - Descriptive requirement title
- **Procedure:** ...
- **Expected result:** ...

## Solution Design

...
```

### Specification Summary

The opening summary uses these labels, in this order:

- **Outcome**
- **Before**
- **After**
- **Changes**
- **Unchanged**
- **Edge cases**
- **Decisions**
- **Next step**

An exact `***` Markdown section break separates the summary from the detailed
specification. Everything above it must accurately and completely represent
the material below it. The summary introduces no requirement, interpretation,
decision, or exception absent from the detailed sections.

### Requirements

- Give every requirement a stable, descriptive identifier.
- State observable, falsifiable outcomes and material constraints.
- Do not reuse retired, archived, or otherwise previously assigned identifiers.
- Keep test procedures and implementation mechanisms out of a requirement
  unless the mechanism is itself part of the required public contract.

### Test definitions

- Give every test a stable, descriptive identifier.
- Categorize it as an automated regression test (`RT`), bounded one-off test
  (`OT`), or human user test (`UT`).
- Trace every test to one or more described requirement identifiers.
- Give every requirement enough credible evidence to prove it.
- Define the procedure and expected result sufficiently for later
  implementation or execution.
- Require every test to earn its place. Do not manufacture brittle automation
  or redundant tests merely to satisfy a count or category.
- Apply TDD only to automated regression tests. The specification defines the
  tests; it does not write their code.

### Solution design

The solution design is a concise technical outline, not an exhaustive
implementation plan. It records:

- affected components and boundaries;
- alignment with operator direction, the project constitution, and the
  established architecture;
- material interfaces, state changes, migrations, failures, and compatibility;
- concrete after-state examples, including actual configuration names and
  representative values where applicable; and
- significant alternatives only when they explain a material decision.

The constitution remains a durable governance and authority document. It names
the project's architecture authority but does not absorb the architecture or
feature solution design.

## Clarification

The drafting skill may ask:

- up to three initial questions needed to establish behaviour and boundaries;
- up to three test questions when operator test direction needs material
  clarification; and
- up to three solution-design questions when operator technical direction needs
  material clarification.

These are maxima, not quotas. Ask only questions whose answers materially alter
the unified specification, preferably in one concise batch for each area.

## Definition audit

Create `audit-definition` as one operator-visible skill invocation and one fresh,
isolated auditor context. It applies the existing criteria from:

- `audit-spec` to the requirements and specification as a whole;
- `audit-design` to the solution design; and
- `audit-tests` to the test definitions.

It returns one combined response. All three component judgements must pass for
the overall definition audit to pass. Findings identify the specialist audit
that produced them. The author may remediate and rerun the combined audit up to
five times. An exact provisional correction follows the existing provisional
receipt contract without consuming another model audit.

After an effective PASS, or after the fifth failed audit, return to the operator
for a hard decision checkpoint. Return earlier only when remediation requires
an operator-owned product, architecture, security, privacy, access, external
contract, or irreversible decision.

Operator approval of the definition authorizes implementation; the audit alone
does not.

## Implementation

After operator approval:

1. Write the approved automated regression tests.
2. Observe them fail for the intended reason.
3. Implement the smallest coherent production change.
4. Observe the automated regression tests pass.
5. Prepare the defined one-off and user-test evidence for later execution.

## Implementation audit

Create `audit-implementation` as one operator-visible skill invocation and one
fresh, isolated auditor context. It applies the existing criteria from:

- `audit-tests` to the implemented tests and their fidelity to the approved test
  definitions; and
- `audit-code` to the implementation change.

It returns one combined response. Both component judgements must pass for the
overall implementation audit to pass. Findings identify the specialist audit
that produced them. The author may remediate and rerun the combined audit up to
five times under the same provisional and operator-owned-decision boundaries.

After an effective PASS, execute and record the approved one-off and user tests.
Then return to the operator for delivery sign-off. If final validation causes a
material test or code change, only the affected evidence becomes non-current.

## Retained specialist audits

Do not remove or weaken `audit-spec`, `audit-design`, `audit-tests`, or
`audit-code`. They remain independently invokable for focused review, paired
development, emergency work, diagnostics, and changes that do not use the
combined delivery path.

The composite skills must consume the four specialist audit contracts as their
single sources of truth. They must not copy and independently maintain those
criteria.

## Delivery shape

```text
Clarify
  -> Unified specification
  -> audit-definition, up to five calls
  -> OPERATOR CHECKPOINT
  -> Write tests
  -> Write code
  -> audit-implementation, up to five calls
  -> One-off and user validation
  -> OPERATOR CHECKPOINT
```

This produces at most ten audit-model calls rather than four independent loops
of up to five calls. The material reduction comes from one approval artefact and
two review boundaries, not merely from a lower numerical maximum.

## Audit mechanics still to settle

Before implementing the two composite skills and runner support, decide:

- the exact combined verdict schema;
- whether a PASS explicitly reports each component result;
- how provisional component findings determine the overall verdict;
- how the runner composes the existing audit prompts without duplicating them;
- which exact artefacts and context each composite skill supplies; and
- whether any high-risk change should require a focused specialist audit in
  addition to the composite audit.
