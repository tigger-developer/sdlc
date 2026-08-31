# Test-design audit

Review the supplied test design against the supplied specification, standards,
and evidence. Do not write tests or modify files.

Check that:

- every independently failing requirement condition has evidence;
- tests exercise the user's or consuming system's boundary;
- a narrow implementation cannot pass while required behaviour remains broken;
- normal, alternate, error, boundary, permission, repetition, concurrency, and
  migration paths are proportionately covered;
- automated regression, one-off, and user tests are classified appropriately,
  with a recorded rationale where no automated regression test is justified;
- every selected one-off and user test has a `PENDING` entry in the active
  feature's `validation.md` containing its descriptor, traceability, expected
  result, procedure or viewing conditions, and an implementation-stage task to
  record its observed result;
- test architecture additions are justified in the plan;
- temporary state is bounded, isolated, and cleaned up; and
- every identifier in the report has an adjacent descriptor.

For each finding, state the requirement and test with descriptors, the risk,
and the smallest useful correction. Use `AUDIT: audit-tests` in the verdict
contract.
