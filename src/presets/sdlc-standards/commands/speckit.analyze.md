## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`,
`~/.agents/sdlc/ISSUES.md`, and
`~/.agents/sdlc/TESTING.md` in full. Check that the specification, plan, and
tasks agree; every material decision is defined at the correct layer; compound
requirements have complete evidence; and no test can pass by inspecting only
source or prose. Give every cited identifier an adjacent descriptor.

This consistency analysis does not replace `audit-spec`, `audit-design`,
`audit-tests`, or `audit-code`. Any change made after an audit invalidates that
artefact's PASS and requires a fresh independent audit.
