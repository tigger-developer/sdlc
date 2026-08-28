---
name: summarize-issues
description: Summarize a project's open work against its specifications and architecture. Advisory only; no file changes.
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` and
`~/.agents/sdlc/ISSUES.md` in full. Read the project's specifications,
constitution, architecture, and open-work source without modifying them.

Summarize each open item with its identifier and descriptor, intended outcome,
dependencies, and verified current state. Then identify missing specification
coverage, contradictions, duplicates, architectural conflicts, and sequencing
constraints. Recommend priorities based on dependency, user impact, risk, and
uncertainty. Do not infer an approval state or invent a delivery lifecycle.
