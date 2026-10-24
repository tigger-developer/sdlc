---
title: Instruction Routing and Rule Preservation
version: 1
last-updated: 2026-10-24
---

# Instruction routing and rule preservation

This change reorganizes the public framework independently of any private
provider configuration. It preserves engineering and approval requirements while
separating the documents required by different roles. Source remains maintained
agent code, not an instruction to change the maintainer's current session.

## Routes

| Invocation | Required route and boundary |
|---|---|
| Explicit independent fresh-context audit, with or without a project marker | AUDIT, READING, REVIEW and the applicable criteria; never MAIN or delivery workflows |
| Workflow review in the author's context | SELF-AUDIT and REVIEW; explicitly internal evidence, no external call |
| Initialized project delivery | MAIN, selected standards and invoked workflow; reuse fully read retained text |
| Advice in an initialized project | READING and retained MAIN/project authorities; no implicit workflow invocation |
| Non-SDLC advice or explicitly requested review | Applicable standards and existing project contracts; no invented SDLC artefacts |
| Incomplete project marker | Report the profile path; diagnosis may proceed, initialization requires authority |

The role must be explicit; an author cannot self-select independence to escape
delivery obligations. A manual fresh reviewer returns findings only. A skill
name or separate provider label alone does not establish independence.

## Preservation map

Baseline: v3.8.9. Existing document histories and Git history remain available.

| Obligation or duplicated material | Canonical owner after this change |
|---|---|
| Complete reads, fixed installed root, absent documents, retained context | READING; skills reference it once rather than restarting the chain |
| Shared scope, evidence, human edits, identifiers, write and command safeguards | READING, selected by every active route |
| Delivery specification, project numbering, initialization, workflow selection | MAIN; initialization remains explicitly authorized |
| Independent reviewer purpose and findings-only authority | AUDIT, selected before MAIN |
| Architecture, test and implementation criteria; evidence gaps and verdict semantics | REVIEW and the existing substantive standards/review skills |
| Test taxonomy, intended RED, scaffolding, final checks, UTs and readiness | TESTING and REVIEW; delivery scheduling in AUDITS |
| Definition/test-code/delivery-code order and variant boundaries | AUDITS; paired/emergency/spike rules remain distinct |
| Annotation blocking, findings as a batch, challenges and scope disputes | REVIEW and AUDITS; no silent waiver or reviewer shopping |
| Internal evidence, remediation, continuation and configured round limit | SELF-AUDIT; no harness-state mutation or independence claim |
| Manual fresh review and human approval | AUDIT/AUDITS; manual results separately labelled, approval remains human-owned |
| External findings, budgets, recovery, reset, cache and session ownership | Preserved former AUDITS body in EXTERNAL-AUDIT plus HARNESS; suspended and excluded from normal review loading |
| Hands-off mode transitions, reversible assumptions, stopping and final reread | deliver-change's shared continuation section, referenced by spike |
| Native goal controls and harness-specific limits | Existing native-goal reference; no child delivery session or budget replenishment |
| Documentation metadata applicability and fields | DOCUMENTATION and MARKDOWN respectively |
| Public operation without private canaries | MAIN/AUDIT provide standalone canaries; no private personality dependency |

The former AUDITS version-12 body, including its document history, was compared
verbatim with EXTERNAL-AUDIT. Its historical routing is explicitly superseded,
not silently retained as active policy. Other consolidated rules are traceable
through the table above and each changed file's diff.

## Evidence transport

Managed, standalone and spike packages now share applicable technology and
format selection. Evidence supplements project declarations; NODE requires
JAVASCRIPT and HUGO requires WEB. Optional references remain review-only. MAIN,
AUDITS, SELF-AUDIT, EXTERNAL-AUDIT, HARNESS and delivery workflows are excluded.
Required missing standards still fail closed. Duplicate evidence paths are not
introduced when an input already names the project profile.

Standards contract version 2 replaces retained version-1 contexts before a new
provider turn, preserving history and budgets. Unchanged standards are still
sent once per retained current context; changed contents are supplied again.
The automatic harness remains suspended. These changes align its future evidence
selection with the routing contract; they do not certify or repair live providers.

## Verification and limits

- Local Go tests exercise selection, profile/evidence union, technology
  prerequisites, mandatory omissions, optional request refusal, all four
  provider adapters' retained-context reuse and legacy-context replacement.
- The full Go suite, including installer and initializer fixtures, passed.
  The Hermes bridge's nine local tests, Ruff, shellcheck and shfmt passed.
- All 15 public skill files passed the skill-creator validator. Its missing
  PyYAML dependency was supplied only in a temporary uv environment.
- `make test` stopped at the installed golangci-lint 2.15.0 rejecting the
  repository's versionless configuration. A temporary migrated configuration
  reported 13 findings in unchanged code: one errcheck, five gosec and seven
  staticcheck findings. A separate check of this change using that temporary
  configuration and `--new-from-rev=HEAD` reported zero new issues.
  The repository lint configuration was not changed;
  this is not a passing full `make test` result.
- One-time document review checks role precedence, retained-rule ownership,
  standalone boundaries and absence of private dependencies. It is author
  review, not an independent audit or a behavioural proof of model compliance.
- No live external audit, provider probe or installation was performed. Runtime
  latency and context retention across deployed harnesses remain unmeasured.

The active entry points are shorter and transport detail is deferred, but the
repository gains explicit shared contracts and preserved history. Total source
size is therefore not a claim about loaded context or response speed.
