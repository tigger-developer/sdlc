## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`,
`~/.agents/sdlc/CODING.md`, and `~/.agents/sdlc/TESTING.md` in full, plus the
constitution's selected standards that apply to the assessment. Distinguish
verified implementation evidence from remaining tasks and give every cited
identifier an adjacent descriptor.

Do not declare convergence while the implementation lacks a current
`audit-code` PASS in the feature's `audits.md`. Changes made during convergence
invalidate that PASS and require a fresh independent audit.
Follow the autonomous convergence and handback rules in `AUDITS.md` for any
resulting implementation and audit-code cycle.
