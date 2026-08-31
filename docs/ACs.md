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
- The seven retired workflow commands, retired drafting or design skills,
  `audit-acs`, root-level technology documents, and obsolete constitution
  addendum are backed up and removed from active discovery when their source
  paths are absent.
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

## Spec Kit project initialization

- The project initializer discovers Markdown technology standards
  alphabetically from `~/.agents/sdlc/technologies/` without a hard-coded
  technology list.
- Universal specification, coding, testing, documentation, Git, and entry-point
  standards are included exactly once.
- CLI values override project `.env` values, which override user SDLC defaults.
- External infrastructure ownership is optional and, when selected, records an
  owner descriptor and integration-contract path without assuming a private
  infrastructure project.
- Rendering is deterministic. When the rendered baseline and selections are
  current, the initializer asks nothing, writes nothing, and launches no agent.
- The baseline is written as the project `constitution-template` override, not
  over Spec Kit's core fallback template, and resolves without parsing preset
  manifests.
- A changed baseline is written atomically before the selected Codex, Claude,
  or Hermes harness receives the constitution-only semantic prompt.

## Independent audit contract

- `audit-spec`, `audit-design`, `audit-tests`, and `audit-code` run in contexts
  independent of the artefact author and never modify the judged artefact.
- Every report names the audit, auditor provider, auditor model, and exact PASS
  or FAIL verdict. Findings are classified as `[BLOCKING]` or `[ADVISORY]`.
- PASS permits numbered advisory findings but no blocking finding. FAIL requires
  at least one numbered blocking finding and may also contain advisories.
- Missing or malformed headers, unclassified findings, PASS with a blocking
  finding, and FAIL without a blocking finding are rejected.
- Specification PASS precedes planning; design PASS precedes tests and tasks;
  test PASS precedes implementation; code PASS precedes completion or
  convergence.
- A change to an audited artefact invalidates its PASS and requires a fresh
  audit whose receipt supersedes the previous result.
- The main authoring context remediates current-phase blocking findings and
  dispatches fresh independent audits without operator handback for at most five
  attempts, counting the first audit.
- Autonomous remediation stops earlier when it would change signed-off upstream
  authority or requires a human-controlled product, scope, security, privacy,
  access, persisted-data, external-contract, or irreversible decision.
- Phase handback occurs after PASS, the fifth FAIL, or an earlier human-controlled
  blocker and reports attempt history, decisions, assumptions requiring
  validation, advisories, and unresolved blockers.
