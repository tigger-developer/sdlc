# Specification audit

Review the supplied specification as a behavioural delta against the supplied
authorities and evidence. Do not modify or rewrite it.

For brownfield work, report a source-coverage finding when the supplied
materials show that the specification omitted a material requirement or design
authority, historical decision, actively protected regression behaviour,
traceability link, or implementation fact. Tests and code are evidence, not
requirement authority. Do not demand that unchanged baseline detail, design
mechanisms, fixtures, or test procedures be copied into a delta specification;
require a concise citation and relationship where that baseline materially
bounds the change.

Check that:

- the user or system outcome is explicit;
- every section, paragraph, and bullet satisfies the Accurate, Brief, and Clear
  presentation contract in `ISSUES.md`;
- each Spec Kit section performs its distinct job: user stories briefly establish
  context and value, acceptance scenarios carry concrete examples, functional
  requirements state authoritative generalized rules, and success criteria
  measure overall feature success;
- repetition is limited to what connects those views, without verbose narrative
  restating detailed behaviour;
- ordinary Markdown headings provide structure while each statement's bold
  semantic spine forms an accurate compressed summary of its distinctive state,
  action, qualifier, quantity, boundary, and outcome;
- paragraphs containing several facts are split into scan-friendly bullets, and
  `GIVEN`, `WHEN`, `THEN`, and `AND` are unbolded capitals on separate lines;
- every requirement is observable, falsifiable, bounded, and internally
  consistent;
- requirements describe behaviour rather than test procedures;
- normal, alternate, error, boundary, permission, repetition, migration,
  security, privacy, and accessibility cases are covered where relevant;
- compound requirements expose every independently failing condition;
- undefined product, architecture, data, or access decisions are identified;
  and
- every identifier in the report has an adjacent descriptor.

For each finding, state the affected requirement with its descriptor, the gap
or contradiction, the consequence, and a concrete proposed correction. Use
`AUDIT: audit-spec` in the verdict contract.
