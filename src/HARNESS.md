---
title: SDLC Audit Harness Guide
version: 10
last-updated: 2026-10-11
---

# SDLC audit harness guide

Internal SDLC helper for running one bounded provider context against
original evidence files with SHA-256 integrity checks. It is installed at
`~/.agents/sdlc/bin/sdlc-harness`. The operator-facing `sdlc-audit` command is a
symlink to this same executable; it accepts the same options and subcommands.
SDLC skills continue to use the resolved absolute harness path.

## Usage

    sdlc-audit --input FILE [--input FILE ...] [options] < context.txt
    sdlc-harness --gate GATE --audit-record FILE --work-item ID [options] < audit-context.txt
    sdlc-harness goal-config [options]

## Standalone standards review

When `--project` has no `.sdlc/project.yaml`, `sdlc-audit` runs a standalone
code and documentation standards review. Supply at least one original project
file with `--input`; omit `--gate`, `--audit-record`, `--work-item` and `--reset`.
The harness prints a `STANDALONE MODE` warning and supplies `STANDALONE-AUDIT.md`
first, followed by the entire installed Markdown standards library as verified
text. Every document must be read once before review; only applicable engineering
rules are enforced. Reading workflow documents does not elect that workflow. It asks for an
`AUDIT: standalone` result. This route does not create SDLC project state or
apply ticket readiness, gate budgets, or closure rules. Supply relevant
project authorities as additional inputs when needed; installed technology
standards are harness-owned.

`--agent-spike` selects a separate spike profile in a project that has
`.sdlc/project.yaml`. The `hands-off-spike` skill uses it so exploratory code
is always independently reviewed without entering an SDLC gate, budget, or
`audits.yaml` record. The harness prints a `SPIKE MODE` warning instead of
`STANDALONE MODE`. The flag accepts only the audit phase and the standalone
options above; every path, anchor and verdict check is unchanged. `SPIKE-AUDIT.md`
routes the review against all coding, architecture, security, Git, technology
and TDD/testing standards. Full SDLC workflow, gates and documentation requirements
do not apply. The response retains the `AUDIT: standalone` envelope for compatibility.

For every audit, the harness derives the project root from its invocation
directory: the enclosing Git root when present, otherwise that directory.
An explicit `--project` must resolve to the same root. Supplied project evidence
and the audit record must remain inside that tree; paths under provider runtime
directories (`.agent`, `.agents`, `.claude`, `.codex`, `.copilot`, `.hermes`,
`.git`) or directories named `tmp`, `temp`, `temporary`, or their dotted,
hyphenated or underscored variants are refused. Other hidden directories,
including `.github`, remain eligible. System temporary roots are also refused,
including symlink aliases. The harness performs these checks before provider
execution.

The auditor receives a bounded inventory of tracked and non-ignored untracked
files in a Git project, or eligible files under a non-Git invocation directory.
It excludes the same runtime and scratch paths. The inventory exposes possible
omissions; it contains paths, not file contents or proof of review. The author
must still supply the relevant source, tests and authorities as `--input`
evidence. A caller that changes the process working directory to another
non-temporary checkout can still present that checkout as the invocation
project; the harness cannot establish its relationship to an earlier session.

## Implementation readiness

Before implementation review, use `sdlc-validate --help` and validate the spec's
records locally. The harness repeats the check using `spec.org` and `validation.org`
beside the required `--audit-record`, adding both to hashed evidence automatically.
Schema/readiness rejection precedes cache lookup and provider startup; it returns
YAML on stdout and consumes no audit round. See `ORG-SCHEMA.md` for the contract.

Use `--gate test-code` to review written tests and minimal execution-enabling
scaffolding before implementing the target behaviour. Resume that context with
`--gate delivery-code` after RTs and OTs are GREEN and affected product docs are
current. UTs may remain AMBER. With no approved RTs, skip test-code and start
delivery-code directly. Paired/emergency routes also start at delivery-code.

