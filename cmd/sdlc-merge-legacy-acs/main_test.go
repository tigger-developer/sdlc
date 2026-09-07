package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpAndVersionExitSuccessfully(t *testing.T) {
	for _, arguments := range [][]string{{"-h"}, {"--help"}, {"--version"}} {
		var output bytes.Buffer
		var diagnostics bytes.Buffer
		if status := commandStatus(arguments, &output, &diagnostics); status != 0 {
			t.Fatalf("%v status = %d: %s", arguments, status, diagnostics.String())
		}
		combined := output.String() + diagnostics.String()
		if !strings.Contains(combined, "sdlc-merge-legacy-acs") {
			t.Fatalf("%v output lacks command name: %s", arguments, combined)
		}
		if arguments[0] != "--version" && !strings.Contains(combined, "One-time migration helper") {
			t.Fatalf("%v help lacks command purpose: %s", arguments, combined)
		}
	}
}
