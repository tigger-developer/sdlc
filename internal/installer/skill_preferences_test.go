package installer

import (
	"strings"
	"testing"

	"github.com/tigger-developer/sdlc/internal/deployment"
)

func TestSkillPreferencesFollowInstalledConfiguration(t *testing.T) {
	const source = "---\nname: deliver-change\ndescription: Delivery workflow\n---\n\nKeep the approved scope.\n"
	plan := installationPlan{tree: deployment.Tree{}}
	path := "skills/deliver-change/SKILL.md"
	if err := plan.tree.Add(path, "source", []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	config := []byte("delivery:\n  test_code:\n    harness: codex\n    provider: openai\n    model: expensive\n    reasoning_effort: high\n  build:\n    harness: goose\n    provider: openai\n    model: economical\n    reasoning_effort: low\n")
	if err := plan.applySkillPreferences(config); err != nil {
		t.Fatal(err)
	}
	updated := string(plan.tree[path].Data)
	for _, expected := range []string{"preferred_harness: goose", "preferred_model: economical", "preferred_reasoning_effort: low", "preferred_test_code_model: expensive", "preferred_test_code_reasoning_effort: high", "Keep the approved scope."} {
		if !strings.Contains(updated, expected) {
			t.Fatalf("missing %q in:\n%s", expected, updated)
		}
	}
	if strings.Contains(source, "preferred_model") {
		t.Fatal("source skill mutated")
	}
	if err := plan.applySkillPreferences([]byte("delivery:\n  build:\n    model: replacement\n")); err != nil {
		t.Fatal(err)
	}
	updated = string(plan.tree[path].Data)
	if !strings.Contains(updated, "preferred_model: replacement") || strings.Contains(updated, "economical") || strings.Contains(updated, "preferred_reasoning_effort") || strings.Contains(updated, "preferred_test_code_model") {
		t.Fatalf("stale preference survived refresh:\n%s", updated)
	}
}

func TestEmptyConfigurationKeepsSkillSource(t *testing.T) {
	const source = "---\nname: code-audit\ndescription: Review code\n---\n\nReview evidence.\n"
	updated, err := skillWithPreferences([]byte(source), skillModelPreference{}, nil)
	if err != nil || string(updated) != source {
		t.Fatalf("empty preference changed source: %q %v", updated, err)
	}
}
