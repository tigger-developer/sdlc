// ABOUTME: Verifies configurable internal failure bounds and project precedence.
// ABOUTME: Keeps malformed configuration from silently changing recovery limits.
package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFailureLimitConfiguration(t *testing.T) {
	for _, value := range []string{"", "2", "0", "-1", "invalid"} {
		t.Run(value, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "global.yaml")
			body := "delivery:\n  audit:\n    harness: codex\n    model: fixture\n"
			if value != "" {
				body += "    max_failures: " + value + "\n"
			}
			if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			config, err := ResolveConfig(ConfigOptions{ProjectRoot: root, GlobalPath: path, Phase: "audit", LookupEnv: noEnvironment})
			invalid := value == "0" || value == "-1" || value == "invalid"
			if (err != nil) != invalid {
				t.Fatalf("config=%#v err=%v", config, err)
			}
			if invalid {
				return
			}
			want := 3
			if value == "2" {
				want = 2
			}
			if config.MaxFailures != want {
				t.Fatalf("limit=%d want=%d", config.MaxFailures, want)
			}
			if err := os.Mkdir(filepath.Join(root, ".sdlc"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".sdlc", "project.yaml"), []byte("delivery:\n  audit:\n    max_failures: 4\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			config, err = ResolveConfig(ConfigOptions{ProjectRoot: root, GlobalPath: path, Phase: "audit", LookupEnv: noEnvironment})
			if err != nil || config.MaxFailures != 4 {
				t.Fatalf("override=%#v %v", config, err)
			}
		})
	}
}
