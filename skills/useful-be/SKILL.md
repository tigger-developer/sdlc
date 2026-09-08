---
name: useful-be
description: Load the minimum project and engineering context needed for useful work.
---

If `~/.agents/sdlc/MAIN.md` is absent or unreadable, report that exact path;
never search for another copy.

1. Read `~/.agents/sdlc/MAIN.md` in full.
2. Read the project's primary README and agent or contributor instructions.
3. If `.sdlc/project.yaml` exists, read it and the active work item's `spec.org`,
   `audits.yaml`, and `validation.org` when relevant.
4. Use the project profile and `MAIN.md` routing table to read only the applicable
   standards. Do not preload every language or domain document.
5. Inspect the repository state and affected area sufficiently to distinguish
   verified facts from assumptions.

Return a concise orientation: the requested outcome, controlling specification
with a descriptor, selected standards, relevant architecture, current evidence,
and unresolved material questions. Do not start implementation merely because
the context has been loaded.
