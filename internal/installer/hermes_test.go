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
