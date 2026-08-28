# Acceptance Criteria

## Public SDLC deployment

The copy-based deployment supersedes the former common-live-tree/provider-link
adapter model.

### Detection and confirmation

- `make install` creates the common `~/.agents` home when it is absent and
  detects existing `~/.claude`, `~/.codex`, `~/.copilot`, and `~/.hermes`
  provider homes.
- It preflights every SDLC-owned destination before writing.
- One `yes` applies the complete detected batch; one refusal writes nothing.
- Explicit `--agent` mode limits provider deployment and retains deliberate
  `--apply` and `--configure` operations.

### Canonical runtime and provider adapters

- The canonical runtime exists only at `~/.agents/sdlc`; provider homes do not
  receive alternate SDLC trees.
- Files discovered recursively beneath `src/` deploy directly to the canonical
  root. Files beneath `skills/` and `hooks/` retain those directory names.
- README, changelog, learnings, project documentation, templates, installer
  source, tests, and build metadata are not runtime deployment candidates.
- Every top-level SDLC skill is discovered from `skills/` and recursively
  copied to each supported skill home.
- A branch that contains the optional legacy `commands/` source copies its
  commands to Claude `commands/` and Codex/Copilot `prompts-commands/`.
- No SDLC-owned deployed destination is a symlink.
- The seven retired workflow commands and four retired drafting or design skills
  are backed up and removed from active discovery when their source paths are
  absent.
- All other destination-only content survives deployment.

### Shared ownership

- Every differing existing artefact, including stale symlinks, wrong types,
  and provider configuration, is renamed to `<path>.<epoch>.bak` before the
  canonical replacement is copied.
- A failed backup stops replacement of that artefact.
- The installer never claims an entire agent, skill, or command directory.
- Deployment decisions are made per managed file. Repository-level changes
  outside the runtime roots do not create deployment variances.
- Legacy retirement decisions are limited to the explicit migration list and
  never infer ownership from an arbitrary destination-only path.
- Hermes configuration changes own only the SDLC command-guard hook and
  preserve private instructions and unrelated configuration.
- A repeated installation is idempotent and requires no confirmation.
