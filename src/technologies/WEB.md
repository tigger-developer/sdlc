# Web Standards

Standards for browser-facing HTML, CSS, JavaScript, and the testing and tooling
concerns that span them. The general coding standards in
`~/.agents/sdlc/CODING.md` apply on top of these. Projects that own JavaScript
or TypeScript source also load
`~/.agents/sdlc/technologies/JAVASCRIPT.md`.

When a Go application serves HTML, CSS, or JavaScript, the server-side patterns in `~/.agents/sdlc/technologies/GO.md` (HTTP server, error handling, timeouts) apply for the server; the standards in this document apply for the response payload.

Hugo projects also load `~/.agents/sdlc/technologies/HUGO.md`. Projects that
use Node.js as an application runtime, or npm-managed packages in their build,
also load `~/.agents/sdlc/technologies/NODE.md`. npm-managed development tooling
does not by itself justify selecting Node.js as the production runtime.

## Source vs Rendered Tier Model

Web content moves through three tiers between authoring and the user:

1. **Source** -- templates, partials, source CSS/JS files. Source code. Tests must never grep, parse, or introspect this tier (see `~/.agents/sdlc/TESTING.md`).
2. **Rendered/built** -- post-build HTML, fingerprinted CSS bundle, transpiled JS, RSS/feed XML, static assets. The artefact that travels over the wire. **Tests may query this tier** with format-aware tools -- the user's browser or feed reader receives this exact content.
3. **Presented** -- post-JS-execution DOM, computed styles, visual layout. What the user actually sees. Requires a browser (Playwright/Cypress) or human validation.

The source-introspection prohibition in `~/.agents/sdlc/TESTING.md` applies to tier 1. Tiers 2 and 3 are legitimate test targets. For mostly-static sites (Hugo), tier 2 is usually sufficient. For JS-heavy SPAs where meaningful content materializes only after browser execution, tier 2 is insufficient -- defer to tier 3.

## HTML

- **Use semantic elements.** `<article>`, `<nav>`, `<header>`, `<main>`, `<footer>`, `<section>` -- never `<div>` soup. Semantic markup is required for accessibility and parser tooling.
- **Document outline.** One `<h1>` per page. Heading levels descend without skipping.
- **Accessibility minimums:**
  - Every `<img>` has meaningful `alt` text (or `alt=""` if purely decorative)
  - Form inputs paired with `<label>` (or `aria-label` if visually labelled elsewhere)
  - Keyboard navigation works without a mouse; focus order is sensible and focus is visible
  - Colour contrast meets WCAG AA (>=4.5:1 for body text, >=3:1 for large text)
