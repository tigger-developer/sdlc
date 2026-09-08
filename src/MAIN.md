# Lean Software Delivery Standards

This repository is a standalone standards and delivery framework for coding
agents. The project and human operator retain authority. SDLC v3 supports Codex,
Claude Code, GitHub Copilot CLI, and Hermes through provider-neutral standards,
canonical global skills, and capability-aware adapters.

The canonical installed root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

## Universal engineering rules

- A question is not an instruction. Answer it without modifying files or
  starting a delivery workflow.
- Never write code without a defined specification. In normal delivery this is
  the active `spec.org`; paired and emergency delivery use the explicit
  alternatives defined below.
- Do not silently decide product behaviour, scope, architecture, security,
  persisted-data formats, access, or irreversible outcomes.
- Never widen access to data or systems without explicit human instruction.
- Preserve human edits. Do not overwrite or revert them.
- Assertions require current evidence. Distinguish verified fact, inference,
  design intent, and unresolved uncertainty.
- Never claim work complete from partial evidence. State what changed, what was
  checked, and what remains unchecked.
- Diagnose before explaining. Correct root causes rather than accumulating
  workarounds.
- Match established project conventions unless the specification records a
  deliberate change.
- Never present a bare identifier to a human. Give every work, requirement,
  test, ticket, finding, or commit ID an adjacent descriptor.
- Never reuse an identifier. Deletion, archival, abandonment, retirement, and
  supersession permanently reserve it.

## Common command prohibitions

- Never invoke `rm`; use a recoverable deletion tool such as `trash`.
- Never invoke `sed` or `awk`. Use a format-aware tool, `rg`, or a direct editor.
- Never use Perl one-liners as shell stream editors.
- Never invoke `python` or `python3`. Python development must use project-owned
  entry points such as its task runner, test target, or environment manager.
- Never read a file whose exact basename is `.env`.
- Never bypass checks with `--no-verify`, disabled hooks, suppressed failures,
  or equivalent shortcuts.
- Do not use SSH, SCP, SFTP, or remote deployment commands without explicit
  authority for the exact operation and target.

## Progressive loading

Read this file first, then only the standards relevant to the work.

| Work | Additional standards |
|---|---|
| Specification, acceptance criteria, bugs, or clarification | `~/.agents/sdlc/ISSUES.md` |
| Test definition or verification | `~/.agents/sdlc/TESTING.md` |
| Audits or gates | `~/.agents/sdlc/AUDITS.md` |
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

## Normal workflow

Normal delivery has two phases and two human gates:

1. **Define:** create one context-independent `spec.org` containing context,
   acceptance criteria, traced test definitions, edge cases, and solution
   design. The `define-change` workflow audits and remediates it locally, runs
   one retained external definition audit, then obtains operator sign-off.
2. **Deliver:** write justified automated tests first, observe RED, implement,
   observe GREEN, then let `deliver-change` audit and remediate the implemented
   tests and code locally before one retained external implementation audit.
   Execute required one-off and user tests, update affected documentation, and
   obtain operator closure.

The definition gate combines specification, design, and test-definition audits.
The implementation gate combines implemented-test and code audits. Read
`AUDITS.md` before either gate.

A specification is a delivery handoff. A new agent must be able to deliver it
using only `spec.org`, `.sdlc/project.yaml`, the named authorities, and the
repository. Uncaptured conversation or hidden assumptions fail the gate.

## Delivery modes

`define-change` is always **ATTENDED**: it may ask the operator for the bounded
decisions needed to make the specification safe and context-independent.

`deliver-change` performs one preflight question set before implementation. The
operator may explicitly select **ATTENDED**; otherwise that invocation proceeds
in **HANDS-OFF** mode. Hands-off is therefore an invoked delivery mode, not a
global default. The cost of stopping prematurely is high: progress reports,
commits, warnings, partial milestones, and ordinary audit remediation are not
terminal handbacks while safe, authorized, executable work remains.

Hands-off delivery records material routine assumptions in the specification's
`Decision Log (Hands-off delivery only)`. It stops only for a human decision or
genuine blocker, a failed mandatory check that cannot be safely remediated, or
the configured audit limit without `PASS` or `PROVISIONAL`. It never authorizes
new product behaviour, scope, architecture, security, access, data, external
contract, or irreversible decisions.

## Variant workflows

- **Paired development:** only an explicit operator selection enables the live
  collaborative route in `PAIRING.md`. Ordinary conversation does not.
- **Emergency delivery:** only the exact operator token `BYPASS-GATE-7` in the
  same request authorizes and invokes `emergency-change`, which applies
  `EMERGENCY.md`. An agent must never suggest, infer, or self-authorize it.

## Project initialization

Run `sdlc-init` exactly once. It creates or migrates the v3 project
profile and Org artefacts, preserves the prior state on a dated archive branch,
and handles SDLC v1 or v2 evidence without installing Spec Kit. The presence of
`.sdlc/project.yaml` means initialization has already occurred.

---

# Canary

Suffix the provider's base coding canary with " SDLC" (with leading space) if
you have read and agree with this document. Then append the suffix from every
task-selected reference document read in full, in the order those documents
are reported to the operator.

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
