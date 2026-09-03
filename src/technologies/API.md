# API Standards

These standards apply when a project provides or consumes a machine interface,
including HTTP APIs, webhooks, remote procedure calls, and service integrations.
They supplement `~/.agents/sdlc/CODING.md`, `~/.agents/sdlc/SECURITY.md`, and
`~/.agents/sdlc/TESTING.md`. Browser-facing documents and interfaces also follow
`~/.agents/sdlc/technologies/WEB.md`.

The HTTP rules below are the default for HTTP and JSON interfaces. Other
protocols must provide equivalent, explicit semantics rather than imitating HTTP
where its methods and status codes do not apply.

## Contract and ownership

- Define the provider, intended consumers, authority for the contract, and
  responsibility for operation and support.
- Specify requests, responses, authentication, authorization, errors, limits,
  compatibility, versioning, deprecation, and relevant service expectations
  before implementation.
- Cross-project and externally consumed HTTP APIs must have a version-controlled
  machine-readable contract such as OpenAPI. Use the established equivalent for
  another protocol. A small project-local interface may use a concise durable
  specification when a separate schema would add no useful assurance.
- Treat generated documentation, clients, and validators as views or products of
  the contract. They do not replace a clear specification or human review.
- Keep transport objects at the boundary. Do not expose persistence models,
  internal error types, or incidental implementation structure as the contract.
- Record data classification, retention, deletion, and cross-system ownership
  where the interface carries personal, confidential, or durable data.
- Make environmental responsibilities explicit. An application contract must
  not silently assume that routing, TLS, credentials, service discovery, or
  availability are owned by another project.

## Provider responsibilities

- Follow the protocol's defined semantics. For HTTP, safe methods must not cause
  requested state changes, and repeated idempotent requests must have the same
  intended effect as one request.
- Specify retry and replay behaviour. Where a non-idempotent operation supports
  an idempotency key, define its scope, lifetime, conflict behaviour, and stored
  result. A header name alone is not an idempotency design.
- Use status codes and content types consistently. Do not return a successful
  status for a failed operation merely because the response body describes an
  error.
- Use a stable, machine-readable error contract. For HTTP APIs without a suitable
  established format, prefer RFC 9457 problem details. Human-readable error text
  must not be the field consumers need to parse.
- Validate authentication and authorization separately. Enforce access for the
  specific object, action, field, and tenant on every request rather than relying
  on possession of a valid identity alone.
- Validate request type, schema, encoding, size, range, and multiplicity at the
  boundary. Define whether unknown fields are rejected or tolerated.
- Bound collection responses. Specify pagination, maximum page size, filtering,
  stable ordering, and continuation behaviour. Continuation tokens should be
  opaque unless their representation is itself contractual.
- Protect updates from lost writes where concurrent modification matters. Use an
  explicit version, validator, precondition, or equivalent conflict mechanism.
- Define rate-limit and quota behaviour, including how a consumer distinguishes
  throttling from permanent rejection and when it may retry. Use `Retry-After`
  where its HTTP semantics fit.
- For asynchronous operations, expose a durable way to observe completion,
  failure, cancellation, and expiry. An accepted request is not evidence that
  the operation completed.

## Consumer responsibilities

- Isolate each external API behind a narrow adapter owned by the consuming
  project. Do not spread provider-specific transport objects or error formats
  through domain code.
- Configure endpoints, credentials, and selected contract versions through the
  project's approved configuration boundary. Never embed secrets or production
  endpoints in source code or tests.
- Set an overall deadline and bounded phase timeouts. Cancellation from the
  calling operation must propagate through the request.
- Retry only failures classified as transient and operations known to be safe to
  repeat. Bound attempts and elapsed time, use backoff with jitter, and honour
  authoritative server retry guidance.
- Validate the response status, content type, size, schema, and required semantic
  fields before trusting it. Treat a syntactically valid but contractually
  incomplete response as failure.
- Distinguish transport failure, timeout, authentication failure, authorization
  failure, throttling, provider rejection, malformed response, and local
  cancellation where the caller can act differently.
- Consume every required page deliberately and guard against repeated or cyclic
  continuation tokens. Do not silently return a partial collection as complete.
- Define caching, stale-data, and fallback behaviour in the specification. Never
  substitute old, default, or alternate-provider data while reporting current
  success unless that behaviour is an explicit part of the contract.
