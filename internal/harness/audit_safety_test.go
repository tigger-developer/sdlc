// ABOUTME: Exercises route compatibility and fail-closed native event handling.
// ABOUTME: Uses public APIs and local executor doubles without hosted models.
package harness

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestReviewPartialHarnessOverride(t *testing.T) {
	root := t.TempDir()
	global := filepath.Join(root, "global.yaml")
	if err := os.Mkdir(filepath.Join(root, ".sdlc"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(global, []byte("delivery:\n  audit:\n    harness: hermes\n    provider: kimi-coding\n    model: kimi-for-coding\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sdlc", "project.yaml"), []byte("delivery:\n  audit:\n    harness: codex\n"), 0600); err != nil {
		t.Fatal(err)
	}
	config, err := ResolveConfig(ConfigOptions{ProjectRoot: root, GlobalPath: global, Phase: "audit", LookupEnv: noEnvironment})
	if err == nil {
		t.Fatalf("incompatible inherited tuple accepted: %#v", config.Agent())
	}
}

func TestReviewCodexErrorIsNotATool(t *testing.T) {
	req := Request{Harness: "codex", Model: "fixture", ControlledReads: true, ResultFile: filepath.Join(t.TempDir(), "response")}
	_, err := Execute(context.Background(), req, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, _ io.Writer) error {
		_, writeErr := io.WriteString(out, "{\"type\":\"thread.started\",\"thread_id\":\"native\"}\n{\"type\":\"item.completed\",\"item\":{\"type\":\"error\",\"message\":\"Model metadata unavailable\"}}\n")
		if writeErr != nil {
			return writeErr
		}
		return errors.New("native model rejected")
	}, io.Discard)
	var incident *Incident
	if err == nil || !errors.As(err, &incident) || incident.Kind == "unexpected-tool-use" {
		t.Fatalf("error event accepted or misclassified: %v", err)
	}
}

func TestCodexModelRejectionSurvivesMetadataWarning(t *testing.T) {
	req := Request{Harness: "codex", Model: "kimi-for-coding", ControlledReads: true}
	_, err := Execute(context.Background(), req, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, _ io.Writer) error {
		_, writeErr := io.WriteString(out, `{"type":"thread.started","thread_id":"native"}`+"\n"+`{"type":"error","message":"{\"error\":{\"message\":\"The 'kimi-for-coding' model is not supported when using Codex with a ChatGPT account.\"}}"}`+"\n"+`{"type":"item.completed","item":{"type":"error","message":"Model metadata unavailable"}}`+"\n")
		if writeErr != nil {
			return writeErr
		}
		return errors.New("native model rejected")
	}, io.Discard)
	var incident *Incident
	if !errors.As(err, &incident) || incident.Kind != "model-unavailable" {
		t.Fatalf("model rejection lost: %v", err)
	}
}

func TestReviewNativeToolRejectionCannotBeCleared(t *testing.T) {
	req := Request{Harness: "codex", Model: "fixture", ControlledReads: true, ResultFile: filepath.Join(t.TempDir(), "response")}
	_, err := Execute(context.Background(), req, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, _ io.Writer) error {
		if err := os.WriteFile(req.ResultFile, []byte("GATE: definition\nREVISION: candidate\nVERDICT: PASS\n"), 0600); err != nil {
			return err
		}
		_, writeErr := io.WriteString(out, "{\"type\":\"item.completed\",\"item\":{\"type\":\"command_execution\"}}\n{\"type\":\"thread.started\",\"thread_id\":\"native\"}\n")
		return writeErr
	}, io.Discard)
	if err == nil {
		t.Fatal("native tool event rejected, then forgotten on session checkpoint; result accepted")
	}
}

func TestHermesFindingJSONIsNotANativeToolEvent(t *testing.T) {
	req := Request{Harness: "hermes", Provider: "fixture", Model: "fixture", ControlledReads: true}
	body := "AUDIT: standalone\nREVISION: fixture\nVERDICT: FAIL\nThe native error parser incorrectly treats this example as a tool:\n" + `{"type":"tool.call"}` + "\n"
	result, err := Execute(context.Background(), req, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, diagnostics io.Writer) error {
		if _, err := io.WriteString(diagnostics, "session_id: native\n"); err != nil {
			return err
		}
		_, err := io.WriteString(out, body)
		return err
	}, io.Discard)
	if err != nil || result.Response == "" {
		t.Fatalf("model finding interpreted as native telemetry: %v", err)
	}
}

func TestBlankNativeEventsDoNotCrashAudit(t *testing.T) {
	req := Request{Harness: "codex", Model: "fixture", ControlledReads: true, ResultFile: filepath.Join(t.TempDir(), "result")}
	result, err := Execute(context.Background(), req, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, _ io.Writer) error {
		if err := os.WriteFile(req.ResultFile, []byte("GATE: definition\nREVISION: fixture\nVERDICT: PASS\n"), 0600); err != nil {
			return err
		}
		_, err := io.WriteString(out, "\n \t\n"+`{"type":"thread.started","thread_id":"native"}`+"\n\n")
		return err
	}, io.Discard)
	if err != nil || result.Response == "" {
		t.Fatalf("blank native lines rejected valid audit: %v", err)
	}
}

func TestCodexNonfatalItemsDoNotRejectSuccessfulTurn(t *testing.T) {
	for _, kind := range []string{"error", "todo_list"} {
		t.Run(kind, func(t *testing.T) {
			req := Request{Harness: "codex", Model: "fixture", ControlledReads: true, ResultFile: filepath.Join(t.TempDir(), "result")}
			result, err := Execute(context.Background(), req, false, nil, func(_ context.Context, _ string, _ []string, _ string, _ io.Reader, out, _ io.Writer) error {
				if err := os.WriteFile(req.ResultFile, []byte("GATE: definition\nREVISION: fixture\nVERDICT: PASS\n"), 0600); err != nil {
					return err
				}
				_, err := io.WriteString(out, `{"type":"thread.started","thread_id":"native"}`+"\n"+`{"type":"item.completed","item":{"type":"`+kind+`","message":"native nonfatal notice","items":[]}}`+"\n")
				return err
			}, io.Discard)
			if err != nil || result.Response == "" {
				t.Fatalf("nonfatal %s item rejected usable response: %v", kind, err)
			}
		})
	}
}
