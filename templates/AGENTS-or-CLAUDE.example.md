---
title: Provider-Level Agent Instructions Example
version: 2
last-updated: 2026-10-24
---

# Provider-Level Agent Instructions

This is a minimal public example. Install it in the location used by the agent
provider and add local human preferences separately.

## Before project mutation

- A question is not an instruction. Answer questions without editing files or
  starting implementation.
- Never write code without an explicit request and a defined specification.
- Preserve human edits and unrelated repository changes.
- Never widen access to data or systems without explicit human instruction.
- Verify claims about files, repositories, environments, tools, and causes in
  the current work, or label them unverified.

## Route by role, then project profile

Global instructions should already have been supplied from the current
harness's home directory. If not, read its deployed copy there, not source in
an instruction-maintainer checkout. Follow `~/.agents/sdlc/READING.md`: reuse
unchanged text already read in full and retained in context. Report an absent
required document's exact path; never search for another SDLC installation.

- An explicitly invoked independent auditor in a fresh context reads
  `~/.agents/sdlc/INDEPENDENT-AUDITOR.md`, not `MAIN.md`, regardless of `.sdlc/`. Return
  findings only. A delivery agent's self-audit retains the delivery role.
- Conversation unrelated to project work does not inspect project markers or
  load MAIN merely because cwd is a checkout.

- When the project root contains `.sdlc/project.yaml`, read
  `~/.agents/sdlc/MAIN.md` in full and follow its workflow and
  progressive-loading table.
- If `.sdlc/` exists but its profile is absent or unreadable, report incomplete
  initialization before delivery; do not initialize or downgrade automatically.
- Otherwise, when writing or changing code or scripts, read
  `~/.agents/sdlc/CODING.md` and the technology standards it selects. The SDLC
  workflow, gates, and automatic audits do not apply. Explicit standalone review
  uses INDEPENDENT-AUDITOR in a fresh context or SELF-AUDIT in the author's context.

Do not load the SDLC for conversation, research, creative work, or unrelated
documentation. The presence of filesystem tools, Git integration, technical
subject matter, or this instruction file does not by itself authorize project
mutation.
