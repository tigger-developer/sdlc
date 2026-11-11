---
name: handover
description: Prepare a detailed Markdown handover of the current work for another agent, after confirming its destination with the user.
---

# Handover

Use this skill only when the user invokes it. Prepare a handover that lets a new agent resume the current work without relying on this conversation.

1. Establish the current working directory and project scope. Propose one destination to the user before writing: `[cwd]/[hyphen-delimited-descriptor-of-what-we-have-been-doing].md`. Use a short, specific, lowercase descriptor drawn from the actual task. Show the expanded absolute path and ask the user to confirm it or provide another permitted destination. Do not write the handover until the user confirms the destination. If the path exists, inspect it and ask before replacing or appending; preserve existing content.
2. Gather as much relevant detail as the current accessible context supports. Inspect current repository state and important source or evidence paths when needed to distinguish current observations from recollection. Do not claim to export hidden instructions, inaccessible earlier turns, private reasoning, or an exact transcript. Do not copy credentials, tokens, private keys, or unnecessary personal data. Identify any material detail omitted for that reason.
3. Write a self-contained Markdown note with the task and intended outcome; project and working directory; governing instructions and workflow route; user decisions and exact constraints; work performed; files changed; current Git branch, revision and worktree state; verification performed with commands, results and limits; open questions, blockers and risks; and concrete next steps in order. Include useful paths, identifiers with descriptors, and dates or revision references where they matter. Distinguish observed facts, user statements, decisions, hypotheses, and unverified recollection. Preserve important wording verbatim when its exact form matters and is available.
4. Follow the governing project's write, documentation, and Git rules for the destination. A handover note is a record, not approval, validation, closure, or authority to perform further actions. Do not change unrelated files to make the note easier to write. Verify that the file exists and contains the intended sections, then report its path and any material limitations.
