# Testing Standards

Tests provide evidence that specified behaviour holds. They do not replace the
specification.

## Test from the user's boundary

Prefer the same entry point and artefact a real user or consuming system sees.

- Test a CLI through its executable interface.
- Test an HTTP service through requests and responses.
- Test a generated site through built output and, where necessary, a browser.
- Test a library through its public API.
- Reserve direct internal tests for isolated logic whose contract is genuinely
  internal.

Do not claim behavioural coverage from grepping source, documentation, prompts,
templates, or configuration text. Text search may support a one-off review, but
it is not a persistent behavioural regression test. Documentation changes that
need sign-off receive a one-time human or unit-level review rather than a
regression test coupled to prose wording.

## Derive tests from the specification

Map each automated or human check to a described requirement. Give any cited
requirement or test identifier an adjacent descriptor.

Never reuse a test identifier within its established namespace. A removed,
abandoned, retired, or superseded test permanently reserves its identifier. For
sequential identifiers, consult active records, archives, and version history
and allocate above the highest value ever assigned. Preserve superseded entries
in `validation.org` rather than recycling them.

### Regression packs as brownfield evidence

For brownfield specification and design work, inspect the maintained regression
test pack for the affected behaviour. Follow its requirement and ticket
references to their source records, and use its assertions to identify
observable behaviour that is still actively protected and compatibility
constraints that a change must address.

A passing or maintained test is strong implementation evidence, but it is not
exhaustive and does not approve a requirement. Code may contain untested
behaviour, and a test may preserve a stale assumption or defect. Reconcile
disagreements among requirements, design, tests, historical work records, and
code explicitly.

Cover each independent condition in a compound requirement. Include relevant:

- happy, alternate, and failure paths;
- empty, missing, malformed, minimum, and maximum values;
- permissions and trust boundaries;
- partial completion and interruption;
- repeated and concurrent operation;
- compatibility and migration; and
- accessibility, layout, readability, or other human judgement.

## Test-driven implementation

Before implementation, select every test type needed to establish the specified
behaviour. A change may require automated regression tests, one-off tests, user
tests, or a combination of them.

Only automated tests follow test-driven development. When a meaningful
automated regression test is justified, write or amend it before implementing
the target production behaviour, observe it fail for the intended reason, implement the smallest
coherent change, observe it pass, then refactor without losing evidence. When no
automated regression test is justified, record the specific reason; urgency,
difficulty, or inconvenience is insufficient. The definition-gate audit harness
confirms test design and traceability; it does not replace a required failing
automated test execution.

At the definition gate, audit only the proposed RT, UT, and OT definitions. Do
not demand test code, execution results, or RED/GREEN evidence before
implementation. At the implementation gates, the audit harness reviews the
actual test changes and current evidence against those signed-off definitions
(or the recorded operator-authorized scope in paired and emergency delivery).
Normal delivery first reviews the complete written RT package at the test-code
gate, after observing RED (or justified initially GREEN preservation evidence),
before implementing the target behaviour. After implementation and RT/OT GREEN, the
delivery-code gate resumes the same implementation-auditor context, reviewing
production code and changed tests. Both stages use the configured shared round
allowance. The author must check that RED fails for the intended reason.
Paired and emergency routes may go directly to delivery-code after their coherent
change; they do not invent a retrospective pre-build test-code gate.
Retrospective test-definition review uses the
separate definition context, not the implementation auditor's session.
Planned tests are not implementation evidence, and implemented tests do not
retroactively repair an inadequate definition.

Before test-code review, the agent may add only the minimum declarations,
interfaces or placeholder bodies needed to compile and execute the specified RTs.
This scaffolding must not implement the target behaviour and must be included in
test-code audit evidence. Compilation and setup failures do not qualify as RED.

If the approved specification contains no RTs, skip the test-code gate. The
definition audit must establish why OT/UT evidence is appropriate. Do not remove
or reclassify approved RTs merely to bypass the gate. Implement, execute required
OTs and invoke delivery-code directly; UTs may await the operator. The validator's
inventory supplies the RT count; it does not authorize changing the test strategy.

