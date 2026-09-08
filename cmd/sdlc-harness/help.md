SDLC harness
============

Internal SDLC helper for running one bounded provider context with an
immutable evidence bundle. It is installed at
`~/.agents/sdlc/bin/sdlc-harness` and is invoked by SDLC skills, not directly
from the global command path.

Usage
=====

    sdlc-harness start|resume [options] < audit-instruction.md

Start and resume
================

`start` creates a fresh external context. Do not pass `--session`; the runner
creates the identity and prints `SESSION_ID: <id>` on stderr. Save that exact
ID.

`resume` continues the same external context. Pass the saved ID as
`--session <id>`. An agent task ID, path, or newly invented value is invalid.

Evidence and instruction
========================

- Repeat `--input <file>` for every exact evidence file.
- Provide the audit instruction on stdin.
- `--phase` is `definition`, `build`, or `audit`.
- Project and global YAML configuration supply harness, provider where
  supported, model, and timeout; explicit flags override them.

Output
======

- **stderr:** provider progress, diagnostics, and `SESSION_ID`.
- **stdout:** the provider's final response.

Start example
=============

    /Users/tigger/.agents/sdlc/bin/sdlc-harness start \
      --phase audit \
      --project . \
      --input specs/001-change/spec.org \
      --input specs/001-change/audits.org \
      --input .sdlc/project.yaml \
      < audit-instruction.md

Resume example
==============

    /Users/tigger/.agents/sdlc/bin/sdlc-harness resume \
      --phase audit \
      --project . \
      --session "<SESSION_ID from start>" \
      --input specs/001-change/spec.org \
      --input specs/001-change/audits.org \
      --input .sdlc/project.yaml \
      < audit-instruction.md

A timeout is a runner incident, not an audit verdict. Record it and do not
silently create a replacement context.
