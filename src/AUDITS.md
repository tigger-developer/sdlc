# Audit and Gate Standards

Audits provide independent findings; they do not replace operator approval,
tests, or implementation evidence. The **audit harness** owns the prompts,
provider invocation, timeout, round limit, and retained session. No individual
audit skill is required or deployed.

For a client-project agent, the harness is an **external tool**, not part of
the work. Use its help, supply the required evidence, and follow its documented
recovery instructions. Do not inspect or troubleshoot its internals, provider
authentication or model behaviour, or ask the operator to choose replacements.
Model selection, fallback and session recovery belong to the harness and its
configuration. If documented recovery cannot proceed, report the exact command
and diagnostic without speculative causes; continue other authorized work.
Harness investigation requires the operator to explicitly place it in scope.

An identical request may return a **cached result** from `audits.yaml`: the same
PASS, FAIL or PROVISIONAL PASS and findings, without a model call or another
round. The harness checks all supplied file paths/hashes, work item, gate,
prompts, caller context and primary/fallback model configuration. Include the
relevant SDLC standards as evidence; independently read files are not covered.
Cache hits retain all verdict conditions. Incidents are not verdicts, and old
records without a request key are not cache entries. Changing or omitting an
input invalidates the key; do not alter inputs merely to evade a cached FAIL.

## Composite gates

Apply `~/.agents/sdlc/ARCHITECTURE.md` at all three gates and include it in the
supplied evidence, alongside the relevant project architecture. Definition
review resolves significant design decisions and their verification; implementation
review checks adherence without substituting the auditor's preferred design.

The **definition gate** reviews the specification, solution design, and RT/UT/OT
test definitions together. The **test-code gate** reviews the written tests and
execution-enabling scaffolding before implementation of the target behaviour.
The **delivery-code gate** reviews the production delta, affected tests, execution
evidence and required product documentation. These two gates share one retained
implementation context. Normal delivery uses the signed-off definition; paired and
emergency delivery use the recorded operator-authorized scope. Neither implementation
review reopens a signed-off definition or demands advance definition approval
from a valid paired or emergency route.

Test definitions belong to the definition context. Normal implementation has two
reviews in one retained context: `--gate test-code` after the written
test package has executed, then `--gate delivery-code` after the solution,
tests and required OTs are GREEN. Review packages, not each file or edit. The test
review must reach PASS or effective PROVISIONAL PASS before normal production work.
If the approved spec has no RTs, skip test-code under TESTING.md; do not reclassify
tests to evade it. Paired/emergency routes retain their implement-first sequence
and may invoke delivery-code directly. Both implementation gates use the same
work-item session and round allowance; the second resumes the first's context.

Before invoking either stage, use `~/.agents/sdlc/bin/sdlc-validate --help` and run
its selected readiness check. `ORG-SCHEMA.md` defines the record contract. The
harness repeats this deterministic check before config/provider startup, cache
lookup or audit-record mutation. It locates `spec.org` and `validation.org` beside
the required `--audit-record` and includes them in hashed evidence. Schema errors
or missing results return YAML and nonzero without a model call or round consumed.
The explicit gate selects its readiness check automatically. There is no default
audit gate. `--readiness-check` belongs to the standalone validator, not the audit
command. Definition review has no deterministic execution-readiness check.

For history compatibility, the harness retains `gate: implementation` as the
shared session key and `readiness_check` as the selected test-code/delivery-code
gate on the entry and each attempt. Returned GATE names use the explicit audit
gate. These are storage details, not additional CLI selectors. A test-code
PASS approves tests only and does not mark delivery passed. Final delivery requires
current delivery-code PASS or effective PROVISIONAL PASS; stage-specific prompts
separate cache keys. Old unlabelled rounds remain history, not current-stage evidence.
Execution states in validation.org are test evidence, not duplicated audit verdicts.

For every gate:

1. Invoke `~/.agents/sdlc/bin/sdlc-harness` with the gate and exact evidence.
2. On `FAIL`, address **every** finding before resuming: remediate all valid
   findings together and supply an evidence-backed challenge for each disputed one.
