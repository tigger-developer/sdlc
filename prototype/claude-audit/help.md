# Claude safe-mode audit prototype

Undeployed paired experiment, not an SDLC gate. Receives synthetic audit evidence
on stdin, invokes Claude once with all model tools disabled, and returns YAML on
stdout. Diagnostics go to stderr. It never writes canonical `audits.yaml`.

Usage: claude-audit --state-dir DIRECTORY [--resume SESSION_ID] < evidence.txt

- `--state-dir`: required existing scratch directory. Claude configuration,
  working directory, logs, native sessions and raw results stay beneath it.
- `--resume`: native session ID from the previous successful output. Omit to
  start. Resume must use the same scratch directory.
- `--dry-run`: print invocation arguments as JSON; do not invoke Claude or
  create files. Stdin is not needed.
- `-h`, `--help`: show this help without side effects.
- `--version`: show the prototype version without side effects.

From the SDLC checkout, build with:

```sh
go build -o bin/claude-audit-prototype ./prototype/claude-audit
```

Create scratch space using `mktemp -d`. Then, replacing the example directory:

```sh
bin/claude-audit-prototype --state-dir /path/to/scratch < evidence.txt
```

```sh
bin/claude-audit-prototype --state-dir /path/to/scratch --resume RETURNED_SESSION_ID < correction.txt
```

`config.yaml` defines Sonnet, the prompt, the two-minute timeout and three-turn
bound; the response schema is an adjacent file. Both are embedded at build time.
There are no retries or automatic provider fallbacks. Exit codes: 0 successful
response (including an audit FAIL), 1 runtime or response failure, 2 invalid CLI.

This slice supports Unix hosts. Authentication remains Claude-owned. An isolated configuration directory may
require its own login; the prototype does not extract or copy credentials.
Do not switch to paid API credentials silently if subscription login is absent.
For an operator-run subscription login, use the same scratch config directory:

```sh
CLAUDE_CONFIG_DIR=/path/to/scratch/config claude auth login
```

Keep scratch state through the paired start/resume test. Afterwards the operator
may remove that exact directory with a recoverable deletion tool. The prototype
does not delete a caller-owned directory. Results contain synthetic evidence only.
Metered invocations must never be put into `make test`, builds, CI or schedules.

This first slice does not provide original-file reading, a filesystem allowlist,
automatic session recovery after timeout, or production audit-record integration.
