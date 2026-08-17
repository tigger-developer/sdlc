# Web Standards

Standards for HTML, CSS, JavaScript, and the testing/tooling concerns that span them. The general coding standards in `[sdlc-home]/CODING.md` apply on top of these.

When a Go application serves HTML, CSS, or JavaScript, the server-side patterns in `[sdlc-home]/GO.md` (HTTP server, error handling, timeouts) apply for the server; the standards in this document apply for the response payload.

## Source vs Rendered Tier Model

Web content moves through three tiers between authoring and the user:

1. **Source** -- templates, partials, source CSS/JS files. Source code. Tests must never grep, parse, or introspect this tier (see `[sdlc-home]/TESTING.md`).
2. **Rendered/built** -- post-build HTML, fingerprinted CSS bundle, transpiled JS, RSS/feed XML, static assets. The artefact that travels over the wire. **Tests may query this tier** with format-aware tools -- the user's browser or feed reader receives this exact content.
3. **Presented** -- post-JS-execution DOM, computed styles, visual layout. What the user actually sees. Requires a browser (Playwright/Cypress) or a human (UT) to verify.

The source-introspection prohibition in `[sdlc-home]/TESTING.md` applies to tier 1. Tiers 2 and 3 are legitimate test targets. For mostly-static sites (Hugo), tier 2 is usually sufficient. For JS-heavy SPAs where meaningful content materializes only after browser execution, tier 2 is insufficient -- defer to tier 3.

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

## JavaScript

This is the JS minimum. When the stack grows substantive TypeScript work, factor a `CODE/TS.md` and reference it from here.

- **ES modules (ESM) only.** No CommonJS in new code.
- **Never `eval`.** Never `new Function(string)`. Never `setTimeout("string")`.
- **Strict equality** (`===`/`!==`). Never `==`/`!=` (except `== null` for the null-or-undefined check, which is allowed but should be deliberate).
- **`async`/`await` over `.then()`** chains. Mix only at module boundaries.
- **Handle promise rejections.** Every promise either has `.catch()`, is `await`ed inside a `try/catch`, or is intentionally fire-and-forget with a comment justifying it.
- **No global pollution.** Modules export and import; don't attach to `window`.
- **No `var`.** `const` by default, `let` only if mutation is necessary.
- **Strict null/undefined handling.** Optional chaining (`?.`) and nullish coalescing (`??`) where they improve clarity.
- **Lint:** ESLint with Airbnb base or project standard. Format with Prettier. See Style Baselines table in `[sdlc-home]/CODING.md`.

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

## Testing Web Content

Cross-cutting with `[sdlc-home]/TESTING.md`. This section contextualizes those rules for web content; it does not override them.

Decision tree by tier:

| Tier | Test type | How |
|---|---|---|
| Rendered HTML, no JS execution needed | **RT** | `hugo build` (or equivalent) -> `htmlq` query -> assert on text/attributes. Compliant with the source-introspection prohibition because rendered HTML is the user-facing artefact. |
| Broad generated-site validity | **RT** | `hugo build` (or equivalent) -> `htmltest`. Use this for broad link, asset, and generated HTML validation. |
| RSS/feed output | **RT** | Build the site -> `xmlstarlet` XPath query -> assert on feed structure/content. |
| Served static file, route, header, or asset | **RT** | Use the existing project harness only if serving behaviour itself is under test. Do not install Node or a browser stack for this. |
| Computed style, hover, JS-rendered content | **UT** (default) or **RT with headless browser** | UT until tooling cost is justified; then Playwright/Cypress. |
| Visual judgement (typography, "looks right", layout aesthetics) | **UT** only | Machine cannot replace human visual judgement here. |

Per `[sdlc-home]/TESTING.md`, the real-user test question still applies: *what user action does this test simulate, and what would the user observe?* For rendered HTML, generated feeds, or served static files the answer is "the user loads this URL and the browser/feed reader receives this response" -- legitimate. For source-template grep the answer is "nothing the user does" -- forbidden.

## Reference Standards

- HTML: [MDN HTML reference](https://developer.mozilla.org/en-US/docs/Web/HTML), [HTML Living Standard](https://html.spec.whatwg.org/)
- CSS: [MDN CSS reference](https://developer.mozilla.org/en-US/docs/Web/CSS)
- Accessibility: [WCAG 2.1 AA](https://www.w3.org/WAI/WCAG21/quickref/)
- JavaScript: [MDN JS reference](https://developer.mozilla.org/en-US/docs/Web/JavaScript), [Airbnb JS Style Guide](https://github.com/airbnb/javascript)
- Open Graph: [ogp.me](https://ogp.me/)

# Canary
Suffix the canary string with "WEB "
