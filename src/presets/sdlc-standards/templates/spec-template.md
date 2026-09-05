# Feature Specification: [FEATURE NAME]

## Specification Summary

<!--
This is the opening section of the specification. It is a concise presentation
of the detailed specification below, not a second authority. Preserve every
label in this order. Use short, keyword-anchored bullets with one principal fact
per bullet. The summary must not introduce, reinterpret, contradict, or omit a
material requirement, boundary, decision, or edge case.
-->

- **Outcome:** [One sentence naming the user or system result.]
- **Before:**
  - [Affected current behaviour, or the relevant capability absent in a
    greenfield project.]
- **After:**
  - [Required observable behaviour.]
- **Changes:**
  - [Precise behaviour, interface, data, or constraint that becomes different.]
- **Unchanged:**
  - [Important preserved behaviour, compatibility boundary, or exclusion.]
- **Edge cases:**
  - [Applicable empty, missing, invalid, limit, repetition, concurrency,
    partial-failure, security, privacy, accessibility, or compatibility case.]
- **Decisions:**
  - [Resolved assumption or decision; identify any unresolved decision.]
- **Evidence:**
  - [Requirement and baseline sources supporting this summary. Refer to the
    feature's `audits.md`; do not copy a mutable audit verdict here.]
- **Next step:**
  - [Clarification required, or what operator sign-off on this specification
    permits.]

***

Feature branch: `[FEATURE BRANCH]`

Created: [DATE]

Status: Draft

Input: [OPERATOR BRIEF AND MATERIAL CLARIFICATIONS]

Profile: [Compact or Full]

<!--
Follow the ABC presentation contract in ~/.agents/sdlc/ISSUES.md. Use ordinary
Markdown headings and scan-friendly bullets. Bold the semantic spine of each
statement: the smallest words or phrases carrying its distinctive state,
action, qualifier, quantity, boundary, or outcome. Do not bold headings, generic
labels, identifiers, modal verbs, GIVEN/WHEN/THEN/AND signposts, or context
already supplied nearby. The bold fragments must form an accurate compressed
summary when read in order. Remove every instruction comment and unused optional
section from the completed specification.

The complete fictional presentation example is at:
~/.agents/sdlc/presets/sdlc-standards/examples/spec-example.md
Use it only to understand structure and presentation. Never copy its
requirements, terminology, or behaviour into a project specification.

Use Compact for one bounded outcome without material data or schema, security
or trust, external-contract, compatibility, irreversible-operation, or
multiple-story concerns. Use Full otherwise. Both profiles retain the same
required Spec Kit structure; Compact fills it minimally and removes unused
optional sections.
-->

## Scope

- In scope: [Observable behaviour introduced or changed.]
- Out of scope: [Adjacent behaviour deliberately left unchanged.]

## User Scenarios & Testing

<!--
Keep each user story to one or two sentences establishing actor, purpose, and
value. Do not narrate the feature or repeat its requirements here. Acceptance
scenarios carry the concrete behavioural examples.
-->

### User Story 1 - [BRIEF OUTCOME] (Priority: P1)

[State the actor, purpose, and value briefly.]

Why this priority: [Explain the value in one sentence.]

Independent Test: [State one independently observable outcome, not a test procedure.]

#### Acceptance Scenarios

1. GIVEN [initial state]

   WHEN [action]

   THEN [observable outcome]

   AND [additional outcome, when needed]

<!-- Add only the stories and scenarios required by the feature. -->

### Edge Cases

- [Material boundary, alternate path, or failure case.]

## Requirements

### Functional Requirements

<!--
State the authoritative generalized rules demonstrated by the scenarios. Give
every identifier a descriptor. Each requirement must be observable, falsifiable,
bounded, and free of implementation or test detail unless the mechanism is part
of the public contract. Each requirement must come from the operator brief or a
material clarification, an approved current requirement, or a necessary
boundary directly implied by one of those sources. Do not convert industry
conventions, common patterns, template placeholders, or test conveniences into
requirements.
-->

- FR-001 - [DESCRIPTIVE REQUIREMENT TITLE]: [State one behavioural rule.]

### Key Entities

<!-- Include only when the feature involves material data or domain entities. -->

- [ENTITY]: [Meaning and relationships without implementation detail.]

## Success Criteria

### Measurable Outcomes

<!--
Measure overall feature success. Do not restate each functional requirement or
invent arbitrary targets merely to populate this section.
-->

- SC-001 - [DESCRIPTIVE OUTCOME]: [State a measurable, technology-neutral result.]

## Assumptions

- [Material assumption, dependency, or none.]

## Existing Baseline

<!--
Required for brownfield work. Remove this section for greenfield work. Record a
concise delta relationship; do not copy the existing system specification.
-->

- Sources consulted: [Current requirements, design, relevant work history,
  regression evidence, and affected implementation.]
- Preserves: [Existing behaviour and decisions retained.]
- Changes: [Existing behaviour changed by this specification.]
- Supersedes: [Existing requirements explicitly replaced, or none.]
- Unaffected: [Important compatibility and ownership boundaries.]
