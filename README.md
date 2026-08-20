# SDLC for AI Coding Agents

An opinionated, issue-driven software development lifecycle for a human working with AI coding agents. The repository is the staging source for process rules, craft standards, workflow commands, and skills. The deployed `~/.agents/sdlc` tree is the canonical live version. Provider configuration remains in the provider's own home directory.

The framework began with Claude Code and now separates the durable SDLC from personal `AGENTS.md` and `CLAUDE.md` files. [LEARNINGS.md](LEARNINGS.md) preserves the failures, design reasoning, and accumulated guidance that produced the current rules.

## Repository Layout

| Path | Purpose |
|---|---|
| `MAIN.md` | SDLC entry point and process rules |
| `CHANGELOG.md` | High-level repository history from the public extraction onward |
| `LEARNINGS.md` | Design rationale, observed failures, and framework evolution |
| `ISSUES.md`, `TESTING.md`, `CODING.md`, `GIT.md`, `DOCUMENTATION.md` | Cross-cutting craft standards |
| `SHELL.md`, `PYTHON.md`, `PERL.md`, `GO.md`, `SWIFT.md`, `WEB.md` | Language-specific standards loaded only when relevant |
| `commands/` | Human-invoked workflow prompts, including gate-adjacent commands |
| `skills/` | Advisory and drafting skills |
| `hooks/` | Optional command guard used by supported provider configuration |
| `templates/` | Provider-instruction and configuration examples |
| `cmd/sdlc-install/` | Installer and configuration analyser |

`[sdlc-home]` always means the directory containing `MAIN.md`.

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

The installer detects the supported provider homes that exist, checks the
shared live deployment once, then checks each detected provider adapter. It
asks separately before changing the shared deployment or any provider. A
declined item is left unchanged, and matching items do not prompt.

Provider-configuration analysis remains an explicit operation. Build the
command, inspect one provider, and add `--apply` or `--configure` only when
that specific operation is wanted:

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

Keep one physical staging clone at a stable path. The installer synchronizes it
into the common live directory at `~/.agents/sdlc`, then points each provider's
`sdlc` adapter to that deployment:

```bash
git clone https://github.com/tigger-developer/sdlc.git ~/code/sdlc
```

From `~/code/sdlc`, run one command:

```bash
make install
```

Accept the shared deployment and whichever detected provider adapters should
be installed. Each accepted provider resolves its framework through
`<agent-home>/sdlc` into the same live copy. A `git pull` updates staging only;
an accepted installer run is required to deploy that change.

### Manual alternative

The installer is optional. To reproduce its main deployment boundary manually,
first synchronize the staging tree while excluding Git metadata:

```bash
mkdir -p ~/.agents/sdlc
```

```bash
rsync -a --delete --exclude=/.git --delete-excluded ~/code/sdlc/ ~/.agents/sdlc/
```

Then create provider adapters to the common live tree, never to staging:

```bash
ln -s ~/.agents/sdlc ~/.claude/sdlc
```

```bash
ln -s ~/.agents/sdlc ~/.codex/sdlc
```

```bash
ln -s ~/.agents/sdlc ~/.copilot/sdlc
```

```bash
ln -s ~/.agents/sdlc ~/.hermes/sdlc
```

The automated installer also creates common and provider skill links. Do not
replace an existing path manually. Inspect it first and use the installer to
migrate recognized staging links safely.

## Connect the Agent to the SDLC

The repository deliberately does not supply a personal `AGENTS.md` or `CLAUDE.md`. Use [templates/AGENTS-or-CLAUDE.example.md](templates/AGENTS-or-CLAUDE.example.md) as the provider-level baseline, add the human-specific working preferences it calls out, and replace `<agent-home>` with the actual provider home.

Common destinations are:

| Agent | Provider-level instructions | SDLC location |
|---|---|---|
| Claude Code | `~/.claude/CLAUDE.md` | `~/.claude/sdlc` |
| Codex | `~/.codex/AGENTS.md` | `~/.codex/sdlc` |
| Hermes | `~/.hermes/SOUL.md` plus `agent.system_prompt` | `~/.hermes/sdlc` |
| Copilot or another agent | Provider-specific instruction file | `<agent-home>/sdlc` |

Project-level `AGENTS.md` or `CLAUDE.md` files should describe only project-specific conventions and should rely on the provider-level file to load this framework.

## Loading Contract

For applicable coding work, the agent reads `MAIN.md` in full, selects the relevant root-level reference documents, and reads every selected document in full before acting.

The first response after loading must state:

- the complete list of selected documents read in full; and
- why each was selected.

