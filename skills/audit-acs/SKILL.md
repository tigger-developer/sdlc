---
name: audit-acs
description: Challenge a specification's requirements and acceptance criteria. Advisory only; no file changes.
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` and
`~/.agents/sdlc/ISSUES.md` in full. Review the active specification without
modifying files.

Check:

- the user or system outcome is explicit;
- every requirement is observable, falsifiable, bounded, and internally
  consistent;
- acceptance criteria describe behaviour rather than test procedures;
- normal, alternate, error, boundary, permission, repetition, migration,
  security, privacy, and accessibility cases are covered where relevant;
- compound requirements expose every independently failing condition;
- undefined product, architecture, data, or access decisions are identified;
  and
- every identifier in the report has an adjacent descriptor.

For each finding, state the affected requirement with its descriptor, the gap
or contradiction, the consequence, and a concrete proposed correction. Do not
rewrite the specification or declare it approved.

End with either `NO FINDINGS` or a numbered findings list ordered by severity.
