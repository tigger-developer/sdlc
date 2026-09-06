# Org Authoring

Use Org as an outliner, not as Markdown with different punctuation. Keep files
readable as plain text, navigable in Emacs, and renderable by Pandoc.

## Structure

- A headline is an addressable node in the document tree. Use it for a subject,
  artefact, requirement, test, decision, or work item.
- A list belongs inside a node. Use it for subordinate facts, steps, or options.
- Keep the hierarchy shallow. Three or four headline levels are normally enough.
- Use description lists for named values: `- Name :: value`.
- Use `#+STARTUP: overview` or `#+STARTUP: content` for initial folding.
- A dense supporting subtree may begin folded:

  ```org
  * Supporting evidence
  :PROPERTIES:
  :VISIBILITY: folded
  :END:
  ```

Folding is a view over explicit structure. Do not hide first-class requirements,
design decisions, or evidence merely to shorten the visible document.

## State and metadata

- Use checkboxes for the completion state of finite list items.
- Use TODO keywords for the lifecycle of a headline. SDLC specifications are
  authority documents, so TODO states belong in `work.org`, not `spec.org`.
- Use tags for cross-cutting classifications, not to repeat the hierarchy.
- Use property drawers for structured attributes such as `CUSTOM_ID`, type,
  priority, dates, and provenance.
- Use ordinary or property drawers for metadata and secondary supporting
  material only.

## Inline text and links

```org
*bold*
/italic/
_underline_
+strikethrough+
=verbatim=
~code~
[[https://example.com][Descriptive link]]
[[file:other.org][Other document]]
[[#stable-id][Named node]]
```

Use `=verbatim=` for filenames, commands, identifiers, and literal values, and
`~code~` for inline code. Do not use Markdown backticks, even when a renderer
accepts them.

## Typed blocks

Use the block that describes the content:

```org
#+begin_src sh
make test
#+end_src

#+begin_example
PASS
#+end_example

#+begin_quote
Quoted prose.
#+end_quote
```

- `src` is source code.
- `example` is literal input, output, or console text.
- `quote` is quoted prose.

Name a block only when another node needs to address it. Do not use Babel
execution, macros, dynamic blocks, generated agendas, or executable Emacs
configuration in SDLC artefacts.

## SDLC traceability

- Give every acceptance criterion and test definition a stable `CUSTOM_ID`.
- Link each criterion to every test that proves it.
- Link each test back to every criterion it covers.
- Give identifiers adjacent descriptors everywhere they appear.
- Preserve never-reused identifiers after deletion, retirement, or abandonment.

Before accepting an Org artefact, confirm that its hierarchy is useful when
folded and that every link and identifier remains understandable in plain text.
