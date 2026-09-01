# Code audit

Review only the supplied implementation change and the adjacent context needed
to assess it. Do not modify or repair files. Do not make unrelated legacy
defects blocking unless the change depends on them, worsens them, or cannot be
assessed safely without resolving them.

Check:

- conformance to the supplied specification, design, constitution, and
  standards;
- correctness, error paths, state transitions, idempotence, and compatibility;
- injection, traversal, secret exposure, unsafe permissions, and widened access;
- hidden failure, disabled verification, broad suppression, and misleading
  diagnostics;
- concurrency, cleanup, resource, timeout, and retry behaviour;
- for deployable applications, the required `make vulncheck` implementation and
  current result, kept separate from behavioural regression evidence;
- canonical Make target names and common variable semantics where the project
  uses Make;
- maintainability and ecosystem idioms;
- implementation of the approved automated tests without reopening the
  effective test-strategy audit; and
- documentation required by changed interfaces or operations.

Final one-off and user tests are completion evidence executed after an
effective code-audit PASS. Their missing results do not fail this audit.
Challenge the approved test strategy only when the implementation deviates from
it or concrete implementation evidence reveals an upstream contradiction.

For each finding, give severity, a file and line reference from the supplied
material, the violated requirement or standard with a descriptor, evidence,
consequence, and concrete remediation. Use `AUDIT: audit-code` in the verdict
contract.
