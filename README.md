# SDLC for AI Coding Agents

An opinionated, issue-driven software development lifecycle for a human working with AI coding agents. The repository is the staging source for process rules, craft standards, workflow commands, and skills. The deployed `~/.agents/sdlc` tree is the canonical live version. Provider configuration remains in the provider's own home directory.

The framework began with Claude Code and now separates the durable SDLC from personal `AGENTS.md` and `CLAUDE.md` files. [LEARNINGS.md](LEARNINGS.md) preserves the failures, design reasoning, and accumulated guidance that produced the current rules.

## Repository Layout

| Path | Purpose |
|---|---|
| `src/` | Deployable SDLC entry point, process rules, and routed standards |
| `CHANGELOG.md` | High-level repository history from the public extraction onward |
| `LEARNINGS.md` | Design rationale, observed failures, and framework evolution |
| `commands/` | Human-invoked workflow prompts, including gate-adjacent commands |
| `skills/` | Advisory and drafting skills |
| `hooks/` | Optional command guard used by supported provider configuration |
| `templates/` | Installer inputs and provider-configuration examples; not deployed into agent instruction trees |
| `cmd/sdlc-install/` | Installer and configuration analyser |

The canonical live SDLC root is exactly `~/.agents/sdlc`.

## Quickstart

Clone the staging repository at a stable development path outside every
provider home:

```bash
git clone https://github.com/tigger-developer/sdlc.git ~/code/sdlc
```

Enter the clone and start the interactive installer:

```bash
make install
```

The installer detects the supported provider homes that exist and recursively
discovers runtime files beneath `src/`, `commands/`, `skills/`, and `hooks/`.
The contents of `src/` map directly to `~/.agents/sdlc`; the other runtime
directories retain their names. Files outside those roots are never agent
deployment candidates.

Each discovered file is compared with its destination by content and
permissions. The installer asks once before applying a batch only when a file
is absent or differs, or a managed provider-configuration fragment varies.
Matching destinations do not prompt.

The default plan lists only destinations that differ. Set `VERBOSE=1` to
include every matching destination:

```bash
VERBOSE=1 make install
```

Explicit provider mode remains available for inspecting or changing one
provider independently:

```bash
make build
```

```bash
bin/sdlc-install --agent codex
```

```bash
bin/sdlc-install --agent codex --configure
```

Replace `codex` with `claude`, `copilot`, or `hermes` as appropriate. For an
unrecognized provider, supply both `--agent custom` and `--agent-home`. Run
`make install-cli` only when a reusable `~/.local/bin/sdlc-install` link is
wanted.

## One Staging Clone, Multiple Agents

Keep one physical staging clone at a stable path. The installer copies it to
the canonical `~/.agents/sdlc` root and installs provider-native adapters; no
provider is linked to staging:

```bash
git clone https://github.com/tigger-developer/sdlc.git ~/code/sdlc
```

From `~/code/sdlc`, run one command:

```bash
make install
```

Accept the single deployment batch. A `git pull` updates staging only; an
accepted installer run is required to deploy that change.

### Manual alternative

The installer is optional. To reproduce its main deployment boundary manually,
first synchronize the staging tree while excluding Git metadata:

```bash
mkdir -p ~/.agents/sdlc
```

```bash
rsync -a --exclude=/.git/ ~/code/sdlc/ ~/.agents/sdlc/
```

Do not create provider-local SDLC copies. The automated installer also copies
skills and supported commands into provider-native locations. It never passes
`--delete` to rsync, so destination-only items are preserved.

## Connect the Agent to the SDLC

The repository deliberately does not supply a personal `AGENTS.md` or `CLAUDE.md`. Use [templates/AGENTS-or-CLAUDE.example.md](templates/AGENTS-or-CLAUDE.example.md) as the provider-level baseline and add the human-specific working preferences it calls out.

Common destinations are:

| Agent | Provider-level instructions | SDLC location |
|---|---|---|
| Claude Code | `~/.claude/CLAUDE.md` | `~/.agents/sdlc` |
| Codex | `~/.agents/AGENTS.md` | `~/.agents/sdlc` |
| Hermes | Private configuration outside this project | `~/.agents/sdlc` |
| Copilot or another agent | Provider-specific instruction file | `~/.agents/sdlc` |

Project-level `AGENTS.md` or `CLAUDE.md` files should describe only project-specific conventions and should rely on the provider-level file to load this framework.

## Loading Contract

For applicable coding work, the agent reads `MAIN.md` in full, selects the relevant root-level reference documents, and reads every selected document in full before acting.

The first response after loading must state:

- the complete list of selected documents read in full; and
- why each was selected.

The canary chain remains present on later responses. It proves which documents remain active, but it is not a substitute for the first-load report.

## Installer Behaviour

The installer has three deliberately separate responsibilities:

1. `make install` always compares the canonical `~/.agents/sdlc` tree and
   detects existing supported provider homes for native skills and commands.
2. It asks once before synchronizing the complete batch with `rsync --archive`,
   excluding `.git` and never using `--delete`.
3. It includes supported provider-configuration changes in the same preflight
   and confirmation batch. Explicit `--agent` mode retains `--apply` and
   `--configure` for automation or one-provider work.

The default installation plan contains only differing destinations. Set
`VERBOSE=1` to include matching destinations in the plan.

Before any existing SDLC-owned artefact is replaced, it is renamed beside the
live path as `<path>.<epoch>.bak`. This includes drifted files, stale links,
wrong destination types, and configuration files. The canonical copy is
written only after the backup succeeds. Provider adapters add only the
integrations their current public interfaces support:

