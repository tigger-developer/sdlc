# Audit and Gate Standards

Audits challenge work before a human gate. They provide evidence; they do not
replace operator approval, test execution, or implementation evidence.

## Audit roles

- `audit-spec` challenges the change-specific title, ABC presentation, context,
  scope, requirements, acceptance criteria, edge cases, authority, and
  context-independent handoff.
- `audit-design` challenges traceability, architecture fit, boundaries,
  trade-offs, security, operability, and failure behaviour.
- `audit-test-definitions` challenges proposed RT/UT/OT coverage,
  classification, observable boundaries, TDD suitability, and gaming
  opportunities without demanding implementation evidence.
- `audit-test-code` challenges implemented tests, RED/GREEN evidence, current
  results, traceability, assertions, and departures from the signed-off test
  strategy.
- `audit-code` challenges the implemented change against the signed-off
  specification, tests, selected standards, and language best practice.

Auditors are findings-only. They do not modify the audited artefact or mark
their own findings resolved.

## Two composite gates

### Definition gate

The authoring context applies `audit-spec`, `audit-design`, and
`audit-test-definitions` together to `spec.org`.

1. Review locally, remediate, and repeat for at most five local rounds.
2. Start no external auditor until all three local reviews pass.
3. Start one external Codex context and run all three audits there in one turn.
4. If it fails, remediate locally, rerun all affected local reviews, then resume
   the same external context. Never create another auditor for the same gate.
5. Stop at PASS, five failed gate rounds, or a human-controlled decision.

The external auditor receives only the current specification, project profile,
named authorities, applicable standards, and focused repository evidence. It
must test whether a new agent can safely deliver without the drafting
conversation.

An effective PASS sets `DEFINITION_GATE` in `spec.org` to `PASS` and moves the
work item in `docs/work.org` to `REVIEW`. Explicit operator sign-off records the
authority and date and moves the work item to `ACTIVE`. Detailed rounds remain
in `audits.org`. Lifecycle-only recording does not invalidate the
audit; any material specification change resets the gate and sign-off to
`PENDING`.

### Implementation gate

The delivery context applies `audit-test-code` and `audit-code` together to the
implemented tests and production code.

Delivery starts only when `docs/work.org` records `ACTIVE` and `spec.org`
records a current definition gate `PASS`, explicit operator sign-off, and
linked audit evidence. It does not rerun definition audits merely because a
delivery context has started.

1. Review locally, remediate, and repeat for at most five local rounds.
2. Start no external auditor until both local reviews pass.
3. Start one external Codex context and run both audits there in one turn.
4. If it fails, remediate locally, rerun affected tests and local reviews, then
   resume the same external context.
5. Stop at PASS, five failed gate rounds, or a human-controlled decision.

Each gate therefore spans at most two contexts: its authoring context and one
retained external auditor context. Do not invoke each focused audit in a new
context and do not replace an unfavourable auditor.

Use the external-audit timeout resolved from `.sdlc/project.yaml`, then
`~/.agents/sdlc.yaml`; the default is five minutes. A timeout is a runner
incident, not an audit finding or PASS. Interrupt the retained auditor when the
timeout expires and record the incident. Resume that same context only when the
operator directs a retry; otherwise return the incident for a human decision.
Never silently create another auditor.

## Findings and verdicts

Classify findings as:

- `[BLOCKING]`: a contradiction, unsafe boundary, missing material decision,
  unverifiable requirement, coverage gap, standards violation, or defect that
  prevents sign-off.
- `[CONDITION]`: an exact mechanical correction with a deterministic check that
  requires no judgement.
- `[ADVISORY]`: a non-blocking improvement.

Use `PASS` when no required correction remains, `PROVISIONAL` when every
required correction is a mechanical condition, and `FAIL` when a blocking
finding exists. A timeout, silence, malformed verdict, or wrong audit name is
not PASS.

Each composite verdict records:

```text
GATE: definition | implementation
REVISION: <audited revision or SHA-256>
VERDICT: PASS | PROVISIONAL | FAIL

1. [audit-name] [classification] <finding with descriptive IDs>
```

A PROVISIONAL verdict becomes effective PASS only when the author applies
exactly the stated conditions, verifies each stated check, and changes nothing
else. Any judgement or additional change requires the retained auditor to
review the new revision.

## Evidence record

Write every local and external gate result to the active change's `audits.org`:

- gate and round;
- artefact and exact revision;
- local or external context;
- all component audits applied;
- verdict and findings;
- remediation or condition receipt; and
- the retained external context identifier.

Preserve superseded verdicts as revision-specific history. A later relevant
change makes an earlier PASS non-current; it does not erase it. Review only the
changed scope and enough adjacent context to judge it safely.

## Brownfield evidence

For brownfield work, confirm the author examined the relevant requirement and
design authorities, traced historical records, maintained regression tests,
and affected implementation. Report material omitted sources or conflicts.
Do not demand that a delta specification copy its entire baseline.

## Human boundaries

Stop the autonomous loop when remediation would:

- change a signed-off requirement or design;
- decide product behaviour, scope, architecture, security, privacy, access,
  persisted data, an external contract, or an irreversible outcome without
  authority; or
- exceed five failed rounds.

At a gate handback, report the verdict, rounds, decisions, assumptions,
advisories, and unresolved findings. Give every ID a descriptor. Only the
operator may sign off the definition or close delivered work.

## Variant workflows

- Paired work uses only the change-scoped audits required by `PAIRING.md`; it
  does not run a gate on every live iteration.
- `BYPASS-GATE-7` skips the definition gate, but runs the implementation gate
  after tests and implementation, then reconciles durable specification,
  design, validation, and documentation evidence.
