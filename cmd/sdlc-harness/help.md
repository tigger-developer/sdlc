SDLC harness
============

Internal SDLC helper for running one bounded provider context with an
immutable evidence bundle. It is installed at
`~/.agents/sdlc/bin/sdlc-harness` and is invoked by SDLC skills, not directly
from the global command path.

Usage
=====

    sdlc-harness start|resume [options] < audit-context.txt

Start and resume
================

`start` creates a fresh external context. For an audit, pass `--gate definition`
or `--gate implementation`; the installed YAML prompt registry supplies the
criteria. Do not pass `--session`; the runner creates the identity and prints
`SESSION_ID: <id>` on stderr. Save that exact ID.

`resume` continues the same external context. Pass the saved ID as
`--session <id>`. An agent task ID, path, or newly invented value is invalid.

Evidence and instruction
========================

- Repeat `--input <file>` for every exact evidence file.
- For audits, stdin is additional bounded context; the gate prompt is loaded
  from `prompts/audits.yaml`.
- `--audit-record <file> --work-item <id> --gate <gate>` makes the harness own
  start/resume selection and writes the session mapping to `audits.yaml`.
- With a record, `start` refuses an existing mapping and `resume` loads its
  recorded session ID. A supplied resume ID must match it.
- Audit status, findings, revisions, round numbers, and session IDs belong only
  in that `audits.yaml` record; do not copy them into a specification or ticket.
- If supplied artefacts duplicate audit state, the audit must return `FAIL` and
  require the duplicate to be removed before rerunning.
- `--phase` is `definition`, `build`, or `audit`.
- Project and global YAML configuration supply harness, provider where
  supported, model, timeout, and `delivery.audit.max_rounds`; explicit flags
  override them.

Output
======

- **stderr:** provider progress, diagnostics, and `SESSION_ID`.
- The harness also emits a liveness line at start and every 30 seconds while
  waiting; this is not provider progress or a verdict.
- **stdout:** the provider's final response.

Start example
=============

    /Users/tigger/.agents/sdlc/bin/sdlc-harness start \
      --phase audit --gate definition \
      --project . \
      --audit-record specs/W001-change/audits.yaml \
      --work-item W001-change \
      --input specs/W001-change/spec.org \
      --input .sdlc/project.yaml \
      < audit-context.txt

Resume example
==============

    /Users/tigger/.agents/sdlc/bin/sdlc-harness resume \
      --phase audit --gate definition \
      --project . \
      --audit-record specs/W001-change/audits.yaml \
      --work-item W001-change \
      --input specs/W001-change/spec.org \
      --input .sdlc/project.yaml \
      < audit-context.txt

A timeout is a runner incident, not an audit verdict. Record it and do not
silently create a replacement context.
