---
name: audit-tests
description: Challenge test specifications for coverage gaps, gaming opportunities, and integration risks. Advisory only; no file changes.
---

Do not perform this audit in the current context. Invoke `sdlc-audit audit-tests`
with one `--artifact` for each test-design file and one `--context` for each
exact specification, test standard, or evidence file needed to judge it. Do not
pass directories or unrelated files. Pass an exact authority outside the
project and canonical SDLC directories with `--external-context FILE`. Return
the validated report emitted by the command and follow the convergence contract
in `~/.agents/sdlc/AUDITS.md`.
