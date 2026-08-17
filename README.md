# SDLC for AI Coding Agents

An opinionated, issue-driven software development lifecycle for a human working with AI coding agents. The repository is the canonical source for process rules, craft standards, workflow commands, and advisory skills. Provider configuration remains in the provider's own home directory.

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

Clone the repository directly into one agent home:

```bash
git clone <repository-url> ~/.codex/sdlc
```

Enter the clone and inspect the proposed installation and configuration recommendations. Analysis is the default and makes no changes:

```bash
go run ./cmd/sdlc-install --agent codex
```

Apply the installation after reviewing the plan:

```bash
go run ./cmd/sdlc-install --agent codex --apply
```

Request provider-configuration changes separately. The installer shows the proposed change and asks for an ordinary yes-or-no confirmation before writing it:

```bash
go run ./cmd/sdlc-install --agent codex --apply --configure
```

Replace `codex` with `claude` or `copilot` as appropriate. For an unrecognized provider, supply both `--agent custom` and `--agent-home`.

## One Clone, Multiple Agents

Keep one physical clone at a stable path, then let each provider home contain a symlink named `sdlc`:

```bash
git clone <repository-url> ~/code/sdlc
```

From `~/code/sdlc`, inspect and then apply each target independently:

```bash
go run ./cmd/sdlc-install --agent claude --agent-home ~/.claude
```

```bash
go run ./cmd/sdlc-install --agent claude --agent-home ~/.claude --apply
```

```bash
go run ./cmd/sdlc-install --agent codex --agent-home ~/.codex --apply
```

```bash
go run ./cmd/sdlc-install --agent copilot --agent-home ~/.copilot --apply
```

Every provider then resolves its framework through `<agent-home>/sdlc`, while a single `git pull` updates the shared clone.

### Manual alternative

The installer is optional. From a stable clone at `~/code/sdlc`, create the same links manually:

```bash
ln -s ~/code/sdlc ~/.claude/sdlc
```

```bash
ln -s ~/code/sdlc ~/.codex/sdlc
```

```bash
ln -s ~/code/sdlc ~/.copilot/sdlc
```

Do not replace an existing path. Inspect it first and choose a different clone or agent-home path if it already contains unrelated material.

## Connect the Agent to the SDLC

The repository deliberately does not supply a personal `AGENTS.md` or `CLAUDE.md`. Use [templates/AGENTS-or-CLAUDE.example.md](templates/AGENTS-or-CLAUDE.example.md) as the provider-level baseline, add the human-specific working preferences it calls out, and replace `<agent-home>` with the actual provider home.

Common destinations are:

| Agent | Provider-level instructions | SDLC location |
|---|---|---|
| Claude Code | `~/.claude/CLAUDE.md` | `~/.claude/sdlc` |
| Codex | `~/.codex/AGENTS.md` | `~/.codex/sdlc` |
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

1. Detect the target provider from `--agent`, `--agent-home`, or a clone already located at `<agent-home>/sdlc`.
2. Analyse installation and known provider configuration, then print recommendations without writing by default.
3. Apply the `sdlc` symlink only with `--apply`, and alter supported provider configuration only with `--configure` plus interactive confirmation.

It never overwrites an existing non-matching destination. Provider adapters add only the integrations that their current public interfaces support:

| Agent | Commands | Skills |
|---|---|---|
| Claude Code | Individual links under `~/.claude/commands/` | Advisory skill links under `~/.claude/skills/` |
| Codex | `~/.codex/prompts-commands` as a reference library, not claimed as slash commands | Advisory skill links under `~/.agents/skills/` |
| Copilot CLI | `~/.copilot/prompts-commands` as a reference library | Advisory skill links under `~/.copilot/skills/` |
| Custom | Repository link only | No provider-specific assumptions |

Only advisory skills are linked automatically. Drafting, state-changing, and gate-adjacent workflows remain in the repository for deliberate, provider-specific invocation.

Claude configuration analysis is based on `settings.json`; the confirmed change adds missing SDLC command restrictions to `permissions.deny` while preserving unknown fields. JSON spacing and key order may be normalized. The installer refuses to replace a symlinked settings file and recommends editing its target manually. Codex analysis checks `config.toml`; the confirmed change creates the managed `rules/sdlc.rules` file only when it does not already exist. The installer never rewrites an existing Codex rules file.

Claude's current personal skill and legacy command locations are documented in its official [skills guide](https://code.claude.com/docs/en/slash-commands). Current Codex configuration, command-rule, and personal-skill semantics are documented in the official [configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference.md), [rules guide](https://learn.chatgpt.com/docs/agent-configuration/rules.md), and [skills guide](https://learn.chatgpt.com/docs/build-skills). Copilot personal skill and instruction locations are documented in GitHub's official [skill guide](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills) and [custom-instruction guide](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-custom-instructions). Copilot and custom targets receive generic configuration recommendations until provider-specific permission changes are explicitly implemented and tested.

## Workflow Safety

Gate-equivalent workflows remain human-invoked commands. They must not be repackaged as implicitly selectable skills. Advisory skills may be installed where the provider supports them, but a skill cannot authorize its own use or advance an SDLC gate.

Configuration recommendations are advisory. Review every proposed permission change against the provider's current documentation and your own threat model before accepting it.

## Updating

Update the canonical clone normally:

```bash
git pull
```

Symlinked agent homes see the update immediately. Re-run the installer without `--apply` to review newly available configuration recommendations before choosing whether to apply them.

## Further Reading

- [CHANGELOG.md](CHANGELOG.md) records the public repository's high-level history.
- [LEARNINGS.md](LEARNINGS.md) explains why the framework is structured this way and records the failures behind the rules.
- [MAIN.md](MAIN.md) is normative for the SDLC process.
- The other root-level standards are normative only when `MAIN.md` selects them for the current task.

## Licence

Apache License 2.0. See [LICENSE](LICENSE).
