---
description: Write automated tests for the current issue. TDD red phase -- no implementation code.
---

Write the automated tests for the current issue now. Do not write any other code -- only tests.

For each test:

1. **State the real-user test.** What user action does this test simulate? What would the user observe? If the answer references internal APIs, database rows, or source code -- rewrite the test before proceeding.

2. **Verify coverage.** For each AC, confirm that every distinct condition is covered by at least one test. Flag any AC with only one test as a potential gap.

3. **Identify weaknesses.** For each test, ask: what is the narrowest shortcut that would make this test pass without the feature actually working? If such a shortcut exists, add a test that blocks it.

4. **Identify gaming opportunities.** How could an implementation agent make these tests pass while building something incomplete or broken? Add tests that close those loopholes.

5. **Isolate mutable state.** Confirm that fixtures, caches, generated artefacts, databases, logs, coverage intermediates, and retained command output use a unique per-run operating-system temporary directory. Confirm explicit teardown covers success, failure, and handled interruption, and that generated data is subject to the project's finite size limit or the 100 MiB default in TESTING.md.

Run the tests. Confirm they fail (TDD red phase). Commit only the test files.

Then proceed to implementation. Do not stop and wait for me to invoke `/implement`.
