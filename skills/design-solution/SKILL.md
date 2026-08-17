---
name: design-solution
description: Document the solution design for an existing issue. No code is written.
---

Document the solution design for the current issue. Do not write any code.

1. Fetch the issue with `gh issue view [n]` and read all comments -- there may have been changes since you last looked.
2. Verify the issue has exactly one AC table. If missing or duplicated, fix before proceeding.
3. Re-run the AC self-audit (see ISSUES.md §AC/Test boundary). If any AC violates the boundary, fix it and repeat the audit in MODE DELIVER; in MODE PAIR, alert me and **STOP.**
4. Document the solution on the issue:
   - Language, frameworks, and libraries to use
   - Patterns to follow
   - Anti-patterns to avoid (cite specific CODING.md sections)
5. Review the solution against the codebase. In MODE DELIVER, resolve safe in-scope contradictions and repeat the review; make a genuine-blocker handback only when resolution requires a material product decision or other human-only input. In MODE PAIR, alert me and **STOP** if the solution is contradictory or unsound.
6. Allocate test file locations and test IDs.

In MODE PAIR, end with `AWAITING PROCEED - issue #NNN`, the issue link, and **STOP**. In MODE DELIVER, return the issue link and completed design to the delivery workflow for the mandatory per-ticket AC and test audits; do not request PROCEED.
