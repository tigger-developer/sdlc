# Documentation Standards

Technical documentation is part of the product contract. Keep it accurate,
direct, and usable without private context.

## Audience and voice

- Write technical documentation impersonally and in the third person. Describe
  the system, behaviour, and rationale, not who did what or who will decide.
- Write for the reader who must understand, operate, maintain, or contribute to
  the project.
- Lead with what the reader can do or needs to know.
- Prefer precise plain language over slogans, filler, or unexplained jargon.
- Define acronyms and project-specific terms on first use.
- Never give an identifier without an adjacent short descriptor.
- Distinguish verified behaviour, design intent, examples, and hypotheses.

Document the product, not the conversation that produced it. State requirements,
decisions, rationale and verified outcomes directly; do not paraphrase operator-agent
requests, permissions, disagreements or exchanges. Do not include personal names
except where required for legal attribution, such as a licence or copyright notice.
Use a role only where technically relevant. Preserve required original feedback
and immutable historical evidence in their designated records, not product prose.

- Instead of "The operator authorized removing Node", write "Node.js is not a
  dependency" when that is the verified state; record removal in the changelog.
- Instead of "The operator requested settings-panel guidance", state the required
  guidance in the specification, or describe the delivered guidance in user docs.
- Record approval as "Approved YYYY-MM-DD", adding scope only where limited.
  Record user tests as status, date, tested revision and observed outcome, with
  method "user validation"; do not narrate who reported the result or imply an
  automated execution. Never infer approval or a test result from this wording.

## Document purpose

- **Vision:** purpose, intended users, context, outcomes, scope and direction.
- **Architecture:** system structure, boundaries, responsibilities, technologies,
  constraints and technical rationale. Link feature requirements; do not absorb them.
- **Specification:** change requirements, tests, solution design and concise approval
  metadata, not a transcript of requirements gathering.
- **Validation:** dated execution status and evidence for the tested revision.
- **Decision logs:** the decision, rationale and consequences, not conversational
  history. **Changelogs:** changes to the product or document, not who requested them.

Put each fact in the document that owns it. Preserve useful technical rationale
without turning Vision or Architecture into a session journal.

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

## Format-specific markup

Use the native syntax for the document format. For Markdown documents, read and
follow `MARKDOWN.md`; do not write raw HTML document markup. For Org documents,
read and follow `ORGMODE.md`; do not write Markdown syntax in an Org document.
New and materially updated Markdown documents must begin with YAML frontmatter
containing the fields defined by `MARKDOWN.md`. Existing documents may adopt
the metadata when they are next updated. Org documents must use the Org-native
metadata syntax defined by `ORGMODE.md`.

## Accuracy and maintenance

- Update affected documentation with behaviour, interface, configuration, and
  operational changes.
- Verify command flags and tool behaviour from the current interface or a probe.
- Mark version-specific instructions and remove stale alternatives.
- Preserve useful history when changing a decision; identify superseded advice
  rather than silently erasing why it existed.
- Report major contradictions that affect the work instead of choosing one
  silently.

New and materially updated Markdown documents must begin with YAML frontmatter
containing `title`, `version`, and `last-updated`. Include `author` when an
operator or other author identifier is applicable, and `schema` when the
document follows a named schema. The version is the document version, not the
Git, product, or SDLC release version. Significant revisions increment that
version and retain a concise document changelog. This standard does not require
a fleet-wide migration of existing documents.

## Public documentation

Public documents must stand on their own. Do not assume access to private agent
instructions, local shell functions, personal tools, private infrastructure, or
an author's filesystem. Where a private convenience exists, describe the public
portable mechanism only.

## Temporary notes

Keep scratch notes outside tracked documentation unless the project explicitly
uses a tracked decision log. Convert durable findings into the appropriate
specification, plan, architecture record, or reference document.

## Delivery closure

Definition identifies affected documentation in the handoff. Update it during
delivery and include it in the final delivery-code audit: verify usage,
configuration, interfaces and architecture wherever the change affects them,
and compliance with this document's voice, purpose and presentation rules.
Missing or materially stale required documentation blocks that gate; optional
editorial improvements do not. Test-code review does not require completed product
documentation. Paired/emergency work updates product docs after the code change
but before the final delivery verdict; retrospective spec consolidation may follow.

Before closing normal, paired, or `BYPASS-GATE-7` work, reconcile every
document made inaccurate, incomplete, or misleading by the change. Record the
documentation outcome as updated or not affected, with the reason when it is
not affected. Known stale documentation blocks closure.

For an emergency change, replace the temporary specification with durable
project authority after verification. Update the specification, design,
operational guidance, and user documentation only where the delivered change
affects them; do not create duplicative prose that merely restates the code.

## Review

Review documentation for document-purpose fit, impersonal factual voice, technical
correctness, missing prerequisites,
ambiguous pronouns, broken references, unsafe copy-and-paste commands, and
unexplained identifiers. A prose grep is not a behavioural regression test;
use a one-time technical review and format-aware link or schema checks where
appropriate.
# Canary

Suffix the canary string with "DOC "