| Agent | Commands | Skills |
|---|---|---|
| Claude Code | Ordinary files under `~/.claude/commands/` | Ordinary skill directories under `~/.claude/skills/` |
| Codex | Ordinary files under `~/.codex/prompts-commands/` | Ordinary skill directories under `~/.agents/skills/` |
| Copilot CLI | Ordinary files under `~/.copilot/prompts-commands/` | Ordinary skill directories under `~/.copilot/skills/` |
| Hermes | No command adapter | Ordinary skill directories under `~/.hermes/skills/` |
| Custom | No command adapter | No provider-specific assumptions |

Every top-level directory under staging `skills/` is recursively copied to its
supported destinations. Installation makes a skill discoverable; it does not
authorize invocation or alter the human-only gate rules.

Claude configuration analysis is based on `settings.json`; the confirmed change adds missing SDLC command restrictions to `permissions.deny`, removes the same restrictions from `permissions.allow`, and preserves unknown fields. JSON spacing and key order may be normalized. The installer refuses to replace a symlinked settings file and recommends editing its target manually. Codex analysis checks `config.toml`; the confirmed change creates `rules/sdlc.rules` when absent and can upgrade a recognized prior SDLC rules file while preserving unrelated rules. Ambiguous or non-regular destinations remain untouched. After migration, repeating the installer reports the current policy unchanged.

Hermes analysis structurally parses `config.yaml` and registers only the SDLC
command guard at the provider-neutral `~/.agents/sdlc` root for the `terminal`
tool. A recognized older registration under the Hermes home is replaced rather
than retained as a second hook. Private instructions and custom prompt text
belong to the user's private agent configuration, not this public project. The
merge compares only the managed hook semantics. A compliant file is retained
byte-for-byte regardless of comments, key order, quoting, or formatting. A
required migration preserves YAML comments, prompt text, unrelated values,
existing hooks, and first-use hook consent. Before rewriting an existing file,
it stores a recovery copy in an operating system temporary directory and prints
the path. Invalid YAML and non-regular configuration paths remain untouched.

Hermes must complete its startup TUI and model selection before installation.
If a Hermes home exists without `config.yaml`, the installer stops the complete
preflight with a visible diagnostic instead of manufacturing an incomplete
provider configuration.

The shared command guard at `hooks/agent-command-guard.sh` speaks Hermes's
`pre_tool_call` hook protocol. It prohibits `rm`, `sed`, `awk`, and direct
`python` and `python3` interpreter commands without blocking project entry
points such as `make test`.

Claude's current personal skill and legacy command locations are documented in its official [skills guide](https://code.claude.com/docs/en/slash-commands). Current Codex configuration, command-rule, and personal-skill semantics are documented in the official [configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference.md), [rules guide](https://learn.chatgpt.com/docs/agent-configuration/rules.md), and [skills guide](https://learn.chatgpt.com/docs/build-skills). Copilot personal skill and instruction locations are documented in GitHub's official [skill guide](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) and [custom-instruction guide](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions). Copilot and custom targets receive generic configuration recommendations until provider-specific permission changes are explicitly implemented and tested.

## Workflow Safety

Gate-equivalent workflows remain human-invoked commands. They must not be repackaged as implicitly selectable skills. Advisory skills may be installed where the provider supports them, but a skill cannot authorize its own use or advance an SDLC gate.

Configuration recommendations are advisory. Review every proposed permission change against the provider's current documentation and your own threat model before accepting it.

## Long-running delivery

MODE PAIR and MODE DELIVER are code-only concepts defined entirely by the SDLC
after `MAIN.md` is loaded for a coding-agent session. Provider instructions
decide when to load the SDLC but do not define its modes. Conversational
sessions do not load the SDLC and do not inherit either mode.

MODE DELIVER records its goal, scope, completion matrix, decisions, quality
checks, child status, and user tests on one master issue. The agent owns the
next action while delivery remains active. Progress reports are non-terminal,
and a final handback in MODE DELIVER is prohibited.

When no safe, authorized, executable agent action remains, the agent records a
delivery exit on the master. `DELIVERY READY` confirms that every autonomous
delivery action is evidenced and accounts for work awaiting a human-only
checkpoint; `DELIVERY BLOCKED` accounts for work stopped by a genuine blocker.
Either verdict atomically returns the session to MODE PAIR before control is
handed to the human. Routine reversible implementation ambiguity is recorded
and resolved without exiting delivery.

After context loss or in a new session, the human can resume an open delivery
master with:

```text
RESUME DELIVER n
```

The agent reloads the durable master state and continues the next unchecked
action without repeating delivery readiness. The directive cannot resume a
closed master or expand its recorded scope.

## Updating

Update the staging clone normally:

```bash
git pull
```

Provider homes continue to see the previous live deployment after `git pull`.
Re-run `make install` to review and accept the shared, provider-adapter, and
managed provider-configuration changes in one batch.

## Further Reading

- [CHANGELOG.md](CHANGELOG.md) records the public repository's high-level history.
- [LEARNINGS.md](LEARNINGS.md) explains why the framework is structured this way and records the failures behind the rules.
- [src/MAIN.md](src/MAIN.md) is normative for the SDLC process and deploys as `~/.agents/sdlc/MAIN.md`.
- The other standards under `src/` are normative only when `MAIN.md` selects them for the current task.

## Licence

Apache License 2.0. See [LICENSE](LICENSE).
