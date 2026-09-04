# JavaScript and TypeScript Standards

These standards apply whenever a project owns JavaScript or TypeScript source,
including libraries, command-line tools, browser code, and host-application
plugins. They supplement `~/.agents/sdlc/CODING.md`.

Also select:

- `~/.agents/sdlc/technologies/NODE.md` when Node.js is the runtime or an
  npm-managed dependency can affect the built artefact; and
- `~/.agents/sdlc/technologies/WEB.md` for browser-facing HTML, CSS, DOM, or
  accessibility behaviour.

## Runtime profiles

- **Web applications and sites:** Apply the common rules in this document to
  source code and add `~/.agents/sdlc/technologies/WEB.md` for browser, DOM,
  accessibility, rendered-output, and visual-validation concerns.
- **Host-application plugins:** Apply the common rules in this document and the
  host-plugin lifecycle rules below. Add `WEB.md` only when the plugin owns
  browser-facing interface or stylesheet behaviour.
- **Node.js applications and build systems:** Apply the common rules in this
  document and add `~/.agents/sdlc/technologies/NODE.md` for runtime,
  package-manager, dependency, server, and vulnerability concerns.

## Language and type safety

- Prefer TypeScript for new multi-file applications, libraries, and plugins.
  Plain JavaScript remains valid for a small script or an established project
  whose build contract does not use TypeScript.
- Enable TypeScript's `strict` mode for new projects. Record and justify any
  disabled strictness option.
- Treat parsed JSON, messages, storage, host APIs, and network responses as
  untrusted boundaries. Narrow `unknown` through validation; do not replace
  validation with `any`, a type assertion, or a non-null assertion.
- Define explicit types at public, persistence, network, and host-API
  boundaries. Allow inference for clear local implementation details.
- Pin the supported ECMAScript target, module format, and host versions. Do not
  emit syntax or APIs unsupported by the declared runtime.

## Modules, state, and ownership

- Use ECMAScript modules for new code unless the runtime or build contract
  requires CommonJS.
- Use `const` by default and `let` only for necessary reassignment. Never use
  `var`.
- Keep mutable state owned by the narrowest practical component. Do not attach
  project state to global objects or rely on ambient mutable singletons.
- Keep entry points thin. Separate host adapters, domain behaviour, persistence,
  and presentation when they can change or be tested independently.
- Give every listener, timer, subscription, observer, worker, and disposable
  resource an explicit owner and teardown path.

## Asynchronous work and errors

- Every promise must be awaited, returned, explicitly combined, or deliberately
  detached with documented error handling.
- Prefer `async` and `await` for sequential control flow. Use promise combinators
  where concurrency is intentional and their failure semantics are understood.
- Apply cancellation, timeouts, and bounded concurrency to work that waits on
  external systems or can outlive its caller.
- Do not catch and discard errors. Add useful context, preserve the original
  cause where supported, and expose failure through the owning interface.
- Never leak secrets, private content, tokens, or unnecessary personal data in
  exceptions, console output, telemetry, or diagnostics.

## Input and execution safety

- Validate external values before using them in paths, commands, selectors,
  URLs, HTML, storage, or dynamic module selection.
- Never use `eval`, `new Function`, string-based timers, or code generated from
  untrusted input.
- Use safe DOM construction and contextual escaping. Never insert untrusted
  content through `innerHTML` or an equivalent raw-HTML sink.
- Do not pass untrusted values to a shell-enabled child process. Prefer a direct
  executable invocation with an argument array when process execution is part
  of the approved design.
- Keep secrets out of browser bundles and distributed plugin artefacts. A value
  embedded in client-side JavaScript is not secret.

## Host-application plugins

- Prefer the host's documented public API to internal APIs, direct filesystem
  access, or assumptions about its private storage layout.
- Register events, timers, views, commands, and other resources through the
  host's lifecycle facilities, and release them when the plugin unloads.
- Make enable, disable, reload, and repeated initialization safe. Do not leave
  duplicate handlers, timers, UI elements, or persistent mutations behind.
- Develop and verify mutating plugins against isolated test data rather than a
  user's primary workspace or vault.
- Declare platform restrictions accurately. Capability-check optional desktop
  or host features instead of allowing an unsupported environment to fail
  unpredictably.
- Treat network access, telemetry, access outside host-managed data, and remote
  content as explicit product and privacy decisions requiring documentation.

For Obsidian plugins, use its `Plugin` and `Component` lifecycle registration,
prefer the Vault and FileManager APIs to storage internals, keep
`manifest.json` compatibility fields accurate, and follow the current community
plugin policies. A plugin using Node.js or Electron APIs must declare itself
desktop-only.

## Tooling, dependencies, and builds

- Use ESLint's current flat configuration and a project formatter. TypeScript
  projects should use type-aware recommended rules where the project can sustain
  their cost.
- Run the TypeScript compiler as a no-emit correctness check separately from
  bundling. Linting, type checking, tests, and production bundling must report
  failures through non-zero exit status.
- Commit the selected package manager's lockfile and build from it
  reproducibly. Apply `~/.agents/sdlc/technologies/NODE.md` to the dependency
  graph and lifecycle scripts.
- Treat generated bundles as build artefacts. Do not hand-edit them or use them
  as the canonical source.
- Minimize dependencies, bundler plugins, polyfills, and runtime shims. Each one
  expands the executable supply chain and must have a current purpose.

## Testing and vulnerability checking

- Follow `~/.agents/sdlc/TESTING.md`. Test observable behaviour through public
  boundaries; do not treat source-text searches as behavioural tests.
- Keep domain logic testable without booting the browser, host application, or
  network. Test the adapter boundary separately where integration risk warrants
  it.
- Use user tests for host-specific interaction or visual judgement that cannot
  be represented honestly by a stable automated test.
- Distributed or deployable projects with package-managed dependencies must
  expose the non-mutating `make vulncheck` contract in
  `~/.agents/sdlc/SECURITY.md`.

## References

- [TypeScript strict mode](https://www.typescriptlang.org/tsconfig/strict)
- [typescript-eslint shared configurations](https://typescript-eslint.io/users/configs/)
- [ESLint configuration](https://eslint.org/docs/latest/use/configure/)
- [MDN JavaScript reference](https://developer.mozilla.org/en-US/docs/Web/JavaScript)
- [Obsidian plugin lifecycle](https://docs.obsidian.md/plugins/guides/lifecycle-management)
- [Obsidian plugin policies](https://docs.obsidian.md/community-directory/developer-policies)
- [Obsidian plugin submission requirements](https://docs.obsidian.md/community-directory/submission-requirements-for-plugins)
