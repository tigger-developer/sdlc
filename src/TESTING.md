---
title: Testing Standards
version: 6
last-updated: 2026-10-24
---

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
and allocate above the highest value ever assigned. Preserve superseded execution
entries in the applicable evidence record rather than recycling them.

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
behaviour. A change may require automated tests (AT), regression tests (RT),
one-off tests (OT), user tests (UT), or a combination of them.

Only automated tests follow test-driven development. When a meaningful
automated test (AT or RT) is justified, write or amend it before implementing
the target production behaviour, observe it fail for the intended reason, implement the smallest
coherent change, observe it pass, then refactor without losing evidence. When no
automated test is justified, record the specific reason; urgency,
difficulty, or inconvenience is insufficient. Review of test design and
traceability does not replace a required failing automated test execution.

Compilation/setup failures do not establish intended RED. Execution-enabling
scaffolding must not implement the target behaviour. Planned tests are not
execution evidence, and implementation cannot retroactively establish adequate
test definitions. Expected OT/UT evidence should be defined before implementation
where practical; neither category requires a pre-change failure.

## Choose durable evidence

Use four distinct test categories. A change may use more than one. AC means
acceptance criterion; it identifies a requirement, not a fifth test category.

- **Automated tests (AT):** repeatable automated verification for the change,
  outside the retained regression pack. Follow the same RED/GREEN TDD and
  execution-evidence rules as RTs.
- **Regression tests (RT):** automated tests retained in the normal regression
  pack because the behaviour could regress.
- **One-off tests (OT):** genuinely one-off verification, such as an irreversible
  migration check, without a repeatable RED/GREEN TDD cycle or regression retention.
- **User tests (UT):** human verification of outcomes requiring visual, editorial,
  ergonomic, operational, or other human judgement.

Any test that requires `sudo` must be classified as a user test (UT) and
executed by the operator. This includes `sudo` invoked indirectly by setup,
fixtures, helper scripts, dependencies, or cleanup. Such tests must never be
included in the automated regression pack or its persistent test targets,
CI, or scheduled automation. Passwordless `sudo` and cached credentials do
not change this classification.

Do not create permanent test code solely to satisfy a test-count rule. Record
one-off or user-test evidence durably, then remove temporary files from the
repository.

## Execution evidence

Evidence identifies the tested candidate, environment, method, dated observations
and actual result. An implementation claim is not execution evidence. A broken
fixture is inconclusive, not an intended test failure. Initially passing
preservation tests need no artificial failure. Historical results apply only
to the unchanged behaviour they assessed; changed candidates need current
evidence. Report missing or stale evidence instead of inferring success.

Live human judgement is first-class UT evidence. Do not replace it with an
automated imitation. Record the observed behaviour, reviewed candidate, material
viewing conditions and whether a later iteration superseded it. Human approval
is required before recording provisional paired observations as accepted UTs.

## Required human acceptance in specifications

Every managed change specification must define at least one meaningful UT with
an observable outcome and explicit human assessment. This includes internal
changes: a representative supported workflow may establish preserved behaviour.
Confirming that automated checks reported success is not a UT. Resolve an absent
credible UT during definition; do not waive the minimum or invent observations.
All required UTs must be GREEN before final acceptance, merge or closure. They
may remain AMBER during unattended delivery and delivery-code review. This
minimum applies to specifications, not exploratory spike briefs or standalone
code reviews, and does not invent advance gates for paired/emergency work.

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

## Document history

This revision separates independent review, author self-audit and delivery-only
responsibilities. Earlier rules remain in their canonical owners and Git history.

Version 4 centralizes review criteria and delivery orchestration while retaining
the testing contract and distinguishing independent review from self-audit.

Version 3 restores the four test categories and applies automated-test readiness
to ATs as well as retained RTs. The earlier three-category wording is superseded.

# Canary

Append " TEST" (with a leading space) to the canary.
