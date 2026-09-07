# Lua Standards

These standards apply to maintained Lua programs, embedded extensions, plugins,
and host configuration such as Hammerspoon. The general rules in
`~/.agents/sdlc/CODING.md` also apply.

## Selection and language boundary

Lua is the preferred scripting language ahead of shell when:

- the host application already embeds Lua, including Hammerspoon, Pandoc,
  Neovim, or another documented plugin interface;
- compact, portable logic benefits from tables, functions, modules, and a real
  parser rather than command pipelines;
- a small embedded extension needs low startup and memory overhead; or
- a standalone script has enough data handling or control flow to be unsafe or
  obscure in shell, while a larger compiled tool is not justified.

Do not add a standalone Lua runtime merely to replace a clear, trivial shell
adapter. When Lua is not supplied by the host, record and justify the runtime,
version, packaging, and deployment dependency. Prefer Go for a substantial
portable CLI or service, Rust for a safety- or performance-critical system, and
the application's existing primary language when it already fits.

Lua versions and hosts are not interchangeable. Record the minimum Lua version
or exact host version. Check whether the host uses standard Lua, LuaJIT, or a
restricted API before selecting syntax, libraries, binary modules, or integer
and numeric behaviour.

## Strengths and weaknesses

Lua is strong at:

- embedding and extension through a small language and C API;
- configuration and data description through tables;
- event-driven plugins and automation inside a capable host;
- coroutines and lightweight cooperative workflows; and
- readable scripts whose logic has outgrown shell.

Its limits must remain visible:

- dynamic typing moves many interface errors to runtime;
- the standard library is deliberately small;
- global variables are easy to create accidentally;
- array-like tables are conventionally one-based and holes make length
  behaviour unsuitable as a general count;
- host APIs and Lua versions can fragment portability; and
- native LuaRocks modules add compiler, ABI, architecture, and packaging risk.

Do not compensate for these limits with hidden conventions. Validate boundaries,
keep modules small, test public behaviour, and declare the runtime contract.

## Structure and modules

- Keep variables and functions `local` unless they are an intentional exported
  interface.
- Put one cohesive responsibility in each module and return an explicit module
  table. Do not mutate a global namespace as an import side effect.
- Keep host integration at the boundary. Put deterministic transformation and
  decision logic in ordinary modules that can be tested without launching the
  host.
- Pass dependencies into modules where clocks, files, networks, UI automation,
  or host state would otherwise become hidden globals.
- Use tables as records only with documented field names and validation at
  external boundaries. Use arrays only when their sequence and density are
  intentional.
- Avoid metatable cleverness unless it provides a small, documented abstraction
  that ordinary tables and functions cannot express clearly.

## Errors and resource handling

- Return `nil, error` or raise an error consistently within one interface. Do
  not mix conventions unpredictably.
- Add context at the boundary that can explain or recover from a failure.
- Use `pcall` or `xpcall` only around a boundary where failure is expected and
  can be handled. Never wrap a whole program merely to suppress a traceback.
- Check file, process, network, and host-API results. Close or release resources
  on every path.
- Put timeouts and explicit bounds on network calls, retries, timers, event
  queues, and external processes.
- Do not use a fixed sleep as synchronization. Observe the host or external
  process state with a bounded timeout.

## Dependencies and common libraries

Prefer the Lua standard library and the host's documented API. Use LuaRocks when
a standalone or embedded project genuinely needs third-party modules; pin the
Lua version, rock versions, and platform assumptions through the project's
reproducible mechanism.

Common candidates require the same dependency review as any other library:

| Concern | Candidate | Notes |
|---|---|---|
| Filesystem | LuaFileSystem | Prefer over shelling out for portable filesystem metadata and directory work. |
| Parsing | LPeg | Use for a real grammar; do not grow complex formats from pattern fragments. |
| JSON | `dkjson` or `lua-cjson` | `dkjson` is pure Lua; `lua-cjson` is faster but introduces a native dependency. |
| Networking | LuaSocket and LuaSec | Use only with explicit timeouts and TLS requirements. |
| Testing | Busted | Appropriate for a maintained Lua module or non-trivial plugin. |
| Formatting | StyLua | Apply one project-owned configuration consistently. |
| Static checks | Luacheck | Reject accidental globals and actionable warnings. |

Do not adopt a broad utility framework merely to gain one helper. Review a
rock's maintenance, licence, supported Lua versions, native code, transitive
dependencies, and installation behaviour before adoption.

## Hammerspoon on macOS

Hammerspoon is a strong fit for operator-owned macOS automation involving
applications, windows, hotkeys, menus, accessibility objects, events, timers,
tasks, notifications, and other APIs exposed through its `hs` modules.
Hammerspoon supplies its own Lua runtime; a Hammerspoon-only configuration must
not acquire a separate system Lua dependency without another requirement.

