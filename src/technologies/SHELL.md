# Shell Standards

Shell-specific standards covering both interactive shell commands and scripts. The general coding standards in `~/.agents/sdlc/CODING.md` apply on top of these; this document does not repeat cross-language rules.

Most rules apply equally to one-shot commands typed at the prompt and to scripts. The Mandatory Safety Header, Portability, Error Handling traps, and Schedulers sections are script-specific and noted as such; everything else applies universally.

## Complexity Limit (scripts)

Shell is glue, not an application language. Once a script starts carrying application logic, stop and tell the user before adding more complexity. Present a recommendation for rewriting it in a more structured language, usually Go for portable CLIs/services, Rust for performance- or safety-critical systems work, Swift for Apple-targeted tools, or Python only where its ecosystem is the reason for the choice.

Treat these as tripwires:

- The script is approaching or exceeding 100 lines.
- More than 3 functions are needed.
- State is shared across multiple phases.
- Control flow requires nested conditionals, retries, backoff, concurrency, or cleanup across several resources.
- The script parses or transforms structured data beyond a simple `jq`, `htmlq`, `xmlstarlet`, or `dasel` query.
- The script talks to multiple external systems or APIs.
- Error handling needs more than simple fail-fast behaviour and local cleanup.

When a tripwire appears, do not silently continue making the shell script
larger. Record the complexity signal and the choice to retain or replace shell
in the plan before adding more application logic. Explain the target language
and migration path when replacement is selected.

## Version Targeting

bash 5+ for all projects. Associative arrays, `[[ =~ ]]` regex with `BASH_REMATCH`, `mapfile`/`readarray`, advanced parameter expansion, and `${var^^}` case conversion are permitted and encouraged where they improve clarity.

If a script genuinely must run on stock macOS bash 3.2 (rare: a Homebrew bootstrap script or similar), declare the requirement explicitly with a comment at the top of the file. Otherwise, scripts may rely on bash 5+ features freely. Public Homebrew formulae should `depends_on "bash"`.

## Mandatory Safety Header (scripts)

```bash
#!/usr/bin/env bash
set -eo pipefail
```

Omit only with documented reason. Does not apply to one-shot commands at the prompt.

**`IFS=$'\n\t'` is forbidden.** A global override breaks legitimate parsing of
space-delimited tool output, including `read A B C < <(cmd)`, even when
variables are quoted. It only mitigates unquoted-variable use, which is already
prohibited. Do not add it to the safety header. When a specific block needs
different word splitting, set `IFS` locally and restore it afterwards.

## Required Practices

- Quote variables: `"$var"`
- Terminate options: `trash -- "$file"`
- Use arrays for command construction, not strings
- Use `command -v`, not `which`
- Never parse `ls`
- `find` pipelines: `-print0` with `read -r -d ''` or `xargs -0`
- All CLI executables must provide `-h`/`--help` and `--version`. Provide `--dry-run` where appropriate
- Help text is documentation: keep it in `./docs/` (e.g. `docs/<command>-help.md`) and have the executable read it at runtime or package it at build time. Do not maintain help text inside the executable as a long inline heredoc -- that puts documentation where it cannot easily be reviewed, edited, or kept consistent with the rest of the docs
- If releasing on Homebrew, always `make release` then install locally via `brew install` or `upgrade` -- never live-patch an existing package install from brew or any other package manager

## Prohibited Shell Patterns

Never execute code from strings or discovery:

- `$cmd` / `"$cmd"` / `${cmd}` as commands
- `eval` (any use)
- String concatenation to build and run commands
- Running scripts found via `find` or globbing
- Wiring discovered paths into schedulers or services

## Safe Alternatives

- Functions or `case` dispatch for behaviour selection
- Bash array for dynamic commands: `cmd=(git status --porcelain); "${cmd[@]}"`
- Script selection: allowlist by name, validate against `^[A-Za-z0-9._-]+$`, verify exists/executable/permissions

## Arithmetic Expression Gotcha

The `(( ))` construct returns exit code 1 when the expression evaluates to 0, triggering `set -e`. The `let` builtin has the same behaviour. Never suppress this with `|| true`.

```bash
# BAD: suppresses the exit code problem
((count++)) || true
((count++)) || :
let count++ || true

# GOOD: arithmetic expansion in assignment (always succeeds)
count=$((count + 1))

# GOOD: null command consumes the expression (always succeeds)
: $((count++))

# GOOD: declare as integer, then use +=
declare -i count=0
count+=1
```

## Data Handling

**Use format-aware tools to read or modify structured data.** Parsing JSON, YAML, TOML, XML, HTML, or program output with grep/sed/awk/perl is brittle and prone to corrupting structure.

**sed, awk, and perl are forbidden in shell scripts and one-shot commands.** Not just for data modification -- entirely. The harness backs this with a permission-layer deny (settings.json) so direct invocations surface as a flag. Substitutes:

- File editing: the Edit tool
- Plaintext reading or filtering of newline-delimited streams or files: `grep`
- Structured data: the format-aware tools in the table below
- Source code refactors: `ast-grep` (`sg`) for cross-file structural changes, direct editing for single-file
- Streaming substitution in pipelines (log redaction, fixture normalization): a purpose-built tool (`detect-secrets`, `gitleaks`) or a small program in Go or Python

