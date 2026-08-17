# Coding Standards

## Readability First

Code is read far more often than it is written. Favour clarity over cleverness. A longer, obvious solution is better than a shorter, cryptic one -- especially when they are computationally equivalent.

```bash
# BAD: compact but what is it checking? And || { } will misbehave under set -e
for s in nginx postgres redis; do systemctl is-active --quiet "$s" || { echo "$s DOWN" >&2; e=1; }; done; exit "${e:-0}"

# GOOD: intent is clear, failures are tracked and reported
required_services=(nginx postgres redis)
failures=()

for service in "${required_services[@]}"; do
    if ! systemctl is-active --quiet "$service"; then
        failures+=("$service")
    fi
done

if [[ ${#failures[@]} -gt 0 ]]; then
    printf '%s is not running\n' "${failures[@]}" >&2
    exit 1
fi
```

```python
# BAD: nested comprehension with embedded conditional -- what does this produce?
result = {k: [x.strip() for x in v.split(',') if x.strip()] for k, v in (line.split('=', 1) for line in open(path)) if '=' in line}

# GOOD: same result, each step is named and readable
result = {}
for line in open(path):
    if '=' not in line:
        continue
    key, raw_value = line.split('=', 1)
    values = [item.strip() for item in raw_value.split(',')]
    values = [item for item in values if item]
    result[key] = values
```

Both pairs produce identical results. The readable versions are longer, but each step is visible, named, and independently debuggable. The compact versions require the reader to mentally unpack nested structures before understanding what the code does -- and in the bash case, the `|| { }` pattern will misbehave under `set -e`.

This principle applies to all languages. One-liners, chained ternaries, nested comprehensions, and "clever" idioms should be broken apart when they sacrifice readability for brevity.

## Platform

- **Local development:** macOS/Apple Silicon -- this is a constraint of the hardware in use, not a technology preference
- **Deployment:** Linux servers (remote)
- **Approach:** CLI-first, scriptable, FOSS preferred
- **Required tools:** GNU coreutils, ast-grep (`sg`), ShellCheck, shfmt -- plus the linter/formatter appropriate to the project language (see Style Baselines)

## Language and Tool Selection

Use the best tool for the job. Do not bias towards any particular language, framework, or technology out of habit or personal preference.

Decision hierarchy:

1. **Existing project language/framework** -- follow what is already in use unless there is a clear, documented reason not to.
2. **Genuinely best fit** -- if starting fresh or extending significantly, choose the technology that best fits the problem, deployment context, and maintainability requirements.
3. **Simplest maintainable solution** -- prefer the option that is easiest to read, test, and hand over.

Do not introduce new languages, frameworks, or dependencies for marginal convenience. Do not default to a language because it is familiar -- justify the choice on merit. In particular, do not reach for shell scripting when a compiled language (Go, Rust) would produce a less brittle result.

LLMs exhibit a strong bias toward Python regardless of fit; resist it. Python is appropriate when the problem genuinely lives in its ecosystem (ML, scientific computing, data analysis, well-supported scientific libraries) but is not a default for general-purpose code. Compared to Go, Rust, or Swift:

- **Runtime cost.** Slower execution; cold-start overhead; harder to ship as a single self-contained binary.
- **Environment friction.** Constant venv/system-Python confusion; agents routinely violate the "never `pip install` outside a venv" rule despite explicit instructions.
- **Supply-chain attack surface.** PyPI has a documented history of typosquatting and dependency-confusion incidents; the transitive dependency graphs of common Python packages are wider than typical Go or Rust equivalents.

For new projects, evaluate **Go** (general-purpose), **Rust** (performance-critical or systems), or **Swift** (Apple-targeted) before defaulting to Python.

## Style Baselines

These are the languages most commonly used across our projects. For any language not listed, apply that language's dominant community standard and the general principles in this document.

