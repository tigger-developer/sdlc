package harness

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the resolved harness configuration for one SDLC phase.
type Config struct {
	Harness  string
	Provider string
	Model    string
	Timeout  time.Duration
}

// ConfigOptions supplies higher-precedence overrides and configuration paths.
type ConfigOptions struct {
	ProjectRoot string
	GlobalPath  string
	Phase       string
	Harness     string
	Provider    string
	Model       string
	Timeout     time.Duration
	LookupEnv   func(string) (string, bool)
}

type phaseDocument struct {
	Harness  string `yaml:"harness"`
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	Timeout  string `yaml:"timeout"`
}

type configDocument struct {
	Delivery struct {
		Definition phaseDocument `yaml:"definition"`
		Build      phaseDocument `yaml:"build"`
		Audit      phaseDocument `yaml:"audit"`
	} `yaml:"delivery"`
}

// ResolveConfig applies CLI, environment, project, global, then default precedence.
func ResolveConfig(options ConfigOptions) (Config, error) {
	phase := strings.ToLower(strings.TrimSpace(options.Phase))
	if phase != "definition" && phase != "build" && phase != "audit" {
		return Config{}, fmt.Errorf("unsupported SDLC phase %q", options.Phase)
	}
	config := Config{}
	explicitProvider := false
	applyDocument := func(path string) error {
		if path == "" {
			return nil
		}
		contents, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading harness configuration %s: %w", path, err)
		}
		var document configDocument
		if err := yaml.Unmarshal(contents, &document); err != nil {
			return fmt.Errorf("parsing harness configuration %s: %w", path, err)
		}
		var value phaseDocument
		switch phase {
		case "definition":
			value = document.Delivery.Definition
		case "build":
			value = document.Delivery.Build
		case "audit":
			value = document.Delivery.Audit
		}
		if value.Harness != "" {
			config.Harness = value.Harness
		}
		if value.Provider != "" {
			config.Provider = value.Provider
		}
		if value.Model != "" {
			config.Model = value.Model
		}
		if value.Timeout != "" {
			parsed, err := time.ParseDuration(value.Timeout)
			if err != nil {
				return fmt.Errorf("invalid %s timeout %q in %s", phase, value.Timeout, path)
			}
			config.Timeout = parsed
		}
		return nil
	}
	globalPath := options.GlobalPath
	if globalPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			globalPath = filepath.Join(home, ".agents", "sdlc.yaml")
		}
	}
	if err := applyDocument(globalPath); err != nil {
		return Config{}, err
	}
	if options.ProjectRoot != "" {
		if err := applyDocument(filepath.Join(options.ProjectRoot, ".sdlc", "project.yaml")); err != nil {
			return Config{}, err
		}
	}
	lookup := options.LookupEnv
	if lookup == nil {
		lookup = os.LookupEnv
	}
	prefix := map[string]string{"definition": "SPEC", "build": "BUILD", "audit": "AUDIT"}[phase]
	if value, ok := lookup("SDLC_" + prefix + "_HARNESS"); ok && strings.TrimSpace(value) != "" {
		config.Harness = value
	}
	if value, ok := lookup("SDLC_" + prefix + "_PROVIDER"); ok && strings.TrimSpace(value) != "" {
		config.Provider = value
	}
	if value, ok := lookup("SDLC_" + prefix + "_MODEL"); ok && strings.TrimSpace(value) != "" {
		config.Model = value
	}
	if value, ok := lookup("SDLC_" + prefix + "_TIMEOUT"); ok && strings.TrimSpace(value) != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid %s timeout %q", phase, value)
		}
		config.Timeout = parsed
	}
	if options.Harness != "" {
		config.Harness = options.Harness
	}
	if options.Provider != "" {
		config.Provider = options.Provider
		explicitProvider = true
	}
	if options.Model != "" {
		config.Model = options.Model
	}
	if options.Timeout != 0 {
		config.Timeout = options.Timeout
	}
	if config.Harness == "" {
		config.Harness = "codex"
	}
	config.Harness = strings.ToLower(strings.TrimSpace(config.Harness))
	if config.Model == "" {
		return Config{}, fmt.Errorf("%s harness configuration requires a model", phase)
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Minute
	}
	if config.Timeout < time.Second {
		return Config{}, errors.New("harness timeout must be at least one second")
	}
	switch config.Harness {
	case "hermes":
		if strings.TrimSpace(config.Provider) == "" {
			return Config{}, errors.New("Hermes harness configuration requires a provider")
		}
	case "codex", "claude", "copilot":
		if strings.TrimSpace(config.Provider) != "" && explicitProvider {
			return Config{}, fmt.Errorf("%s does not accept an explicit provider", config.Harness)
		}
		config.Provider = ""
	default:
		return Config{}, fmt.Errorf("unsupported harness %q", config.Harness)
	}
	return config, nil
}
