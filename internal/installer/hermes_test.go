package installer

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestW004HermesGuardNormalizationCollapsesManagedDuplicates(t *testing.T) {
	var document yaml.Node
	input := `personal: true
hooks:
  pre_tool_call:
    - command: custom-hook
      matcher: terminal
    - command: bash /old/.agents/sdlc/hooks/agent-command-guard.sh
      matcher: old
      timeout: 1
    - command: bash ~/.agents/sdlc/hooks/agent-command-guard.sh
      matcher: .*
      timeout: 5
`
	if err := yaml.Unmarshal([]byte(input), &document); err != nil {
		t.Fatal(err)
	}
	root, err := hermesDocumentMapping(&document)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := ensureHermesCommandGuard(root)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("duplicate managed entries were not normalized")
	}
	encoded, err := yaml.Marshal(&document)
	if err != nil {
		t.Fatal(err)
	}
	contents := string(encoded)
	if strings.Count(contents, "agent-command-guard.sh") != 1 {
		t.Fatalf("managed guard count is not one:\n%s", contents)
	}
	if !strings.Contains(contents, "custom-hook") || !strings.Contains(contents, "personal: true") {
		t.Fatalf("unrelated Hermes configuration was not preserved:\n%s", contents)
	}
}

func TestW004ManagedGuardIdentityRequiresExactCommandVector(t *testing.T) {
	for _, command := range []string{
		"bash ~/.agents/sdlc/hooks/agent-command-guard.sh --extra",
		"bash unrelated ~/.agents/sdlc/hooks/agent-command-guard.sh",
		"printf ~/.agents/sdlc/hooks/agent-command-guard.sh",
	} {
		if isManagedGuardCommand(command) {
			t.Fatalf("unrelated command was treated as SDLC-owned: %q", command)
		}
	}
	for _, command := range []string{
		"bash ~/.agents/sdlc/hooks/agent-command-guard.sh",
		`bash "/Users/operator/.agents/sdlc/hooks/agent-command-guard.sh"`,
		"/bin/bash /Users/operator/.hermes/sdlc/hooks/agent-command-guard.sh",
	} {
		if !isManagedGuardCommand(command) {
			t.Fatalf("managed command was not recognized: %q", command)
		}
	}
}

func TestW004CopilotHookOwnershipIsExact(t *testing.T) {
	managed := `{"version":1,"hooks":{"preToolUse":[{"type":"command","command":"bash ~/.agents/sdlc/hooks/agent-command-guard.sh","timeoutSec":5}]}}`
	if !copilotManagedHook([]byte(managed)) {
		t.Fatal("canonical Copilot hook was not recognized")
	}
	for _, conflicting := range []string{
		`{}`,
		`{"version":1,"personal":true,"hooks":{"preToolUse":[]}}`,
		`{"version":1,"hooks":{"preToolUse":[{"type":"command","command":"bash ~/.agents/sdlc/hooks/agent-command-guard.sh --extra","timeoutSec":5}]}}`,
	} {
		if copilotManagedHook([]byte(conflicting)) {
			t.Fatalf("unknown Copilot conflict was treated as managed: %s", conflicting)
		}
	}
}