Define the expected evidence for one-off and user tests before implementation
where practical. They do not follow TDD and do not require a pre-change failure.
Execute final OTs before the delivery-code audit; every RT and OT must be GREEN.
UTs may remain AMBER until operator participation. Earlier diagnostic executions
count only if their evidence remains current for the candidate. Paired development
uses the live user-validation contract below.

## Choose durable evidence

Use three distinct test categories. A change may use more than one:

- **Automated regression tests:** retained because the behaviour could regress.
- **One-off tests:** temporary evidence for a migration, incident,
  environment, or documentation review that has no lasting regression value.
- **User tests:** human verification of outcomes requiring visual, editorial,
  ergonomic, operational, or other human judgement.

Do not create permanent test code solely to satisfy a test-count rule. Record
one-off or user-test evidence durably, then remove temporary files from the
repository.

## Test execution records and readiness

Every delivery has `validation.org`, with one current heading per RT, OT and UT
defined in `spec.org`. Follow `~/.agents/sdlc/ORG-SCHEMA.md` and the canonical
templates: both documents tag tests `:testdef:`, and only validation headings carry
execution states. Declare `#+TODO: AMBER RED | GREEN`; an absent state is schema-valid
but audit-ineligible. Preserve superseded runs below the same test node.

The deterministic `~/.agents/sdlc/bin/sdlc-validate` reports schema errors, inventory
and readiness as YAML. Read its `--help` and run it before the corresponding audit.
The harness repeats the same checks before spending a round. Schema failures and
ineligibility never invoke a model. Test-code requires RTs RED/GREEN and a recorded
state for every OT/UT; delivery-code requires RT/OT GREEN, permitting outstanding UTs only.

Do not infer execution from implementation claims. Record tested revision,
environment, dated observations and evidence, plus the reviewer for UTs. A broken
fixture is AMBER; RT RED requires an executed intended assertion failure before
implementing the target behaviour. Initially GREEN preservation tests need no artificial RED.
Readiness checks verify recorded structure, not evidence truth or adequate coverage.
Closure requires current GREEN for every required test, including operator UTs.

The explicit legacy ticket-migration skill has one historical-record
exception. It may infer delivery when either its near-complete rule has at least
one recorded pass and leaves at most two unrecorded unit or operator tests, or a
ticket without recorded passing evidence retains at least one RT in the
maintained passing harness. It must mark every AC supported only by inference,
identify the applicable heuristic in a
footnote, and distinguish an assumed migration pass from contemporaneous test
evidence. This creates no precedent for active delivery.

An audit verdict remains historical evidence for the revision and scope it
assessed. If later remediation changes relevant code, the earlier implementation
gate PASS is no longer current for completion. Rerun the affected automated tests
and affected one-off tests before the delivery-code audit; repeat materially affected
user tests when the operator is available. Unaffected results remain current.

## Paired user validation

Explicit operator validation during paired development is first-class user-test
evidence. Do not replace it with an automated imitation of the visual or
subjective judgement that just occurred.

Maintain a provisional ledger during the session. For each validation, retain:

- a plain-language description of the observed behaviour;
- the revision or working-tree state reviewed;
- material viewing conditions such as viewport, browser, or device; and
- whether a later iteration superseded it.

At closure, present the current ledger once and ask whether it may be recorded
as the user tests for the change. On approval, write it to the active change's
`validation.org`. A later change invalidates only entries whose observed behaviour it
materially affects.

Add automated regression coverage only when the behaviour is objective and
stable, regression risk justifies retention, the test uses an appropriate user
boundary, and it adds evidence beyond the paired validation. Report useful gaps
at closure; do not create brittle tests merely to eliminate the report.

## Test architecture

Adding a new framework, browser harness, mock layer, container topology, CI
service, or test-only production interface is an architectural decision. It
must appear in the plan and be justified against simpler existing facilities.
Do not introduce it merely to make one test convenient.

Prefer the project's established framework and helpers. Keep tests deterministic,
parallel-safe, and independent of execution order.

Exercise unit, integration, and end-to-end boundaries in proportion to the
change. Production behaviour needs evidence at every layer that carries a
distinct risk; record why a normally relevant layer is omitted.

## Structure and naming

- Name tests as behaviour: `returns_error_when_config_is_missing`.
- Arrange setup, action, and assertions clearly.
- Keep one behavioural reason for failure per test while allowing several
  assertions about the same result.
