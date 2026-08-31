## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`, and the standards
selected by the constitution that apply to the current tasks. Do not preload
unselected technology or domain documents. Implement only defined specification
behaviour, preserve human edits, and report verification precisely.

Before implementation, verify that test design and traceability have a current
effective `audit-tests` PASS in the feature's `audits.md`. Apply every selected
test type. Before changing production code, write or amend each planned
automated regression test and confirm that it fails for the intended reason.
The audit PASS is not a substitute for this failing test execution. One-off and
user tests do not require a pre-change failure; execute and record them after
implementation. Where no automated regression test applies, preserve the
approved rationale. Explicitly selected paired development follows `PAIRING.md`
instead. After implementation and its verification are complete, the
implementation MUST receive `audit-code` PASS or satisfy a PROVISIONAL receipt
under `AUDITS.md`. Completion or convergence MUST NOT proceed without a current
effective PASS.
The main authoring context owns convergence under `AUDITS.md`: remediate
current-phase blockers and dispatch fresh audits without handback, up to five
total attempts, then request operator sign-off with the required decision and
assumption summary.
