# Testing Standards

This document defines how to write, structure, and organize tests. The TDD workflow sequence - when to enumerate, when to write tests, when to run regression - is in `MAIN.md` §3.

## The real-user test

Tests must verify observable behaviour through the same entry point a user would use. Before marking any RT as passing, state in chat: *what user action does this test simulate, and what would the user observe?* If the answer references internal APIs, database rows, source code grep results, or any artefact the user never sees - the test doesn't match its spec. Rewrite it.

This is not a rule about any specific shortcut. It is the question that all the other testing rules exist to enforce: *"Would this test have caught a real bug?"*

**Anti-patterns this catches:**
- Calling a library function when the spec says "run the CLI command"
- Asserting a database row exists when the spec says "user sees a confirmation page"
- Grepping source code instead of running anything
- Checking that a string appears somewhere without exercising the real boundary behaviour
- Testing the narrowest internal slice that technically touches the AC while leaving the integration between components - and the integration to functionality - unverified

Source code must never be grepped, parsed, or otherwise introspected to make a behaviour test pass. Grep is a discovery tool, not a test oracle. If a test needs to verify HTML, CSS, JavaScript, assets, routes, or generated files, test the rendered/built artefact or the served response that a user would receive.

Do not write tests whose only assertion is that a string exists. String-presence tests are a cheat: they create the impression of broad coverage while proving almost nothing about robustness or functionality. They are no better than source-code greps when they do not exercise the real behaviour through a real boundary entry point. Test the functionality through the CLI, HTTP route, UI, generated artefact, API boundary, or other user/system entry point, then assert on the observable result. The narrow exception is when the exact string is itself the contract being tested (for example help text, an error message, rendered copy, or a feed title), and even then the string must be observed through the real boundary that exposes it.

## Test architecture changes

Do not install or introduce a new test architecture without explicit human approval. This includes new runtimes, package managers, browser frameworks, test runners, and large dependency stacks.

Prefer the project's existing test harness and the smallest tool that verifies real behaviour. For static or mostly-static web output, this usually means:

- build/render the site with the existing project command
- run `htmltest` for broad generated-site HTML/link validation
- use `htmlq` for targeted DOM assertions against rendered HTML
- use `xmlstarlet` for targeted RSS/feed assertions against generated XML
- use an existing project harness only when serving behaviour itself is under test

Do not add Node.js, npm, Playwright, Cypress, or another browser stack merely to check that static files are served. Browser tooling is justified only when the behaviour depends on browser execution, computed styles, interaction state, or JavaScript-rendered content, and the human has approved the added test architecture.

## TDD

We practice test-driven development:

1. Enumerate test cases for each AC (every distinct condition gets a test)
2. Write failing tests defining the desired behaviour
3. Run them, confirm they fail
4. Write minimal code to pass
5. Run them, confirm they pass
6. Refactor while keeping tests green

Each new test becomes part of the **regression test pack** (`make test`).

## Temporary test state

Tests may launch from the repository when required by the project tooling, but all mutable test state must live in a unique per-run directory beneath the operating system temporary directory. This includes fixtures, caches, generated sites, databases, captures, logs, coverage intermediates, build products, and command output retained for later assertions. Writes caused by a test runner or another invoked tool count as test writes.

Use the test framework's temporary-directory fixture or the platform's temporary-directory API. In shell, use `mktemp -d` without hard-coding `/tmp` or assuming `$TMP` is set. Configure tools that normally write caches or generated data elsewhere so they use the per-run directory.

The harness must remove the exact per-run directory on success, failure, and handled interruption using framework teardown, `defer`, `try/finally`, a context manager, or a shell `trap`. Cleanup must never target a shared parent directory, use a wildcard, or depend on a later agent interaction. Tests that intentionally exercise interruption must prove their own teardown behaviour.

Generated test data must have a finite size limit. Each project may document a lower or higher limit appropriate to its domain; where no project limit exists, the maximum is 100 MiB per test run. Exceeding the limit is a visible test failure. Tests must stream, truncate, or discard repetitive output rather than accumulating it silently.

Persistent golden files, snapshots, fixtures, and review reports are not temporary test state. Their project path and retention purpose must be explicit and reviewed. Test logs copied into a temporary review report may be retained only for the active review and must not remain in the project tree.

## Multi-condition coverage

An AC that implies multiple conditions requires a test for each condition.

Example: the AC *"The CLI rejects invalid, missing, and malformed configuration files"* implies three conditions. You must write at least three tests - one for invalid, one for missing, one for malformed. Writing a single test and marking the AC as passing is a process violation.

## Issue tests vs regression tests

- **Issue tests:** the tests written specifically for the issue being worked on. Run these frequently during development.
- **Regression tests:** the full test suite (`make test`). Run these only at batch boundaries, not after every individual issue.

Trust the issue tests during development; the regression pack catches cross-cutting breakage at the end.

## One-off tests

Most tests belong in regression. **This is the default.** If you are writing a new test and are unsure where it belongs, it belongs in regression.

