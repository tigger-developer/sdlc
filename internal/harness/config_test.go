package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestW004ConfigResolutionUsesCLIEnvironmentProjectGlobalOrder(t *testing.T) {
	root := t.TempDir()
	global := filepath.Join(root, "global.yaml")
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(filepath.Join(project, ".sdlc"), 0o700); err != nil {
		t.Fatal(err)
	}
	writeHarnessConfig(t, global, "hermes", "global-provider", "global-model", "4m")
	writeHarnessConfig(t, filepath.Join(project, ".sdlc", "project.yaml"), "claude", "project-provider", "project-model", "3m")

	config, err := ResolveConfig(ConfigOptions{
		ProjectRoot: project, GlobalPath: global, Phase: "audit",
		LookupEnv: func(key string) (string, bool) {
			values := map[string]string{"SDLC_AUDIT_HARNESS": "copilot", "SDLC_AUDIT_MODEL": "env-model"}
			value, ok := values[key]
			return value, ok
		},
		Harness: "codex",
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.Harness != "codex" || config.Model != "env-model" || config.Provider != "" || config.Timeout != 3*time.Minute {
		t.Fatalf("config = %#v", config)
	}
}

func TestW004ProviderIsRequiredOnlyForHermes(t *testing.T) {
	for _, harnessName := range []string{"codex", "claude", "copilot"} {
		config, err := ResolveConfig(ConfigOptions{Phase: "audit", GlobalPath: filepath.Join(t.TempDir(), "absent.yaml"), Harness: harnessName, Model: "model", LookupEnv: noEnvironment})
		if err != nil || config.Provider != "" {
			t.Fatalf("%s config = %#v, %v", harnessName, config, err)
		}
	}
	if _, err := ResolveConfig(ConfigOptions{Phase: "audit", GlobalPath: filepath.Join(t.TempDir(), "absent.yaml"), Harness: "hermes", Model: "model", LookupEnv: noEnvironment}); err == nil {
		t.Fatal("Hermes without provider succeeded")
	}
	if _, err := ResolveConfig(ConfigOptions{Phase: "audit", GlobalPath: filepath.Join(t.TempDir(), "absent.yaml"), Harness: "codex", Provider: "unsupported", Model: "model", LookupEnv: noEnvironment}); err == nil {
		t.Fatal("Codex accepted an explicit provider")
	}
}

func TestW004LowerPrecedenceProviderIsIgnoredWhenHarnessChanges(t *testing.T) {
	root := t.TempDir()
	global := filepath.Join(root, "global.yaml")
	writeHarnessConfig(t, global, "hermes", "global-provider", "global-model", "4m")
	config, err := ResolveConfig(ConfigOptions{
		Phase: "audit", GlobalPath: global, Harness: "claude", Model: "model", LookupEnv: noEnvironment,
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.Provider != "" {
		t.Fatalf("inherited provider was passed to Claude: %#v", config)
	}
}

func TestW004ConfiguredProviderIsIgnoredWhenHarnessDoesNotAcceptIt(t *testing.T) {
	root := t.TempDir()
	global := filepath.Join(root, "global.yaml")
	writeHarnessConfig(t, global, "codex", "openai-codex", "global-model", "4m")
	config, err := ResolveConfig(ConfigOptions{
		Phase: "audit", GlobalPath: global, LookupEnv: noEnvironment,
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.Provider != "" {
		t.Fatalf("configured provider was passed to Codex: %#v", config)
	}
}

func TestW004EveryHarnessResolvesForEveryPhase(t *testing.T) {
	for _, phase := range []string{"definition", "build", "audit"} {
		for _, harnessName := range []string{"codex", "claude", "copilot", "hermes"} {
			t.Run(phase+"/"+harnessName, func(t *testing.T) {
				provider := ""
				if harnessName == "hermes" {
					provider = "test-provider"
				}
				config, err := ResolveConfig(ConfigOptions{
					Phase: phase, GlobalPath: filepath.Join(t.TempDir(), "absent.yaml"), Harness: harnessName,
					Provider: provider, Model: "test-model", Timeout: 17 * time.Second, LookupEnv: noEnvironment,
				})
				if err != nil {
					t.Fatal(err)
				}
				if config.Harness != harnessName || config.Provider != provider || config.Model != "test-model" || config.Timeout != 17*time.Second {
					t.Fatalf("config = %#v", config)
				}
			})
		}
	}
}

func writeHarnessConfig(t *testing.T, path, harnessName, provider, model, timeout string) {
	t.Helper()
	contents := "version: 3\ndelivery:\n  audit:\n    harness: " + harnessName + "\n    provider: " + provider + "\n    model: " + model + "\n    timeout: " + timeout + "\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func noEnvironment(string) (string, bool) { return "", false }

func TestAuditCoolOffPeriod(t *testing.T) {
	for _, tc := range []struct {
		value   string
		want    time.Duration
		invalid bool
	}{
		{"", time.Hour, false}, {"30m", 30 * time.Minute, false}, {"1h", time.Hour, false},
		{"0s", 0, true}, {"-1h", 0, true}, {"tomorrow", 0, true},
	} {
		t.Run(tc.value, func(t *testing.T) {
			root := t.TempDir()
			configPath := filepath.Join(root, "global.yaml")
			body := "delivery:\n  audit:\n    harness: claude\n    model: fixture\n    fallback:\n      harness: hermes\n      provider: nous\n      model: fallback\n"
			if tc.value != "" {
				body += "      cool_off_period: " + tc.value + "\n"
			}
			if err := os.WriteFile(configPath, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			config, err := ResolveConfig(ConfigOptions{Phase: "audit", GlobalPath: configPath, LookupEnv: noEnvironment})
			if tc.invalid {
				if err == nil {
					t.Fatal("accepted invalid period")
				}
				return
			}
			if err != nil || config.CoolOffPeriod != tc.want {
				t.Fatalf("period=%s want=%s error=%v", config.CoolOffPeriod, tc.want, err)
			}
		})
	}
}

func TestProjectCoolOffOverrideRetainsInheritedFallback(t *testing.T) {
	root := t.TempDir()
	global := filepath.Join(root, "global.yaml")
	if err := os.Mkdir(filepath.Join(root, ".sdlc"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(global, []byte("delivery:\n  audit:\n    harness: claude\n    model: primary\n    fallback:\n      harness: hermes\n      provider: nous\n      model: fallback\n      cool_off_period: 2h\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sdlc", "project.yaml"), []byte("delivery:\n  audit:\n    fallback:\n      cool_off_period: 20m\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := ResolveConfig(ConfigOptions{ProjectRoot: root, GlobalPath: global, Phase: "audit", LookupEnv: noEnvironment})
	if err != nil || config.Fallback == nil || config.Fallback.Harness != "hermes" || config.CoolOffPeriod != 20*time.Minute {
		t.Fatalf("config=%#v err=%v", config, err)
	}
}
