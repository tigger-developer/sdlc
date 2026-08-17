---
description: Open a discovery (sketch) session for exploratory or UX work. Opens an issue tagged `discovery` with `[DISCOVERY]` title prefix and NO AC table. Sketches only. See ISSUES.md §"Discovery issues".
---

Open a discovery session for the current task. Read and follow `[sdlc-home]/ISSUES.md` §"Discovery issues".

**This skill is user-invoked only.** Per the provider-level coding safeguards, you may never invoke `/start-discovery` on your own initiative. If you find yourself reaching for it, stop and surface the suggestion in chat instead.

**What discovery is for:** open-ended exploratory work where the spec is not yet known - UX sketching, "I'll know it when I see it" design, prototyping competing approaches. Not for small clearly-scoped tasks (that is what BYPASS-GATE-7 is for) and not for ordinary feature work (that is `/draft-issue`).

1. Confirm the discovery topic with the user in one short line.
2. Check if the `discovery` label exists on the repo. If not, create it with `gh label create discovery --color B7C3F3 --description "Exploratory sketch session"`. Use the `ask` permission if available; surface and pause if it is not.
3. Open the discovery issue with `gh issue create`:
   - Title: `[DISCOVERY] <topic>`
   - Label: `discovery`
   - Body: short problem framing only. Explicitly note: *"No AC table - this is a sketchbook. See ISSUES.md §Discovery issues."*
4. Report the issue URL.
5. State the working rules for the session:
   - Code may be written freely during discovery.
   - Commits during discovery must be prefixed `wip(discovery): ...`.
   - Sketches are throwaway by default; nothing is canonical until promoted.
   - The session ends with `/end-discovery` (promote or rule out) - the user calls that, not you. Discovery is non-destructive: sketch commits stay in history either way.

**End with:** `DISCOVERY ACTIVE - issue #NNN`. **Wait for the user to drive the session.**
