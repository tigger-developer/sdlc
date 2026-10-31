# Defect raising and ticket storage

Status: proposal for operator review. The skill is deliberately outside `skills/`
and is not deployed. No project migration or new workflow routing is activated.

## Recommendation

Treat features, defects and spikes as tickets in one numbered namespace.
`docs/work.org` remains the canonical index and lifecycle authority on primary.
A ticket directory holds the artefacts appropriate to its work:

```text
tickets/W123-example/
  defect.org       # defect report, when applicable
  spec.org         # agreed change/fix specification, when required
  spike.org        # exploratory brief, when applicable
  validation.org   # test execution evidence, when required
  audits.yaml      # the sole review record, when reviews occur
```

This is a menu, not a mandatory bundle. A newly raised defect normally needs
only `defect.org` and its register entry. A feature need not have a defect report;
a spike does not automatically require a product specification. Existing
projects retain their number format, including schemes without a W prefix.
Do not create separate bug and feature number sequences.

`specs/` is the current SDLC directory name. It describes only one kind of
artefact and becomes less useful as the collection includes reports and spikes.
`tickets/` would express the intended ownership more clearly. The filename
`defect.org` distinguishes a report from an approved fix specification without
inventing a separate ticket lifecycle.

## Proposed transition

Draft the skill against the current established root, using
`specs/<existing-number>-description/defect.org`. Decide separately whether to
adopt `tickets/` as the framework default and how existing projects opt in.
Do not move existing files, duplicate directories or break evidence links as
part of raising a defect. A root change needs a coordinated review of routing,
initialization, templates, validators, links and migration guidance.

The same ticket may acquire `spec.org` when correction is designed. It links the
report and existing requirements instead of replacing the original evidence or
allocating a second number. Diagnosis is not required before registration.

## Decision remaining

Whether to keep `specs/` as the storage name or undertake a separate migration
to `tickets/`. The proposed skill's reporting scope and shared ticket identity
do not depend on that decision. This draft does not select a new delivery route
or weaken existing implementation approval requirements.
