SDLC harness
============

Internal SDLC helper for running one bounded provider context against
original evidence files with SHA-256 integrity checks. It is installed at
`~/.agents/sdlc/bin/sdlc-harness`. The operator-facing `sdlc-audit` command is a
symlink to this same executable; it accepts the same options and subcommands.
SDLC skills continue to use the resolved absolute harness path.

Usage
=====

    sdlc-harness start|resume [options] < audit-context.txt
    sdlc-harness goal-config [options]

Goal configuration
==================

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
  output and exit nonzero. Existing start/resume audit behaviour is unchanged.

Start and resume
================

`start` creates a fresh external context. For an audit, pass `--gate definition`
or `--gate implementation`; the installed YAML prompt registry supplies the
criteria. Do not pass `--session`; the runner creates the identity and prints
`SESSION_ID: <id>` on stderr. Save that exact ID.

`resume` continues the same external context. Pass the saved ID as
`--session <id>`. An agent task ID, path, or newly invented value is invalid.

Evidence and instruction
========================

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
  pass. Without `--audit-record`, no cross-invocation hash comparison is possible.
- No document cache is created. The small final-response temporary file is
  cleared before each attempt so fallback cannot reuse failed-provider output,
  then
  removed on normal return, including handled errors and timeouts. Abrupt process
  termination can leave that file behind. Provider-owned session storage is separate.
- For audits, stdin is additional bounded context; the gate prompt is loaded
  from `prompts/audits.yaml`.
- `--audit-record <file> --work-item <id> --gate <gate>` makes the harness own
  the retained session mapping and attempt budget in `audits.yaml`. Use this
  combination for recoverable audits. The record is output, not an `--input`.
- With a record, either operation first checks for an identical cached result.
  On a cache miss, `start` refuses an existing mapping and `resume` loads its
  recorded session ID. A supplied resume ID must match it.
- Audit status, findings, revisions, round numbers, and session IDs belong only
  in that `audits.yaml` record; do not copy them into a specification or ticket.
- If supplied artefacts duplicate audit state, the audit must return `FAIL` and
  require the duplicate to be removed before rerunning.
- `--phase` is `definition`, `build`, or `audit`.
- Project and global YAML configuration supply harness, provider where
  supported, model, timeout, and `delivery.audit.max_rounds`; explicit flags
  override them.

Output
======

- **stderr:** provider progress, diagnostics, and `SESSION_ID`.
- Provider stdout events are decoded as they arrive. Progress shows event types,
  not prompts, reasoning or tool payloads; at most 200 event notices are shown.
  Provider stderr remains live. Claude uses `stream-json --verbose` for both
  start and resume; the other adapters retain their existing output modes.
- Provider error fields are reported even on unsuccessful exits or timeout,
  including a final record without a newline. Error text is bounded to 4 KiB
  before terminal escaping. Plain-text startup failures retain a 4 KiB tail;
  unknown structured payloads are not dumped.
- Captured provider stdout is limited to 16 MiB. Exceeding it fails explicitly
  with `output-limit`; it never yields a silently truncated verdict.
- The harness also emits a liveness line at start and every 30 seconds while
  waiting; this is not provider progress or a verdict.
- **stdout:** the provider's final response, or that exact response from cache.
- A progress event or final-result record is not success until the provider exits
  successfully and evidence and response validation pass. Claude error envelopes
  are rejected even if the provider exits zero. The runner cannot flush output
  that the provider has not emitted; its existing deadline remains in force.

Start example
=============

    /Users/tigger/.agents/sdlc/bin/sdlc-harness start \
      --phase audit --gate definition \
      --project . \
      --audit-record specs/W001-change/audits.yaml \
      --work-item W001-change \
      --input specs/W001-change/spec.org \
      --input .sdlc/project.yaml \
      < audit-context.txt

Resume example
==============

    /Users/tigger/.agents/sdlc/bin/sdlc-harness resume \
      --phase audit --gate definition \
      --project . \
      --audit-record specs/W001-change/audits.yaml \
      --work-item W001-change \
      --input specs/W001-change/spec.org \
      --input .sdlc/project.yaml \
      < audit-context.txt

Cached results
==============

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

Recovery
========

Reset the attempt budget
------------------------

After explicit operator authorization, use `resume --reset-session` to give
one work item/gate a fresh `max_rounds` allowance. It **only resets the budget
and exits**: no prompt, evidence files, provider call or model configuration is
required. It retains the native session, findings, cached results and lifetime
history. It does not create a new provider session or approve the work.

    /absolute/path/to/.agents/sdlc/bin/sdlc-harness resume \
      --reset-session --phase audit --gate definition --project . \
      --audit-record specs/W001-change/audits.yaml --work-item W001-change

Then repeat the normal Resume example **without `--reset-session`**, supplying
the full evidence and context. Never leave the reset flag in a retry loop.
Changing a specification or recovering a provider does not reset the budget.
The reset records a timestamped `budget_resets` boundary; `external_round`
continues to number lifetime attempts. Repeating a reset before another attempt
does nothing. Missing records and running/interrupted attempts are rejected.
Reset does not repair a missing native identity. `--session` and `--input` are
not accepted with the reset flag. Other work items and gates are unchanged.

Session and fallback recovery
-----------------------------

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
A retained fallback audit resumes it while primary and fallback settings match.
Invalid configuration, evidence-integrity failures, record errors and exhausted
rounds stop locally; fallback does not bypass these checks or widen permissions.
Non-audit definition/build invocations retain authentication-only fallback.

Missing-session recovery recognizes Claude's exact `No conversation found with
session ID: <requested ID>` diagnostic. Unknown provider messages fail closed.
Authentication recovery recognizes explicit login, expired-OAuth and invalid-key
diagnostics. Other failed audit invocations do not need a diagnostic allowlist.
At most one missing-session restart and one configured fallback occur per
command. Each launch consumes a round and gets the configured timeout; replacements
do not reset `max_rounds`. Without an audit record, one fallback is still
available, but missing-session replacement and cross-call tracking are not.

Timeout behaviour
-----------------

A timeout is an incident, not PASS or FAIL. The harness checkpoints each
attempt's evidence and records the native session ID as soon as it is observed.
Timeouts count towards `delivery.audit.max_rounds`. Prior verdicts remain in
history; a timed-out attempt has an `incident`, not a fabricated verdict.

The harness first tries an unused configured fallback when budget remains.
If it returns a timeout instead, in HANDS-OFF mode follow that diagnostic and
repeat the invocation with
`resume`, the same configuration, inputs and context. With `--audit-record`, no
manual session argument is needed: the Resume example above loads it. The
harness adds continuation instructions from YAML and compares evidence against
the interrupted attempt so unchanged files can be reused from retained context.
Known FAIL findings must still be remediated; timeout alone needs no artificial edit.

Stop if the native ID is unavailable or the attempt limit is exhausted. Never
invent an ID, switch harness/model, increase bounds, or start a replacement to
evade that stop. The harness reports whether native identity is recorded and
how many attempts remain. It does not spawn a retry itself; the calling agent
executes the next bounded invocation. Without a record, automatic recovery and
attempt accounting are unavailable.
