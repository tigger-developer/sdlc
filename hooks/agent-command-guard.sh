#!/usr/bin/env bash
# ABOUTME: Blocks known-dangerous shell commands before agent tool execution.
# ABOUTME: Intended for providers that pass Bash tool requests to a pre-use hook.
set -eo pipefail

command_json="$(cat)"
command_text="$(printf '%s' "$command_json" | jq -r '.tool_input.command // .toolInput.command // .command // empty')"

if [[ -z "$command_text" ]]; then
    exit 0
fi

block() {
    local reason="$1"

    printf 'Blocked by agent-command-guard: %s\n' "$reason" >&2
    exit 2
}

if [[ "$command_text" =~ (^|[[:space:]])--no-verify([[:space:]]|$) ]]; then
    block 'git hook bypass flags are prohibited; run the hooks or surface the failing hook.'
fi

if [[ "$command_text" =~ (^|[[:space:]])--no-hooks([[:space:]]|$) ]]; then
    block 'hook bypass flags are prohibited; run the hooks or surface the failing hook.'
fi

if [[ "$command_text" =~ (^|[[:space:]])--no-pre-commit-hook([[:space:]]|$) ]]; then
    block 'pre-commit hook bypass flags are prohibited; run the hook or surface the failure.'
fi

if [[ "$command_text" =~ (^|[[:space:];|&])chmod[[:space:]]+777([[:space:]]|$) ]]; then
    block 'chmod 777 is prohibited; use the least permissive mode that works.'
fi

if [[ "$command_text" =~ (^|[[:space:];|&])git[[:space:]]+remote[[:space:]]+add([[:space:]]|$) ]]; then
    block 'git remote add widens repository access; ask the user first.'
fi

if [[ "$command_text" =~ (^|[[:space:];|&])git[[:space:]]+push[[:space:]]+([^;&|]*[[:space:]])?(-f|--force|--force-with-lease)([[:space:]]|$) ]]; then
    block 'force-push is prohibited unless the user explicitly authorizes it.'
fi

if [[ "$command_text" =~ (^|[[:space:];|&])gh[[:space:]]+repo[[:space:]]+(create|edit)([[:space:]]|$) ]]; then
    block 'GitHub repository create/edit widens or changes access; ask the user first.'
fi

if [[ "$command_text" =~ (^|[[:space:];|&])(source|\.)[[:space:]]+ ]]; then
    block 'sourcing arbitrary files bypasses command policy; run an explicit command instead.'
fi
