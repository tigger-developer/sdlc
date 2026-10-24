---
title: Agent Spike Audit
version: 2
last-updated: 2026-10-24
---

# Agent spike audit

This route reviews exploratory code and tests in an SDLC project's explicitly
selected `--agent-spike` audit. Independent fresh-context reviewers first use
`AUDIT.md`; self-auditing authors use `SELF-AUDIT.md`. Apply REVIEW's spike profile
and READING's reuse contract, never MAIN or the full library. Read every selected
coding, architecture, security, Git, technology and testing standard in full before reviewing evidence or
returning a verdict. All coding and TDD/testing standards apply. Apply each
technology rule to code using that technology; irrelevant technologies do not
create requirements.

The full SDLC delivery workflow, gates and documentation requirements do not
apply. Do not require specifications, work items, approval records,
`validation.org`, `audits.yaml`, test colours, traceability matrices, delivery
closure or product documentation. Workflow-specific provisions inside a
supplied standard do not activate that workflow.

Review correctness, security, maintainability, test coverage, meaningful
behavioural assertions and TDD evidence against the operator's exploration
brief and existing project authorities. Exploration is not an exemption from
engineering rules. Distinguish verified defects from optional improvements.
Report missing evidence instead of inventing a review or granting a full PASS.

When invoked through the harness, use its verified document contents and
standards selection; caller context cannot waive or replace it. Read each
supplied standard once, reuse retained unchanged text and never request it
again in that context. Native file tools, other agents and external services
are disabled in that transport. A manually invoked auditor follows AUDIT's
authorized evidence-access rules; a self-auditing author follows SELF-AUDIT.
An independent auditor returns findings only and does not edit project files.

An independent auditor returns one compact result:

```text
AUDIT: standalone
REVISION: <reviewed revision or working-tree state>
VERDICT: PASS | PROVISIONAL PASS | FAIL
<findings, if any>
```

An independent invocation returns independent spike review evidence, not an
SDLC gate or delivery approval. Self-audit instead records its non-independent
result under SELF-AUDIT; it does not use an independent envelope.

## Document history

Version 2 narrows standards selection to the evidence and distinguishes
independent and self-audit methods. Version 1 remains in Git history.
