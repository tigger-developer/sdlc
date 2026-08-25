---
name: draft-design-issue
description: Draft a GitHub issue with full requirements and solution design in one pass. No code is written.
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Do not search, enumerate directories, traverse mounted volumes, inspect network shares, or use `find`, `locate`, Spotlight, or equivalent discovery to resolve this path. If `~/.agents/sdlc/MAIN.md` is absent or unreadable, stop and report that exact path.

Create a GitHub issue with full requirements and solution design in one pass. Read and follow `~/.agents/sdlc/ISSUES.md` for all issue structure and AC quality standards. Do not write any code.

### Phase 1: Requirements

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

### Phase 2: Solution design

9. Verify the issue has exactly one AC table. If missing or duplicated, fix before proceeding.
10. Document the solution on the issue:
    - Language, frameworks, and libraries to use
    - Patterns to follow
    - Anti-patterns to avoid (cite specific CODING.md sections)
11. Review the solution against the codebase. In MODE DELIVER, resolve safe in-scope contradictions and repeat the review; if resolution requires a material product decision or other human-only input, complete other executable work and use the MODE DELIVER exit gate with `DELIVERY BLOCKED`. In MODE PAIR, alert me and **STOP** if the solution is contradictory or unsound.
12. Allocate test file locations and test IDs.

In MODE PAIR, end with `AWAITING PROCEED - issue #NNN`, the issue link, and **STOP**; no code may be written until PROCEED is received. In MODE DELIVER, return the issue link and completed design to the delivery workflow for the mandatory per-ticket AC and test audits; do not request PROCEED or write code before both audits report PASS.
