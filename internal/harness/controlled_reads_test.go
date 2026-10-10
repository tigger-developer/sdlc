// ABOUTME: Checks auditor tool isolation and rejects unexpected native tool use.
// ABOUTME: Uses invocation builders and a local executor without hosted models.
package harness

import (
	"context"
	"io"
	"slices"
	"strings"
	"testing"
)

func TestControlledAuditDisablesNativeToolsAndUsesStdin(t *testing.T) {
	for _, provider := range []string{"codex", "claude", "copilot", "hermes"} {
		t.Run(provider, func(t *testing.T) {
			request := matrixRequest(provider, t.TempDir())
			request.ControlledReads = true
			request.Standards = []EvidenceFile{{Path: "standard.md", SHA256: "captured"}}
			for _, build := range []func(Request) (Invocation, error){BuildStart, BuildResume} {
				invocation, err := build(request)
				if err != nil {
					t.Fatal(err)
				}
				if invocation.Stdin != request.Prompt {
					t.Fatal("audit prompt was not transported over stdin")
				}
				switch provider {
				case "claude":
					i := slices.Index(invocation.Args, "--tools")
					if i < 0 || invocation.Args[i+1] != "" || !slices.Contains(invocation.Args, "--strict-mcp-config") {
						t.Fatalf("Claude tools not isolated: %v", invocation.Args)
					}
				case "codex":
					if !slices.Contains(invocation.Args, "--ignore-user-config") || !slices.Contains(invocation.Args, "shell_tool") || !slices.Contains(invocation.Args, "unified_exec") {
						t.Fatalf("Codex tools not isolated: %v", invocation.Args)
					}
				case "copilot":
					if !slices.Contains(invocation.Args, "--available-tools=") {
						t.Fatal("Copilot tools enabled")
					}
				case "hermes":
					i := slices.Index(invocation.Args, "-t")
					if i < 0 || invocation.Args[i+1] != "none" {
						t.Fatal("Hermes tools enabled")
					}
				}
			}
		})
	}
}

func TestControlledAuditRejectsNativeToolEvents(t *testing.T) {
	request := matrixRequest("claude", t.TempDir())
	request.ControlledReads = true
	executor := func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, _ io.Writer) error {
		_, err := io.WriteString(stdout, `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"TESTING.md"}}]}}`+"\n"+`{"type":"result","result":"VERDICT: PASS"}`+"\n")
		return err
	}
	_, err := Execute(context.Background(), request, false, nil, executor, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "unexpected-tool-use") {
		t.Fatalf("native read accepted: %v", err)
	}
}

func TestControlledAuditRejectsLostContextAndProviderErrors(t *testing.T) {
	cases := []struct {
		name, provider, event, kind string
	}{
		{"compaction", "claude", `{"type":"system","subtype":"compact_boundary","message":"context compacted"}`, "context-lost"},
		{"command", "codex", `{"type":"item.completed","item":{"type":"command_execution"}}`, "unexpected-tool-use"},
		{"model-error", "claude", `{"type":"result","is_error":true,"result":"model unavailable"}`, "provider-rejected"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := matrixRequest(test.provider, t.TempDir())
			request.ControlledReads = true
			delivered := false
			request.OnDelivered = func(string) error { delivered = true; return nil }
			executor := func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, stdout, _ io.Writer) error {
				_, err := io.WriteString(stdout, test.event+"\n"+`{"type":"result","result":"VERDICT: PASS"}`+"\n")
				return err
			}
			_, err := Execute(context.Background(), request, false, nil, executor, io.Discard)
			if err == nil || !strings.Contains(err.Error(), test.kind) || delivered {
				t.Fatalf("invalid invocation acknowledged: error=%v delivered=%v", err, delivered)
			}
		})
	}
}
