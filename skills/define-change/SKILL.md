---
name: define-change
description: Define and audit one lean SDLC v3 change through operator sign-off, without implementing it.
---

Read `~/.agents/sdlc/MAIN.md`, `ISSUES.md`, `TESTING.md`, `CODING.md`, `GIT.md`,
`DOCUMENTATION.md`, and `ORGMODE.md`, then the project profile and its selected
standards and authorities.

Remain on the operator-selected branch throughout definition. Do not create,
switch, merge, or delete a feature branch. The project branch strategy applies
only when an approved specification enters delivery. Synchronize the current
branch at the definition boundaries required by `GIT.md`.

Definition defaults to **ATTENDED**. Ask the bounded clarification questions
required to make the specification safe and context-independent. If the
operator explicitly invokes **HANDS-OFF** definition mode, declare that mode
and assume the operator is unavailable for the remainder of the definition
window. In that mode, make sensible, reversible assumptions to keep the work
moving and record each one in the specification's decision log; do not stop for
routine status, incidental uncertainty, or approval requests. Stop only for a
genuine blocker, an unsafe or unauthorized action, the configured audit limit,
or a decision that cannot be made safely from the available evidence. An
ordinary progress report is not a handback. Any interruption or operator
response returns the skill to ATTENDED until HANDS-OFF is explicitly invoked
again.

Ask no more than three concise free-form questions needed to define product
scope. If operator direction raises unresolved test or design choices, ask no
more than three additional questions for each affected area. Do not use
multiple-choice UI.

Allocate the next never-used project work number. Create
`specs/WNNN-descriptor/spec.org` from
`~/.agents/sdlc/templates/v3/spec.org`. Include context, falsifiable acceptance
criteria, traced RT/UT/OT test definitions, edge cases, solution design,
mandatory security impact, and a context-independent delivery handoff. Create
`audits.yaml` from its canonical template. Create `validation.org` only when
needed. Add or update the descriptive work item in `docs/work.org`. Apply the
change-specific title, Org hierarchy, description-list, and semantic-emphasis
requirements in `ISSUES.md`; generic titles and visually unscannable drafts are
not ready for audit.

The opening summary must accurately and completely represent the detail below
its section break. Invoke the composite definition audit through the installed
audit harness with `--phase audit --gate definition`, using `audits.yaml` for
the retained session mapping. Remediate **all** findings before resuming that
same session; never rerun an unchanged candidate or remediate findings
piecemeal. Do not invoke individual audit skills.

On effective PASS, set the specification's definition-gate status to `PASS` and
move its `docs/work.org` item to `REVIEW`, preserving the detailed evidence in
`audits.yaml`. Open the specification with `sdlc-preview` and return it for
operator sign-off. Preview is presentation, not a gate. After explicit sign-off,
record the authority and date in `spec.org` and move its work item to `ACTIVE`.
Do not implement code.
