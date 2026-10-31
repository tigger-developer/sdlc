# Defect raising and ticket storage

Status: storage layout agreed; skill remains a draft outside installed skills.

## Storage layout

Keep `specs/`. New defect reports go in
`specs/defects/<ticket>-description/defect.org`. Existing feature and spike
locations remain unchanged. All tickets share the existing number namespace
and primary's `docs/work.org` register and lifecycle authority.

```text
specs/
  W123-feature-description/spec.org
  W124-spike-description/spike.org
  defects/
    W125-defect-description/
      defect.org
```

Existing projects retain their number format, including schemes without a W
prefix. The defects subdirectory segregates reports on disk, not in numbering
or lifecycle. Allocation and duplicate checks must cover the shared register
and nested defect directories, not just immediate children of `specs/`.

A newly raised defect needs its report and register entry. If correction later
requires `spec.org`, add it to the same defect directory with the same ticket
number, linking the report and original requirements. Validation and audit
records are added there only when required by the subsequent work. Diagnosis
and fix design are not prerequisites for registration.

## Scope and history

The earlier proposal to introduce `tickets/` is withdrawn. No directory-root
migration or relocation of existing tickets is required. Reuse an existing
defect's current location when adding evidence instead of moving it or creating
a duplicate. The draft skill does not activate a new delivery route or weaken
implementation approval requirements.
