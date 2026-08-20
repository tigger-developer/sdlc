#!/bin/sh

set -eu

repository=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)

test_make_install_starts_interactive_installer_OT4_1() {
    plan=$(make --no-print-directory -n -C "$repository" install)
    case "$plan" in
    *"bin/sdlc-install"*) ;;
    *)
        printf '%s\n' "make install does not start bin/sdlc-install" >&2
        return 1
        ;;
    esac
    case "$plan" in
    *".local/bin"*)
        printf '%s\n' "make install still installs the reusable CLI link" >&2
        return 1
        ;;
    esac
}

test_make_install_propagates_failure_OT4_2() {
    temporary_directory=$(mktemp -d)
    trap 'trash "$temporary_directory"' EXIT
    fake_installer="$temporary_directory/failing-installer"
    printf '%s\n' '#!/bin/sh' 'exit 23' >"$fake_installer"
    chmod +x "$fake_installer"
    if make --no-print-directory -C "$repository" install INSTALLER="$fake_installer" >/dev/null 2>&1; then
        printf '%s\n' "make install hid the installer failure" >&2
        return 1
    fi
    trash "$temporary_directory"
    trap - EXIT
}

test_make_install_starts_interactive_installer_OT4_1
test_make_install_propagates_failure_OT4_2
printf '%s\n' "one-off Makefile installation tests passed"
