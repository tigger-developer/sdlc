---
name: summarize-issues
description: Summarize all open issues, review against architecture and implementation plan, identify gaps, and recommend prioritization.
---

Summarize the current state of open issues for this project. Do not write any code or modify any files.

1. **List open issues.** Fetch all open issues with `gh issue list` and summarize each: issue number, title, current status (which gate it is at, whether it has ACs, whether it has a solution design).

2. **Review project documentation.** Read the project's architecture document, implementation plan, VISION document, and README (if they exist in `./docs/`). Understand the overall goals and planned features.

3. **Gap analysis.** Compare the open issues against the architecture and implementation plan. Identify:
   - Planned features or components that have no corresponding issue
   - Issues that do not align with the documented architecture or vision
   - Dependencies between issues that are not captured
   - Technical debt or infrastructure work implied by the plan but not tracked

4. **Recommend new issues.** For each gap identified, describe what issue should be created and why.

5. **Prioritize.** Recommend a prioritization order for all open issues (existing and proposed), considering:
   - Dependencies (what blocks what)
   - Risk (what is most likely to cause problems if deferred)
   - Value (what delivers the most user-facing impact)
   - Effort (relative size)

6. **Summarize.** Present the full picture: open issues, gaps, proposed additions, and recommended order of work.

In MODE PAIR, wait for the human's response. In MODE DELIVER, return the summary to the delivery workflow and continue only with work inside the confirmed scope; recommendations must not expand that scope.
