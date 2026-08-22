---
description: Migrate ACs from a legacy (closed) issue into the central document `./docs/ACs.md`. See ISSUES.md §"Legacy AC migration".
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Do not search, enumerate directories, traverse mounted volumes, inspect network shares, or use `find`, `locate`, Spotlight, or equivalent discovery to resolve this path. If `~/.agents/sdlc/MAIN.md` is absent or unreadable, stop and report that exact path.

Migrate the ACs from a single legacy issue into `./docs/ACs.md`. Read and follow `~/.agents/sdlc/ISSUES.md` §"Central AC document" and §"Legacy AC migration".

**This skill is user-invoked only.** Per the provider-level coding safeguards, you may never invoke `/migrate-acs` on your own initiative. Migration is the user's call.

**Usage:** `/migrate-acs <issue-number>`

If no issue number was provided, ask which issue to migrate. Do not proceed without one.

1. Read the legacy issue: `gh issue view N`.
2. Locate the AC table in the issue body (or first comment if it lived there).
3. **Validate ACs against current ISSUES.md standards** - do not auto-fix. Surface violations in chat for review:
   - Forbidden language in the AC column (see §"Forbidden language").
   - Single-test coverage where multiple conditions are implied (§"Multi-condition ACs").
   - Pass the litmus test (§"The litmus test").
   - If any AC fails: stop and report. Decide with the user whether to migrate as-is (preserves history), rewrite during migration (loses provenance integrity), or skip until rewritten.
4. Determine the feature area for grouping in `./docs/ACs.md`. If the doc does not yet exist, create it with the header template from ISSUES.md §"Central AC document". Ask the user for the feature-area name if unclear.
5. For each AC:
   - Copy the AC text verbatim.
   - Preserve the original ID (`AC{n}.{m}`) - do not renumber.
   - Add provenance lines:
     ```markdown
     - Introduced: #{issue} (closed YYYY-MM-DD)
     - Migrated: YYYY-MM-DD
     - Tests: <test IDs preserved from the issue>
     - Status: <current status from the issue>
     ```
6. Update the cutover marker at the top of `./docs/ACs.md`:
   ```
   Last migrated: AC{n}.{m} from #{issue} on YYYY-MM-DD
   ```
7. **Do not edit the originating issue body.** It remains the historical record.
8. Commit with `chore(acs): migrate ACs from #N to docs/ACs.md`.
9. Report what was migrated, what was skipped, and any AC quality issues for review.

**End with:** `MIGRATION COMPLETE - <count> ACs migrated from #N`. **STOP.**