- **Validation.** Run `htmltest` on built sites in CI to catch broken links and malformed markup.
- **Metadata.** Include `<meta>` tags for `og:title`, `og:image`, `og:description` on user-facing pages. Reference: [Open Graph protocol](https://ogp.me/).

## CSS

- **Units:** `rem` for anything that scales with the user's font preference (typography, spacing, sizing). `px` only for borders, hairlines, or device-pixel-precise rules. Never `px` for font sizes.
- **Responsive by default.** Mobile-first media queries. Breakpoints in `rem`, not `px`.
- **No inline styles** in HTML attributes. CSS lives in `.css` files.
- **Methodology:** BEM or whatever the project already uses. Pick one; don't mix.
- **Lint with `stylelint`.** Minimum config:

```yaml
extends:
  - stylelint-config-standard
rules:
  declaration-no-important: true
  selector-max-id: 0
```

- **Performance:** prefer CSS Grid / Flexbox over absolute positioning. Avoid selector nesting deeper than 3 levels. Use CSS custom properties (`--var`) for theme values.

## Browser JavaScript and TypeScript

The language-wide rules live in
`~/.agents/sdlc/technologies/JAVASCRIPT.md`. Browser-facing code additionally:

- uses safe DOM APIs and contextual escaping for externally derived content;
- preserves semantic HTML, keyboard operation, visible focus, and accessible
  names when it creates or modifies interface elements;
- feature-detects optional browser APIs within the project's declared support
  range; and
- avoids long synchronous work on the main thread where it would make the
  interface unresponsive.

## Build and Test Tooling

Format-aware tools for web content:

| Use case | Tool | Install | Notes |
|---|---|---|---|
| Broad generated-site validation | `htmltest` | `brew install htmltest` | Broken links, missing assets, malformed markup. Default smoke check for static/generated sites. |
| Targeted DOM assertions | `htmlq` | `brew install htmlq` | CSS-selector queries against rendered HTML. HTML analogue of `jq`. |
| Targeted RSS/feed assertions | `xmlstarlet` | `brew install xmlstarlet` | XPath queries against generated feed XML. |
| Lint CSS | `stylelint` | `npm i -D stylelint` | Industry standard. |
| Lint JS/TS | `eslint` + `prettier` | `npm i -D eslint prettier` | See CODING.md Style Baselines. |

Do not introduce a new web test architecture without explicit human approval. For static-site checks, do not add Node.js, npm, Playwright, Cypress, or another browser stack merely to verify generated HTML, feeds, links, assets, routes, headers, or file serving. Prefer `htmltest`, `htmlq`, and `xmlstarlet` unless the project already has a narrower appropriate harness.

For tests that genuinely require a browser (computed styles, hover behaviour, JS-executed content), reach for Playwright or Cypress -- but only when the body of UI tests justifies the tooling cost (heuristic: roughly 8-10 UTs across the project) and the human has approved the added test architecture.

Any deployable web project with package-managed application or build
dependencies must expose the `make vulncheck` contract defined by
`~/.agents/sdlc/SECURITY.md`. Scan dependencies capable of changing the served
artefact, even when they are absent from the final static output.

## Testing Web Content

Cross-cutting with `~/.agents/sdlc/TESTING.md`. This section contextualizes those rules for web content; it does not override them.

Decision tree by tier:

| Tier | Test type | How |
|---|---|---|
| Rendered HTML, no JS execution needed | **Automated regression** | `hugo build` (or equivalent) -> `htmlq` query -> assert on text/attributes. Compliant with the source-introspection prohibition because rendered HTML is the user-facing artefact. |
| Broad generated-site validity | **Automated regression** | `hugo build` (or equivalent) -> `htmltest`. Use this for broad link, asset, and generated HTML validation. |
| RSS/feed output | **Automated regression** | Build the site -> `xmlstarlet` XPath query -> assert on feed structure/content. |
| Served static file, route, header, or asset | **Automated regression** | Use the existing project harness only if serving behaviour itself is under test. Do not install Node or a browser stack for this. |
| Computed style, hover, JS-rendered content | **Human validation by default; automated browser regression when justified** | Prefer human validation until the project adopts Playwright, Cypress, or an equivalent harness. |
| Visual judgement (typography, "looks right", layout aesthetics) | **Human validation** | Machine checks cannot replace human visual judgement here. |

Per `~/.agents/sdlc/TESTING.md`, the real-user test question still applies: *what user action does this test simulate, and what would the user observe?* For rendered HTML, generated feeds, or served static files the answer is "the user loads this URL and the browser/feed reader receives this response" -- legitimate. For source-template grep the answer is "nothing the user does" -- forbidden.

## Paired visual development

When the operator explicitly selects paired development, follow
`~/.agents/sdlc/PAIRING.md`. Hugo Markdown and YAML may carry accepted content
and configuration, while CSS, templates, and rendered pages carry the design
under review. A separate detailed design document is not required merely to
restate those artefacts.

Automation is not mandatory simply because rendered HTML can be queried. Use
`htmltest`, `htmlq`, browser automation, or another retained test only when it
protects an objective, stable behaviour at proportionate cost. Do not translate
a live visual approval into source-text assertions or a synthetic browser test
whose only purpose is to replay the operator's judgement. Record that judgement
as user-test evidence under `~/.agents/sdlc/TESTING.md`.

## Reference Standards

- HTML: [MDN HTML reference](https://developer.mozilla.org/en-US/docs/Web/HTML), [HTML Living Standard](https://html.spec.whatwg.org/)
- CSS: [MDN CSS reference](https://developer.mozilla.org/en-US/docs/Web/CSS)
- Accessibility: [WCAG 2.1 AA](https://www.w3.org/WAI/WCAG21/quickref/)
- JavaScript: [MDN JS reference](https://developer.mozilla.org/en-US/docs/Web/JavaScript), [Airbnb JS Style Guide](https://github.com/airbnb/javascript)
- Open Graph: [ogp.me](https://ogp.me/)
