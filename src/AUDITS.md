# Independent Audit Standards

Audits provide independent evidence without making the auditor an author or the
operator an intermediary in routine revision cycles.

## Independence

- Run every audit in a fresh context that did not author the audited artefact.
- The main authoring context may dispatch the auditor, wait for its verdict,
  remediate findings, and dispatch the next fresh audit.
- The auditor is findings-only. It must not modify the audited artefact or mark
  its own finding resolved.
- Record the audit name, provider, model, artefact revision, verdict, findings,
  and any superseding attempt in the active feature's `audits.md`.

## Finding classifications and verdicts

Classify each finding as:

- `[BLOCKING]`: a material contradiction, missing decision, unsafe boundary,
  unverifiable requirement or evidence claim, standards violation, or defect
  that prevents phase sign-off.
- `[CONDITION]`: an exact mandatory correction that requires no further
  judgement and has a stated deterministic verification method.
- `[ADVISORY]`: an optional improvement that does not prevent phase sign-off.

Do not fail an audit solely because an advisory exists. Use `PASS` when no
required correction remains; a PASS may include numbered advisory findings.
Use `PROVISIONAL` when every required correction qualifies as a condition and
no blocking finding exists. Use `FAIL` when at least one blocking finding
exists. A PROVISIONAL or FAIL verdict may also include advisories. End with
exactly one of these machine-checkable forms:

```text
AUDIT: <audit name>
AUDITOR_PROVIDER: <provider used for this audit>
AUDITOR_MODEL: <model used for this audit>
VERDICT: PASS

1. [ADVISORY] <optional improvement ordered by severity>
```

or:

```text
AUDIT: <audit name>
AUDITOR_PROVIDER: <provider used for this audit>
AUDITOR_MODEL: <model used for this audit>
VERDICT: PROVISIONAL

1. [CONDITION] <exact required correction> | VERIFY: <deterministic check>
2. [ADVISORY] <optional additional finding>
```

or:

```text
AUDIT: <audit name>
AUDITOR_PROVIDER: <provider used for this audit>
AUDITOR_MODEL: <model used for this audit>
VERDICT: FAIL

1. [BLOCKING] <material finding ordered by severity>
2. [ADVISORY] <optional additional finding>
```

Omit numbered lines when a PASS has no advisories. A changed artefact requires a
fresh independent audit unless every change satisfies an exact PROVISIONAL
condition under the receipt contract below.

## Provisional conditions

A condition is permitted only when it is:

- narrow, unambiguous, and mechanical;
- consistent with every signed-off upstream artefact;
- deterministically verifiable by the authoring context; and
- unrelated to product behaviour, architecture, security, privacy, access,
  persisted data, external contracts, or irreversible outcomes.

The auditor must state the exact correction and its verification method in the
same `[CONDITION]` finding. If the correction requires interpretation, a choice
between alternatives, or a material change, classify it as `[BLOCKING]`
instead.

The authoring context may apply exactly the stated conditions without a fresh
audit. It must verify each condition, ensure no additional change entered the
corrected revision, and append this receipt to the active feature's `audits.md`:

```text
CONDITION_RECEIPT:
AUDIT: <audit name>
AUDITED_REVISION: <revision that received PROVISIONAL>
CORRECTED_REVISION: <revision containing only the conditions>
EFFECTIVE_VERDICT: PASS

1. [SATISFIED] <condition> | EVIDENCE: <verification result>
```

Record every condition in the auditor's order. The receipt matures the
PROVISIONAL verdict to an effective PASS without another model audit. If any
condition cannot be applied or verified exactly, or the corrected revision
contains another change, treat the attempt as FAIL and obtain a fresh audit.

## Autonomous phase convergence

The main authoring context owns convergence within the current phase. The
initial audit is attempt one. After a FAIL, it must remediate blocking findings
that belong to the current phase and dispatch a fresh audit without handing
back to the operator between attempts. It may make and record reasonable,
reversible decisions consistent with signed-off upstream artefacts.

After a PROVISIONAL verdict, apply and verify its conditions under the receipt
contract. This does not consume another audit attempt. A condition that cannot
be satisfied exactly converts the attempt to FAIL for the five-attempt limit.

Stop the autonomous loop when:

- the audit passes or a PROVISIONAL verdict matures to effective PASS;
- five audit attempts in the phase have failed;
- remediation would change a signed-off upstream artefact; or
- a decision affects product behaviour, scope, security, privacy, access,
  persisted data, an external contract, or an irreversible outcome and is not
  already authorized.

Do not change a signed-off upstream artefact merely to obtain a PASS. When a
blocking finding belongs upstream, identify the exact proposed correction and
return it for operator validation. Do not switch auditors merely to seek a more
favourable verdict.

## Phase handback

Return to the operator only after PASS, effective PASS from a satisfied
PROVISIONAL verdict, the fifth failed attempt, or an earlier human-controlled
blocker. Report:

- the current verdict and number of audit attempts;
- decisions made within the signed-off authority and their rationale;
- assumptions or proposed upstream changes requiring operator validation;
- retained advisories; and
- any unresolved blocking findings.

An audit PASS or effective PASS is independent evidence, not operator approval.
Request operator sign-off before advancing to the next phase.

## Emergency changes

`BYPASS-GATE-7` skips the pre-implementation specification, design, and test
audits because its exact surrounding request is the temporary specification and
the applicable test evidence is selected before the fix. It does not skip
`audit-code`.

After implementation, verification, and durable artefact reconciliation, run a
change-scoped `audit-code`. Remediate blocking findings and dispatch fresh
audits under the normal five-attempt convergence contract until the change has
an effective PASS or reaches a defined handback condition. Unrelated legacy
defects are not blocking unless the emergency change depends on them, worsens
them, or cannot be assessed safely without resolving them.

## Paired development

The staged specification, design, test, and code audit sequence applies to Spec
Kit delivery. It is not repeated for each iteration of explicitly selected
paired development under `~/.agents/sdlc/PAIRING.md`.

For paired work:

- run one `audit-code` at closure when the change adds or materially modifies
  code, templates, scripts, or non-trivial CSS;
- scope that audit to the change and only the adjacent context needed to assess
  it;
- do not make unrelated legacy defects blocking unless the change relies on
  them, worsens them, or cannot be assessed safely without resolving them;
- run `audit-design` only for a durable structural, architectural, theme API,
  integration, or deployment decision;
- run `audit-tests` only when material automated-test design exists to judge;
  and
- run `audit-spec` only when the paired work produces or changes a durable
  specification requiring independent review.

When an audit does not apply, state that fact and the reason in the closure
handback. The operator's confirmed user-test record remains evidence rather than
an audit verdict.
