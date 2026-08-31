---
name: audit-design
description: Challenge a technical design for traceability, boundaries, trade-offs, risk, and operability. Advisory only; no file changes.
metadata:
  preferred_provider: openai-codex
  preferred_model: gpt-5.6-luna
---

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md`, `~/.agents/sdlc/AUDITS.md`,
`~/.agents/sdlc/CODING.md`,
`~/.agents/sdlc/TESTING.md`, and `~/.agents/sdlc/DOCUMENTATION.md` in full.
Read the active specification, design or plan, project constitution, external
integration contracts, and only the technology standards selected by the
constitution. For a brownfield project, also read the relevant current design
authorities, historical work records and comments, maintained regression tests
and their requirement or ticket traceability, and affected implementation.
Review without modifying files.

Use relevant quality scenarios to challenge:

- traceability from every material requirement and constraint to a design
  decision;
- components, interfaces, data, state, trust, deployment, and ownership
  boundaries;
- normal operation, failures, recovery, concurrency, resource limits, and
  degraded dependencies;
- security, privacy, authorization, accessibility, and widened access;
- compatibility, migration, rollback, reversibility, and historical contract
  retirement;
- dependencies, complexity, alternatives, assumptions, and explicit
  trade-offs among relevant quality attributes;
- operability, observability, diagnosability, and justified test architecture;
  and
- unresolved decisions that would force implementation to invent behaviour.

Report a source-coverage finding when the design omits a material inherited
constraint, historical decision, actively protected regression behaviour,
traceability link, or implementation fact. Require the design to distinguish
what it inherits, changes, supersedes, and leaves unaffected. Historical work
records provide rationale and lineage but are not automatically current design
authority; tests and code are implementation evidence rather than authority to
change signed-off requirements or design.

For each finding, identify the affected requirement or decision with its
descriptor, the missing or unsafe design reasoning, the consequence, and a
concrete correction. Do not redesign the artefact.

Run in a fresh context that did not author the design. Follow the exact finding
classifications, verdict format, independence rules, and changed-artefact rule
in `~/.agents/sdlc/AUDITS.md`, using `AUDIT: audit-design`.
