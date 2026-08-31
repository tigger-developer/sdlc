# Independent Audit Standards

Audits provide independent evidence without making the auditor an author or the
operator an intermediary in routine revision cycles.

## Independence

- Invoke formal audits through `sdlc-audit`. It starts a new one-shot harness
  process in an empty temporary working directory and does not resume or inherit
  the authoring conversation.
- The audit skill supplies only the candidate files and exact context files
  needed for its judgement. The runner embeds their contents with the canonical
  audit prompt and this contract. Do not pass directories, the whole repository,
  or unrelated files.
- The main authoring context may dispatch the auditor, wait for its verdict,
  remediate findings, and dispatch the next fresh audit.
- The auditor is findings-only. It must not modify the audited artefact or mark
  its own finding resolved.
- Record the audit name, provider, model, artefact revision, verdict, findings,
  and any superseding attempt in the active feature's `audits.md`.

The runner reads `SDLC_AUDIT_HARNESS`, `SDLC_AUDIT_PROVIDER`, and
`SDLC_AUDIT_MODEL` from user defaults in `~/.agents/.env`, then applies values
from the ignored project `.env` with project precedence. This release always
uses Hermes: an unset or `hermes` harness value is silent; any other value emits
a warning and falls back to Hermes. Provider and model are required and passed
explicitly to Hermes. A harness execution error, timeout, malformed verdict,
wrong audit name, or reported provider or model that does not match the request
fails closed. A valid FAIL remains an audit result, not a runner failure.
Each invocation has a 15-minute runtime budget and hard timeout.
Hermes reasoning or display text before the last exact audit header is discarded;
only the validated machine-readable report is returned to the caller.

## Brownfield source coverage

For a brownfield specification or design, verify that the author examined the
relevant requirement and design authorities, historical work records,
maintained regression tests and traceability, and affected implementation. An
auditor that discovers a material source or conflict omitted from the authored
context pass must report a source-coverage finding. Do not require baseline
detail to be copied into a delta artefact merely to prove that it was examined.

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

Return only the selected form. Do not add an introduction, conclusion,
explanation, summary, Markdown fence, or any other text before or after it.

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

Omit numbered lines when a PASS has no advisories. A relevant change within the
audited scope makes the earlier PASS non-current and requires a fresh independent
audit unless every change satisfies an exact PROVISIONAL condition under the
receipt contract below. Preserve the earlier verdict as revision-specific
history.

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

The operator may sign off the current phase and authorize one or more named,
consecutive downstream phases in the same instruction. This authorization does
not pre-approve artefacts that do not yet exist or replace any required audit.
It permits the authoring context to advance through the named phases without an
intermediate handback when each required audit reaches effective PASS and no
autonomous-loop stop condition applies.

After sign-off of an audited design, an instruction to "move on to test and
build" authorizes task and test-traceability generation, `audit-tests`
convergence, cross-artefact analysis, TDD implementation, and `audit-code`
convergence. It does not waive RED/GREEN evidence, authorize an upstream change,
or silently authorize a one-off or user test requiring human participation or
separate external authority. Return after the authorized sequence with the
decisions, assumptions, evidence, advisories, and unresolved matters required
by this section.

## Revision and completion boundaries

An audit verdict remains evidence for the artefact revision and scope it
assessed. A later relevant change does not erase that history, but the earlier
verdict is no longer current for completion. The fresh audit may focus on the
later delta plus the adjacent context needed to assess it safely; it need not
repeat unrelated review.

`audit-tests` owns the selected test strategy before implementation.
`audit-code` owns the implementation and must not reopen an effective
`audit-tests` PASS merely because final one-off or user-test results have not yet
been executed. For staged and emergency delivery, execute final one-off and user
tests after `audit-code` has an effective PASS. Completion and convergence, not
the code audit, require their current passing results in `validation.md`.

If final validation exposes a defect and remediation changes code, obtain a
fresh `audit-code` verdict for the changed implementation and repeat only the
test results materially affected by the change.

## Emergency changes

`BYPASS-GATE-7` skips the pre-implementation specification, design, and test
audits because its exact surrounding request is the temporary specification and
the applicable test evidence is selected before the fix. It does not skip
`audit-code`.

After implementation, verification, and durable artefact reconciliation, run a
change-scoped `audit-code`. Remediate blocking findings and dispatch fresh
audits under the normal five-attempt convergence contract until the change has
an effective PASS or reaches a defined handback condition. Then execute and
record the selected one-off and user tests. Unrelated legacy defects are not
blocking unless the emergency change depends on them, worsens them, or cannot be
assessed safely without resolving them.

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

If audit remediation changes behaviour already covered by a paired validation,
repeat only the materially affected validation before closure. Preserve the
earlier result as superseded evidence for its reviewed revision.

When an audit does not apply, state that fact and the reason in the closure
handback. The operator's confirmed user-test record remains evidence rather than
an audit verdict.