The gate selects the readiness check; the audit command no longer accepts
`--readiness-check` or `--gate implementation`. Use `--readiness-check` only with
`sdlc-validate`. Omitting the audit gate is an error. For record compatibility,
both implementation gates share the existing `gate: implementation` session key
and store the selected gate in `readiness_check` per attempt and on the entry.
Response envelopes use `GATE: test-code` or `GATE: delivery-code`. Test-code PASS
is not delivery approval. Each review has a separate configured verdict budget; their prompts
and cache keys differ. Existing history and session IDs remain intact.

## Goal configuration

`goal-config` is read-only. It returns JSON integers `max_turns` and
`max_token_budget` for the delivery skill; it does not invoke a provider,
activate a goal, write configuration or enforce native continuation itself.

    /absolute/path/to/.agents/sdlc/bin/sdlc-harness goal-config --project .

- `--project DIR`: audit project root; defaults to the invocation Git root or,
  outside Git, the invocation directory.
- `--global-config FILE`: defaults to `~/.agents/sdlc.yaml`.
- `--sdlc-root DIR`: schema root; defaults to the deployed executable's parent
  directory. Source-tree checks may pass `--sdlc-root src`.
- `--goal-max-turns VALUE` and `--goal-max-token-budget VALUE`: explicit overrides.
- Precedence: these overrides, `SDLC_GOAL_MAX_TURNS` / `SDLC_GOAL_MAX_TOKEN_BUDGET`,
  project YAML, global YAML, then `config/project-init.schema.yaml` defaults.
- Defaults: 20 continuations for Hermes/Copilot; 100,000 tokens for Codex.
  Claude's documented eight-consecutive-Stop-hook-continuation cap is separate.
- Accept `100000` or `"100,000"`. Reject periods (including unquoted `100.000`),
  malformed groups, signs, zero, leading zeroes, scientific notation and values
  above 9,007,199,254,740,991, the exact JSON integer limit. Do not infer locale.
- stdout contains JSON only on success. Configuration errors produce no budget
  output and exit nonzero. Audit invocation is independent of goal configuration.

## Normal invocation

Supply the work item, audit record, gate and evidence. The harness starts or resumes
the appropriate provider context automatically. There is no public session-ID option.
Legacy start/resume operation words remain accepted for transition, but both select
the same automatic audit behaviour. The old reset-session flag is replaced by reset.

## Evidence and instruction

- For SDLC gates, supply project evidence only. The harness adds its full installed
  SDLC Markdown inventory to the hashed manifest. Gate-specific core standards
  and the technologies selected in `.sdlc/project.yaml` are mandatory. Source
  extensions supplement technology selection; Markdown and Org evidence add
  their format rules. Node includes JavaScript; Hugo includes Web. The
  auditor judges which other documents apply. Do not pass `~/.agents/sdlc` files as `--input`;
  external caller-supplied paths are rejected. The harness also includes the
  project profile. Supply relevant project authorities separately.
- Repeat `--input <file>` for every exact evidence file.
- Paths are resolved against `--project`. Every provider receives hash-verified
  UTF-8 contents over stdin and is launched from the project. No document copies
  are retained. Supplied project evidence must be sufficient for the review;
  omitted files are an evidence gap rather than permission to use native file tools.
- Claude audits disable built-in tools, use safe mode and strict MCP configuration,
  and deny native reads of installed standards. Codex audits disable shell, external
  integrations and other tool features with invocation-local options. Codex ignores
  user configuration while retaining its credential store: custom model-provider
  definitions in that configuration are unavailable to this isolated audit route.
  The CLI must support these options; an unsupported invocation fails closed.
- Hermes audits use an embedded bridge into the installed native agent interface
  to expose response callbacks hidden by quiet CLI output. The fixed
  `chat --quiet --query-file -` argument contract and native resume identity are
  retained. Native `session_id:` metadata is checkpointed before generation;
  model-written identities are not trusted. The bridge requires `uv` on PATH
  and a Hermes console script with an absolute Python shebang, or the fixed
  Homebrew libexec layout. It runs offline using that existing interpreter and
  packages, without downloading or modifying the Hermes runtime. Unsupported
  layouts or missing native interfaces fail explicitly. Safe mode skips user
  configuration, rules, context files and memory; the bridge clears an inherited
  `HERMES_EPHEMERAL_SYSTEM_PROMPT`. Hermes built-in instructions remain.
  Native tools and background review are disabled. Credentials and native
  session persistence remain owned by Hermes.
