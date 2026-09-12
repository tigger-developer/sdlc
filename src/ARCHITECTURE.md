# Architecture Standards

These principles apply to applications, libraries, command-line tools, plugins,
websites and infrastructure software. They guide the project's architecture and
each change's solution design; they do not replace either document or authorize
a redesign. Apply them proportionately to the authorized change.

## First principle: KISS

**Keep It Simple, Stupid.** Choose the simplest maintainable design that meets
the actual requirements and mandatory standards. Simplicity means understandable
responsibilities and behaviour, not the fewest lines or the least design thought.

- Complexity must earn its place through a present requirement or demonstrated
  constraint, not hypothetical future needs.
- Scope expansion is a red flag. Explain why the authorized outcome cannot
  reasonably be achieved within scope and obtain explicit operator approval
  before expanding it; justification alone is not authorization.
- Reuse established components and dependencies where they fit. Do not create
  a framework, service, configuration layer or extension point without a need.
- Do not duplicate an operation for every case merely to avoid designing its
  shared behaviour. Equally, similar-looking code alone does not establish a
  shared responsibility.
- Apply proportionality to production code, tests, configuration and supporting
  documentation. Simplicity never excuses omitted requirements or safeguards.
- Scale design scrutiny to the consequences of failure and difficulty of
  reversal, not file or line count. Small changes to persisted formats, public
  contracts or authorization may require substantial care.
- Among designs meeting the requirements and standards, prefer fewer new
  concepts, then fewer independently managed moving parts, not fewer lines.

## Responsibilities, reuse and boundaries

- Give each component a coherent purpose and explicit ownership. Separate
  responsibilities where they change independently or require different trust
  or lifecycle boundaries; do not split them merely to add layers.
- Identify genuinely shared behaviour and what varies between cases. Prefer
  straightforward functions with explicit parameters or data when sufficient;
  use classes or other abstractions only when their responsibilities justify them.
- Keep interfaces explicit and dependencies limited and deliberate. Avoid hidden
  global state, duplicated sources of truth and unnecessary coupling.
- Extend the established project architecture. Identify any necessary departure,
  its reason and its consequences for explicit operator approval.
- Respect external ownership. An integration gap does not authorize changes to
  another project or a compensating subsystem inside this one.

For example, processing a known collection of files normally needs one shared
operation accepting a path and per-file data, not a function per filename or a
general-purpose file-processing framework. Distinct behaviour may justify a
separate operation; the design should explain that distinction.

## State, failure and security

- Identify who owns state, its lifetime and source of truth, and the boundaries
  at which it is validated, persisted or transferred.
- When changing existing contracts or stores, describe the transition and
  intended end state, compatibility and recovery constraints, and the retirement
  condition for temporary paths. Planned coexistence is legitimate; unexplained
  parallel old/new paths are not. Apply the compatibility rules in
  `~/.agents/sdlc/CODING.md` and relevant technology standards.
- Define material failure outcomes and recovery. Address atomicity, repetition,
  concurrency and resource bounds where the operation requires them, rather
  than adding every reliability mechanism by default.
- Assess security in every design: trust boundaries, privileges, untrusted input,
  sensitive data and exposure. Record required safeguards or a supported
  no-change rationale. Apply `~/.agents/sdlc/SECURITY.md` and relevant technology
  standards.
- Consider performance, operability, portability and other qualities only where
  they affect this change. Record the design consequence, not generic aspirations.

## Definition must resolve the important decisions

The solution design must let a delivery agent implement the change without
inventing its significant structure. Describe the relevant responsibilities,
component reuse, shared operations and variations, interfaces, state ownership,
and essential algorithms or mechanisms. Link stable project architecture;
describe only the delta and its dependencies.

Use a short flow, after-state example, interface contract or pseudocode where it
clarifies a decision. Do not prescribe every function, filename or statement.
Identify routine details deliberately left open; do not leave significant
design choices disguised as "implementation details".

Record a short rationale beside each non-obvious decision, especially deliberate
complexity. If a compromise is explicitly accepted, record its consequence and
revisit trigger there; this does not waive mandatory standards or grant approval.
Keep these records in the existing specification or project architecture, not a
parallel decision system. Maintain affected authorities under DOCUMENTATION.md.

Require enough detail to explain the chosen solution, not enough machinery to
populate every heading. Omit irrelevant solution-design subsections; retain
mandatory security impact. Do not invent decisions, abstractions or tests to
satisfy the template. A small change may need only a few sentences.

For each important design decision, identify appropriate verification:

- Link traced behavioural tests where the decision has observable consequences.
- State an implementation-review check for structural decisions that behavioural
  tests cannot establish, such as ownership or use of a shared component.
- Do not manufacture product requirements or source-text tests merely to verify
  an internal structure. Passing behavioural tests alone does not prove design
  conformance.
- Establish during definition that the proposed tests can observe the required
  behaviour. Identify necessary control of volatile dependencies under CODING.md
  and TESTING.md; do not invent a mock framework or undocumented third-party
  simulation. Where only live or human evidence is credible, define a bounded
  OT or UT rather than requiring every check to run offline.

## Delivery and review

Implement the approved solution design. Resolve only routine details it leaves
unspecified, within its constraints. If an approved decision cannot reasonably
be followed, explain the conflict and obtain operator approval before departing
from it. Convenience or an auditor's preference does not authorize redesign.
Paired and emergency work follow their recorded operator-authorized scope and
existing approval rules; this standard adds no advance gate to either route.

Definition review checks that the design is proportionate, coherent and
sufficiently resolved for handoff. Implementation review checks adherence to
the approved decisions as well as behaviour and standards. A blocking finding
must identify a violated requirement, approved decision or applicable standard,
or a concrete correctness or security defect. Optional improvements and
hypothetical future needs do not justify FAIL or scope expansion.

Definition review must explicitly assess:

- **Scope:** Does the design introduce behaviour, infrastructure or obligations
  beyond the operator-authorized change?
- **Over-engineering:** Could a materially simpler solution satisfy the same
  requirements and mandatory standards? If so, what justifies the added complexity?
- **Template inflation:** Were components or mechanisms introduced to fill
  sections rather than meet an actual need?

Identify the concrete unresolved decision or unnecessary mechanism. Length,
heading count and an auditor's aesthetic preference are not evidence of a
defect. Implementation review applies proportionality to the implementation
delta without reopening approved design choices.

# Canary

Suffix the canary string with "ARCH "
