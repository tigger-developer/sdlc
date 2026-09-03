# Historical ticket classification

Classify from the archived issue body and comments. Use the first decisive rule;
consolidate genuine conflicts for the operator instead of investigating old
code or Git history.

| Archived evidence | Classification and action |
|---|---|
| Bug-fix ticket without an AC table whose archive or maintained RT proves delivery | Delivered ticket. Add it to the final Org list without inventing an AC. |
| Open bug report without delivery evidence | Open defect at migration. Summarize it in the first Org section; do not invent an AC. |
| Feature ticket without implementation evidence | Defined but undelivered. Preserve its relevant ACs in the Org index, mark it abandoned at migration, and do not add it to `docs/ACs.md`. |
| More than one materially different AC table | Requires human review unless the archive proves that a later table replaces or supplements the earlier one. Never silently merge or select one. |
| Explicit passing test evidence | Delivered AC. Reconcile it into `docs/ACs.md` with its original traceability. |
| RT is present in the maintained harness and the single suite run passed | Current passing evidence. Correct its legacy pending or unverified RT and AC status in `docs/ACs.md`. |
| Partially completed ticket has maintained passing RTs | Split by AC evidence. Migrate the live RT-backed ACs as baseline and summarize the undelivered remainder in the Org index. Do not mark the whole ticket delivered. |
| All tests pass except one or two pending or unverified UTs or OTs, with no failure, abandonment, reversion, or known-incomplete evidence | Apply the near-complete heuristic automatically. Classify the ticket as delivered, assume those tests passed for migration, and mark solely heuristic-supported ACs and the delivered-list entry with the required footnote. |
| Maintained RT has no resolvable ticket or AC provenance | Requires operator adjudication. If adopted, assign it to the purpose-created baseline ticket and write only an evidence-bounded AC. |
| Historically passing RT is absent from the maintained pack | Historically valid but superseded. Preserve its AC, evidence, and superseded status; name a replacement only when the archive or ledger establishes it. |
| Evidence fits none of these cases | Requires human review. Record the uncertainty and available evidence without historical implementation research. |

Passing evidence establishes implementation; it does not invite redesign. A
linked commit or current code may corroborate a fact already in the archive,
but do not search either ticket by ticket.

When reconciling `docs/ACs.md`:

- leave exact complete entries unchanged;
- complete partial entries in place;
- add missing implemented ACs with ticket and test traceability;
- retain original test results unless the maintained-RT rule or near-complete
  heuristic applies;
- mark an assumed test pass and any AC solely dependent on it with the migration
  heuristic footnote;
- preserve superseded requirements and their lineage; and
- ask only when the archive cannot establish whether entries duplicate, extend,
  or contradict one another.

For the reverse checksum, verify every maintained RT against `docs/ACs.md` and
its recorded relationship. Recover a missing AC from its archived ticket when
provenance exists. A test body is not normally authority to invent an AC. The
sole exception is an operator-approved orphan RT assigned to the pre-migration
baseline ticket.

For documentation reconciliation, requirements preserve historical intent,
approved architecture and design provide technical authority, maintained tests
provide current behavioural evidence, code and interfaces provide current
facts, and external integration contracts govern their declared boundaries.