- Keep `init.lua` as a small coordinator and place substantial behaviour in
  named modules or reviewed Spoons.
- Prefer documented `hs` modules over AppleScript, shell commands, simulated
  keystrokes, or Accessibility automation. Use the narrowest capable boundary.
- Use `hs.task` for a necessary external process and pass arguments separately;
  do not construct a shell command string.
- Keep callbacks short. Debounce noisy watchers and retain references required
  to keep watchers, timers, and hotkeys alive.
- Stop or delete watchers, timers, event taps, and hotkeys when reloading or
  replacing them. Reloading must not multiply side effects.
- Treat Accessibility, input monitoring, screen recording, automation, and
  location permissions as explicit security boundaries. Never widen them or
  instruct the user to grant them without a requirement.
- Never edit `~/.hammerspoon` or install Spoons without explicit authority for
  those exact user-owned paths and effects.
- Pin or review external Spoons. A Spoon executes with Hammerspoon's user-level
  authority and is not a harmless configuration fragment.

Use a user test for visual placement, interactive timing, ergonomics, or an
actual application workflow. Automated unit tests may cover extracted pure Lua
logic, but they do not prove macOS permissions or third-party UI behaviour.

## Standalone command-line programs

Every maintained Lua CLI follows the common interface contract:

- equivalent `-h` and `--help` output;
- `--version`;
- `--dry-run` when effects can be previewed faithfully;
- exit status 0 for success, 1 for operational failure, and 2 for invalid
  invocation unless an established external contract differs;
- normal results on stdout and diagnostics on stderr; and
- no side effects for help or version requests.

Do not hand-roll an elaborate parser. Use a small explicit parser for a small
interface or a reviewed library when subcommands and repeated options justify
it. Never evaluate user input as Lua or interpolate it into shell commands.

## Security and data

- Never use `load`, `loadfile`, `dofile`, or host equivalents on untrusted or
  downloaded content. Lua code is executable authority, not a data format.
- Never use Lua patterns as a parser for JSON, YAML, XML, HTML, or another
  structured language. Use a format-aware library.
- Validate paths against their permitted root before file access.
- Keep secrets out of Lua source, logs, global tables, Hammerspoon settings, and
  error messages.
- Escape values for their destination context. Lua strings do not become safe
  shell, HTML, URL, SQL, or regular-expression values merely by quoting them.
- Avoid native modules unless their capability and deployment cost are
  justified and their platform compatibility is tested.

For a deployable Lua application, expose the `make vulncheck` contract from
`~/.agents/sdlc/SECURITY.md`. Scan every supported resolved dependency manifest
with the selected ecosystem-capable scanner. If available tooling cannot assess
the LuaRocks graph, record that limitation explicitly and scan the built
artefact or deployment closure with Trivy or OSV-Scanner where supported. Never
claim dependency coverage from a scanner that does not understand the inputs.

## Verification

- Parse every changed file with the exact Lua or host runtime selected by the
  project, using a syntax-only mode where available.
- Format with StyLua and run Luacheck under committed project configuration when
  the project uses those tools.
- Run Busted or the established behavioural suite for maintained standalone
  modules.
- Test host-independent modules outside the host where that is faithful; test
  host integration through the host's documented reload and diagnostic path.
- Keep metered services and live UI automation out of persistent regression
  tests unless the SDLC test rules explicitly permit the selected evidence type.

## Anti-patterns to avoid

Block or redesign:

- accidental or unexplained globals;
- one giant `init.lua` or script containing unrelated responsibilities;
- string-built commands passed through a shell;
- `load` or `dofile` applied to data or untrusted content;
- ignored `nil` or error returns and blanket `pcall` suppression;
- using `#table` where sparse keys make the result ambiguous;
- implicit dependence on a personal Lua path, installed rocks, or host config;
- native rocks added without an ABI and packaging plan;
- busy timers, unbounded watchers, leaked event taps, or duplicate hotkeys;
- Hammerspoon used for a problem better solved by a supported application API,
  service manager, or small standalone tool; and
- Lua used as sprawling application infrastructure after its embedding or
  scripting advantage has disappeared.

## Primary references

- [Lua 5.4 Reference Manual](https://www.lua.org/manual/5.4/manual.html)
- [Programming in Lua](https://www.lua.org/pil/)
- [LuaRocks](https://luarocks.org/)
- [Hammerspoon API documentation](https://www.hammerspoon.org/docs/)
- [Hammerspoon Spoon documentation](https://github.com/Hammerspoon/hammerspoon/blob/master/SPOONS.md)
