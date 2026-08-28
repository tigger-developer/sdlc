# Coding Standards

## Readability and fit

- Optimize for the next maintainer. Prefer explicit names and straightforward
  control flow over compressed cleverness.
- Match surrounding code and established project conventions unless the plan
  records a deliberate change.
- Use the existing language and framework by default. For new work, choose the
  simplest maintainable technology that fits deployment, performance,
  portability, security, and team constraints.
- Keep functions focused, dependencies visible, and configuration separate from
  logic.
- Prefer pure functions and immutable data where they improve clarity.
- Do not name code `new`, `improved`, `enhanced`, or after the change that
  introduced it. Names and comments must remain accurate over time.

## Style and tooling

Use the project's formatter, linter, type checker, and static-analysis tools.
Where the project has no baseline, adopt the language community's current
standard and record it in project configuration.

| Language | Default baseline |
|---|---|
| Shell | Google Shell Style Guide, ShellCheck, shfmt |
| Python | PEP 8, Ruff |
| Perl | `perlstyle`, Perl::Critic, perltidy |
| Swift | Swift API Design Guidelines, SwiftLint, swift-format |
| Go | Effective Go, `gofmt`, `go vet`, golangci-lint |
| Rust | Rust API Guidelines, Clippy, rustfmt |
| Java or Kotlin | Established Google or Kotlin conventions and project tooling |
| JavaScript or TypeScript | Project standard or Airbnb, ESLint, Prettier |
| HTML and CSS | Semantic HTML, WCAG AA, htmltest, stylelint |
| Other | Dominant community standard and maintained project tooling |

Warnings are work to diagnose, not output to hide. A narrow suppression is
acceptable only when the underlying condition is understood, the specific rule
is named, and an adjacent comment explains why the code is safe. Broad or
file-wide suppression is prohibited.

Technology-specific standards live under `~/.agents/sdlc/technologies/`. Load
only those selected by the project's standards profile.

## Comments and documentation

- Explain purpose, invariants, constraints, and surprising decisions; do not
  narrate obvious syntax.
- Preserve accurate existing comments. Remove commented-out dead code rather
  than treating it as documentation.
- Correct or replace a false comment when behaviour changes; do not leave it to
  mislead the next reader.
- Follow an established file-header convention. If the project has none, each
  source file begins with a concise two-line `ABOUTME` header describing its
  purpose where the language permits it. Record any project-wide deviation in
  the constitution.
- Public APIs, exported symbols, configuration keys, and operational interfaces
  require durable documentation.

## Errors and failure

- Handle errors at the level that can add context or recover meaningfully.
- Preserve the original cause when wrapping an error.
- Fail explicitly on invalid or unsafe state. Do not convert failure to success,
  discard stderr, catch every exception and continue, or temporarily disable
  error handling.
- Anticipated failure may be handled only with a specific branch, useful context,
  and a defined fallback or result.
- Never use `|| true`, empty catch blocks, unchecked return values, or equivalent
  suppression to make checks appear successful.
- Error messages must say what failed, identify the relevant input or operation
  safely, and suggest recovery when the caller can act.
- For CLI programs, use exit status 0 for success, 1 for an operational failure,
  and 2 for invalid invocation unless an established interface says otherwise.
- Never use a fixed sleep as synchronization. Poll an observable readiness
  condition with a timeout.

## Structure and embedded languages

- Keep modules cohesive and interfaces small.
- Avoid hidden global state and action at import or initialization time.
- Treat 50 lines per function and three levels of nesting as review tripwires;
  justify or decompose code that exceeds them.
- Inject clocks, randomness, filesystem roots, networks, and other volatile
  dependencies when deterministic behaviour matters.
- Never assemble one language as an unparseable string inside another when a
  format-aware API or file template exists. Use parameterized queries,
  structured serializers, and appropriate template engines.
- Avoid speculative abstraction. Extract shared code when the shared contract is
  understood, not merely because two blocks look similar.

## Data handling

- Parse structured data with a format-aware parser, not regular expressions or
  line splitting.
- Validate external input at trust boundaries. Prefer allowlists, explicit
  schemas, maximum sizes, and known encodings.
- Preserve unknown fields when the format promises forward compatibility.
- Write durable state atomically where partial output would be harmful.
- Make migrations restartable, observable, and reversible where feasible.
- Define ownership, retention, and deletion behaviour for personal or sensitive
  data.

## Security

- Apply least privilege to files, processes, tokens, users, and network access.
- Never commit, log, display, or embed credentials, tokens, private keys, or
  personal data.
