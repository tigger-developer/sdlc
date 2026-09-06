package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpDoesNotRequireDeployedSchema(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	if err := runCommand([]string{"--help"}, strings.NewReader(""), &output, &output); err != nil {
		t.Fatalf("runCommand(--help): %v", err)
	}
	if !strings.Contains(output.String(), "usage: sdlc-project-init") {
		t.Fatalf("help output missing usage: %q", output.String())
	}
}
