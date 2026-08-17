---
name: audit-acs
description: Challenge acceptance criteria for coverage gaps, missing edge cases, and quality violations. Advisory only -- no code or file changes.
---

Review the acceptance criteria for the current issue. Do not write any code or modify any files. The purpose of the audit is not just to present problems; it is to present solutions. Always include a concrete recommended action for every finding.

1. **Summarize the functionality.** In plain language, describe what this issue delivers from the user's perspective. Post this summary in chat.

2. **Enumerate edge cases.** For each AC, list the edge cases, boundary conditions, error states, and unusual inputs that could falsify it. Be adversarial -- think about what could go wrong, not just what should go right.

3. **Identify missing ACs.** Are there user-facing behaviours, error conditions, or integration points implied by the problem statement that no AC addresses? List them.

4. **Check AC quality.** Run the AC self-audit from ISSUES.md:
   - Does each AC describe a system state, not a test action?
   - Does it contain any forbidden language?
   - Does it pass the litmus test?

5. **Summarize.** Allocate a unique report directory with `mktemp -d` and write `audit-acs-<NNN>.md` there: missing edge cases, missing ACs, quality violations, with specific recommended additions or rewrites for each. Render and open it if a renderer is available: use `AGENT_HTML_RENDERER` when set; otherwise use `pandhtml` if present. If neither exists, report the temporary markdown path. Retain the report only for the active review and remove the exact directory through explicit teardown afterwards. **In chat: one line stating the totals (missing ACs, edge case gaps, quality violations) and the report path.**

End with `PASS` when no unresolved findings remain; otherwise end with `FINDINGS` and the concrete remediation for each finding.

In MODE PAIR, do not proceed past this audit; wait for the human's response. In MODE DELIVER, return control to the delivery workflow. The workflow must remediate every finding and repeat this audit until it reports PASS. No code may be written for the ticket before both `audit-acs` and `audit-tests` report PASS.
