---
name: audit-code
description: Review implementation against the selected engineering standards and language best practice. Advisory only; no file changes.
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` and
`~/.agents/sdlc/CODING.md` in full. Read only the language, testing, Git,
documentation, and domain standards selected by the project constitution or
needed for the changed files.

Review the requested scope and current diff without modifying files. Check:

- conformance to the active specification and project constitution;
- correctness, error paths, state transitions, idempotence, and compatibility;
- injection, traversal, secret exposure, unsafe permissions, and widened access;
- hidden failure, disabled verification, broad suppression, and misleading
  diagnostics;
- concurrency, cleanup, resource, timeout, and retry behaviour;
- maintainability and ecosystem idioms;
- tests at the correct observable boundary; and
- documentation required by changed interfaces or operations.

For each finding, give severity, a file and line reference, the violated
requirement or standard with a descriptor, evidence, consequence, and concrete
remediation. Do not repair the code or declare it approved.

End with either `NO FINDINGS` or a numbered findings list ordered by severity.
