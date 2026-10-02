// ABOUTME: Applies the SDLC shell-command policy to one pre-tool command.
// ABOUTME: Exits without output when allowed and reports a block reason otherwise.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/tigger-developer/sdlc/internal/commandguard"
)

const maxCommandBytes = 1 << 20

var buildRelease = "development"

func main() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "-h", "--help":
			fmt.Fprintln(os.Stdout, "sdlc-guard-shell reads one shell command from stdin and checks SDLC policy.")
			return
		case "--version":
			fmt.Fprintln(os.Stdout, "sdlc-guard-shell "+buildRelease)
			return
		}
	}
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "sdlc-guard-shell accepts no arguments; use stdin.")
		os.Exit(2)
	}
	input, err := io.ReadAll(io.LimitReader(os.Stdin, maxCommandBytes+1))
	if err != nil {
		fmt.Fprintln(os.Stderr, "shell command could not be read safely.")
		os.Exit(2)
	}
	if len(input) > maxCommandBytes {
		fmt.Fprintln(os.Stderr, "shell command exceeds the guard's inspection limit.")
		os.Exit(2)
	}
	if reason := commandguard.CheckShell(string(input)); reason != "" {
		fmt.Fprintln(os.Stderr, reason)
		os.Exit(2)
	}
}
