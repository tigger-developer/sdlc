---
name: define-change
description: Draft one lean SDLC v3 change specification containing requirements, tests, and solution design.
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
`audits.org` and `validation.org` from their canonical templates only when
needed. Add or update the descriptive work item in `docs/work.org`.

The opening summary must accurately and completely represent the detail below
its section break. Open the completed specification with `sdlc-preview`; preview
is presentation, not a gate. Do not implement code. The next action is the
definition gate.
