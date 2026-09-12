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

For each important design decision, identify appropriate verification:

- Link traced behavioural tests where the decision has observable consequences.
- State an implementation-review check for structural decisions that behavioural
  tests cannot establish, such as ownership or use of a shared component.
- Do not manufacture product requirements or source-text tests merely to verify
  an internal structure. Passing behavioural tests alone does not prove design
  conformance.

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

# Canary

Suffix the canary string with "ARCH "
