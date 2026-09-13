---
name: useful-be
description: Load the minimum project and engineering context needed for useful work.
---

If `~/.agents/sdlc/MAIN.md` is absent or unreadable, report that exact path;
never search for another copy.

1. Read `~/.agents/sdlc/MAIN.md` in full.
2. Read the project's primary README and agent or contributor instructions.
3. Read `.sdlc/project.yaml` when present, then read the project's **Vision and
   Architecture** in full, including all configured product and architecture
   authority documents. These are required brownfield context, not optional
   background or substitutes for one another. If paths are not configured, use
   the README's references and locate the current Vision and Architecture in
   the project, allowing filename case and format variations. Report missing
   or ambiguous authorities; never silently substitute the README or a spec.
4. Read the active work item's `spec.org`, `audits.yaml`, and `validation.org`
   when relevant.
5. Use the project profile and `MAIN.md` routing table to read only the applicable
   standards. Do not preload every language or domain document.
6. Inspect the repository state and affected area sufficiently to distinguish
   verified facts from assumptions.

Return a concise orientation: the requested outcome, controlling specification
with a descriptor, selected standards, relevant architecture, current evidence,
and unresolved material questions. Do not start implementation merely because
the context has been loaded.
