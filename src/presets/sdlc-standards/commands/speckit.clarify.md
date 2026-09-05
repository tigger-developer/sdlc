## SDLC standards

If `~/.agents/sdlc/MAIN.md` is absent or unreadable, report that exact path;
never search for another copy.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`, and
`~/.agents/sdlc/ISSUES.md` in full. Prioritize ambiguities that can change
observable behaviour, scope, architecture, security, persisted data, access,
compatibility, or irreversible outcomes. Resolve those facts in the
specification rather than leaving them to implementation.

After applying clarification answers, regenerate the opening Specification
Summary from the detailed specification. Retain its exact labels and `***`
section break. Ensure it neither introduces nor contradicts anything in the
current specification and omits no material content.

Read `.specify/memory/constitution.md` and its `Specification Baseline`. For a
brownfield delta, preserve the specification's context pass and consult the
requirement, design, historical-work, regression-test, or implementation source
that bears on an apparent ambiguity before treating existing behaviour as
undefined. Clarify only the requested change, its compatibility boundaries, or
a genuine conflict in the baseline. Do not reopen, duplicate, or silently
replace an approved existing requirement.

Clarification changes preserve an earlier `audit-spec` PASS as
revision-specific history but make it non-current. After clarification is
complete, run `audit-spec` in a fresh context and record its current verdict
before planning. A satisfied PROVISIONAL receipt is an effective PASS under
`AUDITS.md`.
Converge under the autonomous contract in `AUDITS.md` before requesting
operator sign-off.

After the applicable audit PASS, or validation when none applies, present
approval artefacts with `HTML_PREVIEW_TOOL`; otherwise use an available
non-blocking text editor or report the exact paths. Previewing is not approval
and must not stop the workflow.