- Put shared setup in explicit helpers or fixtures, not hidden global state.
- Make failure messages identify the expected behaviour and observed value.

## Test data and temporary state

- Use realistic but synthetic data. Never copy production secrets or personal
  data into tests.
- Allocate temporary directories through the platform or test framework.
- Bound generated data and clean up state created by the test.
- Do not depend on a developer's home directory, installed private tools,
  clock, locale, timezone, network, or service account unless that dependency
  is the explicit subject of the test.
- Seed randomness when diagnosing reproducibility; retain the failing seed.

## Doubles and boundaries

Prefer real local dependencies when they are deterministic and inexpensive.
Use fakes or mocks only at an actual external boundary or to force a failure
that cannot be produced safely. A double must model the contract that matters;
do not mock the function under test or merely assert that internal calls
occurred.

### External services and human evidence

Test deterministic behaviour owned by the project with automated regression
tests where that evidence is meaningful. A fake or stub for a third-party
service must implement only the relevant part of a published, versioned, or
otherwise verified contract. It proves the project's handling of that contract;
it does not prove the live service behaves as simulated.

Do not invent simulations of undocumented third-party behaviour. Where the
only credible evidence requires the live service or human judgement, specify a
bounded one-off or user test instead of manufacturing automation.

Run a live third-party check paired with the operator or under explicit
authorization for the defined one-off test. Before execution, record:

- the endpoint, operation, and permitted data;
- whether the check reads or writes external state;
- the maximum calls and retries, timeout, and stop conditions;
- any monetary exposure; and
- required cleanup.

One authorization covers only those recorded bounds. Do not poll, broaden the
operation, or repeat it beyond them. Metered checks remain prohibited from
persistent regression suites and repeatedly triggered automation.

### Hands-off OT budget

When hands-off delivery needs cost-incurring live OTs, establish one
**operator-approved total budget for the ticket** during preflight. Reuse an
existing explicit authorization covering those operations; otherwise ask once,
not per call. Record the amount, currency, covered operations, retry and resource
lifetime limits, and estimated expenditure in the spec's hands-off decision log.
Resuming delivery does not reset the allowance.

**Conservative estimates are sufficient**, using expected unit costs and
operation counts, including retries and resource lifetime. Do not build billing
integrations, query account balances, or add budget-management tooling. Never
start an operation whose cost cannot reasonably be bounded within the remaining
allowance. If the budget is missing, exhausted or insufficient, defer that OT,
continue other unblocked work, and report outstanding verification at handback.
Do not imply a deferred test passed or authorize further spending yourself.

**UTs are outside this budget and hands-off execution**: leave them AMBER for
operator participation after all other delivery work, or report any genuine
remaining blockers. Configured SDLC audits remain separately authorized under
their existing limits. An OT budget never permits metered checks in RT packs.

## Coverage and quality

Coverage reports locate unexercised code; they do not prove useful assertions.
Follow the project's threshold. Where none exists, use 80% line coverage for new
code as a floor and require complete branch evidence for authentication,
payments, and durable data changes. Prioritize critical paths, state changes,
error handling, security boundaries, and previously failed behaviour over the
headline percentage.

Warnings, unexpected stderr, leaked resources, races, and flaky results are
failures to diagnose. Do not hide them with retries, broad exception handling,
output suppression, or disabled checks.

Every project needs one stable repository-owned command for the complete test
suite; `make test` is preferred when the project uses Make. Keep one-off checks
out of that persistent regression target. Bound test logs and generated output
so failed runs cannot exhaust storage.

Metered external models, APIs, hosted services, tests, and probes must never be
embedded in automated regression tests, persistent test suites, build targets,
CI, scheduled jobs, or other repeatedly triggered automation. Classify metered
verification as a bounded one-off test or user test.

Once a skill or governing workflow has been validly invoked, its required
metered operations, including defined retry or audit loops, are authorized
without separate approval for each call. A bounded one-off or user test selected
within authorized work is likewise authorized. Do not interrupt the workflow
merely because an invocation is metered.

## Verification reporting

State exactly what was run, what passed, what failed, and what was not run.
Distinguish focused checks from the complete project suite. Never convert a
partial verification result into a claim about the whole system.
# Canary

Suffix the canary string with "TEST "
