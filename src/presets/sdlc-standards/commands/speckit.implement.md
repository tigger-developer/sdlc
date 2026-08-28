## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` and the standards selected by the constitution
that apply to the current tasks. Do not preload unselected technology or domain
documents. Implement only defined specification
behaviour, preserve human edits, and report verification precisely.

Before implementation, verify that test design and traceability have a current
`audit-tests` PASS in the feature's `audits.md`. After implementation and its
verification are complete, the implementation MUST receive `audit-code` PASS
in a fresh agent context. Completion or convergence MUST NOT proceed without a
current PASS.