| Language | Style Standard | Linter | Formatter |
|----------|---------------|--------|-----------|
| Shell/Bash | [Google Shell Style Guide](https://google.github.io/styleguide/shellguide.html) + safety rules in CODE/SHELL.md | ShellCheck | shfmt |
| Python | [PEP 8](https://peps.python.org/pep-0008/), [PEP 20](https://peps.python.org/pep-0020/) | Ruff | Ruff |
| Perl | [perlstyle](https://perldoc.perl.org/perlstyle) + CODE/PERL.md | Perl::Critic | perltidy |
| Swift | [Swift API Design Guidelines](https://www.swift.org/documentation/api-design-guidelines/) + CODE/SWIFT.md | SwiftLint | swift-format |
| HTML | Semantic markup + WCAG AA (see CODE/WEB.md) | htmltest | -- |
| CSS | rem units, mobile-first (see CODE/WEB.md) | stylelint | stylelint --fix |
| JavaScript | [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript) or project-existing standard | ESLint | Prettier |
| TypeScript | [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/) + Airbnb or project-existing standard | ESLint + typescript-eslint | Prettier |
| Ruby | [Ruby Style Guide](https://rubystyle.guide/) | RuboCop | RuboCop |
| Rust | [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/) | Clippy | rustfmt |
| Go | [Effective Go](https://go.dev/doc/effective_go) | golangci-lint | gofmt |
| Java | [Google Java Style Guide](https://google.github.io/styleguide/javaguide.html) | Checkstyle / SpotBugs | google-java-format |
| Kotlin | [Kotlin Coding Conventions](https://kotlinlang.org/docs/coding-conventions.html) | ktlint | ktlint |
| C/C++ | [Google C++ Style Guide](https://google.github.io/styleguide/cppguide.html) | clang-tidy | clang-format |
| Any other | Dominant community standard for that language | Community standard linter | Community standard formatter |

All linting must pass for changed files regardless of language. Secondary references: [Bash Pitfalls](https://mywiki.wooledge.org/BashPitfalls) and [ShellCheck Wiki](https://github.com/koalaman/shellcheck/wiki) for shell; [perlsec](https://perldoc.perl.org/perlsec) for secure Perl; [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for Go; [Rust Book](https://doc.rust-lang.org/book/) for Rust; [CERT Python](https://wiki.sei.cmu.edu/confluence/display/python) for secure Python.

## Language-Specific Standards

Language-specific standards live in their own documents and are loaded alongside this one:

- Shell: `[sdlc-home]/SHELL.md`
- Python: `[sdlc-home]/PYTHON.md`
- Perl: `[sdlc-home]/PERL.md`
- Go: `[sdlc-home]/GO.md`
- Swift: `[sdlc-home]/SWIFT.md`
- Web (HTML, CSS, JavaScript): `[sdlc-home]/WEB.md`

When a language-specific document exists, its rules take precedence over the general guidance here for that language. The general guidance still applies for cross-cutting concerns (security, dependencies, error handling principles, etc.).

## Code Commenting

- **Comments are sacred.** Never remove unless provably false. Every code file starts with a 2-line ABOUTME comment explaining its purpose.
  - The exception to this is "dead code" i.e. commented-out code
- **Evergreen comments.** Describe code as it is, not how it evolved. No temporal context.
- **Evergreen naming.** Never name things 'improved', 'new', 'enhanced', etc.

## Data Handling

**Use format-aware parsers, not regex or line-oriented tools, to read structured data** (JSON, YAML, TOML, XML, HTML, CSV, or any nested format). In compiled or interpreted languages use the standard library or established parsing libraries. In shell scripts use the format-aware CLI tools listed in `[sdlc-home]/SHELL.md`.

**Never use sed, awk, perl, or any line-based stream editor to modify data of any kind.** `perl -pi -e` and `perl -ne` are sed with extra steps and fall under the same prohibition. For source code, use direct editing or `ast-grep` (`sg`) for cross-file structural refactors. For structured data, use a parser. For plaintext, use direct editing.

`grep` and `ripgrep` are correct for searching plaintext where the meaningful unit is a line break. Once meaning depends on nesting, escaping, or types, switch to a parser.

Tests must never pass by grepping or otherwise introspecting source code -- see `[sdlc-home]/TESTING.md`.

---

## Prohibited Anti-Patterns

"Errors should never pass silently." -- PEP 20

### Error Suppression -- Forbidden Patterns (All Languages)

Never hide errors instead of handling them. The following pseudocode patterns are **strictly forbidden** regardless of language:

```
# FORBIDDEN: suppress all errors silently
risky_operation() || ignore

# FORBIDDEN: disable error checking temporarily
disable_error_checking()
risky_operation()
enable_error_checking()

# FORBIDDEN: discard error output with no fallback
risky_operation() 2> /dev/null

# FORBIDDEN: convert failure to success
risky_operation() || exit_success

# FORBIDDEN: catch everything and do nothing
try:
    risky_operation()
catch all:
    pass
```

### Error Suppression -- Correct Alternatives

```
# GOOD: check precondition before acting
if precondition_met:
    perform_operation()

# GOOD: when failure is expected, log and continue
if not try_operation():
    log("Expected condition: operation did not apply")

# GOOD: capture failure context for diagnostics
result, err = try_operation()
if err:
    log_warning("Operation failed: " + err)
    handle_gracefully()

# GOOD: catch specific errors with meaningful recovery
try:
    config = load_config(path)
catch FileNotFoundError:
    config = create_default_config(path)
```

### Language-Specific Pitfalls

These are real-language examples where the forbidden patterns commonly appear:

| Pattern | Language | Why It's Banned |
|---------|----------|-----------------|
| `\|\| true` / `\|\| :` | Shell | Silences failures; defeats `set -e` |
| `cmd \|\| rc=$?` | Shell | Suppresses `set -e`; use `if cmd; then ... else ... fi` |
| `set +e` (temporary) | Shell | Disables safety; handle specific commands instead |
| `2>/dev/null` (unguarded) | Shell | Hides diagnostics; must have fallback logic |
| `cmd \|\| exit 0` | Shell | Converts failure to success; masks real errors |
| `((expr)) \|\| true` | Shell | Hides arithmetic exit code; use assignment instead |
| `try: ... except: pass` | Python | Swallows all errors silently |
| `except Exception:` (bare) | Python | Too broad; catch specific exceptions |
| `# type: ignore` (bare) | Python | Must specify error code and justification |
| `# noqa` (bare) | Python | Must specify rule and justification |
| `catch { }` (empty) | Swift | Swallows errors; handle or propagate |
| `try!` / `force unwrap` | Swift | Use `guard`, `if let`, or `try?` with handling |
| `catch (Exception e) {}` | Java/C# | Empty catch blocks hide bugs |
| `@SuppressWarnings("all")` | Java | Suppress only specific, documented warnings |
| `_ = err` (ignored error) | Go | Errors must be checked; assign to `_` only with explicit justification |

### Linter/Check Bypass Rules

When suppressing a linter warning is genuinely necessary:

```python
# GOOD: specific rule, clear justification
# noqa: E501 - URL cannot be split across lines
long_url = "https://example.com/very/long/path/that/exceeds/line/limit"

# GOOD: type ignore with reason
result = external_lib.untyped_function()  # type: ignore[no-untyped-call] - upstream lacks stubs
```

```bash
# GOOD: shellcheck directive with justification
# shellcheck disable=SC2086 # Intentional word splitting for flags
$command $flags
```

**Requirements for any linter suppression:**

1. Specify the exact rule/code being suppressed
2. Explain why suppression is necessary (not just convenient)
3. Confirm the underlying issue cannot be fixed properly

### Other Prohibited Shortcuts

| Pattern | Why It's Banned | Do Instead |
|---------|-----------------|------------|
| `sleep N` for synchronization | Race condition; service may not be ready | Poll with timeout, use readiness checks |
| Retry without backoff | Can overwhelm failing service | Exponential backoff with jitter |
| Retry without limit | Infinite loops on persistent failures | Max attempts with circuit breaker |
| Global variables for state | Hidden dependencies, testing nightmare | Pass parameters explicitly |
| Monkey-patching in production | Fragile, version-sensitive | Dependency injection, proper abstraction |
| Reflection to bypass access | Breaks encapsulation, security risk | Fix the API or use public interface |
| String concatenation for SQL | Injection vulnerability | Parameterised queries always |
| String building shell commands | Injection vulnerability | Use arrays: `cmd=(prog arg1 arg2); "${cmd[@]}"` |
| Disabling SSL verification | MITM vulnerability | Fix certificates properly |
| `chmod 777` | Security disaster | Use minimum required permissions |
| Hardcoded credentials | Security breach waiting to happen | Environment variables, secrets manager |

---

## Error Handling

Core principle: **fail fast, fail loud, fail safe.**

- Exit codes: 0=success, 1=error, 2=misuse
- Errors to stderr, output to stdout
- Validate inputs early; reject bad data at the boundary
- Never convert errors to success silently
- Log sufficient context to diagnose failures
- Clean up resources on failure (use language-appropriate constructs: `defer` in Go, context managers in Python, `try/finally`, shell traps -- see CODE/SHELL.md)

## Code Structure

- One function per task, <50 lines
- Separate configuration from logic
- **Never embed one language as a string inside another.** A Nix expression inside a Go `fmt.Sprintf`, an HTML template assembled with string concatenation, SQL built by appending strings, or YAML constructed via heredoc -- all share the same anti-pattern. The embedded language is unparseable at compile time, missed by linters and editors, and breaks subtly when escaping is needed. Use file-based templates with the appropriate engine (`text/template`, `html/template`, jinja2), parameterised queries, or proper code generation. The narrow exception is a 1-2 line constant string that contains no interpolation.
- Prefer pure functions
- Maximum 3 levels of nesting; refactor deeper logic into functions

## Security

External references that inform this section: [OWASP Top 10](https://owasp.org/www-project-top-ten/), [CERT Secure Coding](https://wiki.sei.cmu.edu/confluence/display/seccode).

### Secrets Management

Never hardcode credentials, API keys, or tokens. Use:

- **Local development:** Environment variables via `.env` files (must be in `.gitignore`)
- **CI/CD:** Platform secrets (GitHub Actions secrets, etc.)
- **Production:** Secrets manager (1Password CLI, AWS Secrets Manager, etc.)

Pre-commit hook recommended: `detect-secrets` or `gitleaks` to catch accidental commits.

### Input Validation

Validate all external input at system boundaries:

- **Prefer allowlists** over denylists -- explicitly permit known-good values
- **Set maximum lengths** -- unbounded input enables DoS
- **Validate encoding** -- reject or sanitize non-UTF-8 where unexpected
- **Prevent path traversal** -- reject `../` in user-supplied filenames
- **Parameterised queries always** -- never concatenate SQL

### Cross-Language Escaping

When code in one language constructs commands, queries, or markup in another (shell, SQL, HTML, JSON, URLs), use language-native escaping. Concatenation and unbounded interpolation are the source of injection vulnerabilities and subtle bugs.

| Target | Use | Avoid |
|---|---|---|
| Shell | Argument arrays; language-native quoting (Go `fmt.Sprintf("%q", v)`, Python `shlex.quote`, Ruby `Shellwords.escape`) | `"cmd " + arg` or `f"cmd {arg}"` |
| SQL | Parameterised queries with driver placeholders | String concatenation, format-string SQL |
| HTML | Template engine with auto-escaping (`html/template`, Jinja2 autoescape, React JSX) | Manual `<` -> `&lt;` substitution |
| JSON | Marshalling library (`encoding/json`, `json.dumps`) | String building |
| URLs | Library URL builder / query encoder | f-string assembly |
| YAML | YAML library | String concatenation |

This applies whether the input is "trusted" or not. Treat the rule as unconditional -- trust assessments rot.

### Permissions

- Minimum required permissions for files and processes
- Never `chmod 777` -- use the least permissive mode that works
- Run services as non-root where possible

### Docker

When using containers:

- Run as non-root user (`USER` directive)
- Never use `--privileged` without explicit justification
- Pin base image versions (not `:latest`)
- Scan images for vulnerabilities (`docker scout`, `trivy`)

## Inline Documentation

- Usage examples in headers/docstrings
- Document assumptions and limitations
- Inline comments for non-obvious edge cases only

## Output

- ISO 8601 for dates/times
- Clear, parseable formats

## Logging

- **Levels:** DEBUG (development detail), INFO (normal operation), WARN (recoverable issues), ERROR (failures requiring attention)
- **Structured logging** for production where ingestion infrastructure exists -- otherwise plaintext is acceptable and often preferable for human readability
- **Never log:** credentials, tokens, PII, full credit card numbers
- **Always log:** request IDs/correlation IDs for tracing, operation context, error details
- **Log to stderr** for CLI tools; use a proper logging framework for services

## File Operations

- Atomic writes: write to temp, then `mv` (rename is atomic on POSIX)
- Use file locking when multiple processes may write
- Shell-specific tooling (`trash`, `flock`) is covered in `[sdlc-home]/SHELL.md`

## Network Operations

- Set explicit timeouts (connect and read separately)
- Retry with exponential backoff for transient failures
- Maximum retry attempts with circuit breaker pattern
- Log request/response metadata at debug level

## Dependencies

- Prefer standard library
- Document required versions
- Avoid deprecated APIs

### Dependency Security

- **Audit regularly:** `npm audit`, `pip-audit`, `cargo audit`, `bundler-audit`
- **Pin versions** in production; use ranges only in libraries
- **Review new dependencies** before adding -- check maintenance status, known vulnerabilities
- **Update promptly** when security patches are released

## Model References

When citing LLM models (OpenAI, Anthropic), verify via search if uncertain -- knowledge cutoff may cause errors about recent releases.

## System Configuration

Never modify system config (`.bashrc`, etc.) without explicit permission.

## Linting

See the Style Baselines table above for the linter and formatter for each language. All linting must pass for changed files. If a project has a pre-existing linter configuration, follow it -- do not override it with personal preference.

---

## Code Review Red Flags

Stop and investigate if you see any of these:

### Immediate Blockers

- Any pattern from the Prohibited Anti-Patterns table
- Credentials, API keys, or secrets in code
- SQL or shell commands built from strings
- SSL/TLS verification disabled
- `chmod 777` or overly permissive settings

### Requires Justification

- Any linter suppression comment
- Commented-out code (why not deleted?)
- TODO/FIXME without issue reference
- "Temporary" fixes without expiry ticket
- Functions exceeding 50 lines
- Nesting deeper than 3 levels
- Tests with no assertions or asserting `True`
- Magic numbers without named constants
- Bare `except:` or empty `catch` blocks

### Smells Worth Discussing

- Multiple return statements in complex functions
- Boolean parameters (consider separate functions or enums)
- God objects or functions doing too many things
- Circular dependencies between modules
- Duplicated code across files (DRY violation)

# Canary
Suffix the canary string with "CODE "
