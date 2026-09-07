# Rust Standards

These technology-specific standards supplement the general coding standards in
`~/.agents/sdlc/CODING.md`. Also select
`~/.agents/sdlc/technologies/API.md` when the project provides or materially
consumes an API.

Primary references:

- [The Rust Programming Language](https://doc.rust-lang.org/book/)
- [The Cargo Book](https://doc.rust-lang.org/cargo/)
- [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)
- [Clippy](https://doc.rust-lang.org/clippy/)
- [rustfmt](https://github.com/rust-lang/rustfmt)
- [RustSec Advisory Database](https://rustsec.org/)

## Toolchain and reproducibility

- Use stable Rust unless a documented requirement needs nightly. Pin the
  toolchain in `rust-toolchain.toml` when reproducibility depends on an exact
  channel or component set.
- Declare the Rust edition and minimum supported Rust version using `edition`
  and `rust-version` in `Cargo.toml`. Test the declared minimum when it is a
  compatibility promise.
- Commit `Cargo.toml`. Commit `Cargo.lock` for applications, binaries,
  workspaces, and deployable artefacts. A published library must still declare
  accurate compatible dependency ranges because consumers do not inherit its
  lockfile.
- Never edit `Cargo.lock` manually. Dependency updates are deliberate source
  changes reviewed with their lockfile and feature-graph differences.
- Keep Cargo features additive and independently meaningful. Do not use feature
  flags to select mutually incompatible product identities or to conceal
  platform-specific forks.
- Test the supported feature combinations. Do not claim all-feature support
  merely because one default build passes.

## Project layout and typical architectures

Follow an established project layout. For a new project, use the smallest Cargo
structure that expresses its real boundaries:

```text
project/
|-- Cargo.toml
|-- Cargo.lock
|-- rust-toolchain.toml
|-- src/
|   |-- lib.rs
|   `-- main.rs
|-- tests/
`-- Makefile
```

- **Small CLI or service:** keep `main.rs` as composition and process-lifecycle
  code. Put behaviour that needs testing in `lib.rs` or cohesive modules.
- **Library:** expose a deliberately small public API from `lib.rs`; keep
  implementation modules private by default.
- **Service:** separate transport handlers, application operations, domain
  behaviour, and infrastructure adapters when they have distinct reasons to
  change. Do not impose a ceremonial layer for a small service.
- **Provider integration:** isolate each SDK or protocol behind a narrow adapter
  owned by the project. Domain code must not depend on provider request,
  response, pagination, or error types.
- **Workspace:** use a Cargo workspace only for genuinely distinct crates with
  independent compilation, reuse, deployment, or ownership boundaries. A
  workspace is not a substitute for ordinary modules.
- **FFI or systems boundary:** confine platform-specific and unsafe behaviour to
  a small module or crate with a safe project-owned interface.

Prefer feature-oriented modules where they keep related types and behaviour
together. Avoid generic `utils`, `helpers`, `common`, and `manager` modules that
collect unrelated responsibilities.

## Types and API design

- Follow the Rust API Guidelines for public and cross-module interfaces.
- Model valid states in types. Use enums for closed alternatives, newtypes for
  identifiers and validated values, and `Option` only where absence is valid.
- Accept borrowed values when the callee does not need ownership. Return owned
  values when that gives the caller a clear lifetime boundary. Do not expose
  complicated lifetimes merely to avoid a small justified allocation.
- Prefer standard conversion and borrowing traits such as `From`, `TryFrom`,
  `AsRef`, and `Borrow` over bespoke conversion methods.
- Implement common traits such as `Debug`, `Clone`, `Eq`, `Hash`, `Default`,
  `Display`, and `Error` when their semantics are honest and useful. Do not
  derive a trait that exposes secrets or implies an invalid ordering or default.
- Use traits at genuine substitution, testing, or integration boundaries. Do
  not create a trait for every struct or for a single implementation with no
  boundary value.
- Prefer static dispatch and ordinary generics until runtime polymorphism is a
  requirement. Use `dyn Trait` deliberately and keep object-safety constraints
  out of unrelated APIs.
- Mark public APIs `#[must_use]` where silently discarding the value is likely a
  defect. Document panics, errors, safety contracts, and non-obvious complexity.
- Keep public dependency types out of the project API unless adopting that
  dependency's contract is an explicit compatibility decision.

## Ownership, state, and concurrency

- Do not add `.clone()` merely to silence an ownership error. Decide which
  component owns the value and whether borrowing, moving, restructuring, or a
  justified shared owner expresses that decision.
- Do not default to `Arc<Mutex<_>>`. Prefer immutable values, message passing,
  scoped ownership, or a narrower synchronized state object. Document lock
  ordering where more than one lock can be held.
- Never hold a blocking mutex guard or other non-async resource guard across an
  `.await` point.
- Use async only when the workload and selected libraries require cooperative
  concurrency. Synchronous code is simpler for short CLIs, CPU-bound work, and
  linear file transformations.
- Give every spawned task an owner, cancellation path, error path, and join or
  shutdown policy. Detached tasks must be explicitly justified.
- Put timeouts around network and inter-process waits. Bound channels, queues,
  fan-out, retries, and concurrent tasks.
- Keep blocking filesystem, process, database, compression, and CPU-heavy work
  off async executor threads. Use the runtime's bounded blocking facility or a
  deliberate worker boundary.
- Prefer one async runtime per application. Do not create nested runtimes or
  call synchronous runtime entry points from asynchronous code.

## Errors and panics

- Use `Result` for operations that can fail and `Option` only for valid absence.
  Preserve the original cause and add context at boundaries.
- Prefer concrete typed errors for libraries, domain boundaries, and errors a
  caller may handle. `thiserror` is the established derive helper when manual
  implementations add no value.
- `anyhow` is appropriate at an application or CLI coordination boundary where
  callers need context and reporting rather than matching a public error type.
  Do not expose `anyhow::Error` as a reusable library contract.
- Do not use `unwrap`, `expect`, indexing, or `panic!` for recoverable external
  input or ordinary runtime failure. A proven invariant may use `expect` with a
  message that names that invariant.
- Do not convert errors to defaults, empty collections, or successful exit
  status unless that fallback is part of the documented behaviour.
- Panic is not an error-handling strategy. Library code must not panic for
  caller-controlled input.

## Unsafe Rust and FFI

- New projects forbid unsafe code through Cargo lint configuration unless the
  design explicitly requires it. Prefer a maintained safe abstraction over a
  project-owned unsafe implementation.
- Keep every necessary `unsafe` block minimal. Add an adjacent `SAFETY` comment
  that states the invariant the compiler cannot verify and why it holds.
- An `unsafe fn` or `unsafe trait` must document its caller or implementer
  obligations in a `# Safety` section.
- Wrap unsafe implementation behind a safe API that validates its preconditions.
  Audit the safe wrapper and all ways its invariant can be invalidated.
- Use Miri for applicable code that owns unsafe operations. Miri supplements
  reasoning and tests; a clean run does not prove all unsafe code sound.
- Validate lengths, alignment, ownership, lifetimes, nullability, encoding, and
  thread rules at FFI boundaries. Do not let raw pointers or foreign ownership
  conventions leak through the application.

## SDKs and external services

- Prefer the service provider's official maintained Rust SDK when it meets the
  requirement. Otherwise record why a protocol client, generated client, or
  third-party SDK is the safer choice.
- Pin the SDK through `Cargo.lock` and enable only required features. Review its
  default TLS, credential, telemetry, retry, timeout, and proxy behaviour.
- Construct reusable clients once and inject them into the owning adapter. Do
  not create a client per request or hide network access in constructors,
  getters, serialization, or global initialization.
- Keep credentials in the approved runtime secret boundary. Prefer the SDK's
  documented credential provider chain over copying secret values into project
  objects, logs, or error messages.
- Set explicit application-owned timeouts and retry limits. Retry only operations
  known to be idempotent or protected by an idempotency key.
- Handle pagination, rate limits, partial responses, streaming errors,
  cancellation, and provider-specific error classification where applicable.
- Convert SDK results into project-owned types at the adapter boundary. Test the
  conversion and error mapping without attempting to simulate an undocumented
  provider implementation.
- Use bounded one-off or user tests against a real service when contract evidence
  cannot be established locally. Follow the cost and external-action rules in
  `~/.agents/sdlc/TESTING.md`.

## Serialization, configuration, and data

- Use `serde` for established structured serialization formats. Parse into
  typed boundary structures, then validate semantic constraints before creating
  domain values.
- Separate wire, persistence, configuration, and domain types when their
  compatibility or validation rules differ.
- Do not use `serde_json::Value`, untyped maps, or strings as a permanent escape
  from modelling a stable schema.
- Treat deserialization as untrusted input. Bound input size and collection
  length, reject invalid values, and make unknown-field handling an explicit
  compatibility decision.
- Keep defaults visible and deterministic. Do not silently accept misspelled
  security-sensitive configuration fields.
- Never derive or implement `Debug`, `Display`, `Serialize`, or tracing fields
  in a way that exposes secrets or unnecessary personal data.

## Observability

- Use `tracing` and `tracing-subscriber` for structured diagnostics in async
  services and other applications that need spans and correlated events.
- Libraries emit events but do not install a global subscriber. The application
  owns filtering, formatting, and export.
- Do not hold a `tracing::Span::enter` guard across `.await`; instrument the
  future or async function instead.
- Record stable operation, request, and correlation fields. Never record secret
  material, authorization headers, full sensitive payloads, or unbounded data.
- Keep health and error responses free of internal paths, dependency versions,
  credentials, and topology.

## Testing

Follow `~/.agents/sdlc/TESTING.md` and the project's RT, UT, and OT definitions.

- Use Rust's built-in test harness and `cargo test` by default. Add another test
  runner only for a demonstrated project need.
- Keep focused unit tests beside private behaviour and public integration tests
  under `tests/`. Use documentation tests for public examples that should remain
  executable.
- Test behaviour through typed interfaces. Do not grep Rust source, manifests,
  generated code, or fixtures as proof of runtime behaviour.
- Test supported feature and platform boundaries explicitly. A default-feature
  test does not establish no-default-feature or optional-feature support.
- Use property testing for parsers, state machines, and broad invariant spaces
  when generated cases add evidence that examples cannot provide economically.
- Make async tests deterministic. Control clocks and external boundaries where
  timing matters; never replace synchronization with fixed sleeps.
- Use Criterion or another established harness only for a stated performance
  contract. Do not turn unstable microbenchmarks into behavioural gates.

## Tooling and quality gates

Use the project's existing configuration first.

| Tool | Purpose | Required |
|---|---|---|
| `cargo fmt` / rustfmt | Canonical formatting | Yes |
| `cargo clippy` | Rust-aware linting | Yes |
| `cargo test` | Unit, integration, and documentation tests | Yes |
| `cargo audit` | RustSec advisory scan of resolved dependencies | Yes for deployable applications |
| `cargo deny` | Licence, source, ban, and advisory policy | When the project needs explicit supply-chain policy |
| Miri | Undefined-behaviour checking | When project-owned unsafe code is applicable |

Run formatting checks across the workspace. Run Clippy across all maintained
targets and the supported feature matrix, treating warnings as failures. Do not
enable `clippy::pedantic`, `clippy::nursery`, or automatic fixes wholesale;
select additional lints deliberately and justify narrow suppressions beside the
code.

Deployable Rust applications must expose the `make vulncheck` contract defined
by `~/.agents/sdlc/SECURITY.md`. Run `cargo audit` against the committed
`Cargo.lock`, or use the RustSec advisory check in an established `cargo deny`
policy. The target is non-mutating and fail-closed: it must not update the
advisory exception policy, dependencies, or lockfile, install a scanner, or
suppress findings. Use `cargo deny` additionally when licence, allowed source,
duplicate-version, or banned-crate policy is required; those checks do not
replace behavioural tests.

Use this baseline target when `cargo audit` is the selected scanner:

```make
.PHONY: vulncheck
vulncheck:
	@command -v cargo-audit >/dev/null 2>&1 || { \
		printf '%s\n' 'cargo-audit is required for make vulncheck' >&2; \
		exit 1; \
	}
	@test -f Cargo.lock || { \
		printf '%s\n' 'Cargo.lock is required for make vulncheck' >&2; \
		exit 1; \
	}
	cargo audit --file Cargo.lock
```

The target deliberately does not install `cargo-audit` or generate a lockfile.
`cargo audit --file Cargo.lock` checks the named resolved dependency graph and
returns failure for reported vulnerabilities. Record narrowly approved advisory
exceptions in the repository's Cargo Audit configuration under the exception
contract in `~/.agents/sdlc/SECURITY.md`; never add an inline `--ignore` merely
to make the target pass. A workspace normally has one root `Cargo.lock`; add an
explicit scan for every other committed lockfile that contributes to a deployed
artefact.

## Established ecosystem libraries

Prefer the standard library and existing project dependencies when they meet the
need. The following are common, mature reach-fors, not a permanent allowlist or
a reason to add a dependency without reviewing its current maintenance,
licence, features, transitive graph, minimum Rust version, and security history.

| Concern | Established choices | Boundary |
|---|---|---|
| Serialization | `serde`, `serde_json` | Validate after parsing; select only required formats and features. |
| CLI parsing | `clap` | Keep business logic outside argument-parser types. |
| Async runtime | `tokio` | Use only when async is justified; enable required features rather than `full` by habit. |
| Structured diagnostics | `tracing`, `tracing-subscriber` | Applications install subscribers; libraries do not. |
| Typed errors | `thiserror` | Prefer for reusable and matchable error contracts. |
| Application error context | `anyhow` | Keep at executable coordination boundaries. |
| HTTP client | `reqwest` | Reuse a configured client; set timeouts and deliberate TLS, redirect, and proxy policy. |
| HTTP service | `axum`, `tower`, `tower-http` | Appropriate for Tokio services; keep transport concerns outside domain logic. |
| TLS | `rustls` | Prefer when its platform and protocol support fit; never disable certificate or hostname verification. |
| URLs and identifiers | `url`, `uuid` | Parse and validate rather than manipulating structured values as strings. |
| SQLite | `rusqlite` | Keep SQL parameterized and migrations explicit. |
| Async SQL | `sqlx` | Select supported drivers and runtime features narrowly. |
| ORM and query builder | `diesel` | Use when its compile-time model offsets its abstraction and migration cost. |

Choose one coherent crate per concern. For dates and times, follow the existing
project choice between `time` and `chrono`; do not introduce both casually. For
cryptography, use a maintained protocol-level library or provider SDK instead of
assembling primitives. A security-sensitive crate requires specific review even
when it is widely used.

## Anti-patterns

Block or redesign:

- repeated cloning, leaking values, or broad lifetime extension used only to
  appease the borrow checker;
- `Arc<Mutex<_>>` as default application architecture;
- async functions, boxed futures, or spawned tasks where synchronous control
  flow is sufficient;
- blocking work on async executor threads or locks held across `.await`;
- detached tasks without cancellation, error reporting, shutdown, and ownership;
- `unwrap`, `expect`, indexing, or panics on external or recoverable input;
- broad `allow` attributes, disabled Clippy groups, or warnings hidden to make a
  gate pass;
- project-owned unsafe code where a maintained safe implementation exists;
- unsafe blocks without precise invariants and a safe containing abstraction;
- provider SDK types, generated transport types, or ORM models used as domain
  types throughout the application;
- one-implementation traits, excessive generic abstraction, and macros that hide
  straightforward control flow;
- a multi-crate workspace for one cohesive component;
- uncontrolled default features or mutually incompatible feature combinations;
- build scripts or procedural macros that perform network access, depend on
  undeclared host state, or generate non-reproducible output;
- handwritten parsers for established formats, SQL string interpolation, shell
  command strings, or regex parsing of nested structured data;
- exposing a development server, binding publicly by default, terminating TLS
  contrary to the infrastructure contract, or disabling TLS validation; and
- compiling or resolving dependencies on a production host when a controlled
  build can deploy a pinned artefact.
# Canary

Suffix the canary string with "RUST "
