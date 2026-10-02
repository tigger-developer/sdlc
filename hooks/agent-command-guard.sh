#!/usr/bin/env bash
# ABOUTME: Blocks prohibited commands and direct .env reads before agent tool execution.
# ABOUTME: Accepts Claude, Codex, Copilot, and Hermes pre-tool hook payloads.
set -eo pipefail

command_json="$(cat)"
if ! printf '%s' "$command_json" | jq -e 'type == "object"' >/dev/null 2>&1; then
    printf 'agent-command-guard: invalid JSON object payload.\n' >&2
    exit 2
fi
command_text="$(printf '%s' "$command_json" | jq -r '.tool_input.command // .toolInput.command // .toolArgs.command // .command // empty')"
hook_event="$(printf '%s' "$command_json" | jq -r '.hook_event_name // .hookEventName // empty')"
tool_name="$(printf '%s' "$command_json" | jq -r '.tool_name // .toolName // empty')"

block() {
    local reason="$1"

    if [[ "$hook_event" == "pre_tool_call" ]]; then
        jq -cn --arg reason "$reason" '{decision:"block",reason:$reason}'
        exit 0
    fi

    printf 'Blocked by agent-command-guard: %s\n' "$reason" >&2
    exit 2
}

if [[ -z "$hook_event" || -z "$tool_name" ]]; then
    block 'unrecognized pre-tool payload shape is missing its event or tool name.'
fi

names_exact_env_file() {
    local value="$1"

    value="${value#file://}"
    while [[ "$value" == \<* || "$value" == \>* ]]; do
        value="${value:1}"
    done
    value="${value//\\//}"
    [[ "$value" == ".env" || "$value" == */.env ]]
}

tool_reads_files() {
    local normalized

    normalized="$(printf '%s' "$tool_name" | tr '[:upper:]' '[:lower:]')"
    case "$normalized" in
    read | read_file | readfile | view | grep | rg | glob | search | search_files | searchfiles | open_file | *__read* | *__grep* | *__search*)
        return 0
        ;;
    esac
    return 1
}

tool_writes_files() {
    local normalized

    normalized="$(printf '%s' "$tool_name" | tr '[:upper:]' '[:lower:]')"
    case "$normalized" in
    write | write_file | writefile | create_file | createfile | edit | edit_file | editfile | apply_patch | *__write* | *__edit*)
        return 0
        ;;
    esac
    return 1
}

payload_references_env_file() {
    local candidate

    while IFS= read -r candidate; do
        if names_exact_env_file "$candidate"; then
            return 0
        fi
    done < <(printf '%s' "$command_json" | jq -r '
        (.tool_input // .toolInput // .toolArgs // {}) |
        .. | strings
    ')
    return 1
}

read_candidate_count=0
if tool_reads_files; then
    while IFS= read -r candidate; do
        ((read_candidate_count += 1))
        if names_exact_env_file "$candidate"; then
            block 'reading a file whose exact basename is .env is prohibited.'
        fi
    done < <(printf '%s' "$command_json" | jq -r '
        def file_values:
            if type == "object" then
                to_entries[] |
                if (.key | ascii_downcase | test("^(file|files|filename|file_name|filepath|file_path|path|paths|glob|include|target|uri)$")) then
                    .value | .. | strings
                else
                    .value | file_values
                end
            elif type == "array" then
                .[] | file_values
            else
                empty
            end;
        (.tool_input // .toolInput // .toolArgs // {}) |
        if type == "string" then . else file_values end
    ')
    if [[ "$read_candidate_count" -eq 0 ]]; then
        block 'recognized file-reading event is missing a file or path field.'
    fi
    exit 0
fi

if tool_writes_files; then
    exit 0
fi

normalized_tool="$(printf '%s' "$tool_name" | tr '[:upper:]' '[:lower:]')"
case "$normalized_tool" in
bash | shell | terminal | exec | execute | command | exec_command | shell_command)
    if [[ -z "$command_text" ]]; then
        block 'recognized command event is missing its command field.'
    fi
    guard_binary="$(cd "$(dirname "$0")/.." && pwd)/bin/sdlc-guard-shell"
    if [[ ! -x "$guard_binary" ]]; then
        block 'installed shell guard binary is missing or not executable.'
    fi
    if reason="$(printf '%s' "$command_text" | "$guard_binary" 2>&1)"; then
        exit 0
    fi
    block "$reason"
    ;;
*)
    if payload_references_env_file; then
        block 'unrecognized tool payload references a protected .env path.'
    fi
    if [[ -n "$command_text" ]]; then
        block 'unrecognized tool with a command field cannot be classified safely.'
    fi
    exit 0
    ;;
esac
