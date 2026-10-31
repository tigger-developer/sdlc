---
title: Lean Software Delivery Standards
version: 8
last-updated: 2026-10-31
---

# Lean Software Delivery Standards

**Independent auditor: stop reading here.** If explicitly invoked in a fresh
context solely to audit, read `~/.agents/sdlc/INDEPENDENT-AUDITOR.md` instead. Do not load the
rest of this document or activate delivery, even when `.sdlc/` exists. A
delivery agent's self-audit does not select that role.

For other authorized SDLC work, follow `READING.md`: required full reading is
satisfied by unchanged text already read and retained in context. A new skill
does not restart the loading chain. Missing/lost text must be read, not guessed.

## Ticket layout migration check

Before proceeding with project work, check for a root `specs/` directory. If
present, stop normal work and require `sdlc-init` (the SDLC init command) first.
Do not create tickets, edit definitions or implement against the old layout.
Initialization/migration and read-only diagnosis of its blockers are permitted.
For an already-v3 project this renames `specs` to `tickets` and replaces
`specs/` references in all tracked `.org`, `.md` and `.yaml` files. Do not migrate by ad hoc agent edits.
This check applies to consuming SDLC projects, not instruction source maintainers.

## Standing audit mode: internal only

The old external audit harness is retired. `AUDITS.md` owns delivery review
orchestration and routes current-session review through `SELF-AUDIT.md` and the
three review skills. The three review skills never launch another reviewer. An explicitly requested
independent review follows AUDITS' caller contract and INDEPENDENT-AUDITOR.
The retired external harness is not invoked. Review prerequisites
never replace human approval, readiness, tests, scope, holds or closure.

This repository is a standalone standards and delivery framework for coding
agents. The project and human operator retain authority. SDLC v3 supports Codex,
Claude Code, GitHub Copilot CLI, and Hermes through provider-neutral standards,
canonical global skills, and capability-aware adapters.

The canonical installed root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

## Raising defects

An explicit request to register a defect uses `defect-raise`, not
`define-change`. Record observations, expectations and evidence without requiring
root cause or fix design. It uses primary, the shared numeric namespace and
`tickets/DNNN-description/defect.org`. Defect-raise, define-change and
deliver-change all refuse to proceed until initialization and layout migration
have succeeded. Raising a defect does not authorize implementation or an audit.

## Delivery-specific engineering rules

READING owns the shared engineering boundaries and common command prohibitions.
They apply in full to this route as well as to advice and review.

- Never write code without a defined specification. In normal delivery this is
  the active `spec.org`; paired and emergency delivery and the signed-off
  hands-off spike brief are the explicit alternatives defined below.
- New tickets use W for work/specifications, D for defects and S for spikes,
  sharing one numeric namespace under `tickets/`. Follow `ISSUES.md`; preserve
  existing identities during migration.

## Progressive loading

Read this file first, then only the standards relevant to the work.
For feature definition, delivery, paired or emergency work also read DELIVERY.md; it
owns the execution, record and closure duties moved out of shared standards.

Independent audit selection belongs to INDEPENDENT-AUDITOR/REVIEW, not this workflow table.
Do not supply this document or the full library as an auditor preamble.

| Work | Additional standards |
|---|---|
| Specification, acceptance criteria, bugs, or clarification | `~/.agents/sdlc/ISSUES.md` |
| Test definition or verification | `~/.agents/sdlc/TESTING.md` |
| SDLC workflow audits or gates | `~/.agents/sdlc/AUDITS.md` |
| Architecture, solution design, delivery, or any audit gate | `~/.agents/sdlc/ARCHITECTURE.md` |
| Implementation or code review | `~/.agents/sdlc/CODING.md` |
| Explicit paired delivery | `~/.agents/sdlc/PAIRING.md` |
| Exact `BYPASS-GATE-7` emergency delivery | `~/.agents/sdlc/EMERGENCY.md` |
| Git, commits, branches, or hooks | `~/.agents/sdlc/GIT.md` |
| Technical documentation | `~/.agents/sdlc/DOCUMENTATION.md` |
| Org artefacts | `~/.agents/sdlc/ORGMODE.md` |
| Security, dependencies, or vulnerability checking | `~/.agents/sdlc/SECURITY.md` |
| APIs, webhooks, or integrations | `~/.agents/sdlc/technologies/API.md` |
| Go | `~/.agents/sdlc/technologies/GO.md` |
| JavaScript or TypeScript | `~/.agents/sdlc/technologies/JAVASCRIPT.md` |
| Lua or a Lua-hosted framework | `~/.agents/sdlc/technologies/LUA.md` |
| Python | `~/.agents/sdlc/technologies/PYTHON.md` |
| Bash or POSIX-style shell | `~/.agents/sdlc/technologies/SHELL.md` |
| Fish shell | `~/.agents/sdlc/technologies/FISH.md` |
| Perl | `~/.agents/sdlc/technologies/PERL.md` |
| Rust | `~/.agents/sdlc/technologies/RUST.md` |
| Swift | `~/.agents/sdlc/technologies/SWIFT.md` |
| Web interfaces or sites | `~/.agents/sdlc/technologies/WEB.md` |
| Hugo sites | `~/.agents/sdlc/technologies/HUGO.md` and `WEB.md` |
| Node.js or npm applications | `~/.agents/sdlc/technologies/NODE.md` |

