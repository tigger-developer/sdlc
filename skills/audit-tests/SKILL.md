---
name: audit-tests
description: Challenge test specifications for coverage gaps, gaming opportunities, and integration risks. Advisory only; no file changes.
metadata:
  preferred_provider: openai-codex
  preferred_model: gpt-5.6-luna
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`,
`~/.agents/sdlc/ISSUES.md`, and
`~/.agents/sdlc/TESTING.md` in full. Review the active specification and test
design without modifying files.

Check:

- every independently failing requirement condition has evidence;
- tests exercise the user's or consuming system's boundary;
- a narrow implementation cannot pass while the required behaviour remains
  broken;
- normal, alternate, error, boundary, permission, repetition, concurrency, and
  migration paths are proportionately covered;
- regression, one-off, and human validation are classified appropriately;
- test architecture additions are justified in the plan;
- temporary state is bounded, isolated, and cleaned up; and
- every identifier in the report has an adjacent descriptor.

For each finding, state the requirement and test with descriptors, the risk, and
the smallest useful correction. Do not write tests or modify files.

Run in a fresh context that did not author the test design. End with exactly:

```text
AUDIT: audit-tests
AUDITOR_PROVIDER: <provider used for this audit>
AUDITOR_MODEL: <model used for this audit>
VERDICT: PASS
```

or, when any finding exists:

```text
AUDIT: audit-tests
AUDITOR_PROVIDER: <provider used for this audit>
AUDITOR_MODEL: <model used for this audit>
VERDICT: FAIL

1. <finding ordered by severity>
```

Any finding requires `FAIL`. Do not modify the artefact or mark your own
finding resolved. A changed artefact requires a fresh audit.
