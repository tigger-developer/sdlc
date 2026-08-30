## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`,
`~/.agents/sdlc/TESTING.md`, and
`~/.agents/sdlc/DOCUMENTATION.md` in full, plus the standards selected by the
constitution that affect the tasks. Ensure the task set covers specification
evidence, error and boundary behaviour, documentation, migration, security, and
human validation where relevant.

Before generating tasks, verify that the plan and design have a current
`audit-design` PASS in the feature's `audits.md`. After test design and
traceability are complete, they MUST receive `audit-tests` PASS in a fresh
agent context. Implementation MUST NOT begin without that current PASS.

Before presenting a created or revised artefact for operator review, invoke the
executable named by `HTML_PREVIEW_TOOL` with the artefact paths. If it is
unavailable, open the artefacts in an available text editor; otherwise report
their exact paths. Do this once after the required audit PASS, or after
validation when no audit applies. Previewing is a presentation action, not
approval or a reason to stop.
