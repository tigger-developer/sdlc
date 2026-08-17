---
name: audit-code
description: Review implementation code against CODING.md and language best practice. Advisory only -- no file changes.
---

Review the implementation code for the current issue. Do not modify any files. The purpose of the audit is not just to present problems; it is to present solutions. Always include a concrete recommended action for every finding.

1. **Standards check.** Review all changed files against CODING.md and against best practice for the language(s) and stack(s) in use. CODING.md cannot enumerate patterns for every language -- apply community standards, idioms, and known pitfalls for this specific ecosystem alongside our own rules.

2. **Error handling.** Check every error path. Is each failure either a hard fail (let it crash) or properly handled with conditional logic? Flag any error suppression patterns (`|| true`, `|| rc=$?`, `set +e`, bare `except`, empty `catch`, etc.).

3. **Security.** Check for OWASP top 10 vulnerabilities, hardcoded credentials, injection risks, overly permissive permissions.

4. **Complexity.** Flag functions over 50 lines, nesting deeper than 3 levels, god objects, duplicated code.

5. **Summarize.** Allocate a unique report directory with `mktemp -d` and write `audit-code-<NNN>.md` there: each finding with file paths and line numbers, categorized as blockers (must fix) or smells (worth discussing), with a specific recommended fix for each. Render and open it if a renderer is available: use `AGENT_HTML_RENDERER` when set; otherwise use `pandhtml` if present. If neither exists, report the temporary markdown path. Retain the report only for the active review and remove the exact directory through explicit teardown afterwards. **In chat: one line stating blocker count, smell count, and the report path.**

End with `PASS` when no unresolved blocker or introduced finding remains. A pre-existing smell may remain only when it is identified as pre-existing, outside the ticket scope, and accompanied by a concrete justification. Otherwise end with `FINDINGS` and the concrete remediation for each finding.

In MODE PAIR, do not proceed past this audit; wait for the human's response. In MODE DELIVER, run this audit once per ticket after implementation, then return control to the delivery workflow. The workflow must remediate every blocker and introduced finding and repeat this audit until it reports PASS before moving to another ticket.
