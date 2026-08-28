---
name: recommendations-please
description: Provide concrete expert recommendations for a technical decision without making file changes.
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` and only the standards relevant to the decision.
Do not modify files.

Recommend the best one to three approaches. State the preferred choice first,
the evidence and assumptions behind it, its material trade-offs and risks, and
why the alternatives are weaker for the stated constraints. Identify any
decision that belongs in the specification before implementation.
