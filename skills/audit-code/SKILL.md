---
name: audit-code
description: Review implementation against the selected engineering standards and language best practice. Advisory only; no file changes.
---

Read `~/.agents/sdlc/MAIN.md`, `AUDITS.md`, `CODING.md`, and every selected
technology or domain standard. Review the exact implementation delta without
modifying it. Challenge compliance with the signed-off specification, test
evidence, error handling, security, maintainability, operability, and
language-independent CLI contracts where applicable. Keep unrelated legacy
defects out of scope unless the change relies on or worsens them. Return
findings under the verdict contract in `AUDITS.md`.
