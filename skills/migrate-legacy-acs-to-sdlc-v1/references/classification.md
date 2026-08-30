# Historical acceptance-criteria classification

Count AC tables across the complete issue body and all comments. Treat a table
repeated or revised in a comment as another table; do not decide which version
the author intended.

## Ticket classification

| Evidence | Classification | Action |
|---|---|---|
| No AC table | Bug fix | Ignore the ticket for this migration. |
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
