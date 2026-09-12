// ABOUTME: Resolves schema-owned native goal limits without starting an agent.
// ABOUTME: Validates original budget spelling before YAML can coerce monetary values.
package projectinit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var groupedIntegerPattern = regexp.MustCompile(`^([1-9][0-9]*|[1-9][0-9]{0,2}(,[0-9]{3})+)$`)

// parseGroupedInteger allows only positive decimal integers and thousands groups.
// The upper bound preserves exact transport through JSON-based native tools.
func parseGroupedInteger(value string) (int64, error) {
	if !groupedIntegerPattern.MatchString(value) {
		return 0, errors.New("use a positive integer such as 100000 or 100,000; periods, decimals, signs and malformed groups are forbidden")
	}
	integer, err := strconv.ParseInt(strings.ReplaceAll(value, ",", ""), 10, 64)
	if err != nil || integer > 9007199254740991 {
		return 0, errors.New("integer exceeds the exact JSON limit of 9,007,199,254,740,991")
	}
	return integer, nil
}

// validateGoalYAML runs before decoding scalars into Go numbers. In particular,
// unquoted 100.000 must not become the otherwise valid integer spelling 100.
func validateGoalYAML(contents []byte) error {
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	var root map[string]yaml.Node
	if err := decoder.Decode(&root); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("configuration must contain exactly one YAML document")
	}
	delivery, present := root["delivery"]
	if !present {
		return nil
	}
	var fields map[string]yaml.Node
	if err := delivery.Decode(&fields); err != nil {
		return err
	}
	goal, present := fields["goal"]
	if !present {
		return nil
	}
	if goal.Kind != yaml.MappingNode {
		return errors.New("delivery.goal must be a mapping")
	}
	var limits map[string]yaml.Node
	if err := goal.Decode(&limits); err != nil {
		return err
	}
	for name, node := range limits {
		if name != "max_turns" && name != "max_token_budget" {
			return fmt.Errorf("unknown delivery.goal field %q", name)
		}
		if node.Kind != yaml.ScalarNode || (node.Tag != "!!str" && node.Tag != "!!int") {
			return fmt.Errorf("delivery.goal.%s must be a positive integer or comma-grouped integer string", name)
		}
		if _, err := parseGroupedInteger(node.Value); err != nil {
			return fmt.Errorf("delivery.goal.%s: %w", name, err)
		}
	}
	return nil
}

// ResolveGoalLimits returns normalized integer budgets from the shared schema.
// No provider, session storage or project configuration is mutated.
func ResolveGoalLimits(schema ConfigSchema, projectRoot, globalPath string, overrides map[string]string) (map[string]int64, error) {
	global, err := readYAMLConfig(globalPath)
	if err != nil {
		return nil, err
	}
	project, err := readYAMLConfig(filepath.Join(projectRoot, ".sdlc", "project.yaml"))
	if err != nil {
		return nil, err
	}
	limits := make(map[string]int64)
	for _, field := range schema.Fields {
		if !strings.HasPrefix(field.Path, "delivery.goal.") {
			continue
		}
		value := field.Default
		if candidate, ok := yamlPathString(global, field.Path); ok && field.AllowGlobal {
			value = candidate
		}
		if candidate, ok := yamlPathString(project, field.Path); ok {
			value = candidate
		}
		if candidate := os.Getenv(field.Key); candidate != "" {
			value = candidate
		}
		if candidate, ok := overrides[field.Key]; ok {
			value = candidate
		}
		integer, err := parseGroupedInteger(value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", field.Path, err)
		}
		limits[strings.TrimPrefix(field.Path, "delivery.goal.")] = integer
	}
	if len(limits) != 2 || limits["max_turns"] == 0 || limits["max_token_budget"] == 0 {
		return nil, errors.New("goal budget fields are missing from the SDLC schema; update the deployment")
	}
	return limits, nil
}
