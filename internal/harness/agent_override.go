// ABOUTME: Prevents configuration layers from mixing incompatible native routes.
// ABOUTME: Keeps execution limits independent of harness, provider and model identity.
package harness

import (
	"fmt"
	"strings"
)

func applyAgentOverride(config *Config, harness, provider, model, origin string) error {
	harness = strings.ToLower(strings.TrimSpace(harness))
	provider, model = strings.TrimSpace(provider), strings.TrimSpace(model)
	changedHarness := config.Harness != "" && harness != "" && harness != config.Harness
	selected := config.Harness
	if harness != "" {
		selected = harness
	}
	changedProvider := selected == "hermes" && config.Provider != "" && provider != "" && provider != config.Provider
	if (changedHarness || changedProvider) && model == "" {
		return fmt.Errorf("agent route in %s changes harness or provider without a model; specify a complete replacement route or remove the override to inherit it", origin)
	}
	if changedHarness {
		if selected == "hermes" && provider == "" {
			return fmt.Errorf("Hermes route in %s changes harness without a provider; specify harness, provider and model together", origin)
		}
		config.Provider = ""
	}
	if harness != "" {
		config.Harness = harness
	}
	if provider != "" {
		config.Provider = provider
	}
	if model != "" {
		config.Model = model
	}
	return nil
}
