---
title: SDLC Audit Harness Guide
version: 2
last-updated: 2026-09-23
---

# SDLC audit harness guide

Internal SDLC helper for running one bounded provider context against
original evidence files with SHA-256 integrity checks. It is installed at
`~/.agents/sdlc/bin/sdlc-harness`. The operator-facing `sdlc-audit` command is a
symlink to this same executable; it accepts the same options and subcommands.
SDLC skills continue to use the resolved absolute harness path.

## Usage

    sdlc-harness --gate GATE --audit-record FILE --work-item ID [options] < audit-context.txt
    sdlc-harness goal-config [options]

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

- `--project DIR`: project root; defaults to the current directory.
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

- Repeat `--input <file>` for every exact evidence file.
- Paths are resolved against `--project`. Codex and Claude read original absolute
  paths. The provider is launched from the project; no document copies are retained.
- Claude start/resume add invocation-local Read permissions for the exact supplied
  files and their resolved aliases, never their containing directories. Literal
  pattern characters are escaped. Read-only tools, plan mode and existing deny
  rules remain in force; these grants do not bypass native policy. Unsupported
  non-POSIX paths or control characters fail before invocation.
- Hermes uses `chat --quiet --query-file -` with its native resume option and
  reads native `session_id:` stderr metadata. Top-level `-z` is not used because
  it bypasses session options; model-written session identifiers are not trusted.
- Hermes runs with tools disabled. Its supplied UTF-8 evidence is hash-verified
  and sent through stdin on every invocation, including resume, so it needs no
  filesystem tools. This adapter-specific transport resends the full text;
  only paths and hashes are stored in the audit record, not file contents.
- Supply the complete evidence list on every invocation. The harness hashes it
  and, on resume, identifies added, changed, unchanged, and removed inputs against
  the last recorded round. Removed means omitted from the list, not deleted.
- Unchanged files may be reused from retained context, but must be reread when
  context is lost or changes affect their interpretation. Hashes are not proof
  that the auditor read a file.
- Inputs must be regular non-symlink files, at most 16 MiB each and 64 MiB total.
  `.env` inputs are prohibited. Changed, missing, or replaced evidence during
  execution rejects the response; hashes detect changes, not prevent writes.
- Per-attempt paths and hashes live only in `audits.yaml` under `history[].evidence`.
  Older records without hashes retain their session and receive a full evidence
  pass. Every audit requires an audit record.
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

## Output and private diagnostics

Stdout contains only a validated audit response or readiness report. Stderr carries
plain liveness notices and, on failure, a diagnostic reference. Provider names,
session IDs and internal retry counts are not part of the calling-agent contract.

Private files under `.sdlc/audit-diagnostics/` retain provider diagnostics and
unvalidated output, including malformed responses. Files are created with mode
0600 in a 0700 directory with its own Git ignore rule. Each invocation log is
bounded to 16 MiB. Logs may contain reviewed material; they are for operator
investigation, never automatic upload or audit verdicts. Operators control retention.
No request prompts or input-file copies are deliberately logged.

## Audit example

```sh
/absolute/path/to/.agents/sdlc/bin/sdlc-harness --gate definition \\
  --project . --audit-record specs/W001-change/audits.yaml --work-item W001-change \\
  --input specs/W001-change/spec.org --input .sdlc/project.yaml < audit-context.txt
```

Repeat the same command after remediating findings, with the full current evidence.

## Cached results

With `--audit-record`, repeated identical requests return the recorded `PASS`,
`FAIL` or `PROVISIONAL PASS` and findings without launching a provider, changing
the record or consuming a round. Stderr says `AUDIT CACHE HIT` and names the
source round. No new session ID is emitted. A cached FAIL still requires
remediation; a cached provisional pass retains its conditions.

The versioned SHA-256 key covers work item, gate, every supplied path and content
hash, audit prompt registry, caller context and effective primary/fallback model
configuration. File order and timestamps do not matter. Added, changed or omitted
inputs invalidate it. Include all relevant SDLC standards as inputs; ambient
instructions or independently read files are not covered by this cache.

Only validated, incident-free responses are reusable. Running attempts, timeouts,
authentication failures and malformed responses are not cache results. Older
entries without keys remain history and require a real audit before reuse.
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

Three consecutive unusable provider responses for a gate trigger internal lockout
by default. Configure a positive integer as `delivery.audit.max_failures` in global
or project YAML; project configuration overrides the global value.
A valid verdict clears the streak; other gates do not. The harness owns bounded
retries and envelope correction, not the calling agent. For an unusable response,
it first sends a correction prompt in the retained context when available before
trying fallback; a valid corrected verdict consumes one verdict allowance. It returns only
"Audit unavailable: unable to obtain a valid result. Human intervention required"
with a diagnostic reference. Further uncached audits are refused until an authorized
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

Version 2 replaces agent-managed sessions and attempt-based budgets with automatic
invocation, gate-local verdict limits and private bounded failure diagnostics.
Version 1 remains available in Git history.
