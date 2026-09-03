# Specification: [FEATURE NAME]

Feature branch: `[FEATURE BRANCH]`

Created: [DATE]

Status: Draft

Input: [OPERATOR REQUEST]

<!--
Follow the ABC presentation contract in ~/.agents/sdlc/ISSUES.md.
Use ordinary Markdown headings. Bold functional keywords and noun phrases within
the content, not headings, generic labels, or modal verbs. Remove every
instruction comment and unused optional section from the completed specification.
-->

## Outcome

[State the required user or system outcome once, using concise prose or bullets.]

## Scope

- In scope: [Observable behaviour introduced or changed.]
- Out of scope: [Adjacent behaviour deliberately left unchanged.]

## Existing baseline

<!--
Required for brownfield work. Remove this section for greenfield work.
Do not copy the existing system specification into this delta.
-->

- Sources consulted: [Current requirements, design, relevant work history,
  regression evidence, and affected implementation.]
- Preserves: [Existing behaviour and decisions retained.]
- Changes: [Existing behaviour changed by this specification.]
- Supersedes: [Existing requirements explicitly replaced, or none.]
- Unaffected: [Important compatibility and ownership boundaries.]

## Requirements

<!--
State each behavioural rule exactly once. Give every identifier a descriptor.
Each requirement must be observable, falsifiable, bounded, and free of test
procedure or implementation detail unless that mechanism is the public contract.

Presentation example only:

### FR-001 - Additional deployment domains

In addition to the **base URL**, a **Hugo website** may declare in its
**environment configuration** that it is deployed to **additional domains**.

Origin: [Operator request or cited baseline requirement with its descriptor.]
-->

### FR-001 - [DESCRIPTIVE REQUIREMENT TITLE]

[State one requirement. Bold its functional actors, artefacts, states,
configuration, boundaries, and outcomes.]

Origin: [Operator request or cited baseline requirement with its descriptor.]

## Boundaries and failure behaviour

- [State each material limit, alternate path, or observable failure not already
  expressed as a requirement.]
- [Include relevant security, privacy, access, migration, compatibility, and
  accessibility boundaries.]

## Assumptions and unresolved decisions

- Assumption: [Authorized material assumption, or none.]
- Decision required: [Unresolved decision that can change the outcome, or none.]

## Terms

<!-- Include only when the specification introduces or depends on ambiguous domain terms. -->

- [TERM]: [Concise project meaning.]
