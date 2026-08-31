---
name: audit-spec
description: Challenge a specification's requirements and acceptance criteria. Advisory only; no file changes.
---

Do not perform this audit in the current context. Invoke `sdlc-audit audit-spec`
with one `--artifact` for each specification file and one `--context` for each
exact authority or evidence file needed to judge it. Do not pass directories or
unrelated files. Pass an exact authority outside the project and canonical SDLC
directories with `--external-context FILE`. Return the validated report emitted
by the command and follow the convergence contract in
`~/.agents/sdlc/AUDITS.md`.
