#!/usr/bin/env bash
set -eo pipefail

if (($# < 1)); then
    echo "usage: load-sdlc-env.sh FILE [KEY ...]" >&2
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

# The configuration is intentionally evaluated by Bash so shell expansion and
# references between variables retain their documented shell semantics.
# shellcheck disable=SC1090
source "$config_path" 1>&2

for key in "${requested_keys[@]}"; do
    if [[ -v "$key" ]]; then
        printf '%s\0%s\0' "$key" "${!key}"
    fi
done
