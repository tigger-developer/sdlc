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

read_candidate_count=0
if [[ -z "$command_text" ]] && tool_reads_files; then
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
fi

if [[ -z "$command_text" ]]; then
    case "$(printf '%s' "$tool_name" | tr '[:upper:]' '[:lower:]')" in
    bash | shell | terminal | exec | execute | command)
        block 'recognized command event is missing its command field.'
        ;;
    esac
    exit 0
fi

TOKENS=()
TOKEN_TYPES=()
BLOCK_REASON=""

append_word() {
    local word="$1"

    if [[ -n "$word" ]]; then
        TOKENS+=("$word")
        TOKEN_TYPES+=("word")
    fi
}

append_separator() {
    TOKENS+=("$1")
    TOKEN_TYPES+=("separator")
}

tokenize_shell_command() {
    local input="$1"
    local token=""
    local quote=""
    local character next_character
    local index

    TOKENS=()
    TOKEN_TYPES=()
    for ((index = 0; index < ${#input}; index++)); do
        character="${input:index:1}"
        if [[ -n "$quote" ]]; then
            if [[ "$character" == "$quote" ]]; then
                quote=""
            elif [[ "$character" == "\\" && "$quote" == '"' && $((index + 1)) -lt ${#input} ]]; then
                ((index += 1))
                next_character="${input:index:1}"
                token+="$next_character"
            else
                token+="$character"
            fi
            continue
        fi

        case "$character" in
        "'" | '"')
            quote="$character"
            ;;
        "\\")
            if [[ $((index + 1)) -lt ${#input} ]]; then
                ((index += 1))
                next_character="${input:index:1}"
                token+="$next_character"
            fi
            ;;
        " " | $'\t' | $'\n' | $'\r')
            append_word "$token"
            token=""
            ;;
        ";" | "|" | "&" | "(" | ")" | "{" | "}")
            append_word "$token"
            token=""
            append_separator "$character"
            ;;
        *)
            token+="$character"
            ;;
        esac
    done
    append_word "$token"
}

is_assignment() {
    [[ "$1" =~ ^[a-zA-Z_][a-zA-Z0-9_]*= ]]
}

skip_wrapper_options() {
    local wrapper="$1"
    local end="$2"
    local option

    while [[ "$index" -lt "$end" ]]; do
        option="${TOKENS[index]}"
        if is_assignment "$option"; then
            ((index += 1))
            continue
        fi
        case "$wrapper:$option" in
        command:-v | command:-V)
            return 2
            ;;
        env:-i | env:--ignore-environment | env:-0 | env:--null | env:--debug | command:-p | sudo:-n | sudo:--non-interactive | sudo:-E | sudo:--preserve-env | sudo:-H | sudo:--set-home | sudo:-S | sudo:--stdin | sudo:-b | sudo:--background | sudo:-k | sudo:--reset-timestamp | sudo:-K | sudo:--remove-timestamp | sudo:-v | sudo:--validate | sudo:-l | sudo:--list | sudo:-V | sudo:--version | sudo:-e | sudo:--edit)
            ((index += 1))
            ;;
        env:-u | env:--unset | env:-C | env:--chdir | sudo:-u | sudo:--user | sudo:-g | sudo:--group | sudo:-h | sudo:--host | sudo:-p | sudo:--prompt | sudo:-C | sudo:--close-from | sudo:-R | sudo:--chroot | sudo:-T | sudo:--command-timeout)
            index=$((index + 2))
            ;;
        env:--unset=* | env:--chdir=* | sudo:--user=* | sudo:--group=* | sudo:--host=* | sudo:--prompt=* | sudo:--close-from=* | sudo:--chroot=* | sudo:--command-timeout=*)
            ((index += 1))
            ;;
        *:--)
            ((index += 1))
            return 0
            ;;
        *:-*)
            BLOCK_REASON="unrecognized $wrapper wrapper option prevents safe command classification."
            return 3
            ;;
        *)
            return 0
            ;;
        esac
    done
    return 0
}