3. Never split known findings across remediation rounds or manually create a
   replacement session. Unchanged artefacts may be
   resubmitted with a substantive challenge, not merely to seek another verdict.
4. Stop after the configured maximum rounds, or for a decision that cannot be
   made safely from the authorities.

An auditor can be wrong. Challenge incorrect findings or scope overreach in the
resumed audit's caller context, citing the authorized scope, applicable standard
or concrete evidence. Do not implement unauthorized expansion to obtain a PASS.
If unsure, or the dispute remains unresolved, raise it to the operator. A challenge
does not itself dismiss a finding or change the verdict; only the harness records
the auditor's reassessment in `audits.yaml`.

The harness stores separate definition and implementation session mappings per
work item in `audits.yaml`; test-code and delivery-code share the latter.
The first invocation starts a session; later invocations resume its recorded
`session_id`. The round limit is `delivery.audit.max_rounds` (default five) and
applies to the current explicitly bounded audit cycle. Every provider invocation,
including a timeout, consumes one round. To start a new allowance, the operator
must authorize `resume --reset-session`; this resets only the selected work
item/gate's budget and exits without an audit. Run it once, then resume normally
without the flag. Never reset autonomously to evade the limit. The harness retains
native context, findings, cached results and lifetime history, recording a reset
boundary in `audits.yaml`. Spec changes and provider recovery never reset the
budget. Earlier unrecorded timeouts cannot be reconstructed automatically.

Session reuse is an optimization, not a delivery prerequisite. The harness records
the native ID, owning harness/provider/model and effective primary configuration.
Changed configuration or an explicit provider report that the session is missing
causes a fresh context automatically. It preserves old session provenance and all
findings/history, supplies the full evidence and historical findings, and keeps
the existing round budget. It never claims that an unavailable session expired.

An optional `delivery.<phase>.fallback` supplies one alternate harness/provider/model
tuple. For audits, use it once when an attempted provider run ends without a usable
verdict, including usage limits, authentication/launch errors, timeout, or an empty,
malformed or wrong-gate response. PASS, FAIL and PROVISIONAL PASS never trigger
fallback. A PENDING/running attempt must finish or time out before recovery.
A retained fallback audit continues on
that tuple while the configured primary and fallback remain unchanged. Each CLI
invocation permits at most one missing-session restart and one configured
fallback, subject to the remaining audit rounds. Every launched attempt consumes a
round and has the configured timeout. Invalid configuration, evidence-integrity
failures, record errors and exhausted rounds still stop locally. No permission
is widened. Non-audit definition/build retain authentication-only fallback.
Calling agents do not repair
or delete mappings themselves.

The harness tries an unused configured fallback before returning a timeout.
On a returned timeout, follow the recovery diagnostic. In HANDS-OFF mode, resume
automatically while a native session is recorded and the limit permits it.
Resubmitting unchanged evidence after timeout is allowed, but does not waive
earlier FAIL findings. The resumed prompt continues unfinished review and reuses
unchanged evidence still understood. Do not change harness, model, timeout, or
limits to evade a stop. Missing native identity or an exhausted limit requires
operator recovery, not a replacement context.

Audit findings, status, revisions, round numbers, and session IDs MUST be
recorded only in `audits.yaml`. Specifications and tickets may link to that
record but must not duplicate its audit state. Existing duplicated text is
historical input for migration, not a second authority.

If supplied artefacts duplicate audit findings or state, the auditor MUST return
`FAIL` and require the duplicate to be removed before rerun.

The definition audit MUST return `FAIL` if `spec.org` contains any semantically
identified annotation record or block, even resolved or empty. Apply the annotation
tool's documented format, not keyword matching against ordinary prose, examples
or Org comments. Check processed-feedback evidence in `feedback.yaml` when present;
missing dispositions or conflicts between recorded outcomes and the revised spec
also fail. Annotation removal without recorded resolution is not remediation.
The feedback log records human review, not a duplicate audit verdict. This rule
also applies to retrospective definition reviews.

The external auditor returns its envelope and findings only; it does not edit
files. The harness persists that response in `audits.yaml`.

