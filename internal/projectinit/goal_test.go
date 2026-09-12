// ABOUTME: Checks shared initializer budget validation before YAML number coercion.
// ABOUTME: Keeps monetary budget syntax identical between initialization and delivery.
package projectinit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoalBudgetFieldAndOriginalYAMLSpelling(t *testing.T) {
	schema, err := LoadConfigSchema("../../src")
	if err != nil {
		t.Fatal(err)
	}
	var field ConfigField
	for _, candidate := range schema.Fields {
		if candidate.Path == "delivery.goal.max_token_budget" {
			field = candidate
		}
	}
	if field.Key == "" {
		t.Fatal("missing goal token budget schema field")
	}
	for _, good := range []string{"100000", "100,000", "1", "9,007,199,254,740,991"} {
		if err := field.ValidateValue(good, nil); err != nil {
			t.Fatalf("rejected %q: %v", good, err)
		}
	}
	for _, bad := range []string{"100.000", "100,00", "1e5", "+100", "0", "01", "9,007,199,254,740,992"} {
		if err := field.ValidateValue(bad, nil); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
	path := filepath.Join(t.TempDir(), "global.yaml")
	if err := os.WriteFile(path, []byte("version: 3\ndelivery:\n  goal:\n    max_token_budget: 100.000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readYAMLConfig(path); err == nil {
		t.Fatal("initializer silently coerced 100.000")
	}
}
