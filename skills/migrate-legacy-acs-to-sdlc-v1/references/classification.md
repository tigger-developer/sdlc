# Historical ticket classification

Use this classification only after reviewing every maintained regression test
and completing the RT-to-ticket-to-AC evidence map. Classify from that map and
the archived issue body, comments, and recorded commit references. Apply every
relevant integrity and evidence rule; the rules are not mutually exclusive.
Establish delivery from evidence before considering ticket state. Consolidate
genuine conflicts for the operator instead of investigating old code or Git
history.

| Archived evidence | Classification and action |
|---|---|
| AC already exists in `docs/ACs.md` | Delivery evidence even when its archived ticket records no test result. Preserve the AC and reconcile any missing provenance or test relationship. |
| Ticket-linked code commit, with no explicit reversion, abandonment, or partial-delivery statement | Delivery evidence. Classify the ticket and its ACs as delivered unless stronger archived evidence limits the commit's scope. |
| Bug-fix ticket without an AC table whose archive or maintained RT proves delivery | Delivered ticket. Add it to the final Org list without inventing an AC. |
| No delivery evidence remains after applying the evidence rules | Use its snapshot state only now: closed scope is abandoned; open scope is undelivered. Put an open bug in the open-defects section and feature scope in the defined-but-undelivered section with that precise disposition. |
| More than one materially different AC table | Requires human review unless the archive proves that a later table replaces or supplements the earlier one. Never silently merge or select one. |
| Explicit passing test evidence | Delivered AC. Reconcile it into `docs/ACs.md` with its original traceability. |
| Ticket has no recorded passing test evidence but at least one RT in the maintained passing harness | Assume the whole ticket was delivered. Migrate every AC, treat pending or unverified UTs and OTs as assumed passing, and footnote ACs lacking independent evidence plus the delivered-list entry. |
| At least one test has a recorded pass and all others pass except one or two pending or unverified UTs or OTs, with no failure, abandonment, reversion, or known-incomplete evidence | Apply the near-complete heuristic automatically. Classify the ticket as delivered, assume those tests passed for migration, and mark solely heuristic-supported ACs and the delivered-list entry with the required footnote. |
| RT is present in the maintained harness and the single suite run passed | Current passing evidence. Correct its legacy pending or unverified RT and AC status in `docs/ACs.md`. |
| Maintained RT has no resolvable ticket or AC provenance | Automatically assign it to one purpose-created pre-migration baseline ticket and write only an evidence-bounded AC. No operator confirmation is required. |
| Historically passing RT is absent from the maintained pack | Historically valid but superseded. Preserve its AC, evidence, and superseded status; name a replacement only when the archive or ledger establishes it. |
| Evidence fits none of these cases | Requires human review. Record the uncertainty and available evidence without historical implementation research. |

Each RT in the passing maintained harness is delivery evidence for the behaviour
it verifies and every ticket and AC to which it directly traces. Passing
evidence and ticket-linked commits establish delivery without inviting
redesign. Do not inspect current code or search Git history ticket by ticket. An
open ticket may therefore be delivered and a closed ticket may be abandoned;
state never overrides evidence.

When reconciling `docs/ACs.md`:

- leave exact complete entries unchanged;
- complete partial entries in place;
- add missing implemented ACs with ticket and test traceability at their correct
  identifier positions, preserving ledger sequence;
- retain original test results unless the maintained-RT rule or near-complete
  heuristic applies;
- mark an assumed test pass and any AC solely dependent on it with the migration
  heuristic footnote;
- preserve superseded requirements and their lineage; and
- put unresolved duplication, extension, or contradiction in the human-review
  section without interrupting the migration.

For the reverse checksum, verify every maintained RT against `docs/ACs.md` and
its recorded relationship. Recover a missing AC from its archived ticket when
provenance exists. A test body is not normally authority to invent an AC. The
sole exception is an orphan RT automatically assigned to the pre-migration
baseline ticket.

For documentation reconciliation, batch ACs by product area before comparing
them with requirements, design, architecture, maintained tests, and external
contracts. Update clearly stale documentation directly. Inspect only the
relevant implementation when those authorities leave the current behaviour
uncertain; never inspect code once per AC.
