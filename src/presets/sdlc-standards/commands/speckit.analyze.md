## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`,
`~/.agents/sdlc/ISSUES.md`, and
`~/.agents/sdlc/TESTING.md` in full. Check that the specification, plan, and
tasks agree; every material decision is defined at the correct layer; compound
requirements have complete evidence; and no test can pass by inspecting only
source or prose. Give every cited identifier an adjacent descriptor.

This consistency analysis does not replace `audit-spec`, `audit-design`,
`audit-tests`, or `audit-code`. A later relevant change preserves the earlier
PASS as revision-specific history but makes it non-current for phase completion
and requires a fresh independent audit of the delta and necessary context.
Follow `AUDITS.md` when resolving analysis findings within the current phase;
do not hand back between remediable revisions and their required fresh audits.
