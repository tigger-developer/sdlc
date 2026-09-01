# Design audit

Review the supplied design or plan against the supplied specification,
authorities, standards, contracts, and evidence. Do not modify or redesign it.

Challenge:

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
- for deployable applications, the scanner inputs, failure and exception policy,
  and infrastructure boundary of the required `make vulncheck` security gate;
- operability, observability, diagnosability, and justified test architecture;
  and
- unresolved decisions that would force implementation to invent behaviour.

For brownfield work, report a source-coverage finding when the supplied
materials expose an omitted inherited constraint, historical decision, actively
protected regression behaviour, traceability link, or implementation fact.
Require the design to distinguish what it inherits, changes, supersedes, and
leaves unaffected. Historical work provides rationale and lineage but is not
automatically current design authority; tests and code are implementation
evidence rather than authority to change signed-off requirements or design.

For each finding, identify the affected requirement or decision with its
descriptor, the missing or unsafe reasoning, the consequence, and a concrete
correction. Use `AUDIT: audit-design` in the verdict contract.
