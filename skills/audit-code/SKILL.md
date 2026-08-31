---
name: audit-code
description: Review implementation against the selected engineering standards and language best practice. Advisory only; no file changes.
metadata:
  preferred_provider: openai-codex
  preferred_model: gpt-5.6-luna
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`, and
`~/.agents/sdlc/CODING.md`, and `~/.agents/sdlc/TESTING.md` in full. Read only
the language, Git, documentation, and domain standards selected by the project
constitution or needed for the changed files.

Review the requested scope and current diff without modifying files. Check:

- conformance to the active specification and project constitution;
- correctness, error paths, state transitions, idempotence, and compatibility;
- injection, traversal, secret exposure, unsafe permissions, and widened access;
- hidden failure, disabled verification, broad suppression, and misleading
  diagnostics;
- concurrency, cleanup, resource, timeout, and retry behaviour;
- maintainability and ecosystem idioms;
- implementation of the approved automated tests without reopening the
  effective `audit-tests` strategy; and
- documentation required by changed interfaces or operations.

Final one-off and user tests are completion evidence executed after an effective
`audit-code` PASS. Their missing results do not fail this audit. Challenge the
approved test strategy only when the implementation deviates from it or concrete
implementation evidence reveals an upstream contradiction.

For each finding, give severity, a file and line reference, the violated
requirement or standard with a descriptor, evidence, consequence, and concrete
remediation. Do not repair the code.

Run in a fresh context that did not author the implementation. Follow the exact
finding classifications, verdict format, independence rules, and changed-
artefact rule in `~/.agents/sdlc/AUDITS.md`, using `AUDIT: audit-code`.
