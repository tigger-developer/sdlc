---
name: draft-issue
description: Create a GitHub issue with acceptance criteria and test specifications following SDLC standards.
---

Create a GitHub issue for the current task. Read and follow `[sdlc-home]/ISSUES.md` for all issue structure and AC quality standards.

1. Review affected files and project documentation.
2. Draft the issue body: problem statement, AC table.
3. **Self-audit every AC row** (see ISSUES.md §AC/Test boundary):
   - Does it describe a system state, not a test action?
   - Does it contain any forbidden word?
   - Does it pass the litmus test?
   - If any AC fails, rewrite before posting.
4. Enumerate tests for each AC. For each test, justify RT/OT/UT using the decision tree in TESTING.md.
5. Check each AC has more than one test. If any AC has exactly one test, enumerate what's missing.
6. For multi-condition ACs, ensure the Tests column accounts for every condition.
7. Check for forbidden test patterns:
   - Any UT that could be verified by automation -> change to RT or OT
   - Any RT that invokes `make`, `make test`, or the build system -> change to OT or UT
   - Any RT for ephemeral/one-time verification -> change to OT
8. Post the issue.

In MODE PAIR, end with `DRAFT ISSUE CREATED - issue #NNN`, the issue link, and **STOP**. This is the decomposed path; solution design and PROCEED are still required before code may be written. In MODE DELIVER, return the issue link to the delivery workflow for solution design and the mandatory per-ticket AC and test audits; do not request PROCEED.
