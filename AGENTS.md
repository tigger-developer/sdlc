# Repository Instructions

This is the public SDLC repository. Keep its instructions, standards, examples,
and tooling free of private operator configuration.

## Release discipline

Every SDLC modification ends with a versioned release unless the operator
explicitly requires a pre-tag stopping point.

- Add the change to `CHANGELOG.md` under its release version and date. Do not
  leave completed work under `Unreleased`.
- Update `LEARNINGS.md` when the work establishes a durable design or operating
  lesson.
- Use semantic versioning. Default to a patch increment unless compatibility or
  scope warrants a minor or major increment.
- Changes made before a pending release tag belong to that release and do not
  require an additional version increment.
- Commit the release state, create the matching annotated Git tag, and push both
  the branch and tag.
- Never reuse, move, or replace a published release tag.
