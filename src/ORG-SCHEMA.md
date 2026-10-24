---
title: Specification and Validation Record Schema
version: 2
last-updated: 2026-10-24
---

# Specification and validation record schema

`spec.org` owns requirements and test definitions. `validation.org` owns test
execution state and evidence. `audits.yaml` remains harness-owned audit evidence;
test colours are not audit verdicts.

## Addressable records

- Each AC in `spec.org` has one heading: ID, descriptive title, and direct `:ac:` tag.
- Each test in both documents has one heading: ID, descriptive title, and direct
  `:testdef:` tag. Do not rely on inherited tags for these records.
- Preserve established IDs. Supported forms include `AC007.1`, `AT007.1`,
  `AT-007`, and `RT-007.1`. AC identifies an acceptance criterion, not a test.
  The four test types are AT (Automated Test), RT (Regression Test), OT (One-off
  Test) and UT (User Test), as defined by `TESTING.md`. Numeric segments may
  use dots or hyphens. IDs are uppercase and must match exactly across documents.
- Use real Org headlines at column zero. The optional validation state follows
  the stars and space, before the ID. Execution states do not belong in `spec.org`.
- A tagged record without a valid ID/title, an ID-shaped heading without its
  required tag, or duplicate IDs within either document is a schema error.
- Each active spec needs at least one AC and one test. Dense case groups may
  retain one test ID; their evidence must cover every specified case.

```org
** AC007.1 - Invalid input is rejected :ac:
** RT007.1 - Reject invalid input :testdef:
```

## Execution states

Declare exactly once in `validation.org`, outside blocks and drawers:

```org
#+TODO: AMBER RED | GREEN
```

AMBER and RED are not-done states; GREEN is done. Absence of a state is valid
schema but never ready for either implementation audit stage.

- **No state:** no execution result established; test may be unwritten or unrun.
- **AMBER:** not yet executable, deferred, skipped, inconclusive or stale.
  Record the reason. Compilation/setup failures are AMBER, not useful RED.
- **RED:** observed failure. For ATs and RTs, execute the written test and observe the
  intended behavioural failure before implementing the corresponding production
  change. Writing code or predicting failure does not establish RED.
- **GREEN:** observed success for the recorded candidate. An initially GREEN
  preservation AT/RT is legitimate: record its purpose rather than breaking working
  code to manufacture RED. UT and OT colours describe results, not TDD phases.

Keep one current state on each test heading. Retain prior executions in subordinate
run headings or a history drawer without repeating the test ID as another record.
Evidence prose may refer to IDs normally. Use properties immediately under the
test heading for machine-readable current evidence:

```org
** RED RT007.1 - Reject invalid input :testdef:
:PROPERTIES:
:REVISION: abc123 plus the identified test-only working-tree change
:COMMAND: go test ./internal/parser -run TestInvalidInput
:EVIDENCE: Expected rejection; observed acceptance. See the dated run below.
:END:

*** Run 2026-09-16

- Observed assertion and relevant output, or a descriptive evidence link.
```

RED/GREEN require non-empty `REVISION` and `EVIDENCE`; ATs and RTs also require `COMMAND`.
AMBER requires `REASON`. Record dates, environment and validation method in supporting
evidence; UTs use user validation, with a reviewer role only where relevant, not a
personal name. Follow DOCUMENTATION.md. Do not use placeholders as results or
promote states without execution.
Rerun affected tests after changes; keep unaffected evidence with its rationale.

## Deterministic checks

`~/.agents/sdlc/bin/sdlc-validate --help` describes usage. The validator reads both documents,
outputs YAML under `audit-readiness`, and never runs tests or writes files.

- Schema errors prevent readiness evaluation and force `ready: false`.
- `test-code`: every specified test must have a validation heading with RED,
  AMBER or GREEN and the state-specific evidence properties above. ATs and RTs must be
  RED or GREEN; only OTs and UTs may be AMBER at this stage.
- `delivery-code`: additionally requires every AT, RT and OT GREEN. UTs may remain
  AMBER for operator participation; missing UT records/states are not permitted.
- Unknown validation IDs are schema errors. Missing counterparts and states are
  readiness failures. Inventory output includes AC/test counts, IDs and titles,
  current test states, source locations and actionable diagnostic codes.
- Exit 0 means ready, 1 ineligible, 2 invalid schema, 3 invocation/I/O error.
- Inventory counts distinguish acceptance criteria from tests, with separate
  `at`, `rt`, `ot` and `ut` counts. AT support is additive; existing IDs and
  RT/OT/UT evidence remain valid without renumbering or reclassification.

The parser ignores Org blocks, comments and drawers as record sources. It rejects
unclosed blocks/drawers and duplicate properties. It validates this record subset,
not every Org feature or the quality of prose. It cannot prove evidence is true,
current or exhaustive; those remain author obligations and model-review checks.

Implementation-review evidence must conform to this structure without
renumbering, lost history or invented results. Definition-only review does not
require execution states. Existing documents are never rewritten by the validator.

## Document history

This revision separates independent review, author self-audit and delivery-only
responsibilities. Earlier rules remain in their canonical owners and Git history.

Version 1 adds AT recognition, automated-test readiness and separate inventory
counts. AC remains an acceptance-criterion record; earlier schema text is
preserved in Git history.
