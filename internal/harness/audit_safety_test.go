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
	if errors.As(err, &incident) && incident.Kind == "unexpected-tool-use" {
		t.Fatalf("error event misclassified: %v", err)
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
