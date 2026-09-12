# W015 - Reset an exhausted audit allowance

Operator-authorized emergency change, 2026-09-13, under `BYPASS-GATE-7`.
This is the retrospective durable record of the bounded request, not a claim
of advance definition approval.

## Scope and acceptance criteria

- AC015.1 - Renew the allowance: an explicit `resume --reset-session` resets
  the selected work item/gate's used audit allowance to zero, then exits without
  reading stdin, loading provider configuration or invoking a provider.
- AC015.2 - Preserve evidence: reset retains native context, model ownership,
  findings, cached results and lifetime attempt history. Other work items and
  gates remain unchanged. Repeated reset before another attempt is a no-op.
- AC015.3 - Retain bounds: subsequent attempts, including fallback, consume the
  new configured allowance and stop when it is exhausted. Spec edits do not
  replenish it. Missing records, running/interrupted attempts and ambiguous
  reset invocations are rejected without mutation.

The operator will tidy old PATH artefacts. Installer changes and live project
audit-record edits are excluded. No change to provider-failure accounting,
fallback selection, missing-identity recovery, model selection or permissions.

## Design

Extend the existing harness-owned audit record with optional `budget_resets`:
each boundary contains `after_round` and a UTC `updated` timestamp. Keep
`external_round` as the lifetime attempt number. One shared `RoundsUsed` method
subtracts the last boundary, or zero for existing records, at every admission,
exhaustion and timeout-reporting point. This avoids renumbering historical
attempts or losing provenance. Existing records need no rewrite until reset.

CLI contract and example live in `cmd/sdlc-harness/help.md`. Reset is a standalone
`resume` flag requiring the audit record, work item and valid gate. Resolve a
relative record against `--project`. Reject `--input` and `--session`: resetting
the allowance is not running an audit or replacing its native identity.

Use the existing atomic YAML writer. Reset refuses `running` state; this does
not introduce concurrent-writer support. Operators must not run reset while an
audit is launching. No dependency or access changes. Explicit authorization is
a workflow rule, not proof of human identity at the CLI boundary. Known findings
remain binding; reset is never an audit verdict or permission to ignore them.

Architecture: [gate architecture](../../docs/ARCHITECTURE.md#gate-architecture).
Applicable standards: `src/AUDITS.md`, `src/CODING.md`, `src/TESTING.md`,
`src/ARCHITECTURE.md`, `src/technologies/GO.md`.

## Test definitions and evidence

- RT015.1 - Lossless reset (AC015.1, AC015.2): invoke CLI dispatch against a
  disposable exhausted record. Assert zero used rounds, retained session,
  findings and history, untouched other gate, and byte-identical repeated reset.
- RT015.2 - Invalid reset (AC015.3): reject missing/running records and conflicting
  flags, retaining record bytes.
- RT015.3 - Bounded resumed attempts (AC015.2, AC015.3): exhaust a two-attempt
  timeout fixture, reset, resume the same native session twice, then verify
  admission refuses a fifth provider call while retaining all four histories.
- RT015.4 - Fallback after reset (AC015.3): seed an exhausted Claude record,
  reset, then simulate the observed provider limit via a local executable double.
  Verify exactly one primary and one configured fallback call, PASS persistence,
  two used rounds and lifetime numbering continuing beyond five.

All RTs use disposable local files and process doubles; no metered calls belong
in regression automation. The first focused run failed on the absent reset flag.
Audit evidence belongs only in [audits.yaml](audits.yaml).

## Handoff

Update the help, audit standard, audit-record example, changelog and learnings.
Run focused regression tests and project lint, then implementation review and
retrospective definition review. Release without installing or modifying any
client project's records. The operator may install and authorize record resets.