While waiting, the harness emits a start message and 30-second liveness
heartbeats to stderr. A heartbeat is not provider progress, a verdict, or
evidence that the audit has completed.

Codex and Claude receive original absolute evidence paths, not document copies.
Hermes has no filesystem tools: the harness supplies verified UTF-8 contents in
stdin on every invocation. It stores only the paths/hashes, never document copies.
Supply
the complete input list on each invocation; the per-attempt manifest in
`audits.yaml` identifies added, changed, unchanged, and omitted inputs on resume.
Reuse unchanged evidence only while its meaning remains understood; reread it
after context loss or when an affected dependency requires it. Omission never
resolves a finding. Attempts and native IDs are checkpointed before a verdict,
so a timeout retains its evidence baseline. The harness verifies supplied files before and after the
provider runs and rejects changed evidence. Hashes detect mutation; read-only
provider controls remain necessary. Additional files read outside the supplied
list are not covered by the manifest checks.

## Verdict contract

The harness response contains exactly one envelope:

```text
GATE: definition | test-code | delivery-code
REVISION: <audited revision or SHA-256>
VERDICT: PASS | PROVISIONAL PASS | FAIL
```

`PROVISIONAL PASS` is effective only after its exact mechanical conditions are
applied. Timeout, silence, malformed output, or an unavailable provider is not
a pass. Record findings and remediation in `audits.yaml` without deleting
superseded evidence.

## Human boundaries

Only the operator may approve the definition, change signed-off behaviour or
architecture, and close delivered work. A hands-off workflow may continue with
reversible assumptions, but must stop for an unsafe or irreversible decision,
an unremediable mandatory test failure, or the configured round limit.

`BYPASS-GATE-7` skips definition admission for an emergency change. It still
requires TDD where practicable, the delivery-code gate, and a wrap-up that
backfills the specification, design, validation, and documentation evidence.

## Brownfield evidence

A brownfield change must consult the relevant requirements, design, project
authorities, historical records, regression tests, and affected implementation.
The delta need not copy the baseline; material conflicts and omitted authorities
must be reported.

## Variant workflows

Paired work creates a skeleton ticket before code and audits each coherent code
increment promptly, not every live edit. Emergency work fixes first, records the
durable ticket, then audits implementation. Both consolidate and retrospectively
review the definition after the delivery-code gate has an effective PASS.
For all routes, affected product documentation must be current at the final
delivery-code verdict. Include `~/.agents/sdlc/DOCUMENTATION.md` in definition and
delivery-code audit evidence. Review the changed documentation against its voice,
document-purpose and presentation rules as well as technical accuracy; cite concrete
violations, not editorial preferences. The test-code gate does not demand
those updates. A paired/emergency retrospective specification may follow code
review; this does not defer updates to affected product documentation.

Run `~/.agents/sdlc/bin/sdlc-harness --help` before first use. In its stdin audit
context, supply `Workflow: paired` or `Workflow: emergency` for an operator-invoked
variant. For the subsequent `--gate definition` invocation, also supply
`Review: retrospective`. These are prompt instructions, not new CLI flags;
normal delivery omits them. Include the recorded operator scope as evidence.
The YAML prompts select the review contract from these explicit markers, never
from the mere presence of code. Keep the same work ID and audit record:
implementation and definition have separate retained sessions; each gate resumes
its own session on subsequent rounds.

Retrospective review checks fidelity to the authorized outcome, requirement and
test-definition traceability, accurate design and limitations, architecture fit,
and material standards or security defects. It must not fail solely because the
definition was written after implementation, demand an imagined advance plan,
or turn optional improvements into requirements. Code is evidence, not approval
of unrequested behaviour.

Correct documentary defects within the existing audit allowance. Findings that
require changed behaviour or design must be raised to the operator and, if
authorized, handled through normal delivery; do not redesign during wrap-up.
Report immediate safety risks promptly. Unresolved blocking findings remain
blocking; operator review does not manufacture an audit PASS. Audit outcomes
remain solely in `audits.yaml`, and operator closure remains separate.

# Canary

Suffix the canary string with "AUDIT "
