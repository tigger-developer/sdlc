---
description: End a discovery session. Default (promote): formalize the work into the central AC table with tests and close the issue. Alternative (rule out): git revert the discovery commits and close the issue. See ISSUES.md §"Discovery issues".
---

End the current discovery session. Read and follow `[sdlc-home]/ISSUES.md` §"Discovery issues".

**This skill is user-invoked only.** Per the provider-level coding safeguards, you may never invoke `/end-discovery` on your own initiative. The user decides when discovery ends.

Discovery is non-destructive in both paths. The user must indicate one of two outcomes:

## Path A: Promote (default - most discoveries)

The discovery iterations are real implementation. The discovery issue **is** the issue; no separate feature issue is drafted. Formalize the work into the central AC table and ensure test coverage.

1. **Summarize the implementation** in a comment on the discovery issue: what was built, citing specific commits by hash. This becomes the implementation documentation.

2. **Write ACs to `./docs/ACs.md`** capturing the system states the implementation now guarantees:
   - Allocate AC IDs scoped to the discovery issue number: `AC{discovery-issue}.{n}`.
   - Each AC describes a falsifiable system state (see ISSUES.md §"The AC/Test boundary" and §"Forbidden language in the AC column").
   - Add `Introduced: #{discovery-issue} (discovery)` to each AC's provenance lines.
   - Self-audit each AC before posting (see ISSUES.md §"Self-audit procedure").

3. **Write any missing tests** for the new ACs:
   - Test IDs scoped to the discovery issue: `RT-{discovery-issue}.{n}`.
   - Each AC must have at least one test; multi-condition ACs must have a test per condition.
   - Link each test in the central AC table's Tests column.

4. **Run the tests.** Confirm all green. If any fail: either fix the implementation (still under discovery scope) or remove the affected AC. Do not mark anything green without all tests passing.

5. **On all green: commit, close, mark.**
   - Commit the AC and test additions with message `Implement #N: promoted from discovery`.
   - Close the discovery issue: `gh issue close N` with the implementation summary as the closing comment.
   - Update `./docs/ACs.md` to mark each new AC and test ✅.

**End with:** `DISCOVERY ENDED - promoted as #N; ACs and tests in ./docs/ACs.md`. **STOP.**

## Path B: Rule out

The discovery direction turned out to be a dead end. Use `git revert` (non-destructive) to back out the changes; history preserves both the attempt and the revert.

1. **Identify the `wip(discovery):` commits** with `git log --grep "^wip(discovery)"`. Show me the commit hashes that will be reverted. Wait for explicit consent before proceeding.

2. **Revert the commits** in reverse chronological order: `git revert <hash>` for each. This creates new commits that undo the changes; the original commits remain in history.

3. **Summarize** what was tried and why it was ruled out, as a comment on the discovery issue.

4. **Close** the discovery issue: `gh issue close N` with the summary as the closing comment.

The `wip(discovery):` commits and their revert commits both stay in master history - they document the path that was tried and ruled out, which is useful if the question comes back.

**End with:** `DISCOVERY ENDED - ruled out; commits reverted`. **STOP.**
