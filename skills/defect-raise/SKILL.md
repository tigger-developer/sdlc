---
name: defect-raise
description: Capture and register a reported defect with observed and expected behaviour, available evidence and impact, without requiring diagnosis or designing or implementing its fix.
---

# Raise a defect

## Initialization prerequisite

Before creating or editing any ticket or candidate, require a readable
`.sdlc/project.yaml` declaring version 3 and no project-root `specs/`
directory. If either condition fails, stop this skill and tell the operator to
run `sdlc-init` from the project root. Do not allocate an ID, edit a record,
start a review or write code until initialization/migration succeeds. Do not
work around the check, migrate by hand or infer permission to invoke the
initializer. Recheck the profile and layout before resuming.

## Purpose and authority

Produce a useful, traceable defect ticket. Raising a defect does not invoke
`define-change`, approve delivery, start an audit or authorize implementation.
A report may be actionable for investigation before its cause, reproduction or
correct solution is known. Do not turn reporting into a design gate.

Follow the session's project routing and `READING.md`; reuse retained context.
For an initialized SDLC project, read the project's requirement/work authorities,
`ISSUES.md`'s bug and identifier rules, the primary-branch ticket reservation
contract in `DELIVERY.md`, and `ORGMODE.md` for the record. Apply only ticket
registration here, not delivery admission or the full specification template.
The initialization prerequisite above applies to this skill. Do not create a
forge issue without explicit authority.

## Capture the report

Start from the operator's evidence. Ask only questions needed to identify the
problem or make the report useful; record other missing facts as unknown.
Include, where known:

- A specific title and short description of the failed user outcome.
- Observed behaviour and expected behaviour, separately. Cite the requirement,
  documented contract or operator expectation supporting the latter.
- Reproduction steps, inputs and preconditions; frequency or intermittency.
- Affected environment, version or revision, and last known working version.
- Evidence references, such as exact errors, logs or screenshots, with provenance.
- Impact, affected users or operations, and any known workaround.
- Related requirements, tickets and possible duplicates.
- Unknowns and the next useful investigation question.

Distinguish reported observations, independently reproduced behaviour and
hypotheses. Do not claim reproduction or a cause without evidence. A credible
report without reproduction steps is still recordable. If the expected behaviour
is disputed or undocumented, preserve that uncertainty instead of inventing a
requirement, rejecting the report or converting it silently into a feature.
Record operator-assigned priority as such. Keep impact evidence separate from
any proposed severity; do not invent release commitments.

Describe how a later check could distinguish the observed failure from the
expected behaviour when that is already clear. Formal acceptance criteria,
AT/RT/UT/OT definitions, root-cause analysis and solution design are not required
to raise the defect. Do not run intrusive reproductions, change environments,
write regression tests or expand into diagnosis merely to fill the report.

## Register once

Search the authorized register and directly related records for duplicates.
For a confirmed duplicate, update the existing ticket with new evidence and
provenance; do not allocate another number. Record a possible relationship
without merging distinct reports merely because their symptoms resemble each
other. Preserve existing operator edits and evidence.

For a new SDLC ticket, synchronize primary and reserve an ID using DELIVERY's
collision-prevention protocol. Use the same namespace as features and spikes,
using D and one more than the highest number reserved across W, D, S and
legacy tickets. Never use a defect-only counter. Do not create a feature
branch to raise a defect or allocate from a branch copy of the register.

Create `tickets/DNNN-description/defect.org`. All ticket types live directly
under `tickets/`; do not create type subdirectories. If root `specs/` exists,
require `sdlc-init` before normal ticket work. Existing IDs are preserved by
migration; do not renumber an existing defect when adding evidence.
Use a descriptive Org title and concise sections suited to the evidence above.
Do not copy the delivery specification template or add empty ceremonial sections.

Add the ticket to primary's `docs/work.org` with the `defect` tag, the existing
initial lifecycle state and a link to `defect.org`. The register owns lifecycle;
the report owns the defect evidence. Do not duplicate the report in the index.
Commit and synchronize only the authorized ticket changes under the project's
existing policy, preserving unrelated work. Report a publication failure rather
than claiming registration succeeded remotely.

## Handoff

Return the ID with descriptor, report location, whether it was newly registered
or added to an existing ticket, and the material unknowns. Registration means
reported, not confirmed, diagnosed, delivery-approved or corrected.

Further diagnosis or implementation requires its own instruction. If a fix
later needs `spec.org`, add it to the same ticket directory and link the defect
report and original requirement; do not create another ticket merely to change
phase. Preserve the original observations and their provenance. If a genuinely
separate change is needed, link it explicitly rather than silently splitting
identity. Any later requested audit uses canonical `audits.yaml`; raising a
defect alone requires neither an audit nor an audit record.
