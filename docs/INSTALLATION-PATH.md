---
title: SDLC Installation Path and Limited Sudo
version: 1
last-updated: 2026-11-11
---

# Why the SDLC uses /opt/sdlc

The SDLC pre-instructions refer to framework documents by fixed absolute paths.
We tried parameterized paths and environment variables, but in our tests agents
repeatedly explored and read the whole framework while resolving them.

That defeats the routing designed for **cost and accuracy**. For example, a Go
task needs the Go standards; loading the Go, JavaScript, Swift, Python, Perl,
shell, and web standards together weakens the focus on the applicable Go rules.
It also loads substantially more text on **every agent invocation**, multiplying
token use and cost. Fixed paths are therefore a requirement of this design.

The SDLC is also intended for CI builders: restricted, managed users on
ephemeral hosts, potentially with several builders sharing one host. Every
builder needs the same fixed path to the shared framework. This is why the
canonical installation moved from `~/.agents/sdlc` to `/opt/sdlc`.

## Why /opt/sdlc may need sudo

Creating `/opt/sdlc` may require privileged access, particularly on macOS.
The installer first tries to create it as the current user. If the operating
system refuses, it links to this explanation and asks whether it may run only
`sudo mkdir /opt/sdlc` and `sudo chown <current-user> /opt/sdlc`. It then
continues as the normal user; further privileged access should not be needed.

**Do not run the installer with sudo:** it rejects root execution. To avoid the
installer's sudo prompt, pre-create `/opt/sdlc` and grant the installing user
write access before running it. If `/opt/sdlc` already exists but is not
writable, the installer stops.

## Installed layout

The public installer places standards in `/opt/sdlc/docs`, skills in
`/opt/sdlc/skills`, commands in `/opt/sdlc/bin`, hooks in `/opt/sdlc/hooks`, and
global configuration in `/opt/sdlc/sdlc.yaml`. It links each SDLC skill
directory into `~/.agents/skills` and supported provider skill directories for
harness discovery; personal skills remain separate. The public installer does
not construct global `AGENTS.md` or `CLAUDE.md` files. A separate personal
installer can deploy those files and route agents to `/opt/sdlc/docs/MAIN.md`.

On an existing installation, the installer carries values from
`~/.agents/sdlc.yaml` into `/opt/sdlc/sdlc.yaml` when the latter is absent.
The old file remains available for recovery but is no longer read by SDLC
commands.

For a different fixed location, fork the project and replace the paths in its
installer and instructions, including the tests.
