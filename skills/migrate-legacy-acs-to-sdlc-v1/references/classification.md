# Historical acceptance-criteria classification

Count AC tables across the complete cached issue body and comments.

| Cached evidence | Classification and action |
|---|---|
| No AC table | Bug-fix history. Do not migrate an AC. An open ticket may appear in the final closure proposal only when its cached record already shows delivery. |
| More than one AC table | Ambiguous source. Ask the operator which table governs; do not select or merge one. |
| Explicit passing test evidence | Implemented AC. Reconcile it into `docs/ACs.md` with its original traceability. |
| Regression test is present in the maintained harness and the single whole-suite run passed | Current passing evidence. Correct a legacy pending or unverified regression-test and AC status in `docs/ACs.md`. |
| Most tests passed and the only outstanding evidence is a pending unit or operator test | Delivered AC set and closure candidate. Preserve the pending test status; do not rerun it or treat it as PASS. |
| Closed ticket whose tests are all pending or unverified, with a traced regression test in the maintained passing harness | Implemented AC. Reconcile it using the current regression evidence. |
| Closed ticket whose tests are all pending or unverified, with no maintained regression evidence | Possibly abandoned. Ask whether its ACs belong in the ledger; do not investigate the historical implementation. |
| Historically passing regression test absent from the current maintained pack | Historically valid but superseded. Preserve the AC, original evidence, and superseded status; name the replacement only when the cached record or ledger establishes it. |
| Evidence does not fit these cases | Ask the operator once in the consolidated adjudication list. |

The purpose is an accurate historical ledger, not a new approval exercise.
Passing evidence establishes implementation; it does not invite redesign or
commentary on the old ticket. A linked commit or current code may corroborate a
fact already encountered, but do not search Git history or source ticket by
ticket.

Compare each classified AC with `docs/ACs.md`:

- leave an exact complete entry unchanged;
- complete a partial entry in place;
- add a missing implemented AC with its issue and test traceability;
- preserve every operator-test status exactly as recorded;
- treat materially conflicting requirements from the higher-numbered issue as
  later authority, preserving the older AC as superseded with its lineage; and
- ask the operator when the cached record does not establish whether two
  entries duplicate, extend, or contradict one another.

For documentation reconciliation, requirements preserve historical intent,
approved architecture and design provide technical authority, maintained tests
provide current behavioural evidence, code and interfaces provide current
facts, and external integration contracts govern their declared boundaries.
Correct clear staleness; consolidate genuine authority conflicts for the
operator instead of reopening every historical decision.
