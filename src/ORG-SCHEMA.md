# Specification and validation record schema

`spec.org` owns requirements and test definitions. `validation.org` owns test
execution state and evidence. `audits.yaml` remains harness-owned audit evidence;
test colours are not audit verdicts.

## Addressable records

- Each AC in `spec.org` has one heading: ID, descriptive title, and direct `:ac:` tag.
- Each test in both documents has one heading: ID, descriptive title, and direct
  `:testdef:` tag. Do not rely on inherited tags for these records.
- Preserve established IDs. Supported forms include `AC007.1`, `RT007.1`,
  `RT-007`, and `RT-007.1`; categories are AC, RT, OT and UT. Numeric segments may
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
- **RED:** observed failure. For RTs, execute the written test and observe the
  intended behavioural failure before implementing the corresponding production
  change. Writing code or predicting failure does not establish RED.
- **GREEN:** observed success for the recorded candidate. An initially GREEN
  preservation RT is legitimate: record its purpose rather than breaking working
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

RED/GREEN require non-empty `REVISION` and `EVIDENCE`; RTs also require `COMMAND`.
AMBER requires `REASON`. Record dates, environment and UT reviewer in the supporting
evidence. Do not use placeholders as results. Never promote states without execution.
Rerun affected tests after changes; keep unaffected evidence with its rationale.

## Deterministic checks

Run `~/.agents/sdlc/bin/sdlc-validate --help` for usage. It reads both documents,
outputs YAML under `audit-readiness`, and never runs tests or writes files.

- Schema errors prevent readiness evaluation and force `ready: false`.
- `test-code`: every specified test must have a validation heading with RED,
  AMBER or GREEN and the state-specific evidence properties above. RTs must be
  RED or GREEN; only OTs and UTs may be AMBER at this stage.
- `delivery-code`: additionally requires every RT and OT GREEN. UTs may remain
  AMBER for operator participation; missing UT records/states are not permitted.
- Unknown validation IDs are schema errors. Missing counterparts and states are
  readiness failures. Inventory output includes AC/test counts, IDs and titles,
  current test states, source locations and actionable diagnostic codes.
- Exit 0 means ready, 1 ineligible, 2 invalid schema, 3 invocation/I/O error.

The parser ignores Org blocks, comments and drawers as record sources. It rejects
unclosed blocks/drawers and duplicate properties. It validates this record subset,
not every Org feature or the quality of prose. It cannot prove evidence is true,
current or exhaustive; those remain author obligations and model-review checks.

Before the next implementation audit, adapt older sheets to this structure without
renumbering, losing history or inventing results. Definition-only review does not
require execution states. Existing documents are never rewritten by the validator.
