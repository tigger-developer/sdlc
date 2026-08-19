#!/usr/bin/env bash
# ABOUTME: Blocks known-dangerous shell commands before agent tool execution.
# ABOUTME: Intended for providers that pass Bash tool requests to a pre-use hook.
set -eo pipefail

command_json="$(cat)"
command_text="$(printf '%s' "$command_json" | jq -r '.tool_input.command // .toolInput.command // .command // empty')"
hook_event="$(printf '%s' "$command_json" | jq -r '.hook_event_name // empty')"
tool_name="$(printf '%s' "$command_json" | jq -r '.tool_name // empty')"

if [[ -z "$command_text" ]]; then
    exit 0
fi

block() {
    local reason="$1"

    if [[ "$hook_event" == "pre_tool_call" && "$tool_name" == "terminal" ]]; then
        jq -cn --arg reason "$reason" '{decision:"block",reason:$reason}'
        exit 0
    fi

    printf 'Blocked by agent-command-guard: %s\n' "$reason" >&2
    exit 2
}

TOKENS=()
TOKEN_TYPES=()

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

segment_invokes_python() {
    local start="$1"
    local end="$2"
    local index="$start"
    local executable basename argument

    while [[ "$index" -lt "$end" ]] && is_assignment "${TOKENS[index]}"; do
        ((index += 1))
    done
    while [[ "$index" -lt "$end" ]]; do
        executable="${TOKENS[index]}"
        basename="${executable##*/}"
        case "$basename" in
        python | python3)
            return 0
            ;;
        env | command | sudo)
            ((index += 1))
            while [[ "$index" -lt "$end" && ("${TOKENS[index]}" == -* || "${TOKENS[index]}" =~ ^[a-zA-Z_][a-zA-Z0-9_]*=) ]]; do
                ((index += 1))
            done
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
                    if (command_invokes_python "${TOKENS[index + 1]}"); then
                        return 0
                    fi
                    return 1
                fi
            done
            return 1
            ;;
        *)
            return 1
            ;;
        esac
    done
    return 1
}

command_invokes_python() {
    local command="$1"
    local segment_start=0
    local index

    tokenize_shell_command "$command"
    for ((index = 0; index <= ${#TOKENS[@]}; index++)); do
        if [[ "$index" -eq ${#TOKENS[@]} || "${TOKEN_TYPES[index]}" == "separator" ]]; then
            if segment_invokes_python "$segment_start" "$index"; then
                return 0
            fi
            segment_start=$((index + 1))
        fi
    done
    return 1
}

if command_invokes_python "$command_text"; then
    block 'direct python and python3 interpreter commands are prohibited; use the task-appropriate tool or a project-owned entry point.'
fi

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
