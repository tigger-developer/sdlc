# Audit and Gate Standards

Audits provide independent findings; they do not replace operator approval,
tests, or implementation evidence. The **audit harness** owns the prompts,
provider invocation, timeout, round limit, and retained session. No individual
audit skill is required or deployed.

## Composite gates

The **definition gate** reviews the specification, solution design, and RT/UT/OT
test definitions together. The **implementation gate** reviews implemented tests
and the production delta against the signed-off definition. The implementation
gate never reopens a signed-off definition.

For either gate:

1. Invoke `~/.agents/sdlc/bin/sdlc-harness` with the gate and exact evidence.
2. On `FAIL`, remediate **all** findings before resuming the recorded session.
3. Never submit an unchanged revision or create a replacement session.
4. Stop after the configured maximum rounds, or for a decision that cannot be
   made safely from the authorities.

The harness stores one session mapping per work item and gate in `audits.yaml`.
The first invocation starts a session; later invocations resume its recorded
`session_id`. The round limit is `delivery.audit.max_rounds` (default five) and
applies to the current invoking session only.

Audit findings, status, revisions, round numbers, and session IDs MUST be
recorded only in `audits.yaml`. Specifications and tickets may link to that
record but must not duplicate its audit state. Existing duplicated text is
historical input for migration, not a second authority.

## Verdict contract

The harness response contains exactly one envelope:

```text
GATE: definition | implementation
REVISION: <audited revision or SHA-256>
VERDICT: PASS | PROVISIONAL PASS | FAIL
```

`PROVISIONAL PASS` is effective only after its exact mechanical conditions are
applied. Timeout, silence, malformed output, or an unavailable provider is not
a pass. Record findings and remediation in `audits.yaml` without deleting
superseded evidence.

## Human boundaries

Only the operator may approve the definition, change signed-off behaviour or
architecture, and close delivered work. A hands-off workflow may continue with
reversible assumptions, but must stop for an unsafe or irreversible decision,
an unremediable mandatory test failure, or the configured round limit.

`BYPASS-GATE-7` skips definition admission for an emergency change. It still
requires TDD where practicable, the implementation gate, and a wrap-up that
backfills the specification, design, validation, and documentation evidence.

## Brownfield evidence

A brownfield change must consult the relevant requirements, design, project
authorities, historical records, regression tests, and affected implementation.
The delta need not copy the baseline; material conflicts and omitted authorities
must be reported.

## Variant workflows

Paired work records live user validation and uses the same harness only when a
gate is appropriate. It does not audit every live iteration.

# Canary

Suffix the canary string with "AUDIT "
