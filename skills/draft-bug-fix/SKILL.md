---
name: draft-bug-fix
description: Draft a bug-fix issue that references existing ACs. No new AC table is created. See ISSUES.md §"Bug-fix issues reference existing ACs".
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Do not search, enumerate directories, traverse mounted volumes, inspect network shares, or use `find`, `locate`, Spotlight, or equivalent discovery to resolve this path. If `~/.agents/sdlc/MAIN.md` is absent or unreadable, stop and report that exact path.

Draft a bug-fix issue for the current task. Read and follow `~/.agents/sdlc/ISSUES.md` - in particular the §"Bug-fix issues reference existing ACs" section.

**Critical rule:** a bug-fix issue does **not** contain its own AC table. It references the AC(s) the bug violates, by ID, from the central document `./docs/ACs.md` (or the originating issue if not yet migrated).

1. Identify the bug. Reproduce or confirm reproduction. Treat the user's report as authoritative (`MAIN.md` §"Bug Reports").
2. **Locate the violated AC.** Search `./docs/ACs.md` first, then closed issues if needed.
   - If found in `./docs/ACs.md`: cite as `AC{n}.{m}` with a link.
   - If found in a legacy issue (not yet migrated): cite as `AC{n}.{m} (legacy - see #N)`. Consider whether to migrate it first via `/migrate-acs N` - migrate when the AC is load-bearing and likely to be cited again.
   - If no AC covers the bug, choose between these paths from the evidence in MODE DELIVER; in MODE PAIR, STOP and surface both options to the user:
     - **Backfill:** the underlying feature has an AC table but the bug exposes a gap. Backfill the AC on the original feature, then proceed.
     - **New feature:** the underlying behaviour was never specified. Use `/draft-issue` instead - this is a new feature, not a bug fix.
   - Do not invent a new AC inside the bug-fix issue.
3. Draft the bug issue body:
   - Problem statement: what is observed, what is expected.
   - Reproduction steps.
   - Violated AC(s) cited by ID with link.
   - Solution outline (see ISSUES.md §"Solution section rules").
   - Test plan: regression tests to add under the original AC's coverage. These prove the violated AC now holds. Allocate test IDs scoped to the bug issue number (e.g. `RT-{bug-issue}.1`).
4. Self-audit:
   - No new AC table in this issue body. ✓
   - Violated AC cited by ID with link. ✓
   - Regression tests target the original AC, not a new parallel one. ✓
5. Post the issue.

In MODE PAIR, end with `AWAITING PROCEED - issue #NNN`, the issue link, and **STOP**; no code may be written until PROCEED is received. In MODE DELIVER, return the issue link and completed design to the delivery workflow for the mandatory per-ticket AC and test audits; do not request PROCEED. If choosing between backfill and a new feature requires a material product decision, make a genuine-blocker handback instead of guessing.
