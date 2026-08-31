# Historical acceptance-criteria classification

Count AC tables across the complete issue body and all comments. Treat a table
repeated or revised in a comment as another table; do not decide which version
the author intended.

## Open-ticket readiness classification

| Evidence | Classification | Action |
|---|---|---|
| Exactly one AC table and every AC has passing evidence | Delivered feature candidate | Present the ticket, AC evidence, and documentation variances for operator-authorized reconciliation and closure. |
| Exactly one AC table and most ACs have passing evidence | Probably or partially delivered | Present every outstanding AC and test. Do not infer closure or migrate any AC until the operator adjudicates the remainder. |
| Exactly one AC table without sufficient delivery evidence | Undelivered or unresolved feature | Mark `NO ACTION`; leave the ticket and its ACs untouched. |
| No AC table, with implementation and maintained regression evidence | Delivered bug-fix candidate | Present the evidence for operator-authorized closure; no AC migration is required. |
| No AC table without sufficient delivery evidence | Undelivered or unresolved bug fix | Mark `NO ACTION`; leave the ticket untouched. |
| More than one AC table | Ambiguous source | Do not select, merge, or migrate a table. Ask the operator which table is authoritative. |

Passing evidence establishes implementation, not approval. No classification
authorizes ticket closure, an AC change, or a documentation change. An
operator may approve a clearly enumerated batch after reviewing the proposed
actions.

Treat every unapproved partially delivered, ambiguous, or undelivered ticket as
untouched open scope. Do not create a Spec Kit draft or successor ticket for it.

## Historical ticket source classification

Use this source classification for the complete open-and-closed ticket pass:

| Evidence | Classification | Action |
|---|---|---|
| No AC table | Bug-fix history | No AC migration is required. |
| Exactly one AC table | AC-bearing ticket | Classify every AC independently. |
| More than one AC table | Ambiguous source | Do not select, merge, or migrate a table. Ask the operator which table is authoritative. |

## AC classification

Classify using the strongest applicable evidence in this order:

1. **Current:** At least one referenced unit or regression test remains in the
   supported checked-in test harness. Preserve every other associated test and
   its status. A regression test may be recorded as passing only when the
   supported whole-suite command has current passing evidence.
2. **Superseded:** A regression test recorded as passing historically is absent
   from the current regression pack. Preserve the AC as historically valid and
   superseded. Name the replacement AC when the evidence establishes it;
   otherwise record that the replacement is unresolved.
3. **Historically valid:** No current live test or historical regression test
   establishes a later classification, but at least one associated unit test is
   recorded as passing. Preserve the AC as valid historical behaviour, retain
   every companion test status, and record its current regression-coverage gap.
4. **Possibly abandoned:** The only associated tests are unit tests recorded as
   pending, or every associated automated test remains pending. Do not migrate
   the AC as valid without operator adjudication.
5. **Unresolved evidence:** The AC has no associated test, only operator tests,
   contradictory statuses, or evidence that does not fit the preceding cases.
   Preserve every observed status in the report and ask the operator.

The presence of any passing unit or regression test establishes that the AC was
implemented. It does not override stronger evidence that the AC was later
superseded.

Preserve operator-test statuses exactly as recorded. Never infer, promote, or
otherwise change them. A linked code commit is supporting implementation
evidence only; it does not resolve pending-only or conflicting test evidence.

## Existing centralized entries

Compare every classified AC with the centralized record:

- An exact existing entry needs no content change; add only missing provenance
  or traceability.
- A partial entry is completed in place.
- Different wording is not automatically a contradiction. Both requirements
  may remain valid when one narrows or extends the other.
- When two requirements materially conflict, the higher-numbered issue takes
  precedence. Preserve the older AC as superseded, including its original
  wording, test relationships, and both issue descriptors.
- When the evidence does not establish whether entries duplicate, extend, or
  contradict one another, leave both unchanged and ask the operator.

These comparisons produce proposed actions only. Add, complete, supersede,
retire, or otherwise change an AC only after the operator authorizes that exact
action, individually or in an enumerated batch. Preserve the old wording,
status, provenance, and replacement relationship whenever an AC is superseded
or retired.

## Documentation readiness

For each delivered closure candidate and again after the historical AC pass,
compare the project documentation with the strongest applicable evidence:

1. approved current requirements;
2. explicit later supersession decisions;
3. current design and architecture authority;
4. maintained regression behaviour and traceability;
5. current implementation and interfaces as evidence of facts; and
6. external ownership or integration contracts.

Report stale or contradictory documentation as a variance. Do not silently
choose between conflicting authorities, treat code as approval of a requirement,
or rewrite a document merely because its language is old. Apply only
operator-authorized corrections and preserve useful historical decisions.
