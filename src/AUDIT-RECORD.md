---
title: Canonical Audit Record
version: 1
last-updated: 2026-10-29
---

# Canonical audit record

For each managed ticket, `audits.yaml` beside `spec.org` is the **single
canonical audit record** for self-audits and independent external audits.
It is owned by the review process, not by a particular harness. The old external
audit harness is retired. A future CI harness may implement this contract;
its existence is not a prerequisite for recording or accepting a review.

The authorized self-reviewing agent records its own assessment. An independent
reviewer remains findings-only and returns its result; the authorized caller
records that result without changing its verdict or claiming greater separation.
Both methods use REVIEW's criteria. Read this contract without loading delivery
workflow when operating as an independent reviewer.

## No parallel audit documents

Do not create or continue `internal-review.md`, `reviews.md`, independent-review
Markdown files, or other parallel audit ledgers. Do not copy findings, verdicts,
round counts or reset state into `spec.org`, `work.org` or `validation.org`.
Those documents link to the relevant entry in `audits.yaml`. Validation results,
operator sign-off, lifecycle state and implementation decision logs retain their
own documented owners; they are not duplicate audit verdicts.

Raw evidence may be referenced as an attachment, not maintained as another audit
ledger. Historical review files are retained unchanged and linked as history;
they are not continued in parallel. Do not fabricate a conversion or turn an
old PASS into current evidence. Before further review, reconcile known prior
rounds, reset authority and unresolved findings into the canonical record with
source references. If history is incomplete, record the uncertainty and obtain
operator direction before treating an exhausted or unknown allowance as fresh.

Existing legacy `audits.yaml` formats must not be overwritten or silently
reinterpreted. Preserve the original as a clearly marked historical snapshot
before an explicit conversion, link it, and retain the original method and
provenance. Unknown legacy fields/results are historical, not automatically
valid under this schema. No automatic migration or runtime validator is
implemented by this document.

## Versioned structure

Use YAML mappings and arrays, unique keys and UTF-8. Quote timestamps and revision
identifiers. Required top-level fields are `schema`, `ticket`, `history`,
`resets` and `reviews`. Empty collections are `[]`; do not invent a verdict
or placeholder review. The schema discriminator is `sdlc-audit-record/v1`.

Each review has a unique `id`, timestamp, profile, gate, cycle, round, method, reviewer,
recorder, candidate, supervision status, verdict, criteria, evidence, findings,
limitations, prior-finding dispositions and provisional-condition evidence.
Gates are `definition`, `test-code`, `delivery-code`; methods are `self`
and `independent`; verdicts are `PASS`, `FAIL`, `PROVISIONAL PASS`.
Profiles are `SDLC`, `paired`, `emergency`, `spike` or `non-SDLC`.
For managed gates, cycle and round are positive integers and rounds range from
1 to 5 per gate/cycle. A bounded standalone or spike review records its profile
and authorized budget; gate and cycle may be null, and round counts assessments
within that invocation. This does not invent managed gates. An unticketed review
uses ticket: null and an explicit scope reference; keep one canonical record.
Independent method requires truthful fresh-context separation. Recorder and
reviewer identities must describe actual actors, not invented provider identities.

`history` entries identify a path/reference, original method where known, and
any uncertainty. Unticketed records also require `scope`; standalone/spike
records require `budget` with the authorizing instruction and assessment limit. Criteria, evidence and limitations are arrays of strings or
immutable evidence references. Prior-finding dispositions identify the finding
ID, disposition, reason and evidence. Provisional conditions identify the exact
mechanical correction, verification evidence and whether it was satisfied;
unverified conditions cannot support effective PASS.

Candidate identity includes the commit and any reviewed uncommitted delta's
immutable digest or evidence reference. A commit alone cannot identify dirty
inputs. Findings have stable IDs, severity (blocking or advisory), location,
requirement, evidence and required correction. Subsequent entries record each
prior finding's disposition (open, resolved or withdrawn), reason and evidence;
never edit an old FAIL into PASS or erase a disputed finding.

Example shape (illustrative assessment, not evidence for a real ticket):

```yaml
schema: sdlc-audit-record/v1
ticket: W123
history: []
resets: []
reviews:
  - id: definition-c1-r1
    timestamp: "2026-10-29T12:00:00Z"
    profile: SDLC
    gate: definition
    cycle: 1
    round: 1
    method: self
    reviewer: current author session
    recorder: current author session
    separation: same context; not independent
    candidate:
      commit: "<actual commit>"
      delta: null # null only when reviewed inputs match that commit
    live_operator_supervision:
      requester_confirmed: true
      reviewer_confirmed: true
    verdict: FAIL
    criteria: ["ISSUES.md at <actual revision>"]
    evidence: ["spec.org at <candidate reference>"]
    findings:
      - id: F1
        severity: blocking
        location: "spec.org: User tests"
        requirement: "TESTING.md: meaningful UT"
        evidence: "No user test is defined in the reviewed candidate."
        required_correction: "Define a meaningful user acceptance test."
    prior_finding_dispositions: []
    provisional_conditions: []
    limitations: []
```

A reset entry requires `id`, `timestamp`, `gate`, `from_cycle`, `to_cycle`,
`operator`, `instruction`, `authority_reference` and `reason`. Preserve the
actual instruction or a faithful attributed record; do not invent a quote.
An explicit reset increments that gate's cycle by one; its next review is round
one. A reset alone has no verdict and leaves unresolved findings outstanding.
Cycle 1 requires no reset. Later cycles require the matching authorization.
A reviewer or transport change is not a reset event.

## Deterministic consumption

Derive current state from the applicable gate/cycle's latest assessment; do not
maintain a second mutable summary verdict. Check schema, required fields, unique
IDs, contiguous rounds, authorized cycle transitions, candidate identity,
method/separation, current evidence and unresolved blocking findings. Malformed,
missing, ambiguous or stale records never satisfy a gate. A reset invalidates
any assumption that an earlier cycle supplies the new cycle's verdict.

PASS must have no unresolved blocking findings and adequate evidence for its
scope. PROVISIONAL PASS is eligible only with both live-supervision confirmations
and verified, recorded mechanical conditions under REVIEW; otherwise its effective
verdict is FAIL. Preserve the original verdict. A newer FAIL cannot be bypassed
by selecting an older PASS. A record is evidence, not operator approval or
proof that a test ran. Future automation must validate the evidence and these
contracts rather than trusting a status string alone.

Historical inputs and existing review entries remain append-only. Corrections
are new attributed entries or linked history notes, never silent rewrites.
Use the repository's existing serialization/locking mechanism for a single
writer; re-read before updating and preserve concurrent changes. Do not let
reviewers and callers independently write competing copies.
