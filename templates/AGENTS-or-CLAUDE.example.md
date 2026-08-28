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

## Load the standards only for coding work

The canonical SDLC root is exactly `~/.agents/sdlc`. Do not search, enumerate
directories, traverse mounted volumes, inspect network shares, or use `find`,
`locate`, Spotlight, or equivalent discovery to resolve it. If
`~/.agents/sdlc/MAIN.md` is absent or unreadable, report that exact path.

When handling code, scripts, software configuration, builds, tests, deployment,
or systems work:

1. Read `~/.agents/sdlc/MAIN.md` in full.
2. Follow its progressive-loading table.
3. Read only the additional standards selected by the project constitution or
   needed for the current work.

Do not load the SDLC for conversation, research, creative work, or unrelated
documentation. The presence of filesystem tools, Git integration, technical
subject matter, or this instruction file does not by itself authorize project
mutation.

Spec Kit or the project's own durable artefacts define its workflow. The SDLC
adds engineering standards; it does not add approval gates, operating modes, or
a ticket lifecycle.
