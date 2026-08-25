# DOCUMENTATION STANDARDS

## Voice and tone

Documentation must be written impersonally, in the third person. Describe the system, the behaviour, and the rationale -- not who did what or who will decide. Documentation outlives conversations and must stand on its own.

**Wrong:**
- "I added a new flag to handle this case."
- "We should run the linter before committing."
- "The project owner decides which output mode to use."

**Right:**
- "The `--verbose` flag enables debug output."
- "The linter must pass before a commit is accepted."
- "Output mode is configured via the `--mode` flag."

## Versioning

Project documentation files (under `./docs/` in a project) should include a version header:

```markdown
<!-- Version: 1.2 | Last updated: 2026-01-20 -->
```

When making significant changes, increment the version and note what changed in a brief changelog at the end of the document.

Global framework documents under `~/.agents/sdlc` are version-controlled via git and do not require version headers. Provider-specific live configuration under `~/.claude`, `~/.codex`, `~/.copilot`, or another agent home is not a documentation source.

## Process

- Before devising any solution, make sure that you have read and digested the documentation for this project which may have changed since the last time you looked.
- When changing any code, update the relevant documentation after (and not before).
- All scratch files must live in a unique directory allocated by the operating system's temporary-directory mechanism. Never create project-local agent scratch directories. Cleanup logic must remove the exact allocated directory when the documentation operation ends.

## Structure
- Each project should have README.md or README.org (but not both)
- Documentation should be in .md or .org format. Markdown (`.md`) is the default; use `.org` only where a project already standardizes on it
- All other documentation should reside in ./docs/
- README should contain:
  - A high-level overview of the project
  - A Quickstart section
  - Document any project dependencies
  - A list of the important files in the project and their purpose
  - A brief overview of each doc file and a jumping off point to it
- ./docs/ should contain
  - VISION: A document outlining the overall vision and goals of the project.
  - architecture: A document describing the high-level architecture of the project.
  - testing: A document detailing the testing strategy and procedures.
  - implementation_plan (if relevant)
  - a separate doc for each significant feature or area
  - help text for each executable (e.g. `<command>-help.md`), which the executable reads at runtime or has packaged in at build time. Help text is documentation and belongs alongside the rest of the docs, not embedded in the executable as an inline heredoc.

## Problems
- Stop and alert me if there are any major inconsistencies in the documentation that impact the task at hand
- Warn me if there are any other inconsistencies in the documentation that do not impact the task at hand

## License

License is Apache 2.0 with Copyright Tadhg O'Brien unless otherwise specified

# Canary
Suffix the canary string with "DOC "
