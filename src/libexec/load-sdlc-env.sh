#!/usr/bin/env bash
set -eo pipefail

if (($# < 1)); then
    echo "usage: load-sdlc-env.sh [--filter] FILE [KEY ...]" >&2
    exit 2
fi

mode=load
if [[ "$1" == "--filter" ]]; then
    mode=filter
    shift
fi

if (($# < 1)); then
    echo "usage: load-sdlc-env.sh [--filter] FILE [KEY ...]" >&2
    exit 2
fi

config_path=$1
shift
requested_keys=("$@")
readonly -a requested_keys

if [[ ! -e "$config_path" ]]; then
    exit 0
fi
if [[ ! -f "$config_path" ]]; then
    echo "configuration is not a regular file: $config_path" >&2
    exit 1
fi

for key in "${requested_keys[@]}"; do
    if [[ ! "$key" =~ ^[A-Z][A-Z0-9_]*$ ]]; then
        echo "invalid configuration key: $key" >&2
        exit 2
    fi
    unset "$key"
done

if [[ "$mode" == "filter" ]]; then
    while IFS= read -r line || [[ -n "$line" ]]; do
        candidate=${line#"${line%%[![:space:]]*}"}
        if [[ "$candidate" == export[[:space:]]* ]]; then
            candidate=${candidate#export}
            candidate=${candidate#"${candidate%%[![:space:]]*}"}
        fi
        managed=false
        for key in "${requested_keys[@]}"; do
            if [[ "$candidate" =~ ^${key}[[:space:]]*= ]]; then
                managed=true
                break
            fi
        done
        if [[ "$managed" == false ]]; then
            printf '%s\n' "$line"
        fi
    done <"$config_path"
    exit 0
fi

# The configuration is intentionally evaluated by Bash so shell expansion and
# references between variables retain their documented shell semantics.
# shellcheck disable=SC1090
source "$config_path" 1>&2

for key in "${requested_keys[@]}"; do
    if [[ -v "$key" ]]; then
        printf '%s\0%s\0' "$key" "${!key}"
    fi
done
