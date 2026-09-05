# Lean Delivery Redesign

Status: Agreed process direction; combined-verdict details remain under review.

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

Before invoking an independent auditor, the author reviews and corrects the
unified specification in its existing context against:

- `audit-spec` to the requirements and specification as a whole;
- `audit-design` to the solution design; and
- `audit-tests` to the test definitions.

This author preflight uses context already loaded for drafting. It is not an
independent audit, produces no audit verdict or separate artefact, and does not
count against the external-attempt limit. Its purpose is to correct avoidable
problems before paying the context and model cost of independent review.

Create `audit-definition` as one operator-visible skill invocation. Its first
external attempt starts one fresh, isolated, phase-scoped auditor session and
applies the same three specialist contracts. It returns one combined response.
All three component judgements must pass for the overall definition audit to
pass. Findings identify the specialist audit that produced them.

If the independent audit fails, the author assesses its findings, remediates
those supported by the specification and authorities, repeats the author
preflight, and resubmits the complete revised candidate to the same independent
auditor session. Do not accept an out-of-scope or contradictory finding merely
because an auditor emitted it. Permit at most five independent attempts. An
exact provisional correction follows the existing provisional receipt contract
without consuming another model audit.

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

Before invoking an independent auditor, the author reviews and corrects the
implemented tests and code in its existing context against:

- `audit-tests` to the implemented tests and their fidelity to the approved test
  definitions; and
- `audit-code` to the implementation change.

As in the definition phase, this is an author preflight rather than independent
audit evidence. It produces no separate record and does not consume an external
attempt.

Create `audit-implementation` as one operator-visible skill invocation. Its
first external attempt starts a new fresh, isolated, implementation-scoped
auditor session. It returns one combined response. Both component judgements
must pass for the overall implementation audit to pass. Findings identify the
specialist audit that produced them.

After a failure, remediate supported findings, repeat the author preflight, and
resubmit the complete current test and code candidate to the same implementation
auditor session. Permit at most five independent attempts under the same
provisional and operator-owned-decision boundaries.

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

## Auditor session reuse

Use one independent auditor session for the definition phase and a separate one
for the implementation phase. The first attempt in each phase begins without
the authoring conversation. Later attempts resume that phase's auditor rather
than starting again.

On every resumed attempt, provide:

- the complete current candidate;
- a concise description or diff of the remediation; and
- any authority newly made relevant by the change.

Do not resend unchanged project authorities merely to reconstruct a new
conversation. The auditor must reassess the current candidate; its previous
findings and verdict do not bind the new judgement.

Store session identifiers only as ignored project runtime state, never as
tracked delivery artefacts. If a session cannot be resumed, start a new isolated
auditor and supply the complete required context rather than silently treating
the lost session as a verdict.

Session reuse avoids repeated discovery, file loading, and prompt construction.
It does not assume that a provider stops processing or charging for retained
conversation history; provider-side context and caching behaviour remain an
implementation concern.

## Delivery shape

```text
Clarify
  -> Unified specification
  -> Author definition preflight
  -> audit-definition in one reusable phase session, up to five calls
  -> OPERATOR CHECKPOINT
  -> Write tests
  -> Write code
  -> Author implementation preflight
  -> audit-implementation in one reusable phase session, up to five calls
  -> One-off and user validation
  -> OPERATOR CHECKPOINT
```

This produces at most ten audit-model calls rather than four independent loops
of up to five calls. The material reduction comes from one approval artefact and
two review boundaries, not merely from a lower numerical maximum.

## Paired delivery shape

Paired development retains a specification before every edit without requiring
a complete advance document. The bounded objective and each explicit operator
instruction authorize one reviewable slice. The durable unified specification
is consolidated from the accepted final result at closure:

```text
Bounded objective
  -> explicit instruction
  -> reviewable slice
  -> objective checks and operator validation
  -> repeat as directed
  -> consolidate unified specification and validation record
  -> focused change-scoped audits
  -> OPERATOR CHECKPOINT
```

Do not impose the normal pre-implementation `audit-definition` checkpoint after
the implementation already exists. Apply the specialist audits according to
materiality under `~/.agents/sdlc/PAIRING.md` and
`~/.agents/sdlc/AUDITS.md`. The operator confirms once that the consolidated
specification and user-test record represent the paired decisions.

## Emergency delivery shape

Emergency delivery begins only with the exact human invocation
`BYPASS-GATE-7`. Its sufficiently bounded request is the temporary
specification. It preserves evidence selection and automated TDD, but defers the
durable unified specification until the immediate fix has been implemented and
verified:

```text
Exact invocation and temporary specification
  -> select evidence
  -> RED where automated regression applies
  -> implement bounded fix
  -> GREEN and immediate verification
  -> reconcile unified specification, design, and documentation
  -> audit-code convergence, up to five attempts
  -> one-off and user validation
  -> OPERATOR CHECKPOINT
```

The emergency path does not run `audit-definition` retrospectively. Its durable
record must truthfully distinguish the temporary authority from the reconciled
as-built specification. `~/.agents/sdlc/EMERGENCY.md` remains authoritative for
the exact invocation and exception boundaries.

## Audit mechanics still to settle

Before implementing the two composite skills and runner support, decide:

- the exact combined verdict schema;
- whether a PASS explicitly reports each component result;
- how provisional component findings determine the overall verdict;
- how the runner composes the existing audit prompts without duplicating them;
- which exact artefacts and context each composite skill supplies; and
- whether any high-risk change should require a focused specialist audit in
  addition to the composite audit.
