# Goose integration

Select `goose` in the installer or the top-level `active-harness` list.
For explicit installation, use `--agent goose`; `--agent-home` overrides
the default `~/.config/goose`. The canonical installation is `/opt/sdlc`,
outside `~/.config`. SDLC skill directories are linked into Goose's `skills/`
directory and `~/.agents/skills` from `/opt/sdlc/skills`.
Personal instructions are a separate maintainer's responsibility.

The installer owns only its skill projections, the canonical
`/opt/sdlc/plugins/sdlc-command-guard-goose` plugin and its discovery link at
`~/.agents/plugins/sdlc-command-guard-goose`. It preserves
`config.yaml` and unrelated plugins. Existing conflicting plugin files
follow the installer's configuration-conflict handling.

## Command execution

The plugin uses Goose's native `PreToolUse` event, passes its JSON payload
to the shared command guard and sets `on_failure: block`. Policy rejection
returns exit status 2 before the tool runs. Hook startup errors and timeouts
must also block. `BeforeShellExecution` cannot veto execution and is not
a substitute. No legacy external-audit adapter is installed for Goose.

This requires a Goose build supporting user-level plugins, `PreToolUse`
and `on_failure: block`, as described in the
[upstream hook contract](https://github.com/aaif-goose/goose/blob/main/documentation/docs/guides/context-engineering/hooks.md).
No minimum released version has been established. The installer verifies the
local guard projection, not Goose's runtime discovery or hook execution. Its
report deliberately leaves native hook execution UNVERIFIED. Before relying
on the guard, verify in the intended Goose runtime that a harmless command
is permitted, a prohibited command is rejected and an unavailable hook blocks
execution. These are bounded acceptance checks, not metered regression tests.

Goose's [global instruction loader](https://github.com/aaif-goose/goose/blob/main/crates/goose/src/hints/load_hints.rs)
also reads shared `~/.agents/AGENTS.md`. If an installation contains the
same personal instructions there and in `~/.config/goose/AGENTS.md`, both
can enter context. The installers do not change Goose's context-file settings
or remove another harness's instructions.