A **one-off test** is the exception: a test tied to a specific, non-recurring activity - verifying a data migration, reproducing a production incident, checking a one-time environment state.

### Storage

```
tests/
  regression/    ← default; run by `make test`
  one_off/       ← quarantine; never run by `make test`
```

`make test` must only reference `tests/regression/`.

### Marker

All one-off tests must carry a marker with a mandatory `issue` reference, using whichever annotation mechanism the project's test framework provides:

```python
# Python/pytest - illustrative only
@pytest.mark.one_off(issue="#123")
def test_migration_user_records_backfilled():
    ...
```

If the framework has no native marker, encode the issue reference in the test name: `test_migration_user_records_backfilled_OT123_1`.

A one-off test without an `issue` reference must fail linting. Both directory placement and marker are required.

### Running one-off tests

```bash
make test-one-off          # run all one-off tests
make test-one-off ISSUE=123  # run one-off tests for a specific issue
```

### Decision rule

Work through the decision tree in [Choosing the right test type](#choosing-the-right-test-type). The default among *automated* tests is RT, but a test that requires human verification is always a UT regardless of how simple it seems, and a test tied to a one-time event is always an OT.

### Lifecycle

One-off tests are temporary. They must be reviewed and deleted once their associated issue is closed and verified.

## Test naming

```
test_<unit>_<scenario>_<expected_result>
```

Examples:
- `test_login_with_invalid_password_returns_401`
- `test_cart_add_item_increases_total`
- `test_export_empty_list_returns_empty_file`

The test name should tell you what failed without reading the code.

## Test IDs

All tests carry a unique ID. IDs are scoped to the issue that created them, following the same namespacing pattern as ACs. The prefix indicates type:

| Format | Type | Location | Run by |
|--------|------|----------|--------|
| `RT-{issue}.{n}` | Regression test | `tests/regression/` | `make test` |
| `OT-{issue}.{n}` | One-off test | `tests/one_off/` | `make test-one-off` |
| `UT-{issue}.{n}` | User test | originating issue; delivery master while rolled up | human |

`{issue}` is the GitHub issue number. `{n}` is a sequential integer starting at 1 within each issue and prefix. For example, issue #12 might allocate RT-12.1, RT-12.2, RT-12.3, OT-12.1, and UT-12.1.

**User tests (`UT-{issue}.{n}`)** require a human to perform and verify manually. They have no corresponding test file. Description, steps, and expected outcome are documented with the AC while its issue is open. Only I can mark a UT as passing or failing.

### MODE PAIR user-test presentation

Every UT must be presented to the human and receive an explicit passing result before the agent seeks approval to close its issue. The agent must start the application or tool, create or select representative data, apply the required configuration, and place the system in a ready-to-inspect state. The human must not be asked to run commands, perform setup, navigate through a sequence of screens, or reconstruct the test state.

The human normally performs only visual or subjective inspection. When the behaviour under test inherently requires interaction, the presentation may ask for at most one reasonable action per UT, such as selecting the prepared control or submitting a prepared form. If independent actions are required, define and present separate UTs. A pending, unpresented, or failing UT blocks `READY FOR REVIEW` and the request for APPROVED.

### MODE DELIVER user-test lifecycle

When an audited MODE DELIVER child is otherwise eligible to close, each pending UT moves from the child AC table to the master issue's user-test roll-up. The transferred row preserves the UT ID, originating AC and child issue, user action, expected result, and pending status. The master becomes the single active definition and resolution location for that UT; the migrated central AC entry retains only its ID, status, and a link to the master row.

All rolled-up UTs are resolved together at the end of the MODE DELIVER cycle. Only the human may mark them passing or failing. A pending or failed UT blocks master closure. A failed UT is diagnosed and remediated through an audited child when the correction remains within the confirmed scope. When the human resolves a UT, its central AC provenance status is updated to match while its full definition remains on the master. Resolved UT rows remain on the master as delivery history, and **APPROVED n**, where n is the master issue number, accepts the results for the complete delivery tree.

An **APPROVED n** directive received while any rolled-up UT is pending or failing does not carry forward. Keep the master open, report the unresolved UTs, and require a fresh approval directive after every rolled-up UT has a human-confirmed passing result.

### Choosing the right test type

Ask two questions in order:

1. **Can a machine verify this?** If the outcome requires human judgement - visual correctness, subjective quality, natural-language readability - it is a **UT**. No code is written; the test is documented in the AC table only.

2. **Will this test matter after the issue closes?** If it verifies a one-time event - a data migration, an incident reproduction, a transient environment state - it is an **OT**. If it guards ongoing behaviour that could regress, it is an **RT**.

| Can a machine verify it? | Matters after issue closes? | Type |
|---|---|---|
| No | - | **UT** (human test; rolled up to the delivery master when applicable) |
| Yes | No | **OT** (one-off, `tests/one_off/`) |
| Yes | Yes | **RT** (regression, `tests/regression/`) |

**Examples:**
- "The menu icon looks correct at all display scales" -> **UT** - requires human visual judgement
- "Migration backfills all legacy rows" -> **OT** - once migrated, the test has no purpose
- "Login rejects empty password with 401" -> **RT** - must never regress

**Anti-pattern: tests that invoke the build system.** Never write an RT (or any automated test) that invokes `make`, `make test`, or any build/test harness target. Since RTs run inside `make test`, this creates infinite recursion. Tests that verify Makefile behaviour, CLI entry points, or build outputs belong as **OTs** or **UTs**, not RTs.

### Usage in markers

```python
# Python/pytest - illustrative only
@pytest.mark.regression(test_id="RT-12.1")
def test_login_with_invalid_password_returns_401():
    ...

@pytest.mark.one_off(issue="#123", test_id="OT-123.1")
def test_migration_user_records_backfilled():
    ...
```

If the framework has no native marker, encode the ID in the test name as a suffix:
```
test_login_with_invalid_password_returns_401_RT12_1
```

### Usage in GitHub issues

```
| ID | AC | Tests |
|---|---|---|
| AC12.1 | Login rejects invalid passwords | ✅ RT-12.1: Empty password returns 401<br>✅ RT-12.2: Wrong password returns 401 |
| AC12.2 | Migration backfills user records | ✅ OT-12.1: Legacy rows backfilled<br>⏳ UT-12.1: Spot-check migrated accounts |
| AC12.3 | ~~Chooser filters by app name~~ | ~~🚫 RT-12.3: Filtered list matches query~~ |

**Key:** ✅ passing · ⏳ pending · ❌ failing · ~~🚫 removed~~
```

### ID allocation

Test IDs are scoped to the issue that creates them. No central counter file is needed.

To allocate a new test ID for issue #N:

1. Check the AC table and any existing tests for that issue to find the highest allocated number for the relevant prefix (RT, OT, or UT).
2. Increment by 1.

If no tests of that prefix exist yet for the issue, start at 1 (e.g. `RT-N.1`).

### Mid-project migration

#### Directory structure

Projects with a flat `tests/` directory must be migrated before any other work:

1. Move all existing tests into `tests/regression/`
2. Create `tests/one_off/.gitkeep`
3. Update `make test` to point at `tests/regression/`
4. Commit: `chore: migrate test suite to regression/one_off layout`
5. Then proceed with the issue

This must be its own commit.

#### Test IDs

1. **Do not retrofit IDs.** Pre-existing tests without IDs are not modified solely to add one.
2. **Migrate on touch.** If a pre-existing test is modified as part of an issue, add an ID then.
3. **New tests always get an ID.** No exceptions.
4. **Do not flag missing IDs unprompted.** Note in the issue comment but do not modify the test.

## Test structure

Follow Arrange-Act-Assert (AAA) regardless of language:

```
test "discount applied to eligible order":
    # Arrange
    order = Order(items=[item_over_threshold])
    # Act
    result = apply_discount(order)
    # Assert
    assert result.discount_percent == 10
```

## Coverage

- **Minimum:** 80% line coverage for new code
- **Critical paths:** 100% for authentication, payment, data persistence
- Coverage is a floor, not a ceiling - high coverage with weak assertions is worthless

## Test boundaries

| Type | Scope | Speed | When to run |
|------|-------|-------|-------------|
| Unit | Single function/class | Fast (<100ms) | Every save |
| Integration | Multiple components | Medium (<5s) | Pre-commit |
| End-to-end | Full system | Slow | CI pipeline |

- Unit tests: no external dependencies (no network, filesystem, database)
- Integration tests: may use local resources (test database, mock servers)
- E2E tests: exercise the real system

## Test data

- Use factories for generating test objects (factory_boy, FactoryBot, fishery, etc.)
- Fixtures for shared setup - minimal and obvious
- No production data - synthetic only
- Deterministic: same result every run (no random without seed)

## The "no mocks" rule

The rule in CODING.md means **do not mock your way out of testing real behaviour**.

**Acceptable:** mocking external HTTP APIs, time/dates, filesystem for unit tests (use the framework's system-temporary fixture).

**Not acceptable:** mocking the class under test, mocking so much the test verifies nothing real, using mocks instead of fixing design to be testable.

If removing the mock would make the test fail for reasons *other than* external dependencies, you're mocking too much.

## Warnings and errors

- Never introduce new warnings or errors
- If an action generates a warning or error, stop and resolve it
- Build-time errors are equivalent to compile errors - fix immediately

## Policy

- All production code requires unit, integration, and end-to-end tests
- Test coverage must be runnable via `make test`
- Update regression tests; create them if none exist

**Exception:** requires explicit approval - I must say **"I AUTHORIZE YOU TO SKIP WRITING TESTS THIS TIME"**

## Makefile test targets

```makefile
lint:
	<linter> .

test: lint
	<test-runner> tests/regression/

test-one-off:
ifdef ISSUE
	<test-runner> tests/one_off/ <filter-flag> "$(ISSUE)"
else
	<test-runner> tests/one_off/
endif
```

`make test` must never reference `tests/one_off/`.

# Canary
Suffix the canary string with "TEST "
