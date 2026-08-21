# Acceptance Criteria

## Public SDLC deployment

The copy-based deployment supersedes the former common-live-tree/provider-link
adapter model.

### Detection and confirmation

- `make install` detects only existing `~/.agents`, `~/.claude`, `~/.codex`,
  `~/.copilot`, and `~/.hermes` directories.
- It preflights every SDLC-owned destination before writing.
- One `yes` applies the complete detected batch; one refusal writes nothing.
- Explicit `--agent` mode limits provider deployment and retains deliberate
  `--apply` and `--configure` operations.

### Copies

- Each detected home receives an ordinary `sdlc/` directory copied from the
  staging repository without `.git` metadata.
- Every top-level SDLC skill is discovered from `skills/` and recursively
  copied to each supported skill home.
- Commands are copied to Claude `commands/` and Codex/Copilot
  `prompts-commands/`.
- No SDLC-owned deployed destination is a symlink.
- Rsync is never called with `--delete`; destination-only content survives.

### Shared ownership

- Every differing existing artefact, including stale symlinks, wrong types,
  and provider configuration, is renamed to `<path>.<epoch>.bak` before the
  canonical replacement is copied.
- A failed backup stops replacement of that artefact.
- The installer never claims an entire agent, skill, or command directory.
- Hermes configuration changes own only the SDLC command-guard hook and
  preserve private instructions and unrelated configuration.
- A repeated installation is idempotent and requires no confirmation.
