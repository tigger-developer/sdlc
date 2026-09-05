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
- human validation is accepted as first-class evidence when visual, ergonomic,
  editorial, operational, or other human judgement is the only credible test,
  without replacing it with an automated imitation;
- third-party doubles model only a published, versioned, or otherwise verified
  contract, and are not claimed as evidence of undocumented or live provider
  behaviour;
- every live third-party one-off test records its endpoint or operation,
  permitted data, read or write effect, maximum calls and retries, timeout,
  stop conditions, monetary exposure, cleanup, and paired or explicit operator
  authorization;
- no metered or live third-party check is embedded in a persistent regression
  suite or repeatedly triggered automation;
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
