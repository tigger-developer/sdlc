# Session recovery one-off checks

These checks are **OT**, not regression-pack entries. `testdata/` is excluded
from `go test ./...`. They use local CLI doubles and never contact a provider.
Run them explicitly with a temporary Go overlay mapping
`cmd/sdlc-harness/recovery_ot_test.go` to the absolute path of
`testdata/recovery_test.go`, then select
`TestRecordedSessionRecovery|TestFallbackConfigOneOff|TestHermesEvidenceTransportOneOff|TestAuditCacheOneOff|TestCacheIdentityAndIncidentsOneOff`
in `./cmd/sdlc-harness`.
Go's `-overlay` flag takes a JSON object with a `Replace` path mapping.

The checks cover context replacement, history/ownership, round limits, retained
fallback reuse, configuration precedence, verified inline evidence, cache hits,
cache invalidation and non-recoverable failures. Native
provider qualification is separate, bounded, metered OT and is never invoked
by these checks, build targets or CI.
