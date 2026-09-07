---
name: define-change
description: Define and audit one lean SDLC v3 change through operator sign-off, without implementing it.
---

Read `~/.agents/sdlc/MAIN.md`, `ISSUES.md`, `TESTING.md`, `CODING.md`,
`DOCUMENTATION.md`, and `ORGMODE.md`, then the project profile and its selected
standards and authorities.

Ask no more than three concise free-form questions needed to define product
scope. If operator direction raises unresolved test or design choices, ask no
more than three additional questions for each affected area. Do not use
multiple-choice UI.

Allocate the next never-used project work number. Create
`specs/NNN-descriptor/spec.org` from
`~/.agents/sdlc/templates/v3/spec.org`. Include context, falsifiable acceptance
criteria, traced RT/UT/OT test definitions, edge cases, solution design,
mandatory security impact, and a context-independent delivery handoff. Create
`audits.org` from its canonical template. Create `validation.org` only when
needed. Add or update the descriptive work item in `docs/work.org`. Apply the
change-specific title, Org hierarchy, description-list, and semantic-emphasis
requirements in `ISSUES.md`; generic titles and visually unscannable drafts are
not ready for audit.

The opening summary must accurately and completely represent the detail below
its section break. Apply the complete definition gate from `AUDITS.md` in this
context: review and remediate locally for at most five rounds, then use one
retained external context for the composite independent audit. Do not invoke
the focused auditors through separate agents or scripts.

On effective PASS, set the specification's definition-gate status to `PASS` and
move its `docs/work.org` item to `REVIEW`, preserving the detailed evidence in
`audits.org`. Open the specification with `sdlc-preview` and return it for
operator sign-off. Preview is presentation, not a gate. After explicit sign-off,
record the authority and date in `spec.org` and move its work item to `ACTIVE`.
Do not implement code.
