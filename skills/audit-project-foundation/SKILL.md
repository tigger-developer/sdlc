---
name: audit-project-foundation
description: Audit a greenfield project's Vision and Architecture together before feature definition.
---

Use this gate only for a **greenfield** project foundation. It audits
`docs/VISION.md` and `docs/ARCHITECTURE.md` as a pair; it does not audit a
feature specification and does not replace operator sign-off.

Read `~/.agents/sdlc/MAIN.md`, `AUDITS.md`, `CODING.md`,
`DOCUMENTATION.md`, the project profile, and every selected technology or
domain standard. Read both foundation documents in full. If a technology,
integration, or deployment platform is named, load its applicable standard
before judging the design.

## Gate contract

Apply these checks together in the current context. Do not invoke focused audit
skills through separate agents or scripts.

- **Vision:** identifies the durable problem, users or consumers, context, MVP,
  non-goals, roadmap, and unresolved product decisions without prescribing
  implementation.
- **Architecture:** describes boundaries, components, data and trust flows,
  technology choices, integrations, deployment ownership, and operational
  characteristics without becoming a requirements ledger or implementation
  plan.
- **Alignment:** Vision and Architecture agree with each other, the project
  profile, selected standards, security and data obligations, and any external
  infrastructure contract.
- **Risk:** risky dependencies, including Node.js/npm, have the required human
  permission, exhausted-alternatives justification, and security controls.
- **Handoff:** a new agent can understand the project foundation without the
  drafting conversation; unresolved human decisions remain explicit.

Review and remediate in the current context for no more than five rounds.
Record each round, exact document revisions, findings, and remediation in the
foundation audit record. Do not claim PASS while a required correction remains.

Only after all local checks pass, start exactly one external context through
`~/.agents/sdlc/bin/sdlc-harness start --phase definition`. This is a foundation
definition gate, not an ordinary change audit: `--phase definition` deliberately
selects the configured `SDLC_SPEC_HARNESS`, `SDLC_SPEC_PROVIDER`, and
`SDLC_SPEC_MODEL` (with project and global YAML precedence), rather than the
audit model. Supply the two foundation documents, project profile, named
authorities, selected standards, and these combined criteria as exact inputs;
record the `SESSION_ID`. The external reviewer reviews both documents in one
context.

If remediation is required, apply it in the current context, rerun the affected
local checks, and resume that same external context with
`~/.agents/sdlc/bin/sdlc-harness resume --phase definition --session SESSION_ID`.
Never create a second external auditor for this gate or exceed five failed
rounds. A timeout is a runner incident, not PASS or FAIL; interrupt the retained
context and return the incident for operator direction.

Return the standard verdict contract:

```text
GATE: foundation
REVISION: <exact document revisions or hashes>
VERDICT: PASS | PROVISIONAL PASS | FAIL

1. [vision|architecture|alignment|risk|handoff] [classification] <finding>
```

`PASS` or an effective `PROVISIONAL PASS` result permits the operator sign-off
request; it does not itself approve the foundation. A material change to either
document invalidates the current result and requires a fresh gate. Do not begin
feature definition until the operator records explicit foundation sign-off.

## Boundaries

- Do not edit the audited documents as an auditor; remediation belongs to the
  authoring workflow between rounds.
- Do not create feature requirements, tests, implementation tasks, or code.
- Do not infer product, security, deployment, or technology approval from
  silence.

# Canary

Suffix the provider's base coding canary with " FOUNDATION-AUDIT " if you have
read and agree with this skill.