- Do not depend on undocumented response fields, field order, presentation text,
  timing accidents, or permissive behaviour observed from one provider version.

## Compatibility and lifecycle

- Select and document a compatibility strategy; a version number in a path or
  header is not a complete strategy.
- Prefer additive evolution. Providers must not silently remove or rename fields,
  narrow accepted inputs, change meanings, repurpose enum values, or alter error
  classification for existing consumers.
- Consumers should tolerate additional response fields while continuing to
  validate the fields and invariants they require.
- Breaking changes require a deliberate new contract, migration guidance, a
  defined coexistence or transition period, and an observable retirement plan.
- Deprecation must identify the affected operation or field, replacement,
  consumer action, support window, and removal condition. Do not retain dead
  versions indefinitely without an explicit decision.
- Preserve contract history. A current schema alone is insufficient evidence for
  understanding a compatibility regression or previously supported consumer.

## Webhooks and callbacks

- Treat inbound webhooks as untrusted API requests. Authenticate the sender,
  verify integrity before parsing, and enforce a bounded replay window.
- Assume duplicate delivery and define idempotent processing. Record which event
  identifier or business key establishes uniqueness and how long it is retained.
- Do not assume ordering unless the provider contract guarantees it. Define how
  delayed, missing, repeated, or out-of-order events are reconciled.
- Acknowledge only according to the documented delivery contract. Define retry,
  backoff, terminal failure, and operator recovery on both sides.
- Do not place credentials or sensitive payload data in callback URLs. Validate
  callback destinations against the approved network and access boundary.

## Observability and operation

- Carry a safe request or correlation identifier across the boundary and return
  it in errors where it helps support without exposing internals.
- Record latency, request volume, error classification, throttling, and saturation
  at a level proportionate to the service. Do not rely on raw payload logging for
  observability.
- Redact credentials, authorization headers, cookies, personal data, and
  confidential request or response fields from logs and diagnostics.
- Define health and readiness from the service's real ability to accept work.
  Do not report readiness merely because a process exists.
- Bound failure amplification. Retries, fan-out, polling, callbacks, and circuit
  breakers must not turn a dependency failure into an avoidable outage.

## Testing API contracts

- Test providers through their request and response boundary and consumers
  through their adapter boundary. Internal handler or serializer tests do not by
  themselves establish API compatibility.
- Verify conformance with the machine-readable contract where one exists, while
  retaining behavioural assertions for semantics a schema cannot express.
- Cover authentication, object and action authorization, malformed and oversized
  input, error contracts, pagination, concurrency, throttling, timeout, retry,
  replay, idempotency, and partial dependency failure where applicable.
- Test provider and consumer compatibility in the direction each project owns.
  A contract-faithful fake may isolate a third-party boundary; it must not encode
  assumptions absent from the provider's published contract.
- Do not put calls to metered or third-party hosted APIs in the regression pack.
  Use bounded one-off verification and record it as required by
  `~/.agents/sdlc/TESTING.md`.
- Do not claim behavioural coverage by grepping an OpenAPI document, source code,
  routes, fixtures, or configuration. Schema linting and generated-code checks
  are useful static checks, not substitutes for request-response evidence.

## Antipatterns

Reject or redesign:

- undocumented endpoints or implementation code treated as the only contract;
- blanket successful responses containing an error-shaped body;
- consumers parsing human-readable error prose;
- stack traces, query details, credentials, or internal identifiers in errors;
- authentication without object-level, action-level, field-level, or tenant-level
  authorization where those boundaries exist;
- unbounded requests, uploads, collections, fan-out, polling, or pagination;
- retries on every failure, retries without deadlines, or retries of operations
  whose repeatability has not been established;
- breaking changes without versioning, migration, deprecation, and consumer
  coordination;
- consumers coupled to undocumented fields, ordering, formatting, or permissive
  provider accidents;
- transport or persistence models used directly as the domain model;
- webhook handling without integrity checks, replay protection, deduplication,
  and failure recovery;
- mocks that reproduce the current implementation rather than the published
  boundary; and
- a bespoke client, parser, authentication scheme, or error format where a
  maintained standard implementation meets the documented need.

## References

- [RFC 9110: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html)
- [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457.html)
- [OpenAPI Specification](https://spec.openapis.org/oas/latest.html)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