- Hermes and Copilot also run with file tools disabled. The harness selects and
  supplies mandatory standards automatically for each route. Further references
  use a bounded `STANDARDS-REQUEST` response with exact inventory names. All
  adapters use this same protocol. Already supplied, duplicated or unknown names
  are rejected; an invalid batch does not consume any document. Each provider
  turn has its own configured timeout, with at most four additional fetch turns.
  Only paths and hashes are stored in the audit record, not file contents.
- Supply the complete evidence list on every invocation. The harness hashes it
  and, on resume, identifies added, changed, unchanged, and removed inputs against
  the last recorded round. Removed means omitted from the list, not deleted.
- Unchanged files are reused from retained context. Changed contents are supplied
  again; context loss or unconfirmed delivery requires a replacement context.
  An auditor reports lost mandatory contents with `CONTEXT-LOST` instead of a
  verdict or a duplicate fetch. Hashes are not proof
  that the auditor read a file.
- Inputs must be regular non-symlink files, at most 16 MiB each and 64 MiB total.
  `.env` inputs are prohibited. Changed, missing, or replaced evidence during
  execution rejects the response; hashes detect changes, not prevent writes.
- Per-attempt paths and hashes live only in `audits.yaml` under `history[].evidence`.
  `standards_delivered` records successful transport separately from findings.
  Older records without the current standards contract receive a replacement
  context and full package. Every SDLC gate audit requires an audit record;
  standalone review does not.
- No document cache is created. The small final-response temporary file is
  cleared before each attempt so fallback cannot reuse failed-provider output,
  then
  removed on normal return, including handled errors and timeouts. Abrupt process
  termination can leave that file behind. Provider-owned session storage is separate.
- For audits, stdin is additional bounded context; the gate prompt is loaded
  from `prompts/audits.yaml`.
- `--audit-record <file> --work-item <id> --gate <gate>` makes the harness own
  the retained session mapping and verdict budgets in `audits.yaml`. Use this
  combination for recoverable audits. The record is output, not an `--input`.
- An identical request returns its cached verdict; otherwise context selection is automatic.
- Audit status, findings, revisions, round numbers, and session IDs belong only
  in that `audits.yaml` record; do not copy them into a specification or ticket.
- If supplied artefacts duplicate audit state, the audit must return `FAIL` and
  require the duplicate to be removed before rerunning.
- `--phase` is `definition`, `build`, or `audit`.
- Project and global YAML configuration supply harness, provider where
  supported, model, timeout, and `delivery.audit.max_rounds`; explicit flags
  override them.

## Audit execution deadlines

Configure these duration strings under `delivery.audit` in project
`.sdlc/project.yaml` or global `~/.agents/sdlc.yaml`:

| Key | Default | Meaning |
|---|---|---|
| `timeout` | `10m` | Maximum duration of one provider turn. |
| `total_timeout` | `15m` | Shared budget for the invocation's provider turns, reference fetches, recovery and fallback. |
| `response_start_timeout` | `3m` | Maximum wait for the first native model response activity. |
| `response_idle_timeout` | `2m` | Maximum gap between native response events after activity begins. |
| `max_failures` | `3` | Consecutive unusable attempts before internal lockout. |
| `max_rounds` | `5` | Valid verdict allowance for each managed gate. |

Durations require units and must be at least one second: `90s`, `3m`, and
`1m30s` are accepted; bare `90`, zero, negative and subsecond values are
rejected. A personal `response_start_timeout: 90s` overrides the `3m` product
default. Project YAML overrides global YAML; process environment overrides both.
The additional environment keys are `SDLC_AUDIT_TOTAL_TIMEOUT`,
`SDLC_AUDIT_RESPONSE_START_TIMEOUT` and `SDLC_AUDIT_RESPONSE_IDLE_TIMEOUT`.
These three limits apply only to audit execution, including standalone and
agent-spike reviews. Definition/build retain their existing timeout defaults.
The overall budget begins when provider recovery starts; earlier local readiness,
configuration and evidence preparation are outside it. Cached verdicts invoke no
provider. The earliest applicable deadline wins; recovery never extends the
shared budget, even when the provider or model changes.

