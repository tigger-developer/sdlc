# Node.js and npm Standards

Node.js as an application runtime is discouraged. Selecting it for a new
application requires a recorded justification, including the alternatives
considered and why a simpler compiled or static architecture is inadequate.
Incidental use of npm-managed development tooling does not by itself select
Node.js as the application runtime.

These standards supplement `~/.agents/sdlc/CODING.md` and, for web interfaces,
`~/.agents/sdlc/technologies/WEB.md`.

## Runtime and package manager

- Pin a supported Node.js release in project configuration and deployment
  inputs.
- Select one package manager. Record and pin its version through the project's
  established mechanism.
- Commit exactly the selected package manager's lockfile. Do not retain
  competing lockfiles.
- Use frozen, lockfile-respecting installation such as `npm ci`; never resolve
  unconstrained production dependencies during build or deployment.
- Do not install packages globally or directly on the production host.
- Separate runtime, development, and optional dependencies accurately.
- Use ES modules for new code unless an established project contract requires
  CommonJS.

## Dependency discipline

Every direct dependency must have a coherent purpose that the platform or a
small project-owned implementation cannot reasonably provide. Before adding one,
review its maintenance activity, ownership, release history, transitive graph,
licence, install scripts, and alternatives.

Treat package installation and lifecycle scripts as arbitrary code execution.
Use `--ignore-scripts` when the project does not require lifecycle scripts. When
scripts are required, record which packages run them and why. Never add a
package solely to avoid writing a small, clear function.

Do not run automatic dependency-fix commands in a build, audit, or deployment
target. Dependency updates are reviewed source changes with lockfile diffs and
normal verification.

## Internet-facing services

- Use a production server entry point; never deploy a development server,
  inspector, hot reload, or an interactive console.
- Run without root privileges and with the minimum filesystem and network
  access required by the application.
- Bind, route, terminate TLS, receive secrets, and expose health checks according
  to the applicable infrastructure contract.
- Set explicit request-body, header, connection, upstream, shutdown, and
  resource limits at the application-owned boundary.
- Handle rejected promises and uncaught exceptions deliberately. Do not keep a
  process alive after its state may be inconsistent.
- Validate external input before using it in paths, commands, templates,
  redirects, database queries, or outbound requests.
- Do not use `eval`, `new Function`, string-based timers, shell-enabled child
  processes, or dynamically selected modules from untrusted input.
- Keep production logs structured and free of secrets, tokens, request bodies,
  and unnecessary personal data.
- Do not publish source maps, development manifests, package caches, lockfile
  credentials, or server source with the production artefact unless the design
  explicitly requires them.

## Testing and quality

- Follow `~/.agents/sdlc/TESTING.md` and the project's established test runner.
- Keep tests independent of global packages, home-directory state, live
  registries, network access, locale, and execution order.
- Use ESLint and a project formatter. Do not weaken security or correctness
  rules merely to pass existing code.
- Test process startup, graceful shutdown, input limits, failed dependencies,
  and authorization boundaries proportionately for an internet-facing service.

## Vulnerability checking

Expose the `make vulncheck` contract defined by
`~/.agents/sdlc/SECURITY.md`. Use the selected package manager's audit command
against the committed lockfile. Include dependencies that can alter the build
artefact, not only packages present at runtime.

The target must not run `npm audit fix`, install packages, update the lockfile,
or suppress the audit exit code. If registry disclosure of the dependency graph
is not acceptable, select and document Trivy, OSV-Scanner, or another compatible
local scanner rather than silently omitting the check.

## References

- [Node.js security best practices](https://nodejs.org/learn/getting-started/security-best-practices)
- [npm dependency selectors and dependency types](https://docs.npmjs.com/specifying-dependencies-and-devdependencies-in-a-package-json-file/)
- [npm audit](https://docs.npmjs.com/cli/v11/commands/npm-audit/)
