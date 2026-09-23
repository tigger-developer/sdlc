---
title: W019 Emergency Harness Self-Review
version: 1
last-updated: 2026-09-23
---

# Emergency harness self-review

Author review under the operator's release-specific exception recorded in
[the specification](spec.org). This is not an independent audit or a replacement
entry in the harness-owned [audit record](audits.yaml).

## Implementation checks

- Gate accounting: only incident-free valid verdicts count. Unlabelled legacy
  implementation evidence does not acquire a test-code or delivery-code label.
  Tests compare counts before and after history migration.
- Recovery: malformed output receives format clarification within the configured
  bound. Authentication selects fallback immediately, including with a limit of
  one. Failed fallback or absent fallback stops for intervention.
- Lockout and reset: raising configuration limits does not clear a recorded stop.
  Reset retires context and clears both selected-gate counters, preserving history,
  logs and other gate budgets. Pre-reset cache entries are not reused.
- Caller output: no lifetime attempt numbers or native identities in ordinary
  audit diagnostics. Anonymous liveness continues after the log cap. Intervention
  supplies a log reference, including on a refused repeat invocation.
- Persistence: the current valid verdict is recorded before deriving exhausted
  status. Evidence checks, read-only provider controls and atomic record writes
  remain in place. Logs have private permissions and bounded retention.

## Test and document checks

The full local suite and race checks passed. The compiled v3.5.0 executable also
passed a local-double smoke for correction, caching, lockout, refusal and fresh
context after reset. No metered provider was embedded in tests or called during
this self-review.

Reset tests now assert the intended error, and cooldown tests verify that a
retired primary identity is not resumed on fallback. Newly added lockout,
diagnostic-reference and authentication boundary assertions passed against the
existing behaviour; they are preservation coverage, not claimed RED results.
Observed RED/GREEN defects and commands are recorded in [validation](validation.org).

Help, standards, README, quickstart, architecture and the four workflow skills
were checked against the normal-invocation and authorized-reset contract. Cache,
lockout, reset output and liveness descriptions were reconciled. The specification
retrospectively captures the authorized behaviour, test mapping and release exception.

## Remaining boundary

Self-review lacks independent challenge. It does not establish an external PASS
for this revised candidate. The existing audit record is preserved unchanged.
Installation, real-provider operation and operator closure remain unverified.
The installed-use test remains AMBER. This exception changes no shared audit policy.

The existing CLI and recovery coordinator remain larger than the function-size
guideline. The specification records why this emergency change does not include
a wider orchestration refactor. Abrupt process death and competing writers to the
same audit record are not new recovery guarantees supplied by this change.
