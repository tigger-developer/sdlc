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

## Route by project profile

If `~/.agents/sdlc/MAIN.md` is absent or unreadable, report that exact path;
never search for another copy.

- When the project root contains `.sdlc/project.yaml`, read
  `~/.agents/sdlc/MAIN.md` in full and follow its workflow and
  progressive-loading table.
- Otherwise, when writing or changing code or scripts, read
  `~/.agents/sdlc/CODING.md` and the technology standards it selects. The SDLC
  workflow, gates, and audits do not apply.

Do not load the SDLC for conversation, research, creative work, or unrelated
documentation. The presence of filesystem tools, Git integration, technical
subject matter, or this instruction file does not by itself authorize project
mutation.
