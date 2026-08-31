---
name: audit-spec
description: Challenge a specification's requirements and acceptance criteria. Advisory only; no file changes.
metadata:
  preferred_provider: openai-codex
  preferred_model: gpt-5.6-luna
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`, and
`~/.agents/sdlc/ISSUES.md` in full. Read the project constitution. When its
`Specification Baseline` classifies the project as brownfield, read the named
requirement and design authorities relevant to the active change, consult the
relevant historical work records and comments, inspect the maintained
regression tests and their requirement or ticket traceability, and inspect the
affected implementation. Review the active specification as a delta against
that baseline without modifying files.

Do not report a finding solely because an unchanged approved requirement,
design mechanism, fixture, oracle, or test procedure is not copied into the
delta specification. Require a baseline citation with an identifier and
descriptor where unchanged behaviour materially bounds the delta. Report a
finding when the delta contradicts its baseline, changes behaviour without
saying so, relies on an unresolved authority conflict, or lacks an observable
boundary necessary to distinguish compliance.

Report a source-coverage finding when the specification's brownfield context
pass omits a material authority, historical decision, actively protected
regression behaviour, traceability link, or implementation fact. Treat tests
and code as evidence rather than requirement authority. Do not require source
detail to be copied into the specification when a concise citation and baseline
relationship establish that it was considered.

Check:

- the user or system outcome is explicit;
- every requirement is observable, falsifiable, bounded, and internally
  consistent;
- acceptance criteria describe behaviour rather than test procedures;
- normal, alternate, error, boundary, permission, repetition, migration,
  security, privacy, and accessibility cases are covered where relevant;
- compound requirements expose every independently failing condition;
- undefined product, architecture, data, or access decisions are identified;
  and
- every identifier in the report has an adjacent descriptor.

For each finding, state the affected requirement with its descriptor, the gap
or contradiction, the consequence, and a concrete proposed correction. Do not
rewrite the specification.

Run in a fresh context that did not author the specification. Follow the exact
finding classifications, verdict format, independence rules, and changed-
artefact rule in `~/.agents/sdlc/AUDITS.md`, using `AUDIT: audit-spec`.
