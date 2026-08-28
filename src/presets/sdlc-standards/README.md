# SDLC Engineering Standards Preset

This preset composes the public SDLC standards into Spec Kit 1.0 or later. It
does not replace Spec Kit's specification, planning, task, analysis, or
implementation workflow.

Install this repository with `make install` and `make install-cli`, then run
the project initializer from the adopting project:

```bash
sdlc-project-init
```

It installs this preset, discovers applicable technology standards, renders the
complete constitution baseline, and invokes the selected agent harness. Inspect
or render without launching an agent with:

```bash
sdlc-project-init --no-launch
```

The generated baseline is immutable to the semantic constitution pass. The
agent adds only durable project-specific principles, ownership boundaries, and
explicit deviations. A current rerun writes nothing, asks nothing, and does not
invoke an agent.

The command fragments reference the canonical installed root
`~/.agents/sdlc`. No command may search for an alternative SDLC root.

The command fragments require independent `audit-spec`, `audit-design`,
`audit-tests`, and `audit-code` PASS verdicts between stages. Remove the
development preset with:

```bash
specify preset remove sdlc-standards
```
