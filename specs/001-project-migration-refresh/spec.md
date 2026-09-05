# Specification Summary

- **Outcome:** SDLC v1 projects can migrate during initialization, and SDLC v2 projects can refresh their local skill adapters.
- **Before:** Migration required a separate manual invocation, used the specification harness in the pending implementation, and an unchanged preset could leave stale project-local skills.
- **After:** Initialization offers migration through the configured audit harness; project update always recomposes project-local adapters against canonical global instructions.
- **Changes:** Add the migration offer, audit execution configuration, updater refresh, adopted-release update, documentation, and verification.
- **Unchanged:** Migration remains operator-authorized; project update never launches an agent or offers legacy migration; specialist skill instructions remain globally maintained.
- **Edge cases:** Declined migration, unavailable GitHub issue probe, failed migration, incomplete issue closure, missing migrated ledger, unsupported harness, and unversioned updater build fail without continuing initialization.
- **Decisions:** Project-local adapters contain only stable pointers to `~/.agents/sdlc`; evolving instructions are never copied into each project.
- **Next step:** Verify, audit, release as SDLC v2.1.0, then deploy.

***

## Requirements

### R-001 - Offer legacy migration during brownfield initialization

When a brownfield project contains `docs/ACs.md` without `docs/ACs.org`, initialization MUST identify whether GitHub issues remain and ask whether to run the legacy migration skill before creating Spec Kit infrastructure. Declining MUST stop without continuing initialization.

### R-002 - Use configured audit execution

Accepted migration MUST invoke the skill using `SDLC_AUDIT_HARNESS`, `SDLC_AUDIT_PROVIDER` when supported, and `SDLC_AUDIT_MODEL`, resolved through normal configuration precedence.

### R-003 - Verify migration outcome

Initialization MUST continue only after the migration command succeeds, no GitHub issues remain open, `docs/ACs.md` is absent, and `docs/ACs.org` exists.

### R-004 - Refresh project-local skill adapters

`sdlc-project-update` MUST recompose project-local Spec Kit skills even when the stored preset is unchanged. Every adapter MUST direct the invoked agent to the corresponding current instruction under `~/.agents/sdlc`.

### R-005 - Preserve updater boundaries

Project update MUST never offer legacy migration or launch an agent. It MUST update exactly one adopted SDLC revision in the project constitution to the updater's tagged release and refuse ambiguous or unversioned input.

## Tests

### RT-001 - Migration uses audit configuration

- **Requirements:** R-001 - offer legacy migration; R-002 - use configured audit execution; R-003 - verify migration outcome
- **Procedure:** Exercise brownfield migration with a stubbed issue probe and Hermes audit configuration.
- **Expected result:** The offer names a described issue, Hermes receives the configured provider and model, and initialization continues only after issue closure and ledger replacement.

### RT-002 - Update recomposes an unchanged preset

- **Requirements:** R-004 - refresh project-local skill adapters
- **Procedure:** Run preset installation in update mode with equal source and destination preset content.
- **Expected result:** The existing preset is removed and added again, causing Spec Kit to recompose its project-local skills.

### RT-003 - Updater remains bounded and versioned

- **Requirements:** R-005 - preserve updater boundaries
- **Procedure:** Exercise command dispatch, no-launch enforcement, release resolution, and constitution revision validation.
- **Expected result:** The updater cannot launch an agent or enter migration, rejects unsafe revision states, and idempotently records a tagged release.

## Solution Design

Reuse the schema-resolved audit configuration already held by `sdlc-project-init`. A shared harness launcher passes provider selection only to Hermes and passes the selected model to all supported harnesses. The initializer performs a bounded one-item `gh` probe, invokes the migration skill only after confirmation, and verifies its durable outcomes before proceeding.

The updater is the same executable selected by its invocation name. Update mode forces preset removal and re-addition so Spec Kit recomposes project-local skills, but backs up the preset only when its owned source content differs. Stable project-local adapters point to canonical command instructions under `~/.agents/sdlc`, allowing later `make install` runs to update behaviour globally.
