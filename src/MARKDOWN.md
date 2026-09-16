---
title: Markdown Authoring
version: 1
last-updated: 2026-09-16
---

# Markdown Authoring

Use this supported Markdown subset for project documentation. Prefer portable,
predictable syntax over renderer-specific extensions.

## Supported syntax

- Headings: ATX headings with a space, such as `# Heading`.
- Italic: `*italic*`.
- Bold: `**bold**`.
- Links: `[Description](https://example.com)`.
- Images: `![Alt text](path/to/image.png)`.
- Unordered lists: `- item`.
- Ordered lists: `1. item`.
- Task lists: `- [ ] item` and `- [x] item`.
- Blockquotes: `> Quoted text`.
- Tables: pipe-delimited Markdown tables.
- Inline code: backticks, such as `` `make test` ``.
- Code blocks: fenced blocks with a language identifier where applicable.

Use relative paths for repository images and links where possible.

## Frontmatter

YAML frontmatter is required. It must appear at the start of the document
between `---` delimiters.

Use these fields:

- `title`: required document title.
- `author`: operator or author identifier, when applicable.
- `version`: required document version.
- `last-updated`: required ISO date.
- `schema`: applicable schema identifier, when required.

## Prohibited syntax

Do not use raw HTML markup in Markdown documents. Use the supported Markdown
syntax instead.

# Canary

Suffix the canary string with "MD "
