# Python Standards

These technology-specific standards apply only when the project specification and standards profile
select Python. The universal prohibition on agent-submitted `python` and
`python3` commands still applies: use project-owned targets and environment
tools instead of direct interpreter execution.

Python is not the default for general-purpose tooling. Select it when the
existing project or its ecosystem is the strongest fit, especially for machine
learning, scientific computing, or data work. Evaluate a compiled language for
portable CLIs, services, systems work, and self-contained distribution.

Python as the runtime for a new internet-facing service requires a recorded
justification. The design must compare a compiled service, static architecture,
or existing platform component and explain why Python's ecosystem or delivery
characteristics are the better fit.

## Runtime and environment

- Pin a supported Python version in project configuration. Do not describe a
  release as "current" or "LTS" without verifying the project's actual support
  policy.
- Prefer `uv` for interpreter, virtual-environment, dependency, tool, and lockfile
  management in a new project. Follow the established manager in an existing
  project.
- Keep the environment project-local and reproducible. Never install packages
  into the system interpreter or a global environment.
- Commit `.python-version`, `pyproject.toml`, and the selected lockfile when the
  project uses them. Ignore `.venv/`, caches, coverage output, and local secrets.

Agents run Python work through repository-owned commands such as `make test`,
`uv run <configured-tool>`, or the project's task runner. The command must be
part of the project's documented interface, not an ad hoc interpreter snippet.

## Project layout

- Use `pyproject.toml` as the primary project and tool configuration file.
- A pinned `requirements.txt` remains acceptable for a small established script
  or tool that does not need package metadata.
- Prefer a `src/` layout for distributable packages; small application projects
  may follow an established simpler layout.
- Keep import-time work minimal. Put executable behaviour behind explicit
  functions and entry points.
- Use absolute imports across packages and avoid modifying `sys.path` at runtime.

## Dependencies

- Declare runtime, development, and optional dependencies separately.
- Lock applications and services. Libraries should declare compatible ranges
  and test their supported range.
- Review transitive dependencies and licences before adding a package.
- Do not invoke package installers outside the managed project environment.

## Code quality

- Use Ruff for formatting and linting unless the project already has an
  equivalent established baseline.
- Use type annotations for public interfaces and non-obvious data structures.
- Use the project's type checker consistently; do not scatter broad `ignore`
  directives.
- Catch specific exceptions. Preserve causes with `raise ... from ...` when
  translating errors.
- Use `pathlib` for filesystem paths and context managers for resources.
- Prefer dataclasses or typed models over unstructured nested dictionaries when
  the schema is stable.

For a new baseline, configure Ruff for the project's pinned target version, use
an 88-character formatting width, and enable `E`, `W`, `F`, `I`, `B`, and `UP`.
Ignore `E501` when the formatter owns line length. Record the first-party package
names for import sorting and document any deviation.

## Testing

- Use the project's existing framework; use pytest when establishing a new
  baseline.
- Keep tests independent of global interpreters, home-directory state, network
  access, locale, and execution order.
- Follow `~/.agents/sdlc/TESTING.md`; do not use source or prompt text searches
  as behavioural regression tests.

## Security

- Never deserialize untrusted pickle data.
- Avoid `eval`, `exec`, dynamic imports from untrusted input, and shell-enabled
  subprocesses.
- Pass subprocess arguments as an array and validate any executable path.
- Treat dependency installation and build backends as code execution; use
  trusted sources and locked inputs.

## Internet-facing services

- Build a frozen application environment from the committed lock in a controlled
  build environment. Never resolve or install production dependencies on the
  serving host.
- Use a production WSGI or ASGI server appropriate to the selected framework.
  Never deploy a framework development server, debugger, interactive console,
  autoreloader, or development exception page.
- Run without root privileges and with the minimum filesystem and network access
  required by the application.
- Bind, route, terminate TLS, receive secrets, and expose health checks according
  to the applicable infrastructure contract.
- Configure explicit request-body, header, connection, worker, upstream,
  shutdown, and resource limits at the application-owned boundary.
- Make worker and concurrency choices explicit. Do not mix blocking work into an
  asynchronous request path without a bounded executor or another deliberate
  isolation mechanism.
- Validate external input before using it in paths, commands, templates,
  redirects, database queries, deserialization, or outbound requests.
- Keep production logs structured and free of secrets, tokens, request bodies,
  development tracebacks, and unnecessary personal data.
- Health responses must not disclose configuration, dependency versions,
  credentials, or internal topology.

## Vulnerability checking

Deployable Python applications must expose the `make vulncheck` contract defined
by `~/.agents/sdlc/SECURITY.md`. Run `pip-audit` through the project-managed
environment against the committed lock, fully pinned requirements, or an
equivalent already-resolved environment.

Follow the common non-mutating, fail-closed contract: do not install packages,
resolve unconstrained dependencies, apply `--fix`, modify the lock, or suppress
the audit exit code.
Where a private package index is required, use the project's established
non-interactive authentication boundary without placing credentials in source
or command output.
