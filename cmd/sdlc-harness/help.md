Usage
=====

    sdlc-harness start|resume [options] < audit-instruction.md

`sdlc-harness` is an internal SDLC workflow helper. It runs one bounded
provider harness with an immutable evidence bundle. It is installed at
`~/.agents/sdlc/bin/sdlc-harness` for skills; it is not an operator-facing
global command.

Internal SDLC helper: this command is workflow infrastructure, not an
operator-facing global command.

Operations
==========

`start` creates a new retained external context. **Do not pass `--session` to
`start`**: the runner creates the external identity. Capture the exact
`SESSION_ID` line it prints on stderr; that is the value required for a later
resume. The invoking agent's own conversation or task ID is not an external
session identity.

`resume` sends a revised evidence bundle to the same retained external context.
It requires `--session` set to the exact previously reported `SESSION_ID`. Do
not substitute an agent task name, a filesystem path, or a newly generated ID.

Evidence and instruction
========================

Repeat `--input` once for every exact evidence file. The audit instruction is
read from stdin. Inputs are copied into an immutable bundle for the invocation;
the runner does not interpret a prose list of filenames as evidence.

Configuration
=============

The runner resolves harness, provider where supported, model, and timeout from
the project `.sdlc/project.yaml`, then `~/.agents/sdlc.yaml`, with explicit
flags taking precedence. `--phase` is `definition`, `build`, or `audit`.

Output channels
===============

Provider progress and diagnostics are written to stderr. The final provider
response is written to stdout. The stable session identity is written to
stderr as:

    SESSION_ID: <provider session identity>

Example: start an audit
=======================

    /Users/tigger/.agents/sdlc/bin/sdlc-harness start \
      --phase audit \
      --project . \
      --input specs/001-change/spec.org \
      --input specs/001-change/audits.org \
      --input .sdlc/project.yaml \
      --timeout 5m \
      < audit-instruction.md

Record the `SESSION_ID` emitted on stderr before continuing.

Example: resume that audit
==========================

    /Users/tigger/.agents/sdlc/bin/sdlc-harness resume \
      --phase audit \
      --project . \
      --session "<SESSION_ID from start>" \
      --input specs/001-change/spec.org \
      --input specs/001-change/audits.org \
      --input .sdlc/project.yaml \
      --timeout 5m \
      < audit-instruction.md

Use the same retained context for the gate. A timeout is a runner incident, not
an audit verdict; record it and do not silently create a replacement context.
