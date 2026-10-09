---
title: Standalone Standards Audit
version: 4
last-updated: 2026-10-10
---

# Standalone standards audit

Use this route only when the selected project has no `.sdlc/project.yaml`, or
when the operator's `hands-off-spike` workflow selects `--agent-spike` in an
SDLC project. A spike review is deliberately outside SDLC delivery: the
presence of an SDLC profile does not make its workflow artefacts required, and
exploratory code receives the same code and documentation scrutiny.
The audit reviews the supplied original project files against `CODING.md`,
`DOCUMENTATION.md`, other applicable installed standards, and the project's
existing authorities. The full installed Markdown inventory is available, but
SDLC workflow documents do not become standalone requirements. It reports concrete, evidence-backed code and documentation
findings. The harness must reject a project or input under a system temporary
directory before calling a provider. The invocation project inventory shows
eligible paths, not their contents. Identify relevant files absent from the
evidence package; do not claim to have reviewed a listed file without reading
it. Report insufficient access or missing context instead of granting a full
standards PASS.

This review does not adopt the SDLC delivery process. Do not require a work
item, specification, audit gate or record, validation record, test colour,
traceability matrix, sign-off, or delivery closure. Do not report the absence
of SDLC process artefacts as a finding. Requirements in the supplied standards
that explicitly address SDLC-managed delivery apply only when that workflow
has been elected. Apply other rules to the extent relevant to the supplied
code and documentation and the project's established conventions.

Assess correctness, security, maintainability, documentation accuracy, and
applicable standards. Name the exact file and concrete evidence for each
finding. Distinguish a verified defect from an optional improvement or an
unknown project decision. Do not edit the reviewed files.

Return one compact result:

```text
AUDIT: standalone
REVISION: <reviewed revision or working-tree state>
VERDICT: PASS | PROVISIONAL PASS | FAIL
<findings, if any>
```

This result is review evidence, not an SDLC gate verdict or operator approval.

## Document history

- Version 4 (2026-10-10): Apply the route to `--agent-spike` reviews of hands-off spikes in SDLC projects.
- Version 3 (2026-10-06): Distinguish full standards availability from standalone applicability.
- Version 2 (2026-10-06): Distinguish project inventory from reviewed contents.
- Version 1 (2026-10-06): Define standards-only review outside SDLC delivery.
