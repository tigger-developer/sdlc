// ABOUTME: Verifies duration units, defaults and project overrides for audit deadlines.
// ABOUTME: Rejects invalid response limits before a provider can be invoked.
package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResponseTimeoutConfiguration(t *testing.T) {
	for _, value := range []string{"3m", "90s", "90", "0s", "500ms", "invalid"} {
		t.Run(value, func(t *testing.T) {
			root := t.TempDir()
			global := filepath.Join(root, "global.yaml")
			if err := os.WriteFile(global, []byte("delivery:\n  audit:\n    model: fixture\n    response_start_timeout: "+value+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			config, err := ResolveConfig(ConfigOptions{Phase: "audit", GlobalPath: global, LookupEnv: noEnvironment})
			valid := value == "3m" || value == "90s"
			if (err == nil) != valid {
				t.Fatalf("config=%#v err=%v", config, err)
			}
			if valid {
				want := 3 * time.Minute
				if value == "90s" {
					want = 90 * time.Second
				}
				if config.ResponseStartTimeout != want || config.ResponseIdleTimeout != 2*time.Minute || config.TotalTimeout != 15*time.Minute {
					t.Fatalf("limits=%#v", config)
				}
			}
		})
	}
}

func TestAuditDeadlinePrecedence(t *testing.T) {
	root := t.TempDir()
	global := filepath.Join(root, "global.yaml")
	if err := os.WriteFile(global, []byte("delivery:\n  audit:\n    model: fixture\n    response_start_timeout: 90s\n    total_timeout: 20m\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".sdlc"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sdlc", "project.yaml"), []byte("delivery:\n  audit:\n    response_start_timeout: 2m\n"), 0600); err != nil {
		t.Fatal(err)
	}
	config, err := ResolveConfig(ConfigOptions{Phase: "audit", ProjectRoot: root, GlobalPath: global, LookupEnv: noEnvironment})
	if err != nil || config.ResponseStartTimeout != 2*time.Minute || config.TotalTimeout != 20*time.Minute {
		t.Fatalf("project/global resolution=%#v err=%v", config, err)
	}
	lookup := func(key string) (string, bool) {
		values := map[string]string{"SDLC_AUDIT_RESPONSE_START_TIMEOUT": "45s", "SDLC_AUDIT_TOTAL_TIMEOUT": "12m", "SDLC_AUDIT_RESPONSE_IDLE_TIMEOUT": "30s"}
		value, found := values[key]
		return value, found
	}
	config, err = ResolveConfig(ConfigOptions{Phase: "audit", ProjectRoot: root, GlobalPath: global, LookupEnv: lookup})
	if err != nil || config.ResponseStartTimeout != 45*time.Second || config.TotalTimeout != 12*time.Minute || config.ResponseIdleTimeout != 30*time.Second {
		t.Fatalf("environment resolution=%#v err=%v", config, err)
	}
}
