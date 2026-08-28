# Documentation Standards

Technical documentation is part of the product contract. Keep it accurate,
direct, and usable without private context.

## Audience and voice

- Write technical documentation impersonally and in the third person. Describe
  the system, behaviour, and rationale rather than who performed an action.
- Write for the reader who must understand, operate, maintain, or contribute to
  the project.
- Lead with what the reader can do or needs to know.
- Prefer precise plain language over slogans, filler, or unexplained jargon.
- Define acronyms and project-specific terms on first use.
- Never give an identifier without an adjacent short descriptor.
- Distinguish verified behaviour, design intent, examples, and hypotheses.

## Structure

- Keep one primary `README.md` or `README.org`, following project convention.
- Put durable topic documentation under the project's established documentation
  directory.
- The README should provide an overview, quickstart, prerequisites, important
  files, and links into the detailed documentation.
- Mature projects should document vision, architecture, testing strategy,
  implementation planning where relevant, and significant feature areas.
- Executable help text is documentation. Keep it reviewable beside the other
  project documentation and package it at build time where runtime access is
  required; do not maintain a large duplicate inline heredoc.
- Use descriptive headings and short sections that can be linked directly.
- Put commands in executable order and explain destructive or environment-
  specific effects before the command.
- Keep examples minimal, valid, and free of secrets or private machine paths.

## Accuracy and maintenance

- Update affected documentation with behaviour, interface, configuration, and
  operational changes.
- Verify command flags and tool behaviour from the current interface or a probe.
- Mark version-specific instructions and remove stale alternatives.
- Preserve useful history when changing a decision; identify superseded advice
  rather than silently erasing why it existed.
- Report major contradictions that affect the work instead of choosing one
  silently.

Project documents under `docs/` should carry a version and last-updated header
when the project uses document-level versioning. Significant revisions increment
that version and retain a concise document changelog. Git-versioned global
framework documents do not require duplicate version headers.

## Public documentation

Public documents must stand on their own. Do not assume access to private agent
instructions, local shell functions, personal tools, private infrastructure, or
an author's filesystem. Where a private convenience exists, describe the public
portable mechanism only.

## Temporary notes

Keep scratch notes outside tracked documentation unless the project explicitly
uses a tracked decision log. Convert durable findings into the appropriate
specification, plan, architecture record, or reference document.

## Review

Review documentation for technical correctness, missing prerequisites,
ambiguous pronouns, broken references, unsafe copy-and-paste commands, and
unexplained identifiers. A prose grep is not a behavioural regression test;
use a one-time technical review and format-aware link or schema checks where
appropriate.
