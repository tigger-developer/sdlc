# Fish Shell Standards

These standards apply to maintained Fish scripts, functions, completions, and
configuration. The general rules in `~/.agents/sdlc/CODING.md` still apply.
`~/.agents/sdlc/technologies/SHELL.md` governs Bash and POSIX-style shell work
and is not an automatic prerequisite for Fish; load both only when the project
maintains both kinds of shell source.

## Language boundary

Fish is intentionally not POSIX-compatible. Treat it as a distinct language:

- use the `.fish` suffix for sourced files, functions, and completions;
- use `#!/usr/bin/env fish` for directly executable Fish scripts;
- never label Fish source as `sh`, source Bash or POSIX shell code into Fish, or
  mechanically translate syntax between them;
- invoke a script for another shell through its declared interpreter instead of
  sourcing it; and
- declare the minimum supported Fish version when the project relies on
  version-specific behaviour.

Fish is appropriate for interactive configuration, user-owned functions,
completions, and short glue explicitly targeting Fish users. Do not introduce a
Fish runtime dependency into portable bootstrap, deployment, service, or
installer paths unless the specification and target environment own that
dependency.

## Complexity boundary

Fish is glue, not an application language. Stop and make the architecture
choice visible before extending Fish source that:

- approaches 100 lines or needs more than three substantial functions;
- owns persisted application state or a multi-phase workflow;
- needs nested retries, concurrency, backoff, or multi-resource cleanup;
- parses or transforms structured data beyond one format-aware command; or
- coordinates several external systems or APIs.

Prefer Go for portable tools and services, Rust for performance- or
safety-critical systems work, Swift for Apple-specific tools, or another
project language whose ecosystem justifies the choice. Interactive shell
integration may remain a thin Fish adapter around that program.

## Preferred builtins, libraries, and tools

Fish has a plugin ecosystem rather than a conventional application-library
ecosystem. Prefer its maintained builtins before adding executable third-party
shell code.

