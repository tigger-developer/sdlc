## SDLC standards

The canonical SDLC root is exactly `~/.agents/sdlc`. Never search the
filesystem to locate it. If `~/.agents/sdlc/MAIN.md` is absent or unreadable,
report that exact path.

Read `~/.agents/sdlc/MAIN.md` in full. Determine the project's languages,
interfaces, test strategy, documentation needs, and source-control constraints
from verified project evidence and the user's stated intent. Read only the
applicable standards routed by `MAIN.md`.

Populate the Engineering Standards Profile in the constitution. Include the
universal document, applicable documents, the current SDLC release or Git
revision, project-specific additions, and explicit deviations. Do not import
SDLC delivery process: Spec Kit owns orchestration.
