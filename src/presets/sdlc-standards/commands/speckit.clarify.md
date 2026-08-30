## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` and
`~/.agents/sdlc/ISSUES.md` in full. Prioritize ambiguities that can change
observable behaviour, scope, architecture, security, persisted data, access,
compatibility, or irreversible outcomes. Resolve those facts in the
specification rather than leaving them to implementation.

Read `.specify/memory/constitution.md` and its `Specification Baseline`. For a
brownfield delta, consult the named requirement authorities before treating an
existing behaviour as ambiguous. Clarify only the requested change, its
compatibility boundaries, or a genuine conflict in the baseline. Do not reopen,
duplicate, or silently replace an approved existing requirement.

Clarification changes invalidate an earlier `audit-spec` PASS. After
clarification is complete, run `audit-spec` in a fresh context and record its
current verdict before planning.

Before presenting a created or revised artefact for operator review, invoke the
executable named by `HTML_PREVIEW_TOOL` with the artefact paths. If it is
unavailable, open the artefacts in an available text editor; otherwise report
their exact paths. Do this once after the required audit PASS, or after
validation when no audit applies. Previewing is a presentation action, not
approval or a reason to stop.
