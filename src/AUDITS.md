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
- `[ADVISORY]`: an optional improvement that does not prevent phase sign-off.

Do not fail an audit solely because an advisory exists. Use `PASS` when no
blocking finding exists; a PASS may include numbered advisory findings. Use
`FAIL` when at least one blocking finding exists; a FAIL may also include
advisories. End with exactly one of these machine-checkable forms:

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
VERDICT: FAIL

1. [BLOCKING] <material finding ordered by severity>
2. [ADVISORY] <optional additional finding>
```

Omit numbered lines when a PASS has no advisories. A changed artefact requires a
fresh independent audit of that artefact.

## Autonomous phase convergence

The main authoring context owns convergence within the current phase. The
initial audit is attempt one. After a FAIL, it must remediate blocking findings
that belong to the current phase and dispatch a fresh audit without handing
back to the operator between attempts. It may make and record reasonable,
reversible decisions consistent with signed-off upstream artefacts.

Stop the autonomous loop when:

- the audit passes;
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

Return to the operator only after PASS, the fifth failed attempt, or an earlier
human-controlled blocker. Report:

- the current verdict and number of audit attempts;
- decisions made within the signed-off authority and their rationale;
- assumptions or proposed upstream changes requiring operator validation;
- retained advisories; and
- any unresolved blocking findings.

An audit PASS is independent evidence, not operator approval. Request operator
sign-off before advancing to the next phase.