The canary chain remains present on later responses. It proves which documents remain active, but it is not a substitute for the first-load report.

## Installer Behaviour

The installer has three deliberately separate responsibilities:

1. `make install` detects existing supported provider homes and compares only
   the shared deployment and adapters owned by this repository.
2. It asks once before synchronizing staging into `~/.agents/sdlc` with
   `rsync --archive --delete`, excluding `.git`, and once for each changed
   provider adapter set.
3. Explicit `--agent` mode analyses one provider and retains `--apply` and
   `--configure` for automation or deliberate provider-configuration work.

Interactive multi-agent installation never changes provider configuration.

It never overwrites an existing non-matching destination. Provider adapters add only the integrations that their current public interfaces support:

| Agent | Commands | Skills |
|---|---|---|
| Claude Code | Individual links under `~/.claude/commands/` to live command files | All skill links under `~/.claude/skills/` point to common skill entries |
| Codex | `~/.codex/prompts-commands` links to the live command library, not claimed as slash commands | All live SDLC skill entries are linked under `~/.agents/skills/` for direct discovery |
| Copilot CLI | `~/.copilot/prompts-commands` links to the live command library | All provider skill links point to common skill entries |
| Hermes | Live repository adapter only | All provider skill links point to common skill entries |
| Custom | Live repository adapter only | Common skill entries are installed without provider-specific assumptions |

Every top-level directory under staging `skills/` is deployed. Each common
entry at `~/.agents/skills/<skill>` points to
`~/.agents/sdlc/skills/<skill>`. Installation makes a skill discoverable; it
does not authorize invocation or alter the human-only gate rules.

Claude configuration analysis is based on `settings.json`; the confirmed change adds missing SDLC command restrictions to `permissions.deny`, removes the same restrictions from `permissions.allow`, and preserves unknown fields. JSON spacing and key order may be normalized. The installer refuses to replace a symlinked settings file and recommends editing its target manually. Codex analysis checks `config.toml`; the confirmed change creates `rules/sdlc.rules` when absent and can upgrade a recognized prior SDLC rules file while preserving unrelated rules. Ambiguous or non-regular destinations remain untouched. After migration, repeating the installer reports the current policy unchanged.

Hermes analysis structurally parses `config.yaml`, adds or updates one delimited operations-bootstrap block in `agent.system_prompt`, and registers the shared command guard for the `terminal` tool. It preserves custom prompt text, unrelated YAML values, existing hooks, and first-use hook consent. Before rewriting an existing file, it stores a recovery copy in an operating system temporary directory and prints the path. Invalid YAML, ambiguous management markers, and non-regular configuration paths remain untouched.

The shared command guard at `hooks/agent-command-guard.sh` speaks Hermes's `pre_tool_call` hook protocol. It prohibits direct `python` and `python3` interpreter commands, including path-qualified, compound, pipeline, and shell-wrapper forms, without blocking project entry points such as `make test`.

Claude's current personal skill and legacy command locations are documented in its official [skills guide](https://code.claude.com/docs/en/slash-commands). Current Codex configuration, command-rule, and personal-skill semantics are documented in the official [configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference.md), [rules guide](https://learn.chatgpt.com/docs/agent-configuration/rules.md), and [skills guide](https://learn.chatgpt.com/docs/build-skills). Copilot personal skill and instruction locations are documented in GitHub's official [skill guide](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) and [custom-instruction guide](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions). Copilot and custom targets receive generic configuration recommendations until provider-specific permission changes are explicitly implemented and tested.

## Workflow Safety

Gate-equivalent workflows remain human-invoked commands. They must not be repackaged as implicitly selectable skills. Advisory skills may be installed where the provider supports them, but a skill cannot authorize its own use or advance an SDLC gate.

Configuration recommendations are advisory. Review every proposed permission change against the provider's current documentation and your own threat model before accepting it.

## Updating

Update the staging clone normally:

```bash
git pull
```

Provider homes continue to see the previous live deployment after `git pull`.
Re-run `make install` to review and accept the shared and provider deployment
changes. Review provider-configuration recommendations separately with an
explicit `bin/sdlc-install --agent <name>` invocation.

## Further Reading

- [CHANGELOG.md](CHANGELOG.md) records the public repository's high-level history.
- [LEARNINGS.md](LEARNINGS.md) explains why the framework is structured this way and records the failures behind the rules.
- [MAIN.md](MAIN.md) is normative for the SDLC process.
- The other root-level standards are normative only when `MAIN.md` selects them for the current task.

## Licence

Apache License 2.0. See [LICENSE](LICENSE).
