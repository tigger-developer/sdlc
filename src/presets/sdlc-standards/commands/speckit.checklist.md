## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` and
`~/.agents/sdlc/ISSUES.md` in full. Build checklist items that assess the
quality, completeness, consistency, and boundaries of requirements. Do not turn
the checklist into implementation steps or source-text assertions.

Before presenting a created or revised artefact for operator review, invoke the
executable named by `HTML_PREVIEW_TOOL` with the artefact paths. If it is
unavailable, open the artefacts in an available text editor; otherwise report
their exact paths. Do this once after the required audit PASS, or after
validation when no audit applies. Previewing is a presentation action, not
approval or a reason to stop.
