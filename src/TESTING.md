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
automated regression test is justified, write or amend it before changing
production code, observe it fail for the intended reason, implement the smallest
coherent change, observe it pass, then refactor without losing evidence. When no
automated regression test is justified, record the specific reason; urgency,
difficulty, or inconvenience is insufficient. An `audit-tests` PASS confirms
test design and traceability; it does not replace a required failing automated
test execution.

Define the expected evidence for one-off and user tests before implementation
where practical. They do not follow TDD and do not require a pre-change failure.
For staged Spec Kit and `BYPASS-GATE-7` work, execute their final verification
after `audit-code` has an effective PASS. Earlier diagnostic executions may
inform implementation but are not final evidence unless they remain current for
the audited candidate. Paired development uses the live user-validation contract
below.

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

## Non-automated test results

When a Spec Kit feature selects any one-off or user test, its active feature
directory must contain `validation.md`. Create the record before implementation
with each selected test marked `PENDING`, then replace that status with the
observed `PASS` or `FAIL` result after execution. Non-Spec Kit projects must use
an equivalent durable project record.

Every planned entry must include:

- a test identifier with an adjacent descriptor and its one-off or user-test
  category;
- the requirement or task it verifies;
- the expected result and procedure or human viewing conditions;
- `PENDING`, `PASS`, `FAIL`, or `SUPERSEDED` status.

A completed or superseded entry must also include:

- the tested revision and relevant environment;
- the observed result and supporting evidence;
- the tester or human authority for a user test; and
- any later result that supersedes it.

Do not infer PASS from a completed task, an agent report, or an implementation
claim. Required one-off and user tests remain incomplete until their results are
recorded in `validation.md`. Completion and convergence require a current PASS
for every required entry; missing, `PENDING`, `FAIL`, or materially stale results
block closure, not `audit-code`.

An audit verdict remains historical evidence for the revision and scope it
assessed. If later remediation changes relevant code, the earlier `audit-code`
PASS is no longer current for completion. Rerun the affected automated tests and
`audit-code`, then repeat only the one-off and user tests materially affected by
the change. Unaffected results remain current.

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
as the user tests for the change. On approval, write it to the active feature's
`validation.md`. Outside Spec Kit, use the project's equivalent durable evidence
record. A later change invalidates only entries whose observed behaviour it
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