Only native model response events count as activity. Session initialization,
process liveness notices, heartbeats and retry notices do not. Hermes reports
content/reasoning callback metadata; Claude exposes partial streaming events;
Codex and Copilot use their native assistant/reasoning events. Private progress
metadata contains no partial Hermes prose. A provider that cannot expose events
may hit the start deadline despite computing internally; configure limits with
that interface in mind.

A turn or response deadline produces an unusable attempt, no verdict and no
verdict allowance consumed. It may recover or select fallback within the remaining
budget and failure bound. A timeout notice reports recovery as pending; it does
not itself require human intervention. Exhausting the shared budget stops all
recovery and, for a managed audit, records internal lockout requiring an authorized
`--reset`. Standalone reviews have no persisted lockout.

## Output and private diagnostics

An audit writes only its validated response or readiness report to stdout;
`--reset` instead writes a reset confirmation and exits. Stderr carries
an input-freeze and delivery-readiness reminder at each provider start, anonymous
start and 30-second liveness notices and, on failure, a diagnostic reference. The
reminder requires completed checks and settled GREEN RT/OT evidence before
delivery-code review, and warns that edits to reviewed inputs during or after
review invalidate the audit for the changed candidate. Mutation detection occurs
after provider execution, without immediate cancellation. Provider names,
session IDs and internal retry counts are not part of the calling-agent contract.

Private files under `.sdlc/audit-diagnostics/` retain provider diagnostics and
unvalidated output, including malformed responses. Files are created with mode
0600 in a 0700 directory with its own Git ignore rule. Each invocation log is
bounded to 16 MiB; excess bytes are discarded without stopping liveness notices.
Logs may contain reviewed material; they are for operator
investigation, never automatic upload or audit verdicts. Operators control retention.
No request prompts or input-file copies are deliberately logged.

## Audit example

```sh
/absolute/path/to/.agents/sdlc/bin/sdlc-harness --gate definition \
  --project . --audit-record specs/W001-change/audits.yaml --work-item W001-change \
  --input specs/W001-change/spec.org --input .sdlc/project.yaml < audit-context.txt
```

Repeat the same command after remediating findings, with the full current evidence.

## Cached results

With `--audit-record`, repeated identical requests return the recorded `PASS`,
`FAIL` or `PROVISIONAL PASS` and findings without launching a provider, changing
the record or consuming a round. Stderr says `AUDIT CACHE HIT` and names the
work item, gate and record, without exposing provider attempt numbers or session IDs. A cached FAIL still requires
remediation; a cached provisional pass retains its conditions.

The versioned SHA-256 key covers work item, gate, every supplied path and content
hash, audit prompt registry, caller context and effective primary/fallback model
configuration. File order and timestamps do not matter. Added, changed or omitted
inputs invalidate it. Installed standards are automatically included in the
manifest; ambient instructions or independently read files are not covered.

Only validated, incident-free responses are reusable. Running attempts, timeouts,
authentication failures and malformed responses are not cache results. Older
entries without the current cache key and required standards-delivery evidence
remain history and require a real audit before reuse.
The lookup precedes round exhaustion and native-session recovery: a recorded
result needs no live provider session. Execution limits alone do not invalidate
it, unless their configuration file is itself a supplied, changed input.

## Recovery

### Verdict budget and internal failures

Each ticket and explicit gate has its own `delivery.audit.max_rounds` allowance,
default five. Only validated PASS, PROVISIONAL PASS and FAIL results count.
Cached results, malformed output and failed provider calls consume none.
Lifetime `external_round` numbers still identify attempts for historical traceability.
Historical labelled verdicts count for their gate; unlabelled implementation
history remains preserved but does not invent test-code or delivery-code verdicts.

