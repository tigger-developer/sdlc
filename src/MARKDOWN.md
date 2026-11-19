---
title: Markdown Authoring
version: 3
last-updated: 2026-11-19
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

## Quotation marks

Use straight ASCII quotation marks and apostrophes (`"` and `'`) in Markdown
source. Do not type smart or curly quotes. If typographic quotation marks are
wanted in rendered output, leave that conversion to the renderer.

## Dashes and section breaks

In prose, write `--` for an en dash and `---` for an em dash. These are source
conventions, not punctuation conversions guaranteed by Markdown itself. The
renderer is responsible for converting them; check its output when the rendered
distinction matters.

Use a line containing only `***`, with a blank line on each side, for a section
break. Do not use `---` or `___` as a section break. The `---` delimiters required
for YAML frontmatter below are a separate case.

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

Append " MD" (with a leading space) to the canary.