- Keep local secrets in ignored environment files or the project's approved
  secret store; use platform secret facilities in CI and production.
- Prevent path traversal and unsafe archive extraction. Resolve and validate
  paths against their allowed root before use.
- Use parameterized database queries. Never concatenate untrusted SQL.
- Escape at the output boundary for the destination context; HTML, URLs, shell,
  SQL, JSON, and regular expressions have different rules.
- Use secure defaults. A user must opt into weaker behaviour explicitly, with
  the risk documented.
- Do not widen data or system access without explicit human instruction.
- Never disable TLS verification or use world-writable permissions such as
  `chmod 777`.

For containers, pin base images, run as a non-root user, avoid privileged mode,
and scan the resulting image with a maintained vulnerability scanner.

Never modify shell startup files, operating-system configuration, global
package-manager state, or service configuration without explicit authority for
the exact target.

## Output and logging

- Keep normal CLI output concise and actionable. Send diagnostics to stderr.
- Provide a documented verbose mode for full detail rather than printing every
  unchanged item by default.
- Use stable machine-readable output only when it is part of the interface;
  version that interface when compatibility matters.
- Use appropriate log levels and structured fields where logs are ingested.
- Include operation and correlation context. Never log secrets or unnecessary
  personal data.
- Do not use colour as the only carrier of meaning; respect non-interactive and
  no-colour environments.
- Use ISO 8601 for machine-readable dates and timestamps unless an established
  external interface requires another format.

## Files and networks

- Use platform temporary-directory facilities; do not assume a hard-coded
  temporary path.
- Set deliberate permissions when creating sensitive files.
- Use atomic replacement for configuration and durable state when readers could
  observe partial writes.
- Put timeouts on network calls. Bound retries, back them off, and retry only
  operations safe to repeat.
- Use exponential backoff with jitter and a circuit breaker where repeated
  failure could overload a dependency.
- Validate status, content type, size, and integrity before trusting downloaded
  content.
- Do not add remote access, telemetry, uploads, or external services without the
  specification's authority.
- Before designing or changing external deployment, read the project's
  infrastructure integration contract. Report contract or tooling gaps instead
  of compensating with hidden project-side behaviour.

## Dependencies

- Prefer the standard library and existing dependencies when they meet the need.
- Review a new dependency's maintenance, licence, security history, transitive
  graph, binary size, and operational cost.
- Pin applications through the ecosystem's lock or checksum mechanism. Libraries
  may use compatible ranges according to ecosystem convention.
- Commit reproducibility files, including `go.sum` for Go modules.
- Remove unused dependencies and respond promptly to relevant security fixes.
- Do not install packages globally or outside the project's managed environment.
- Run the ecosystem's maintained dependency and vulnerability audit tooling as
  part of the project verification strategy.

## Repository entry points and packaging

- Provide stable repository-owned entry points for build, test, lint, and
  installation. A Makefile is preferred when it fits the project and no
  ecosystem-native task interface already exists.
- Keep help text reviewable as documentation and make `--help` and `--version`
  available on CLI executables. Provide dry-run behaviour for operations whose
  effects merit preview.
- For Homebrew releases, update the formula or cask version, source URL, and
  checksum together. Install or upgrade through Homebrew to validate packaging;
  do not live-patch a package-manager installation.

## Configuration and compatibility

- Validate configuration before acting and identify invalid fields precisely.
- Define precedence between defaults, files, environment, and command-line
  values.
- Preserve backward compatibility unless the specification explicitly changes
  it. Provide migration guidance for intentional breaks.
- Keep platform assumptions visible. Test or document supported operating
  systems, architectures, shells, runtimes, and service versions.

## Review red flags

Block or redesign code that introduces:

- hidden failure or disabled verification;
- injection, traversal, secret exposure, or widened access;
- unbounded input, recursion, concurrency, retries, or resource use;
- irreversible mutation without a specified recovery path;
- a new architecture or public contract absent from the specification;
- tests coupled only to implementation details; or
- claims of compatibility or correctness unsupported by evidence.

Require explicit justification for new dependencies, global state, duplicated
logic, complex concurrency, custom parsers, test-only production interfaces,
and exceptions to established project conventions.

Also investigate commented-out code, TODO or FIXME markers without a durable
descriptor, magic numbers, boolean control parameters, circular dependencies,
god objects, monkey-patching, reflective access bypasses, multiple complex
return paths, tests without meaningful assertions, and temporary fixes without
a removal condition.