**Direct `python` and `python3` interpreter commands are forbidden for agent-submitted shell work.** Use the task-appropriate format-aware tool or a project-owned entry point such as `make test`. This restriction applies to the command the agent submits; it does not prohibit a Make target, test runner, or other approved project entry point from launching Python internally.

If none of the substitutes solve a case, report the constraint instead of
bypassing it. The project may then define a safe, reviewable entry point.

**`grep` is for plaintext streams or files where a newline character is the delimiter** -- logs, command output, single-file pattern matches. **`ripgrep` (`rg`) is for finding files**, not for extracting content. In shell scripts, the only routine use of `rg` is `rg -l` to enumerate files for further processing by a format-aware tool. Reaching for `rg` to pull content out of a file is usually a sign the wrong tool is being used downstream.

Tests must never pass by grepping or otherwise introspecting source code -- see `~/.agents/sdlc/TESTING.md`.

### Tools

| Format | Tool | Brew package | Notes |
|---|---|---|---|
| Source code (structural refactor) | `ast-grep` (`sg`) | `ast-grep` | Preferred for bulk renames and signature changes across files. Single-file edits: direct editing is fine. |
| JSON | `jq` | `jq` | The standard. |
| YAML | `yj \| jq` | `yj` + `jq` | Default. Converts YAML to JSON then queries with jq. One query language, no naming conflict. |
| YAML (in-place edit preserving comments) | `yq` (Mike Farah) | `yq` | Reserved for cases where comments and anchors must survive the edit. v4+. Verify install: `yq --version 2>&1 \| grep -q mikefarah`. |
| TOML | `dasel` | `dasel` | |
| XML | `xmlstarlet` | `xmlstarlet` | For edit/validate. `dasel` acceptable for simple reads. |
| HTTP / API content | `curl \| jq` | `curl` + `jq` | Standard pattern for fetching and processing API responses, including in tests. |
| Format conversion | `yj` | `yj` | YAML <-> TOML <-> JSON <-> HCL |
| CLI output -> JSON | `jc` | `jc` | Wrap `ps`/`df`/`mount`/`netstat`/etc. before parsing. |
| Multi-format | `dasel` | `dasel` | One selector syntax across JSON/YAML/TOML/XML/CSV. |

For HTML and CSS tooling such as htmlq, htmltest, and stylelint, see
`~/.agents/sdlc/technologies/WEB.md`. JavaScript and TypeScript linting follows
`~/.agents/sdlc/technologies/JAVASCRIPT.md`.

### yq pitfall

Two unrelated tools share the name `yq`:

- Mike Farah's (Go, dominant) -- `mikefarah/yq`
- Kislyuk's (Python, jq wrapper) -- `kislyuk/yq`

Their syntaxes are incompatible. Any script depending on `yq` must verify the right one is installed:

```bash
if ! yq --version 2>&1 | grep -q 'mikefarah'; then
    echo "Requires mikefarah/yq v4+" >&2
    exit 1
fi
```

Prefer the `yj | jq` pipeline above to avoid the naming conflict entirely.

### Data Format Policy

| Use case | Format |
|---|---|
| Plaintext data -- newline-delimited streams or files (logs, command output) | Read with `grep`; never modify with sed/awk/perl |
| Program data persisted to disk or emitted for another program to consume | **JSON** |
| User-authored configuration | **YAML** (nested) |
| User-authored tabular data | **CSV** |
| Output a human will read directly (status, summaries, reports) | **YAML** rendered from program data at the display boundary |
| Externally-standardized formats (`pyproject.toml`, `Cargo.toml`, GitHub workflows, etc.) | Follow upstream |

User config is parsed into the language's native types at load time and must not depend on YAML features that have no JSON equivalent (anchors with side effects, complex tags, non-string keys) -- this keeps configs portable and prevents loader-specific behaviour.

## Portability

- Cross-platform Darwin/Linux where feasible
- `#!/usr/bin/env` shebangs
- Handle macOS paths (`~/Library`, etc.)
- Consider ARM64 quirks
- Prefer Unix built-ins (exceptions: coreutils, ast-grep)

## Error Handling

The general principles (fail fast, fail loud, fail safe; exit codes 0/1/2; errors to stderr) are in `~/.agents/sdlc/CODING.md`. Shell-specific patterns:

- Validate inputs early; reject bad data at the boundary
- Clean up resources on failure using traps:

```bash
cleanup() {
    # remove temp files, close connections, etc.
}
trap cleanup EXIT
```

## File Operations

- Use `trash` instead of `rm` to allow recovery
  - **Exception:** a script or program may delete the exact directory it created with `mktemp -d` through explicit cleanup/teardown logic. Do not use globs, shared parent directories, or ad hoc `rm` commands as an agent shell action.
- Create scratch directories with `mktemp -d`, which selects the operating system temporary directory. Do not hard-code `/tmp`, assume `$TMP` is set, or create scratch directories inside the project.
- Register cleanup immediately after allocation and handle success, failure, and signals. Tests must also enforce the output limits in `~/.agents/sdlc/TESTING.md`.
- Atomic writes: write to temp, then `mv`
- Use `flock` when multiple processes may write

## Schedulers and Services

For systemd, launchd, cron:

- Never generate definitions from runtime discovery (unless job-dedicated directory with validated scripts)
- Use build/install-time path resolution
- `ExecStart=` must use absolute, fixed paths
- Avoid `sh -c`/`bash -c` in units; pass arguments directly
- Logs must go somewhere predictable; failures must be visible
