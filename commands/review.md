---
description: Full review of the current issue's implementation -- run tests, check standards, demo user tests.
---

Perform a full review of the current issue's implementation. Allocate a unique review directory with `mktemp -d` before running tests. Long-form output (test logs, standards check, summary) is written there; the terminal stays short. Retain the directory only for the active review and remove that exact directory through explicit teardown afterwards.

1. Run `make test` (includes lint) and capture the full output in the review directory -- unless it has just been run and no changes have been made since then, in which case skip and note this. **In chat: one line stating pass/fail and total counts.** Save the full output for the review file in step 10.
2. Verify: zero errors, no new warnings. These are hard blocks -- no exceptions. **In chat: one line confirming hard blocks pass.**
3. Verify repository hygiene: `.agent/`, `.agents/`, `.claude/`, and `.codex/` are ignored; no paths beneath them are tracked; and no staged or work-generated file exceeds the documented project size limit, or 100 MiB when none is documented. Any violation is a hard block. **In chat: one line confirming runtime-state and size checks pass.**
4. Identify each coding standard section you checked against, by name and section number. Save the list for the review file in step 10 -- do not paste it in chat.
5. For each UT: launch the application/tool, prepare representative data and configuration, and place it at the exact state I must inspect. Show me what is on screen and ask "Does this pass UT-{issue}.{n}?" as a yes/no question. Never ask me to run commands, perform setup, or navigate through a series of actions. I should need only visual or subjective judgement and, when the UT inherently requires interaction, at most one reasonable action. Split independent actions into separate UTs. **(This step stays in the terminal -- it is interactive.)**
6. Update the AC table **in place**. Update automated test statuses. Leave UTs as pending until I answer them in step 5 (they will already be answered by this point).
7. Update project documentation as appropriate. **Help text counts as documentation** -- if this issue changed any executable's flags, behaviour, defaults, or user-visible output, update the corresponding `docs/<command>-help.md` (or wherever the help text lives) before APPROVED.
8. Commit with message `Implement #[n]: [short description]` and push.
9. Add a comment to the issue: implementation details, testing instructions, commit link.
10. Allocate a unique report directory with `mktemp -d` and write `review-<NNN>.md` there. Retain it only for the active review and remove the exact directory through explicit teardown afterwards. Include:
   - Issue restatement (title, summary, link)
   - Full `make test` output from step 1
   - Standards checked: each section by name and number (from step 4)
   - AC table snapshot or link to the updated issue
   - UT results from step 5
   - Summary of all actions taken
   - Commit hash and push confirmation
11. Handle the report according to the active mode:
    - **MODE PAIR:** If `HTML_PREVIEW_TOOL` names an available executable, use it to render and open the Markdown report. Treat its value as one executable path, not shell code or a compound command. If it is unset or unavailable, choose an available text editor and open the Markdown report there. **In chat: one line stating the rendered or Markdown report path.**
    - **MODE DELIVER:** Do not render or open the report. Record the Markdown report path and relevant evidence on the affected issue, then continue with the next delivery action. A review report is an artefact, not a checkpoint or terminal condition.

In MODE PAIR, do not seek APPROVED or end with READY FOR REVIEW while any UT is pending, unpresented, or failing.

In MODE PAIR, end with `READY FOR REVIEW - issue #NNN` and the issue link, then stop at the human review gate. In MODE DELIVER, ignore `HTML_PREVIEW_TOOL`, do not open the report in a renderer or editor, and do not emit READY FOR REVIEW or stop here; return control to the delivery workflow and continue.
