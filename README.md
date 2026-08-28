# Engineering Standards for AI Coding Agents

This public repository provides a provider-neutral engineering standards
library for coding agents. It is designed to complement
[GitHub Spec Kit](https://github.com/github/spec-kit): Spec Kit owns
specification and delivery orchestration; this repository supplies the coding,
testing, Git, documentation, language, and domain standards applied within it.

The library does not require private agent instructions. Its canonical installed
root is exactly `~/.agents/sdlc` for every supported provider.

## What is included

| Path | Purpose |
|---|---|
| `src/MAIN.md` | Compact universal rules and progressive-loading routes |
| `src/ISSUES.md` | Specification and acceptance-criteria standards |
| `src/TESTING.md` | Behavioural testing and evidence standards |
| `src/CODING.md` | Cross-language implementation standards |
| `src/GIT.md` | Source-control and recoverability standards |
| `src/DOCUMENTATION.md` | Public technical-documentation standards |
| `src/{GO,PYTHON,SHELL,PERL,SWIFT,WEB}.md` | Stack-specific standards |
| `src/presets/sdlc-standards/` | Spec Kit preset that selects standards progressively |
| `skills/` | Findings-only audits and advisory tools |
| `hooks/` | Optional provider-integrated command safeguard |
| `cmd/` and `internal/` | Installer implementation |

Only runtime material from `src/`, `skills/`, and `hooks/` is
deployed. Repository documentation, installer source, tests, and build metadata
do not enter the live instruction tree.

## Install the standards

Prerequisites are Go, `rsync`, and the tools required by the repository's build
targets.

```bash
git clone https://github.com/tigger-developer/sdlc.git ~/code/sdlc
```

```bash
cd ~/code/sdlc
```

```bash
make install
```

The installer:

- synchronizes `src/` into `~/.agents/sdlc` and retains the other runtime
  directory names;
- installs common and provider-native skill copies for detected agents where
  required for discovery;
- compares source and destination before prompting;
- lists only missing or differing destinations by default;
- backs up owned drift before replacement;
- backs up and retires the known command and drafting-skill paths removed by
  the standards-only model; and
- performs no deployment prompt or write when every detected destination is
  current.

Set `VERBOSE=1` to include matching destinations in the plan:

```bash
VERBOSE=1 make install
```

To inspect installer options:

```bash
make build
```

```bash
bin/sdlc-install --help
```

Provider-specific homes are adapters, not alternate SDLC roots. Agents must
load standards from `~/.agents/sdlc` and must never search the filesystem for a
different copy.

## Use with Spec Kit

Install Spec Kit according to its upstream documentation. In the project that
will use it, initialize the desired agent integration. For example:

```bash
specify init --here --integration codex --script sh --non-interactive
```

`--integration codex` selects the project-local Codex command adapter. It does
not launch Codex or make the standards Codex-specific. Choose the integration
for the agent interface through which the Spec Kit commands will be invoked;
the resulting `spec.md`, `plan.md`, `tasks.md`, and constitution remain project
artefacts.

Add the deployed standards preset:

```bash
specify preset add --dev ~/.agents/sdlc/presets/sdlc-standards
```

Inspect the composed template if desired:

```bash
specify preset resolve constitution-template
```

Invoke the active integration's rendering of `speckit.constitution` once. In the
current Codex skills integration this is `$speckit-constitution`; other
integrations render the same logical command in their native syntax. The command
reads `~/.agents/sdlc/MAIN.md`, selects only the standards applicable to the
project, and records them in the constitution's Engineering Standards Profile.
This is agent-assisted project setup; the operator does not hand-write a fresh
constitution.

Installing a preset does not rewrite an existing
`.specify/memory/constitution.md`. Invoke `speckit.constitution` to adopt or
update the composed template deliberately.

Thereafter:

- `speckit.specify` and `speckit.clarify` load specification standards;
- `speckit.plan` loads cross-language and selected stack standards;
- `speckit.tasks` loads testing and documentation standards;
- `speckit.analyze` checks consistency and coverage; and
- `speckit.implement` and `speckit.converge` load the standards named by the
  project profile.

`speckit.taskstoissues` retains Spec Kit's task artefacts as the source of truth
and applies the rule that human-facing identifiers always include descriptors.

The preset adds standards to these commands without replacing Spec Kit's
orchestration. The shared documents remain the single source of truth.

## Use without Spec Kit

Provider or project instructions may direct an agent to read
`~/.agents/sdlc/MAIN.md` when coding work begins. `MAIN.md` then routes only the
documents relevant to that work. An equivalent durable project specification is
required before code is written.

For a small emergency change before project Spec Kit artefacts exist, the human
may include `BYPASS-GATE-7` in the request. The request then serves as a
temporary specification for its stated outcome and scope. The exception does
not restore the retired SDLC modes, gates, or ticket lifecycle and does not
override safety or engineering standards.

The public template at
[`templates/AGENTS-or-CLAUDE.example.md`](templates/AGENTS-or-CLAUDE.example.md)
shows the minimum standalone integration. Adapt provider discovery filenames and
locations to current provider documentation.

## Skills and safeguards

The retained audit skills are findings-only:

- `audit-acs` challenges requirements and acceptance criteria;
- `audit-tests` challenges evidence and coverage; and
- `audit-code` reviews implementation against the selected standards.

Advisory skills diagnose problems, recommend technical decisions, summarize
open work, or load minimal project context. Skills do not approve their own
findings and do not define a delivery lifecycle.

The optional Hermes hook and provider rules reinforce the common prohibitions
on `rm`, `sed`, `awk`, and direct `python` or `python3` interpreter commands.
They do not prevent a documented project target from invoking its own runtime.

Hermes must create `~/.hermes/config.yaml` through its startup configuration
before the installer can safely amend it. If the Hermes home exists without that
file, installation stops with a prerequisite diagnostic.

## Updating and rollback

Update the staging clone, inspect the branch, and redeploy:

```bash
git pull --ff-only
```

```bash
make install
```

The `spec-kit-prototype` branch contains the standards-only model. Switching the
staging clone to another branch and rerunning `make install` redeploys that
branch's runtime files. Arbitrary destination-only files are preserved. The
prototype installer recognizes only the known commands and drafting skills
retired by this migration, renames them to adjacent `<path>.<epoch>.bak`
backups, and leaves all other destination-only material untouched.

## Further reading

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) describes source, deployment,
  loading, and Spec Kit composition.
- [`LEARNINGS.md`](LEARNINGS.md) records the design lessons behind the current
  standards model.
- [`CHANGELOG.md`](CHANGELOG.md) records repository changes.

## Licence

Apache License 2.0. See [`LICENSE`](LICENSE).
