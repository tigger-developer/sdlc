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
	if !strings.Contains(output.String(), "usage: sdlc-init") {
		t.Fatalf("help output missing usage: %q", output.String())
	}
	for _, explanation := range []string{"Run it once from the project root", "does not install or update the global SDLC framework"} {
		if !strings.Contains(output.String(), explanation) {
			t.Errorf("help output missing %q: %q", explanation, output.String())
		}
	}
	for _, option := range []string{"--override-global-config", "--no-agent-scan"} {
		if !strings.Contains(output.String(), option) {
			t.Errorf("help output missing %q: %q", option, output.String())
		}
	}
}
