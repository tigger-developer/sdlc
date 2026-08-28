# SDLC Engineering Standards Preset

This preset composes the public SDLC standards into Spec Kit 1.0 or later. It
does not replace Spec Kit's specification, planning, task, analysis, or
implementation workflow.

Install this repository with `make install`, then add the prototype from an
initialized Spec Kit project:

```bash
specify preset add --dev ~/.agents/sdlc/presets/sdlc-standards
```

Inspect the composed constitution template:

```bash
specify preset resolve constitution-template
```

Then invoke the active agent integration's `speckit.constitution` command once.
The agent selects the applicable standards and records a concise Engineering
Standards Profile in `.specify/memory/constitution.md`. Existing constitutions
are not rewritten merely because the preset was installed.

The command fragments reference the canonical installed root
`~/.agents/sdlc`. No command may search for an alternative SDLC root.

Remove the development preset with:

```bash
specify preset remove sdlc-standards
```
