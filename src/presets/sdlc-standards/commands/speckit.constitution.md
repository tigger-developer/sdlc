## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` in full. Use the constitution template generated
by `sdlc-project-init` as an immutable baseline. Read the named standards and
verified project documentation. Add only durable project-specific principles,
ownership boundaries, and explicit deviations.

Do not copy feature requirements or detailed architecture into the
constitution. Do not remove or weaken generated clauses, select an additional
standard without recording it, or resolve an unsupported fact by guessing.
Leave unresolved matters explicit. Produce only the constitution during this
operation. Spec Kit owns orchestration; the SDLC supplies engineering standards
and independent audit requirements.