The tracked project profile is `.sdlc/project.yaml`. It names project facts,
standards, authorities, and explicit local overrides. Global non-secret
defaults and the installed SDLC release live in `~/.agents/sdlc.yaml`. The
schema `version` and deployed `release` are separate values. Project profiles do
not pin the globally installed SDLC. Agents never read `.env`.

For non-SDLC advice use ADVICE's advisory route. For an explicit review choose
INDEPENDENT-AUDITOR (fresh independent context) or SELF-AUDIT (current author context), using
the standalone profile. Neither initializes SDLC or requires its artefacts.

When a document must be presented outside the terminal, use the executable named
by `HTML_PREVIEW_TOOL`. If unset or empty, use `htmlpreview` from
[tigger-developer/HTML-Preview](https://github.com/tigger-developer/HTML-Preview).
The override is an executable name or path, not a shell command string.
`make install` installs the default previewer when missing.
Preview is presentation, not a workflow gate.

Commands and executables under the canonical SDLC root MUST be invoked by their
resolved absolute filesystem path, not by relying on `PATH`.

## Normal workflow

Normal delivery has two phases and two human gates:

1. **Define:** create one context-independent `spec.org` containing context,
   acceptance criteria, traced test definitions, edge cases, and solution
   design. The `define-change` workflow checks and corrects its own draft, runs
   the `definition-audit` skill in-session, then obtains operator sign-off.
2. **Deliver:** write and execute the AT/RT package, validate its records and obtain
   the test-code review. Implement, obtain AT/RT/OT GREEN, then validate and review
   delivery-code through the in-session review skills. `deliver-change`
   follows `TESTING.md` and `AUDITS.md`; UTs may await the operator. Update affected
   documentation before the final delivery-code verdict and obtain operator closure.
   Follow TESTING.md for minimal test scaffolding; skip test-code only when neither ATs nor RTs apply.

Read `AUDITS.md` before any gate for its purpose, evidence ownership, finding
resolution and continuation contract. Ordinary author checks are not extra
review rounds. Test-code PASS is not delivery approval. Record every review method in canonical `audits.yaml` under `AUDIT-RECORD.md`;
never duplicate audit state in parallel documents.

A specification is a delivery handoff. A new agent must be able to deliver it
using only `spec.org`, `.sdlc/project.yaml`, the named authorities, and the
repository. Uncaptured conversation or hidden assumptions fail the gate.

## Delivery modes

`define-change` defaults to **ATTENDED** unless the operator explicitly invokes
hands-off definition: it may ask the operator for the bounded
decisions needed to make the specification safe and context-independent.

`deliver-change` owns the delivery-mode transitions and continuation contract.
Hands-off delivery is enabled by invoking that workflow, not a global default.
The explicitly invoked spike workflow reuses that contract with its signed-off
brief and non-merge/non-deployment boundary. Definition retains its distinct
interruption rule in `define-change`; ordinary conversation selects neither.

## Variant workflows

- **Paired development:** only an explicit operator selection enables the live
  collaborative route in `PAIRING.md`. Ordinary conversation does not.
- **Hands-off spike:** only an explicit operator invocation of
  `hands-off-spike` starts an unattended exploration. It answers a signed-off
  question on a never-merged branch, and its tests and implementation receive
  the in-session review skills in spike mode. It never delivers product behaviour.
- **Emergency delivery:** only the exact operator token `BYPASS-GATE-7` in the
  same request authorizes and invokes `emergency-change`, which applies
  `EMERGENCY.md`. An agent must never suggest, infer, or self-authorize it.

## Project initialization

Only when project initialization is explicitly requested, run `sdlc-init` once.
Reading these standards or giving advice does not authorize it. It creates or migrates the v3 project
profile and Org artefacts, preserves the prior state on a dated archive branch,
and handles SDLC v1 or v2 evidence without installing Spec Kit. The presence of
`.sdlc/project.yaml` means initialization has already occurred.

---

# Canary

Suffix the provider's base coding canary with " SDLC" (with leading space) if
you have read and agree with this document. Then append the suffix from every
task-selected reference document read in full, in the order those documents
are reported to the operator.
If no base canary is defined, use `SDLC` alone followed by the reference suffixes;
the public framework requires no private personality or operations markers.

On the first interaction for a task after reading `MAIN.md` and every
task-selected reference document in full, immediately follow the canary with
this statement:

`I have read the relevant SDLC documents in full. I pledge to uphold their rules, the spirit of these same rules, and that I will not attempt to game these same rules.`

Do not repeat this statement with later canaries for the same task. If anything
in this document or a selected reference is unclear, countermands a previous
instruction, or contradicts itself internally, say so now. If you are not
prepared to follow them, say so now. If the above is all true, include `SDLC`
and every applicable reference suffix in the canary greeting for every
interaction with the operator.

## Document history

This revision separates independent review, author self-audit and delivery-only
responsibilities. Earlier rules remain in their canonical owners and Git history.

Version 5 redirects independent auditors before workflow loading, standardizes
reading reuse and delegates audit/mode details to their canonical owners.

Version 1 records the document metadata and strengthens mandatory standards
reading for harness audits. Earlier revisions remain in Git history.

Version 2 adds standing internal review after external-auditor execution failure.

Version 3 makes internal review the standing route without an external attempt.

Version 4 includes ATs in automated-test readiness and the test-code admission rule.
