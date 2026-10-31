# Defect raising and ticket storage

Status: storage and numbering agreed; the skill source is now in
`skills/defect-raise/SKILL.md`. Deployment remains a separate operator action.

Use a flat `tickets/` root. Work items use W, defects D and spikes S, sharing
one numeric namespace: W123, D124, S125, W126. Allocate above the highest ever
reserved number across all types, not from independent counters. All types
share primary's `docs/work.org` register and lifecycle.

- `tickets/W123-description/spec.org`
- `tickets/D124-description/defect.org`
- `tickets/S125-description/spike.org`

A defect's later fix specification stays in its D-prefixed directory under the
same identity. Diagnosis and fix design remain unnecessary for initial reporting.

For initialized v3 projects, `sdlc-init` renames `specs/` to `tickets/` and
replaces all `specs/` references in all tracked `.org`, `.md` and `.yaml` files, including those
inside the renamed directory. It does not reclassify or
renumber existing identities. Normal work stops until migration is performed.

The earlier specs/defects proposal is superseded by this flat layout. No separate
defect tree or numeric namespace remains in the proposal.