| Concern | Preferred reach-for | Standard |
|---|---|---|
| Options | `argparse` | Define and validate a real CLI contract; do not hand-roll option parsing. |
| Strings | `string` | Use explicit matching, replacement, escaping, splitting, and collection. |
| Paths | `path`, `fish_add_path` | Manipulate paths structurally and avoid duplicate path entries. |
| Arithmetic | `math` | Do not delegate arithmetic through another shell or text processor. |
| Command and terminal checks | `type -q`, `status`, `isatty` | Query capabilities and invocation state explicitly. |
| Functions and completions | `function`, `functions`, `complete`, `abbr` | Use native autoloading and declarative completions; reserve abbreviations for interactive typing. |
| Formatting and parsing | `fish_indent`, `fish --no-execute` | These form the minimum static verification baseline. |
| Plugin management | [Fisher](https://github.com/jorgebucaran/fisher) | Use only when multiple maintained plugins justify a manager; pin dependencies and track `fish_plugins`. |
| Fish-native tests | [Fishtape](https://github.com/jorgebucaran/fishtape) | Optional for a substantial Fish library or plugin; do not add it merely to wrap trivial command checks. |
| Editor diagnostics | [fish-lsp](https://github.com/ndonfris/fish-lsp) | Optional development aid; it does not replace parser, formatter, or behavioural checks. |

Treat every plugin as code executing with the user's shell authority. Review its
source, licence, maintenance, installation and update hooks, and transitive
behaviour. Pin a released tag or commit where reproducibility matters. Never
install Fisher or a plugin by piping a network response into `source`; download,
inspect, verify, and install through a controlled project or user-owned process.

## Best practice: structure and startup

- Keep `config.fish` small. Put reusable functions in
  `~/.config/fish/functions/<name>.fish` and completions in
  `~/.config/fish/completions/<command>.fish` or their project-owned package
  equivalents.
- Use `conf.d` snippets for independently owned startup configuration, not as an
  unordered collection of hidden application logic.
- Guard interactive-only setup with `status is-interactive`. Configuration read
  by non-interactive Fish must not print, prompt, launch programs, access the
  network, or depend on a terminal.
- Keep startup deterministic and fast. Defer expensive work until the function
  or command that needs it runs.
- Prefer an autoloaded named function to a large `config.fish` function block.
  Give maintained public functions a useful description and explicit argument
  names where that improves their interface.
- Use an abbreviation for interactive typing convenience and a function for
  behaviour that must also work non-interactively.
- Keep completions declarative, fast, and free of side effects. Do not run an
  expensive command merely to render every prompt or completion candidate.
- Do not modify user or system Fish configuration without explicit authority
  for the exact file and effect.

## Best practice: variables, lists, and expansion

- Declare scope deliberately: normally `set -l` inside a function, `set -g` for
  session state, and `set -gx` only for values child processes require.
- Use `set -q` to test whether a variable exists. Do not import Bash assignment,
  default-expansion, or `unset` syntax.
- Treat Fish variables as lists. Preserve argument boundaries and use list
  operations rather than joining values into a command string.
- Fish does not perform Bash-style word splitting. When command output needs a
  delimiter other than newline, use `string split`, `string split0`, or another
  explicit format-aware operation.
- Use `$(command)` or `(command)` for command substitution; never use backticks.
  Check the substituted command's result when failure matters.
- Use Fish's `string`, `path`, and `math` builtins for their intended operations
  instead of constructing fragile external pipelines.
- Universal variables persist across Fish sessions. Use them only for an
  intentional user preference, never for project state, transient state, or
  secrets. Never edit `fish_variables` directly or append to a universal
  variable during every shell startup.

## Best practice: commands and failure

- Use Fish-native `if`, `switch`, `for`, `while`, `function`, `test`, and
  `begin`/`end` structure. Do not paste Bash `[[ ... ]]`, brace groups, process
  substitutions, arrays, or parameter expansions into Fish.
- Use `argparse` for a non-trivial script or function interface. Reject unknown
  options and validate required values before causing side effects.
- Follow the complete common CLI switch contract: equivalent `-h` and `--help`,
  `--version`, and `--dry-run` where the command's effects can be previewed
  faithfully. Help and version requests exit successfully without side effects.
  Document normal output, send diagnostics to stderr, and use exit status 0 for
  success, 1 for operational failure, and 2 for invalid invocation unless an
  established interface requires otherwise.
- Capture `$status` immediately after the command it describes. Inspect
  `$pipestatus` when every pipeline component matters; the last command's status
  alone does not prove the whole pipeline succeeded.
- Test command availability with a Fish-native query such as `type -q` before
  relying on an optional external dependency.
- An unmatched glob can fail a command. Do not use a broad glob where an empty
  input set is valid without handling that case explicitly.
- Never suppress or overwrite a meaningful failure merely to continue startup
  or make a check appear successful.

## Best practice: safety and data handling

- Never execute constructed command strings or use `eval`. Build commands as
  Fish lists and pass each argument as a distinct list element.
- Never `source` downloaded, generated, or otherwise untrusted content.
- Use format-aware tools for JSON, YAML, TOML, XML, and HTML. Fish string
  processing is not a structured-data parser.
- Quote literal wildcard, variable-expansion, command-substitution, and
  redirection characters when they are data rather than syntax.
- Create temporary resources through the operating system and register exact,
  bounded cleanup. Do not weaken the common prohibition on agent-submitted
  destructive commands.
- Functions and configuration must be safe when loaded more than once. Avoid
  duplicate path entries, repeated event handlers, and state that grows on each
  shell start.

## Portability and packaging

- Support Darwin and Linux where the requirement does not name one platform.
  Test platform-specific commands and paths rather than assuming GNU or BSD
  behaviour.
- Make Fish an explicit runtime dependency. Verify the minimum version at a
  stable entry point and report an actionable error when it is unavailable.
- Do not rely on the operator's prompt theme, abbreviations, aliases, universal
  variables, plugin set, or personal startup configuration.
- Package functions, completions, and `conf.d` snippets in their conventional
  directories. Installation must be repeatable and must not overwrite unrelated
  user configuration.
- Keep plugin dependencies declared, pinned where required, and reproducible.
  Do not update them implicitly during shell startup.
- Keep secrets outside Fish source, history, universal variables, and committed
  plugin configuration.

## File operations

- Use `path` operations for path components, normalization, filtering, and
  extension handling. Do not treat structured paths as arbitrary strings.
- Create scratch directories with the operating system's temporary-directory
  mechanism. Register exact cleanup immediately and cover success, failure, and
  signals.
- Use `trash` rather than `rm` for recoverable user or project files. A script
  may remove only the exact temporary directory it created under the common
  temporary-resource exception.
- Write durable output to a sibling temporary file and replace it atomically
  when partial content would be harmful.
- Treat a requirement for locking, concurrent writers, or rollback across
  several files as a complexity signal rather than improvising an application
  transaction in Fish.

## Schedulers and services

- Prefer an executable script file with a Fish shebang to a constructed
  `fish -c` command string.
- Declare Fish as a runtime dependency and resolve its executable path during
  installation. Do not assume an interactive user's `PATH` in systemd,
  launchd, cron, or another scheduler.
- Run without personal configuration and pass required environment explicitly.
- Use absolute, fixed program and script paths in service definitions. Do not
  discover executable jobs dynamically at runtime.
- Keep stdout, stderr, exit status, timeout, restart, and cleanup behaviour
  visible to the owning service manager.
- Do not install a Fish service when a direct executable invocation removes the
  shell dependency and command-construction layer.

## Formatting and verification

For maintained Fish source:

- format with `fish_indent` and require `fish_indent --check` in the project
  lint path;
- syntax-check with `fish --no-config --no-execute <file>`;
- run behavioural script tests with `fish --no-config` so personal startup
  configuration cannot affect the result;
- test configuration or completion behaviour with an isolated, test-owned Fish
  configuration root; and
- test interactive behaviour only when the requirement is genuinely
  interactive, using a user test or a controlled pseudo-terminal test rather
  than pretending a non-interactive invocation proves it.

ShellCheck and `shfmt` do not replace Fish's parser and formatter. Apply them
only to Bash or POSIX-style files maintained by the same project.

## Anti-patterns to avoid

Block or redesign:

- Fish source presented as portable POSIX shell;
- copied Bash syntax whose Fish meaning differs or is invalid;
- a monolithic `config.fish` containing unrelated functions and side effects;
- startup-time network access, prompts, output, or expensive discovery;
- universal variables used as a hidden configuration database;
- unpinned plugin collections or framework adoption for one small function;
- network-to-`source` installation commands;
- string-built commands, `eval`, or sourced untrusted output;
- ignored `$status` or pipeline failures at a consequential boundary;
- scripts that parse structured data with ad hoc text operations; and
- `fish -c` strings embedded in schedulers or service definitions;
- Fish glue that owns substantial application logic or exceeds the general
  function and nesting tripwires in `CODING.md` and should become a small
  program in a structured language.

## Primary references

- [The Fish language](https://fishshell.com/docs/current/language.html)
- [Fish for Bash users](https://fishshell.com/docs/current/fish_for_bash_users.html)
- [`fish` command options](https://fishshell.com/docs/current/cmds/fish.html)
- [`argparse`](https://fishshell.com/docs/current/cmds/argparse.html)
- [`fish_indent`](https://fishshell.com/docs/current/cmds/fish_indent.html)
- [Fisher](https://github.com/jorgebucaran/fisher)
- [Fishtape](https://github.com/jorgebucaran/fishtape)
- [fish-lsp](https://github.com/ndonfris/fish-lsp)