After explicit operator authorization, invoke `--reset` with the project, gate,
audit record and work item, without inputs. It records a gate-specific boundary
and exits without a provider call. Repeat the normal audit without the flag.
The selected gate's verdict allowance and failure streak reset, its provider context
is retired, and pre-reset verdicts are not reused from cache. Findings, history and
logs remain intact. Other gates' counters are unchanged.

Reset accepts every recorded status, including stale or active `running` state;
no process-name search or proof of termination is required. It advances the
selected context's generation. Earlier invocations cannot write identity,
failure, or verdict checkpoints after that boundary, or retry by adopting the
new generation. They may finish their provider request but return no verdict.
Repeated reset before another attempt remains a no-op.

Short read-modify-write transactions use a process-released file lock; provider
execution never holds it. Locks live in a self-ignored `.audit-locks/` directory
beside the resolved record. Other records and projects have independent locks.
Test-code and delivery-code retain their existing shared implementation context;
reset retires that shared context but resets only the selected gate's counters.
Legacy records without a generation begin at zero. Late-write protection applies
to invocations of this release; an already-running older binary must exit before
concurrent use because it does not implement the generation check.

Three consecutive unusable provider responses for a gate trigger internal lockout
by default. Configure a positive integer as `delivery.audit.max_failures` in global
or project YAML; project configuration overrides the global value.
A valid verdict clears the streak; other gates do not. The harness owns bounded
retries and envelope correction, not the calling agent. For an unusable response,
it first sends a correction prompt in the retained context when available before
trying fallback; a valid corrected verdict consumes one verdict allowance. It returns only
"Audit unavailable: unable to obtain a valid result. Human intervention required"
with a diagnostic reference. Further audits, including cached requests, are refused until an authorized
`--reset`. Operators inspect the private log before authorizing recovery.
Authentication failure immediately selects the configured fallback, without format
correction. If that fallback fails, or no fallback is configured, the harness locks
out immediately. An authorized reset is required after operator investigation.
Local configuration, readiness, evidence and record failures stop without recovery.

### Session and fallback recovery

With `--audit-record`, the harness records the session's harness/provider/model.
On resume, a changed effective tuple or an explicit missing-session diagnostic
starts a new context automatically. All history, previous IDs and findings remain
in `audits.yaml`; replacements receive the full current evidence and historical
findings. A missing session is not assumed to have expired. Never delete mappings.

Configure an optional fallback beside the primary:

    delivery:
      audit:
        harness: claude
        model: claude-opus-5
        fallback:
          harness: codex
          provider: openai
          model: gpt-5.6-sol

The same `fallback` mapping works under `definition` and `build`. Project YAML
replaces the entire global fallback tuple; `fallback: {}` disables inheritance.
Omission inherits. There are no fallback CLI/environment overrides. Harness and
model are required; provider is required by Hermes and ignored by other adapters.
For audits, an attempted provider run ending without a usable verdict selects
the configured fallback once: usage limits, authentication, launch failures,
timeouts, empty/malformed responses or a wrong-gate response qualify.
PASS, FAIL and PROVISIONAL PASS are usable verdicts and never trigger fallback.
PENDING/running is not a final verdict: wait for completion or the harness timeout.
A retained fallback audit resumes it while primary and fallback settings match,
unless a recorded primary cooldown has expired.
Invalid configuration, evidence-integrity failures, record errors and exhausted
rounds stop locally; fallback does not bypass these checks or widen permissions.
Non-audit definition/build invocations retain authentication-only fallback.

