// ABOUTME: Projects supplied evidence into exact Claude Read permission rules.
// ABOUTME: Keeps permissions invocation-local and never grants directory or write access.
package harness

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

func claudeEvidenceSettings(files []EvidenceFile) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	var rules []string
	seen := make(map[string]bool)
	for _, file := range files {
		canonical, err := filepath.EvalSymlinks(file.Path)
		if err != nil {
			return nil, fmt.Errorf("resolving Claude evidence permission %q: %w", file.Path, err)
		}
		// Claude checks both the supplied alias and its resolved target.
		for _, path := range []string{file.Path, canonical} {
			if !filepath.IsAbs(path) || !strings.HasPrefix(path, "/") || filepath.Base(path) == ".env" || strings.ContainsFunc(path, unicode.IsControl) {
				return nil, fmt.Errorf("Claude evidence permission requires a non-secret absolute POSIX path without control characters: %q", path)
			}
			if seen[path] {
				continue
			}
			seen[path] = true
			// Read uses gitignore patterns. Escape metacharacters so file names
			// cannot become wildcard grants. The extra slash means absolute path.
			literal := strings.NewReplacer("\\", "\\\\", "*", "\\*", "?", "\\?", "[", "\\[", "]", "\\]", "{", "\\{", "}", "\\}").Replace(path)
			if strings.HasSuffix(literal, " ") {
				literal = strings.TrimRight(literal, " ") + strings.Repeat("\\ ", len(literal)-len(strings.TrimRight(literal, " ")))
			}
			rules = append(rules, "Read(/"+literal+")")
		}
	}
	settings, err := json.Marshal(map[string]map[string][]string{"permissions": {"allow": rules}})
	if err != nil {
		return nil, fmt.Errorf("encoding Claude evidence permissions: %w", err)
	}
	return []string{"--settings", string(settings)}, nil
}