segment_invokes_guarded_action() {
    local start="$1"
    local end="$2"
    local index="$start"
    local executable basename argument subcommand wrapper_status

    while [[ "$index" -lt "$end" ]] && is_assignment "${TOKENS[index]}"; do
        ((index += 1))
    done
    while [[ "$index" -lt "$end" ]]; do
        executable="${TOKENS[index]}"
        basename="${executable##*/}"
        case "$basename" in
        python | python3)
            BLOCK_REASON='direct python and python3 interpreter commands are prohibited; use the task-appropriate tool or a project-owned entry point.'
            return 0
            ;;
        rm)
            BLOCK_REASON='rm is prohibited; use recoverable deletion such as trash.'
            return 0
            ;;
        sed | awk)
            BLOCK_REASON='sed and awk are prohibited; use format-aware tools or an explicit patch.'
            return 0
            ;;
        env | command | sudo)
            ((index += 1))
            if skip_wrapper_options "$basename" "$end"; then
                :
            else
                wrapper_status="$?"
                [[ "$wrapper_status" -eq 2 ]] && return 1
                [[ "$wrapper_status" -eq 3 ]] && return 0
            fi
            continue
            ;;
        if | then | elif | else | while | until | do | time | "!")
            ((index += 1))
            while [[ "$index" -lt "$end" && "${TOKENS[index]}" == -* ]]; do
                ((index += 1))
            done
            continue
            ;;
        bash | sh | zsh)
            for ((index += 1; index < end; index++)); do
                argument="${TOKENS[index]}"
                if [[ "$argument" =~ ^-[a-zA-Z]*c[a-zA-Z]*$ && $((index + 1)) -lt "$end" ]]; then
                    if command_invokes_prohibited "${TOKENS[index + 1]}"; then
                        return 0
                    fi
                    return 1
                fi
            done
            return 1
            ;;
        source | .)
            BLOCK_REASON='sourcing arbitrary files bypasses command policy; run an explicit command instead.'
            return 0
            ;;
        chmod)
            for ((index += 1; index < end; index++)); do
                [[ "${TOKENS[index]}" == "777" ]] || continue
                BLOCK_REASON='chmod 777 is prohibited; use the least permissive mode that works.'
                return 0
            done
            return 1
            ;;
        git)
            ((index += 1))
            while [[ "$index" -lt "$end" && "${TOKENS[index]}" == -* ]]; do
                ((index += 1))
            done
            [[ "$index" -lt "$end" ]] || return 1
            subcommand="${TOKENS[index]}"
            if [[ "$subcommand" == "remote" && $((index + 1)) -lt "$end" && "${TOKENS[index + 1]}" == "add" ]]; then
                BLOCK_REASON='git remote add widens repository access; ask the user first.'
                return 0
            fi
            for ((index += 1; index < end; index++)); do
                case "${TOKENS[index]}" in
                --no-verify | --no-hooks | --no-pre-commit-hook)
                    BLOCK_REASON='git hook bypass flags are prohibited; run the hooks or surface the failing hook.'
                    return 0
                    ;;
                -f | --force | --force-with-lease | --force-with-lease=*)
                    if [[ "$subcommand" == "push" ]]; then
                        BLOCK_REASON='force-push is prohibited unless the user explicitly authorizes it.'
                        return 0
                    fi
                    ;;
                esac
            done
            return 1
            ;;
        gh)
            if [[ $((index + 2)) -lt "$end" && "${TOKENS[index + 1]}" == "repo" && ("${TOKENS[index + 2]}" == "create" || "${TOKENS[index + 2]}" == "edit") ]]; then
                BLOCK_REASON='GitHub repository create/edit widens or changes access; ask the user first.'
                return 0
            fi
            return 1
            ;;
        *)
            if [[ "$basename" != "echo" && "$basename" != "printf" && "$basename" != "rg" ]]; then
                for ((index += 1; index < end; index++)); do
                    case "${TOKENS[index]}" in
                    --no-hooks | --no-pre-commit-hook)
                        BLOCK_REASON='hook bypass flags are prohibited; run the hooks or surface the failing hook.'
                        return 0
                        ;;
                    esac
                done
            fi
            return 1
            ;;
        esac
    done
    return 1
}

command_invokes_prohibited() {
    local command="$1"
    local segment_start=0
    local index

    tokenize_shell_command "$command"
    for ((index = 0; index <= ${#TOKENS[@]}; index++)); do
        if [[ "$index" -eq ${#TOKENS[@]} || "${TOKEN_TYPES[index]}" == "separator" ]]; then
            if segment_invokes_guarded_action "$segment_start" "$index"; then
                return 0
            fi
            segment_start=$((index + 1))
        fi
    done
    return 1
}

command_references_env_file() {
    local command="$1"
    local token executable basename
    local index=0
    local pattern_seen=false
    local option_value=""
    local -a command_tokens

    tokenize_shell_command "$command"
    command_tokens=("${TOKENS[@]}")
    while [[ "$index" -lt ${#command_tokens[@]} ]] && is_assignment "${command_tokens[index]}"; do
        ((index += 1))
    done
    [[ "$index" -lt ${#command_tokens[@]} ]] || return 1
    executable="${command_tokens[index]}"
    basename="${executable##*/}"
    ((index += 1))

    if [[ "$basename" == "bash" || "$basename" == "sh" || "$basename" == "zsh" ]]; then
        while [[ "$index" -lt ${#command_tokens[@]} ]]; do
            token="${command_tokens[index]}"
            if [[ "$token" =~ ^-[a-zA-Z]*c[a-zA-Z]*$ && $((index + 1)) -lt ${#command_tokens[@]} ]]; then
                command_references_env_file "${command_tokens[index + 1]}"
                return
            fi
            ((index += 1))
        done
        return 1
    fi
    if [[ "$basename" == "rg" || "$basename" == "grep" ]]; then
        while [[ "$index" -lt ${#command_tokens[@]} ]]; do
            token="${command_tokens[index]}"
            if [[ -n "$option_value" ]]; then
                if [[ "$option_value" == "file" ]] && names_exact_env_file "$token"; then
                    return 0
                fi
                pattern_seen=true
                option_value=""
                ((index += 1))
                continue
            fi
            case "$token" in
            -e | --regexp)
                option_value="pattern"
                ;;
            -f | --file)
                option_value="file"
                ;;
            --regexp=*)
                pattern_seen=true
                ;;
            --file=*)
                names_exact_env_file "${token#*=}" && return 0
                ;;
            -*)
                ;;
            *)
                if [[ "$pattern_seen" == false ]]; then
                    pattern_seen=true
                elif names_exact_env_file "$token"; then
                    return 0
                fi
                ;;
            esac
            ((index += 1))
        done
        return 1
    fi

    for (( ; index < ${#command_tokens[@]}; index++)); do
        token="${command_tokens[index]}"
        if names_exact_env_file "$token"; then
            return 0
        fi
        if [[ "$token" == *[[:space:]]* ]] && command_references_env_file "$token"; then
            return 0
        fi
    done
    return 1
}

if command_references_env_file "$command_text"; then
    block 'reading a file whose exact basename is .env is prohibited.'
fi

if command_invokes_prohibited "$command_text"; then
    block "$BLOCK_REASON"
fi
