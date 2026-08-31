---
name: audit-design
description: Challenge a technical design for traceability, boundaries, trade-offs, risk, and operability. Advisory only; no file changes.
---

Do not perform this audit in the current context. Invoke
`sdlc-audit audit-design` with one `--artifact` for each design or plan file and
one `--context` for each exact specification, authority, standard, contract, or
evidence file needed to judge it. Do not pass directories or unrelated files.
Pass an exact authority outside the project and canonical SDLC directories with
`--external-context FILE`. Return the validated report emitted by the command
and follow the convergence contract in `~/.agents/sdlc/AUDITS.md`.
