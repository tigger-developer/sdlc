---
name: useful-be
description: Get up to speed with the SDLC and project documentation before starting work.
---

This skill loads the minimum durable context needed to start useful work, then pulls task-specific reference docs on demand. Do not load every craft document speculatively.

## 1. Load the project context

Read in full:

- The repo's `README.{md,org}` if present
- Local project instructions in `AGENTS.md`, `AGENT*.{md,org}`, `CLAUDE.md`, or `.claude/CLAUDE.md` if present.
- If the project depends on infrastructure documented in a separate repo (e.g. an infra/deployment repo), and you have been told this, review that repo's documentation read-only.

Do not recurse through `./docs/` unless the user's task needs it. The point of this skill is orientation, not context flooding.

## 2. Load the SDLC layer

Read `[sdlc-home]/MAIN.md` in full first. Then apply its §"Reference Documents" routing table and read every selected reference in full.

Typical selections are:

- Issue-tracked implementation: `ISSUES.md`, `TESTING.md`, `CODING.md`, `GIT.md`, `DOCUMENTATION.md`, and every relevant language standard.
- Documentation-only tracked changes: `GIT.md` and `DOCUMENTATION.md`.
- Read-only diagnosis or advice: only the craft and language standards needed to ground the answer.

These examples do not replace the routing table. A task spanning several concerns selects every applicable row; never stop after reading `MAIN.md` or the first matching reference.

## 3. Read the infrastructure contract
If we are in a project that will be deployed outside the local machine, follow `MAIN.md` §"Externally deployed projects". The infrastructure repo must be supplied by the user's own instructions or the project; this distributable does not assume a personal infrastructure repository.

## 4. Confirm mode

On first load, name every selected document read in full and give a one-line reason for selecting it. Do not use the canary chain as a substitute for that report.
The user's provider-level `{CLAUDE,AGENTS}.md` rules apply in all cases and are *never* superseded by supplementary SDLC, project, or skill documentation.

In MODE PAIR, ask what to do next and await instruction. In MODE DELIVER, return the loaded context to the delivery workflow and continue towards the confirmed goal without asking what to do next.
