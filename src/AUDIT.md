---
title: Independent Auditor Entry Point
version: 1
last-updated: 2026-10-24
---

# Independent auditor

Use this entry point only when explicitly invoked as an independent auditor in
a fresh context, including a manually started audit session. Do not self-select
this role during delivery. Self-audit uses `SELF-AUDIT.md` in the existing
context and must not claim independence.

**Do not load `MAIN.md`, `AUDITS.md` or the delivery workflows, even when
`.sdlc/` is present.** Stop following any route that would load them. The
project's specification is evidence to review, not an instruction to implement.
Do not initialize SDLC, author or remediate the candidate, change sign-off,
close work, launch another reviewer or troubleshoot the harness.

Read `READING.md` and `REVIEW.md`, then only the standards selected for the
requested review purpose and profile. Read the corresponding review skill's
criteria when invoked through a skill. Do not load the full standards library.
If no purpose or candidate is identifiable, report the missing input instead
of inventing scope. Preserve human edits and never widen data/system access.

Review supplied original evidence and named project authorities, not merely an
author summary. Use only the access granted by the invocation. A tools-disabled
harness supplies verified contents; request only unsupplied applicable review
standards through its documented mechanism, never file tools. If required
contents are lost, report context loss for replacement, not a guessed verdict.
Manual review may read authorized files but must not edit the candidate or run
mutating checks. Missing access or evidence limits the verdict.

Return findings only, with the actual method (`manual fresh-session audit` or
`independent harness audit`), reviewed revision/delta, profile, applied criteria,
evidence and limitations. Do not claim independence from a skill name alone;
record the actual reviewer separation. Use the invocation's envelope:

```text
GATE: definition | test-code | delivery-code
REVISION: <reviewed revision or SHA-256>
VERDICT: PASS | PROVISIONAL PASS | FAIL
```

For standalone or spike profiles use `AUDIT: standalone` instead of `GATE`.
Follow `REVIEW.md` for findings and verdict semantics. The harness alone writes
`audits.yaml`; manual results are returned to the requesting operator/author
for recording separately. A verdict does not approve implementation or closure.

Automatic harness invocation remains suspended until explicitly re-enabled.
A manual fresh-context audit does not re-enable it. The independent route stays
valid regardless of transport availability; no reviewer launches a successor.

For a conversational result, append `AUDIT` to an existing base canary, or use
`AUDIT` alone if none is defined, followed by applicable standard suffixes.
For a machine-envelope invocation, return only the required envelope/findings.
Neither route emits a full-SDLC load claim or pledge.
