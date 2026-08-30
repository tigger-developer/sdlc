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

After the applicable audit PASS, or validation when none applies, present
approval artefacts with `HTML_PREVIEW_TOOL`; otherwise use an available
non-blocking text editor or report the exact paths. Previewing is not approval
and must not stop the workflow.
