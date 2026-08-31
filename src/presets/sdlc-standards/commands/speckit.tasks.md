## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`,
`~/.agents/sdlc/TESTING.md`, and
`~/.agents/sdlc/DOCUMENTATION.md` in full, plus the standards selected by the
constitution that affect the tasks. Ensure the task set covers specification
evidence, error and boundary behaviour, documentation, migration, security, and
human validation where relevant.

Before generating tasks, verify that the plan and design have a current
effective `audit-design` PASS in the feature's `audits.md`. After test design
and traceability are complete, they MUST receive `audit-tests` PASS or satisfy
a PROVISIONAL receipt under `AUDITS.md`. Implementation MUST NOT begin without
that current effective PASS.
The main authoring context owns convergence under `AUDITS.md`: remediate
current-phase blockers and dispatch fresh audits without handback, up to five
total attempts, then request operator sign-off with the required decision and
assumption summary.

After the applicable audit PASS, or validation when none applies, present
approval artefacts with `HTML_PREVIEW_TOOL`; otherwise use an available
non-blocking text editor or report the exact paths. Previewing is not approval
and must not stop the workflow.
