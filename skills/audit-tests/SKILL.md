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

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`,
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
- automated regression, one-off, and user tests are classified appropriately,
  with a recorded rationale where no automated regression test is justified;
- every selected one-off and user test has a `PENDING` entry in the active
  feature's `validation.md` containing its descriptor, traceability, expected
  result, procedure or viewing conditions, and an implementation-stage task to
  record its observed result;
- test architecture additions are justified in the plan;
- temporary state is bounded, isolated, and cleaned up; and
- every identifier in the report has an adjacent descriptor.

For each finding, state the requirement and test with descriptors, the risk, and
the smallest useful correction. Do not write tests or modify files.

Run in a fresh context that did not author the test design. Follow the exact
finding classifications, verdict format, independence rules, and changed-
artefact rule in `~/.agents/sdlc/AUDITS.md`, using `AUDIT: audit-tests`.
