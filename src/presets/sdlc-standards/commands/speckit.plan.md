## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`,
`~/.agents/sdlc/CODING.md`, `~/.agents/sdlc/GIT.md`, and
`~/.agents/sdlc/DOCUMENTATION.md` in full. Read only the additional language
and domain documents selected by the constitution's Engineering Standards
Profile. Apply them to architecture, dependencies, interfaces, migration,
security, test architecture, compatibility, and operational design.

If the plan needs a standard not yet selected, amend the standards profile. Do
not duplicate the standards text in `plan.md`; record project decisions and
references.

Before planning, verify that the active specification has a current
`audit-spec` PASS in the feature's `audits.md`. After the plan and design are
complete, they MUST receive `audit-design` PASS in a fresh agent context. Test
design and task generation MUST NOT begin without that current PASS.

Before presenting a created or revised artefact for operator review, invoke the
executable named by `HTML_PREVIEW_TOOL` with the artefact paths. If it is
unavailable, open the artefacts in an available text editor; otherwise report
their exact paths. Do this once after the required audit PASS, or after
validation when no audit applies. Previewing is a presentation action, not
approval or a reason to stop.
