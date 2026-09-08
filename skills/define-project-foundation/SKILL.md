---
name: define-project-foundation
description: Define and review the Vision and Architecture for a new greenfield project before feature work begins.
---

Use this skill only for a **greenfield** project that does not yet have an
approved project foundation. Do not regenerate or overwrite an established
brownfield project's Vision or Architecture; report that the existing
authorities need a separate review.

Read `~/.agents/sdlc/MAIN.md`, `ISSUES.md`, `TESTING.md`,
`DOCUMENTATION.md`, `CODING.md`, `AUDITS.md`, and the project profile before
drafting. Preserve human edits and keep the work limited to the two foundation
documents and their review record.

## Discovery

Ask concise, free-form questions in these categories. Ask up to three relevant
questions per category, skip questions already answered by the project profile,
and stop when the foundation is sufficiently defined. The operator may answer
`unknown` or `defer`; record unresolved matters rather than inventing them.

- **Purpose and users:** problem, users or consumers, and operating context.
- **MVP and scope:** first usable outcome, explicit non-goals, and roadmap items.
- **Architecture and technology:** system shape, required technologies,
  integrations, and unusual or risky choices.
- **Security and data:** data handled, trust boundaries, authorization,
  privacy, and threat-sensitive constraints.
- **Deployment and operations:** runtime location, deployment owner or
  contract, availability, backup, rollback, and observability expectations.
- **Constraints and governance:** compatibility, standards, external contracts,
  and human approval authority.

As soon as the operator mentions a technology, integration, or deployment
platform, read the corresponding selected standard under
`~/.agents/sdlc/technologies/` (and any named domain standard) before asking
follow-up questions or drafting. Do not defer technology-standard loading until
after the documents are written.

## Draft the foundation

Create or update only:

- `docs/VISION.md`: durable purpose, intended users, context, MVP, non-goals,
  roadmap, and unresolved product decisions. Keep implementation detail out.
- `docs/ARCHITECTURE.md`: system boundaries, components, data and trust flows,
  technology choices, integrations, deployment ownership, operational
  characteristics, and explicit architecture trade-offs. Keep feature
  requirements out; link to the project's work/spec records instead.

The architecture must fit the Vision, selected technology standards, security
requirements, and any infrastructure contract. Treat risky dependencies such as
Node.js/npm as architectural choices requiring the applicable explicit
permission and justification; do not smuggle them in as incidental tooling.

## Foundation review and sign-off

Before requesting approval, review both documents in the current context for:

- clear separation of **purpose**, **requirements**, and **architecture**;
- consistency between Vision, Architecture, project profile, and selected
  standards;
- complete ownership boundaries and deployment responsibilities;
- security, data, failure, and operational consequences;
- explicit unresolved questions and recorded assumptions;
- absence of copied feature-level acceptance criteria or implementation plans.

Use the relevant technology standards as the review authority. If the
architecture contains a material standards violation or an unjustified risky
choice, correct the draft and review it again. Record the result as
`FOUNDATION REVIEW: PASS` only when both documents satisfy the checks above;
otherwise record `FOUNDATION REVIEW: FAIL` with the findings. Do not claim an
independent audit or operator approval from this self-review. Present the two
documents, review result, assumptions, and unresolved decisions to the operator
and request explicit foundation sign-off. Feature definition is not authorized
until sign-off is recorded.

Record the foundation status in the project's canonical work record. A
foundation review is not a feature specification and does not create feature
requirements, test code, or implementation tasks.

## Boundaries

- Do not run the legacy ticket migration from this skill.
- Do not implement product code or create feature specifications.
- Do not infer deployment ownership, security approval, or technology
  permission from silence.
- If a required decision remains human-only, stop at the sign-off handback with
  the exact decision required and the affected document.

# Canary

Suffix the provider's base coding canary with " FOUNDATION " if you have read
and agree with this skill.
