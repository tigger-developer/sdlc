---
name: audit-code
description: Review implementation code against CODING.md and language best practice. Advisory only -- no file changes.
---

Review the implementation code for the current issue. Do not modify any files. The purpose of the audit is not just to present problems; it is to present solutions. Always include a concrete recommended action for every finding.

1. **Standards check.** Review all changed files against CODING.md and against best practice for the language(s) and stack(s) in use. CODING.md cannot enumerate patterns for every language -- apply community standards, idioms, and known pitfalls for this specific ecosystem alongside our own rules.

2. **Error handling.** Check every error path. Is each failure either a hard fail (let it crash) or properly handled with conditional logic? Flag any error suppression patterns (`|| true`, `|| rc=$?`, `set +e`, bare `except`, empty `catch`, etc.).

3. **Security.** Check for OWASP top 10 vulnerabilities, hardcoded credentials, injection risks, overly permissive permissions.

4. **Complexity.** Flag functions over 50 lines, nesting deeper than 3 levels, god objects, duplicated code.

5. **Documentation impact validation.** Review the implemented changes against the relevant project documentation and `DOCUMENTATION.md`. Confirm that the documentation accurately describes every affected user-observable behaviour, interface, configuration, operational procedure, architecture, dependency, and documented contract. This is a semantic review, not a test: do not accept documentation greps as evidence. Record the documentation reviewed, updates required or made, and any justified determination that no update was required.

6. **Summarize.** Allocate a unique report directory with `mktemp -d` and write `audit-code-<NNN>.md` there: each finding with file paths and line numbers, categorized as blockers (must fix) or smells (worth discussing), with a specific recommended fix for each. Include the documentation-impact validation and its evidence. In MODE PAIR, use the executable path in `HTML_PREVIEW_TOOL` when it is set and available; treat its value as one executable, never shell code or a compound command. If it is unset or unavailable, choose an available text editor and open the Markdown report there. In MODE DELIVER, ignore `HTML_PREVIEW_TOOL`, do not open the report in a renderer or editor, record the Markdown evidence on the affected issue, and continue. Retain the report only for the active review and remove the exact directory through explicit teardown afterwards. **In chat: one line stating blocker count, smell count, and the report path.**

End with `PASS` when no unresolved blocker or introduced finding remains. A pre-existing smell may remain only when it is identified as pre-existing, outside the ticket scope, and accompanied by a concrete justification. Otherwise end with `FINDINGS` and the concrete remediation for each finding.

In MODE PAIR, do not proceed past this audit; wait for the human's response. In MODE DELIVER, post a `QUALITY CHECK: CODE AUDIT` comment on the affected issue using the verdict, evidence, findings, remediation, and attempt fields defined in `ISSUES.md`, then return control to the delivery workflow. PASS advances immediately to the remaining child-closure work. FAIL returns to remediation and re-evaluation; it is not a handback by itself. Run this audit once per ticket after implementation, remediate every blocker and introduced finding, and repeat the audit until it reports PASS before moving to another ticket.
