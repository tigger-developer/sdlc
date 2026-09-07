package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestW004HelpAndVersionExitSuccessfully(t *testing.T) {
	for _, arguments := range [][]string{{"-h"}, {"--help"}, {"start", "-h"}, {"--version"}} {
		var output bytes.Buffer
		var diagnostics bytes.Buffer
		if err := run(arguments, strings.NewReader(""), &output, &diagnostics); err != nil {
			t.Fatalf("%v returned %v: %s", arguments, err, diagnostics.String())
		}
		combined := output.String() + diagnostics.String()
		if arguments[0] != "--version" && !strings.Contains(combined, "Internal SDLC helper") {
			t.Fatalf("%v help lacks command purpose: %s", arguments, combined)
		}
	}
}
