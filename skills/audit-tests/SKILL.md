---
name: audit-tests
description: Challenge test specifications for coverage gaps, gaming opportunities, and integration risks. Advisory only; no file changes.
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
the smallest useful correction. Do not write tests, modify files, or approve the
test design.

End with either `NO FINDINGS` or a numbered findings list ordered by severity.
