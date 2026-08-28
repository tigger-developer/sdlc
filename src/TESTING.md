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

Cover each independent condition in a compound requirement. Include relevant:

- happy, alternate, and failure paths;
- empty, missing, malformed, minimum, and maximum values;
- permissions and trust boundaries;
- partial completion and interruption;
- repeated and concurrent operation;
- compatibility and migration; and
- accessibility, layout, readability, or other human judgement.

Tests should be specified before implementation when practical. A useful loop
is to observe the test fail for the intended reason, implement the smallest
coherent change, observe it pass, then refactor without losing evidence.

## Choose durable evidence

Use three distinct forms of verification:

- **Regression tests:** retained because the behaviour could regress.
- **One-off verification:** temporary evidence for a migration, incident,
  environment, or documentation review that has no lasting regression value.
- **Human validation:** outcomes requiring visual, editorial, ergonomic, or
  other subjective judgement.

Do not create permanent test code solely to satisfy a test-count rule. Record
one-off or human evidence in the feature artefacts or final report, then remove
temporary files from the repository.

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

## Verification reporting

State exactly what was run, what passed, what failed, and what was not run.
Distinguish focused checks from the complete project suite. Never convert a
partial verification result into a claim about the whole system.
