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

The installer detects the supported provider homes that exist, compares every
SDLC-owned destination, and asks once before applying the complete batch.
Matching destinations do not prompt.

The default plan lists only destinations that differ. Set `VERBOSE=1` to
include every matching destination:

```bash
VERBOSE=1 make install
```

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

Keep one physical staging clone at a stable path. The installer copies it into
the existing common and provider homes; no provider is linked to staging or to
another provider's copy:

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

Repeat the same ordinary copy for each installed provider:

```bash
rsync -a --exclude=/.git/ ~/code/sdlc/ ~/.claude/sdlc/
```

```bash
rsync -a --exclude=/.git/ ~/code/sdlc/ ~/.codex/sdlc/
```

```bash
rsync -a --exclude=/.git/ ~/code/sdlc/ ~/.copilot/sdlc/
```

```bash
rsync -a --exclude=/.git/ ~/code/sdlc/ ~/.hermes/sdlc/
```

The automated installer also copies skills and supported commands. It never
passes `--delete` to rsync, so destination-only items are preserved.

## Connect the Agent to the SDLC

The repository deliberately does not supply a personal `AGENTS.md` or `CLAUDE.md`. Use [templates/AGENTS-or-CLAUDE.example.md](templates/AGENTS-or-CLAUDE.example.md) as the provider-level baseline, add the human-specific working preferences it calls out, and replace `<agent-home>` with the actual provider home.

Common destinations are:

| Agent | Provider-level instructions | SDLC location |
|---|---|---|
| Claude Code | `~/.claude/CLAUDE.md` | `~/.claude/sdlc` |
| Codex | `~/.agents/AGENTS.md` | `~/.codex/sdlc` |
| Hermes | Private configuration outside this project | `~/.hermes/sdlc` |
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
   the copies, skills, and commands owned by this repository.
2. It asks once before synchronizing the complete batch with `rsync --archive`,
   excluding `.git` and never using `--delete`.
3. Explicit `--agent` mode analyses one provider and retains `--apply` and
   `--configure` for automation or deliberate provider-configuration work.

The default installation plan contains only differing destinations. Set
`VERBOSE=1` to include matching destinations in the plan.

Interactive multi-agent installation never changes provider configuration.

Before any existing SDLC-owned artefact is replaced, it is renamed beside the
live path as `<path>.<epoch>.bak`. This includes drifted files, stale links,
wrong destination types, and configuration files. The canonical copy is
written only after the backup succeeds. Provider copies add only the
integrations their current public interfaces support:

| Agent | Commands | Skills |
|---|---|---|
| Claude Code | Ordinary files under `~/.claude/commands/` | Ordinary skill directories under `~/.claude/skills/` |
| Codex | Ordinary files under `~/.codex/prompts-commands/` | Ordinary skill directories under `~/.agents/skills/` |
| Copilot CLI | Ordinary files under `~/.copilot/prompts-commands/` | Ordinary skill directories under `~/.copilot/skills/` |
| Hermes | No command adapter | Ordinary skill directories under `~/.hermes/skills/` |
| Custom | Repository copy only | No provider-specific assumptions |

Every top-level directory under staging `skills/` is recursively copied to its
supported destinations. Installation makes a skill discoverable; it does not
authorize invocation or alter the human-only gate rules.

Claude configuration analysis is based on `settings.json`; the confirmed change adds missing SDLC command restrictions to `permissions.deny`, removes the same restrictions from `permissions.allow`, and preserves unknown fields. JSON spacing and key order may be normalized. The installer refuses to replace a symlinked settings file and recommends editing its target manually. Codex analysis checks `config.toml`; the confirmed change creates `rules/sdlc.rules` when absent and can upgrade a recognized prior SDLC rules file while preserving unrelated rules. Ambiguous or non-regular destinations remain untouched. After migration, repeating the installer reports the current policy unchanged.

Hermes analysis structurally parses `config.yaml` and registers only the SDLC
command guard for the `terminal` tool. Private instructions and custom prompt
text belong to the user's private agent configuration, not this public
project. The merge preserves prompt text, unrelated YAML values, existing
hooks, and first-use hook consent. Before rewriting an existing file, it stores
a recovery copy in an operating system temporary directory and prints the
path. Invalid YAML and non-regular configuration paths remain untouched.

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
