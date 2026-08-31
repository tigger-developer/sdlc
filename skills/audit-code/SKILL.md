---
name: audit-code
description: Review implementation against the selected engineering standards and language best practice. Advisory only; no file changes.
---

Do not perform this audit in the current context. Materialize the change as an
exact diff file, then invoke `sdlc-audit audit-code` with that file as an
`--artifact`. Add one `--context` for each exact specification, design, test,
standard, or adjacent source file needed to judge the change. Do not pass
directories, the whole repository, or unrelated legacy files. Pass an exact
authority outside the project and canonical SDLC directories with
`--external-context FILE`. Return the validated report emitted by the command
and follow the convergence contract in `~/.agents/sdlc/AUDITS.md`.
