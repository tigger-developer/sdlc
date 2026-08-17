# Go Standards

Go-specific standards. The general coding standards in `[sdlc-home]/CODING.md` apply on top of these.

For Go applications that serve HTML/CSS/JavaScript (lean web apps, Hugo-adjacent services), the patterns here cover the server side; the response payload itself follows `[sdlc-home]/WEB.md`.

## Project Layout

Standard module layout for an application:

```
project/
├── cmd/<app>/main.go     # entry point per binary
├── internal/             # private packages (cannot be imported externally)
├── pkg/                  # optional: public packages for external library use
├── go.mod
├── go.sum
└── Makefile
```

Single-binary projects may keep `main.go` at the repo root. Avoid `pkg/` unless the code is genuinely meant for external consumption -- premature `pkg/` is a smell.

For deployable applications, the project also contains `config/`, `secrets/`, and `dependencies/` directories; conventions for those live with the deployment tooling, not here.

## Configuration

Follow [12-factor](https://12factor.net/config) -- runtime configuration comes from environment variables, not flags or build-time constants. Hardcoded paths to config or secret files are a smell; let the operator point at them via environment variables.

```go
addr := os.Getenv("ADDR")
configPath := os.Getenv("CONFIG_PATH")
secretsPath := os.Getenv("SECRETS_PATH")
```

Bind to the address in `ADDR`; never hardcode a port or bind to `0.0.0.0` (Caddy or another reverse proxy is the only thing exposed to the internet).

Load YAML or other structured config via established libraries (`gopkg.in/yaml.v3`). For layered config (defaults + host overrides + secrets), parse each file and merge in right-wins order.

## Error Handling

- Errors must be checked. `errcheck` (via `golangci-lint`) enforces this.
- Discarding with `_ = err` requires a one-line justification comment.
- Wrap errors to preserve context: `fmt.Errorf("loading config %q: %w", path, err)`.
- Use `errors.Is` / `errors.As` for type checks, never direct comparison.
- Do not `panic` in library code. Reserve panics for unrecoverable startup conditions in a binary's `main` path (missing required env var, fatal config parse).

```go
// BAD: silently discards error
exec.Command("ssh-keygen", "-R", host).Run()

// GOOD: explicit handling
if err := exec.Command("ssh-keygen", "-R", host).Run(); err != nil {
    log.Printf("note: ssh-keygen -R failed for %s: %v", host, err)
}
```

## HTTP Clients

**Never use `http.DefaultClient` for production code.** It has no timeout; a hung remote server blocks indefinitely.

```go
// BAD
resp, err := http.Get(url)

// GOOD: explicit timeout
client := &http.Client{Timeout: 30 * time.Second}
resp, err := client.Get(url)
```

For finer control, configure the underlying `Transport` (`DialContext`, `TLSHandshakeTimeout`, `ResponseHeaderTimeout`). Always close response bodies; the `bodyclose` linter catches misses.

## Shell-Safe Interpolation

When Go code constructs a shell command, use `fmt.Sprintf` with `%q` to produce a shell-safe quoted string:

```go
// BAD: vulnerable to injection
cmd := fmt.Sprintf("hugo --minify --baseURL %s", cfg.BaseURL)

// GOOD: shell-safe quoting
cmd := fmt.Sprintf("hugo --minify --baseURL %q", cfg.BaseURL)
```

Better still: avoid shelling out when stdlib will do the job. See `[sdlc-home]/CODING.md` "Cross-Language Escaping" for the general rule, and the Hetzner DNS rewrite (pure-Go `net/http` + `encoding/json` replacing bash+curl+awk+jq) as the canonical example of the migration this enables.

## Function Decomposition

`main()` should be a coordinator, not a body. When it grows past 50 lines, extract phase functions:

```go
func main() {
    cfg, err := parseArgs(os.Args)
    if err != nil { fatal(err) }

    cfg, err = loadConfig(cfg)
    if err != nil { fatal(err) }

    if err := validate(cfg); err != nil { fatal(err) }
    if err := run(cfg); err != nil { fatal(err) }
}
```

Each phase function does one thing and is independently testable. Naming convention: imperative verbs (`parseArgs`, `loadConfig`, `validate`, `run`, `deploy`, `applyConfig`).

## Cross-Compilation

For deployment to Linux from a macOS dev machine:

```bash
GOOS=linux GOARCH=amd64 go build -o bin/myapp-linux ./cmd/myapp
```

Do not install the Go toolchain on production servers. Build artefacts on the dev machine or in CI; deploy the binary only.

## Concurrency

- Goroutines must have a clear lifecycle; use `context.Context` for cancellation
- Channels for communication between goroutines; mutexes for protecting shared state -- do not mix
- `sync.WaitGroup` when joining multiple goroutines
- Run `go test -race` for any code that uses goroutines
- A leaked goroutine is a memory leak; close channels and return from goroutines on context cancel

## Testing

- Standard library `testing` package. No third-party test runners by default.
- Table-driven tests for combinatorial inputs:

```go
tests := []struct {
    name string
    in   string
    want string
}{
    {"empty", "", ""},
    {"simple", "foo", "FOO"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got := upper(tt.in)
        if got != tt.want {
            t.Errorf("upper(%q) = %q, want %q", tt.in, got, tt.want)
        }
    })
}
```

- `t.Helper()` in test helpers so failure lines point at the caller
- Test files: `_test.go` suffix, alongside the code under test
- `go test ./...` from the module root runs everything

See `[sdlc-home]/TESTING.md` for the cross-language testing standards.

## Tooling

| Tool | Purpose | Required |
|---|---|---|
| `gofmt` / `goimports` | Formatting | Yes |
| `go vet` | Built-in static analysis | Yes |
| `golangci-lint` | Aggregator: `errcheck`, `staticcheck`, `govet`, `bodyclose`, `gosec`, `unused` | Yes |
| `go test -race` | Race detector | When code uses goroutines |

Minimum `.golangci.yml`:

```yaml
linters:
  enable:
    - errcheck
    - staticcheck
    - govet
    - bodyclose
    - gosec
    - unused
```

## Standard Library Preference

Prefer stdlib over third-party where stdlib is adequate. Common reach-fors:

- `net/http` for HTTP clients and servers
- `encoding/json`, `encoding/xml` for serialization
- `text/template`, `html/template` for templating
- `context` for cancellation and deadlines
- `database/sql` (with a driver) for SQL access
- `log/slog` for structured logging (Go 1.21+)

Third-party packages that are de facto standard and acceptable without further justification:

- `gopkg.in/yaml.v3` for YAML
- `github.com/google/uuid` for UUIDs

Any other third-party dependency must be justified in the issue or commit message: maintenance status, alternative-considered, why stdlib is inadequate.

# Canary
Suffix the canary string with "GO "