Missing-session recovery recognizes Claude's exact `No conversation found with
session ID: <requested ID>` diagnostic. Unknown provider messages fail closed.
Authentication recovery recognizes explicit login, expired-OAuth and invalid-key
diagnostics. Other failed audit invocations do not need a diagnostic allowlist.
At most one missing-session restart and one configured fallback occur per
command. Each launch gets the configured timeout and is subject to the internal
failure bound; replacements do not reset the verdict allowance.

## Primary audit cooldown

After an unusable primary audit attempt, a configured distinct fallback enables
a project-local cooldown. The deadline is the local failure time plus
`delivery.audit.fallback.cool_off_period`, defaulting to `1h`. It does not
parse provider messages or infer reset dates. The policy applies to every
supported primary harness, including launch/authentication failures, timeouts
and unusable responses. PASS, FAIL and PROVISIONAL PASS do not trigger it.
Local configuration, evidence-integrity and record failures remain blockers.

```yaml
delivery:
  audit:
    harness: claude
    provider: claude
    model: claude-opus-5
    timeout: 15m
    max_rounds: 5
    max_failures: 3
    fallback:
      cool_off_period: 1h
      harness: hermes
      provider: nous
      model: z-ai/glm-5.3
```

The period must be a positive Go duration, such as `30m`, `1h` or `2h`;
zero, negative and malformed values are rejected. A project may override only
`cool_off_period` while retaining an inherited fallback. A replacement fallback
tuple defaults to one hour unless it declares its own period; `fallback: {}`
still disables fallback. Non-audit phases retain their existing recovery policy.

The harness writes versioned JSON state under the selected project's
`.sdlc/cooldowns/`, keyed by the effective primary harness/provider/model.
Other projects and different primary tuples are unaffected. The directory
contains its own `.gitignore`, so timestamps and locks remain runtime metadata.
Only the route and UTC deadline are stored, never credentials, prompts or verdicts.
Atomic updates and a per-route lock preserve the later concurrent deadline.

Before expiry, uncached audits skip the primary and use the configured fallback.
The private log records the cooldown and deadline. Skips and failed calls consume
no verdict allowance. Expiry makes the primary eligible
on the next requested audit, not guaranteed healthy. There is no background
retry or sleep. Replacements retain audit history and budgets; cached verdicts
remain reusable. A failed fallback does not extend the primary's deadline.

Without a configured distinct fallback, new failures do not create cooldowns.
If fallback is removed during an existing cooldown, invocation stops until
expiry rather than calling the cooled primary. Invalid or unreadable state stops
explicitly; missing state means no cooldown. Expired records remain small and
are replaced on later failures. Changing the period affects future failures,
not an already recorded deadline.

This supersedes v3.4.4's global Claude-message-based policy. The old global store
and `SDLC_HARNESS_STATE_DIR` override are no longer used. Existing global files
are neither migrated nor deleted. A retained fallback with no project-local
cooldown returns to the primary on its next uncached request. Installation remains operator-run and does
not initialize cooldown state.

## Timeout behaviour

Timeouts preserve the attempt manifest and native identity when available, but
consume no verdict allowance. The harness performs bounded continuation and
fallback internally. Calling agents must not build a second retry loop around an
infrastructure failure. Known FAIL findings remain unresolved until reassessed.

## Document history

- Version 10 (2026-10-11): Add shared, response-start and response-idle deadlines,
  native Hermes progress callbacks and terminal-only intervention reporting.

Version 9 supplies mandatory standards deterministically for three audit profiles,
isolates provider tools, rejects caller routing overrides and reuses unchanged
contents through a context-bound delivery ledger. Caller context is limited to
2 MiB and each complete provider prompt to 8 MiB; oversized input is rejected,
never silently trimmed.
Version 8 adds `--agent-spike` standalone review for hands-off spikes in SDLC
projects.
Version 7 adds bounded on-demand standards retrieval for tools-disabled auditors.
Version 6 exposes the full installed Markdown inventory while distinguishing
mandatory core standards from additional applicable documents.
Version 5 makes SDLC gate standards harness-owned evidence and limits caller
inputs to the invocation project.
Version 4 binds audits to the invocation project and supplies a bounded file
inventory while excluding runtime and scratch directories. Version 3 adds
standalone standards review and temporary project/evidence
refusal. Version 2 replaces agent-managed sessions and attempt-based budgets
with automatic invocation, gate-local verdict limits and private bounded
failure diagnostics.
Version 1 remains available in Git history.
