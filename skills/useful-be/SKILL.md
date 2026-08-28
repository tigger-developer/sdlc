---
name: useful-be
description: Load the minimum project and engineering context needed for useful work.
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Do not search, enumerate
directories, traverse mounted volumes, inspect network shares, or use `find`,
`locate`, Spotlight, or equivalent discovery to resolve it. If
`~/.agents/sdlc/MAIN.md` is absent or unreadable, report that exact path.

1. Read `~/.agents/sdlc/MAIN.md` in full.
2. Read the project's primary README and agent or contributor instructions.
3. If the project uses Spec Kit, read
   `.specify/memory/constitution.md` and the active feature's `spec.md`,
   `plan.md`, and `tasks.md` when those artefacts are relevant to the
   requested work.
4. Use the constitution or `MAIN.md` routing table to read only the applicable
   standards. Do not preload every language or domain document.
5. Inspect the repository state and affected area sufficiently to distinguish
   verified facts from assumptions.

Return a concise orientation: the requested outcome, controlling specification
with a descriptor, selected standards, relevant architecture, current evidence,
and unresolved material questions. Do not start implementation merely because
the context has been loaded.
