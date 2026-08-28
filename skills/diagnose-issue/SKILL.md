---
name: diagnose-issue
description: Diagnose an observed software problem and recommend a fix without making file changes.
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` and only the standards relevant to the system
under diagnosis. Do not modify files.

Reproduce or otherwise isolate the observation when safe. Separate observed
facts from hypotheses. Trace the failure to its root cause, identify the
affected requirement with a descriptor, and recommend the smallest coherent
fix plus verification at the observable boundary.

If the cause is not established, say what remains unverified and name the next
diagnostic that would distinguish the leading hypotheses. Do not present a
plausible cause as a fact.
