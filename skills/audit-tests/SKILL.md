---
name: audit-tests
description: Challenge test specifications for coverage gaps, gaming opportunities, and integration risks. Advisory only -- no code or file changes.
---

Review the test specifications for the current issue. Do not write any code or modify any files. The purpose of the audit is not just to present problems; it is to present solutions. Always include a concrete recommended action for every finding.

1. **Coverage check.** For each AC, confirm that every distinct condition implied by the AC is covered by at least one test. Flag any AC with only one test as a potential gap. Flag any multi-condition AC where not all conditions have tests.

2. **Real-user test.** For each specified test, state: what user action would this test simulate, and what would the user observe? If the answer would reference internal APIs, database rows, source code, or any artefact the user never sees -- flag it.

3. **Gaming analysis.** For each test, ask: what is the narrowest shortcut an implementation could take to make this test pass without the feature actually working? If such a shortcut exists, recommend an additional test that blocks it.

4. **Integration gaps.** Are there interactions between components that no test covers? Could all tests pass individually while the integrated system is broken?

5. **Test architecture creep.** Check whether the proposed tests add a new runtime, package manager, browser framework, test runner, or other test architecture. Flag this unless the issue includes explicit human approval. Prefer the project's existing test harness and the narrowest real-behaviour test. For web/static-site checks, source files must never be grepped; prefer generated-site tools (`htmltest` for broad HTML/link validation, `htmlq` for targeted DOM assertions, `xmlstarlet` for targeted RSS/feed assertions) before proposing Node, Playwright, Cypress, or a new browser stack.

6. **Temporary-state check.** Confirm that all mutable test state uses a unique per-run operating-system temporary directory, teardown covers success, failure, and handled interruption, and generated data has a finite size limit. Flag project-local scratch, writes to home-directory caches, missing cleanup, shared temporary paths, and unbounded output. Recommend the smallest concrete correction for each violation.

7. **Type check.** Verify each test is correctly typed as RT/OT/UT per the decision tree in TESTING.md. Flag any UT that could be automated, any RT that invokes the build system, any OT for ongoing behaviour.

8. **Summarize.** List all findings: coverage gaps, gaming opportunities, integration risks, test-architecture creep, temporary-state violations, and mistyped tests. Recommend specific additions or rewrites for each finding.

End with `PASS` when no unresolved findings remain; otherwise end with `FINDINGS` and the concrete remediation for each finding.

In MODE PAIR, do not proceed past this audit; wait for the human's response. In MODE DELIVER, post a `QUALITY CHECK: TEST AUDIT` comment on the affected issue using the verdict, evidence, findings, remediation, and attempt fields defined in `ISSUES.md`, then return control to the delivery workflow. PASS advances immediately to implementation when the AC audit has also passed. FAIL returns to remediation and re-evaluation; it is not a handback by itself. The workflow must remediate every finding and repeat this audit until it reports PASS. No code may be written for the ticket before both `audit-acs` and `audit-tests` report PASS.
