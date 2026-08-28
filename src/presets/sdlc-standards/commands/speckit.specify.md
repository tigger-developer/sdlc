## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` and
`~/.agents/sdlc/ISSUES.md` in full. Apply their defined-specification,
requirement-quality, acceptance-criteria, traceability, and identifier rules
throughout this command. Do not put implementation or test procedures into
requirements.

After specification and clarification are complete, the specification MUST
receive `audit-spec` PASS in a fresh agent context. Record the verdict in the
active feature's `audits.md`. Planning MUST NOT begin without a current PASS.
