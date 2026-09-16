// Package readiness validates the SDLC Org test-record subset without invoking a model.
package readiness

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type Issue struct {
	Code    string `yaml:"code"`
	ID      string `yaml:"id,omitempty"`
	File    string `yaml:"file,omitempty"`
	Line    int    `yaml:"line,omitempty"`
	Message string `yaml:"message"`
}

type Record struct {
	ID             string            `yaml:"id"`
	Title          string            `yaml:"title"`
	State          string            `yaml:"state,omitempty"`
	File           string            `yaml:"file"`
	Line           int               `yaml:"line"`
	ValidationLine int               `yaml:"validation-line,omitempty"`
	Properties     map[string]string `yaml:"-"`
}

type Counts struct {
	ACs   int `yaml:"acceptance-criteria"`
	Tests int `yaml:"tests"`
	RT    int `yaml:"rt"`
	OT    int `yaml:"ot"`
	UT    int `yaml:"ut"`
}

type Inventory struct {
	Counts Counts   `yaml:"counts"`
	ACs    []Record `yaml:"acceptance-criteria"`
	Tests  []Record `yaml:"tests"`
}

type SchemaResult struct {
	Valid  bool    `yaml:"valid"`
	Errors []Issue `yaml:"errors"`
}

type CheckResult struct {
	Evaluated bool    `yaml:"evaluated"`
	Ready     bool    `yaml:"ready"`
	Errors    []Issue `yaml:"errors"`
}

type AuditReadiness struct {
	Schema    SchemaResult           `yaml:"schema"`
	Checks    map[string]CheckResult `yaml:",inline"`
	Inventory Inventory              `yaml:"inventory"`
}

type Report struct {
	Audit AuditReadiness `yaml:"audit-readiness"`
}

func ValidMode(mode string) bool { return mode == "test-code" || mode == "delivery-code" }

func (r Report) ExitCode() int {
	if !r.Audit.Schema.Valid {
		return 2
	}
	for _, check := range r.Audit.Checks {
		if !check.Ready {
			return 1
		}
	}
	return 0
}

// Check reads files only. It neither writes execution states nor runs tests.
func Check(specPath, validationPath, mode string) (Report, error) {
	r := Report{Audit: AuditReadiness{
		Schema:    SchemaResult{Errors: []Issue{}},
		Checks:    map[string]CheckResult{mode: {Errors: []Issue{}}},
		Inventory: Inventory{ACs: []Record{}, Tests: []Record{}},
	}}
	if !ValidMode(mode) {
		return r, fmt.Errorf("unsupported readiness check %q; use test-code or delivery-code", mode)
	}
	spec, issues, err := parse(specPath, false)
	if err != nil {
		return r, err
	}
	r.Audit.Schema.Errors = append(r.Audit.Schema.Errors, issues...)
	validation, issues, err := parse(validationPath, true)
	if err != nil {
		return r, err
	}
	r.Audit.Schema.Errors = append(r.Audit.Schema.Errors, issues...)
	expected, actual := map[string]bool{}, map[string]Record{}
	for _, record := range spec {
		if strings.HasPrefix(record.ID, "AC") {
			r.Audit.Inventory.ACs = append(r.Audit.Inventory.ACs, record)
			continue
		}
		expected[record.ID] = true
		r.Audit.Inventory.Tests = append(r.Audit.Inventory.Tests, record)
	}
	for _, record := range validation {
		actual[record.ID] = record
		if !expected[record.ID] {
			r.Audit.Schema.Errors = append(r.Audit.Schema.Errors, issue("UNKNOWN_TEST", record, "Validation ID is not a test in spec.org."))
		}
	}
	inv := &r.Audit.Inventory
	inv.Counts.ACs, inv.Counts.Tests = len(inv.ACs), len(inv.Tests)
	if inv.Counts.ACs == 0 || inv.Counts.Tests == 0 {
		r.Audit.Schema.Errors = append(r.Audit.Schema.Errors, Issue{Code: "EMPTY_INVENTORY", File: specPath, Message: "Specification requires at least one AC and one test."})
	}
	for i := range inv.Tests {
		test := &inv.Tests[i]
		switch test.ID[:2] {
		case "RT":
			inv.Counts.RT++
		case "OT":
			inv.Counts.OT++
		case "UT":
			inv.Counts.UT++
		}
		if result, ok := actual[test.ID]; ok {
			test.State, test.ValidationLine = result.State, result.Line
		}
	}
	r.Audit.Schema.Valid = len(r.Audit.Schema.Errors) == 0
	if !r.Audit.Schema.Valid {
		return r, nil
	}
	check := CheckResult{Evaluated: true, Errors: []Issue{}}
	for _, test := range inv.Tests {
		result, found := actual[test.ID]
		if !found {
			check.Errors = append(check.Errors, issue("MISSING_TEST", test, "No matching :testdef: heading in validation.org."))
			continue
		}
		if result.State == "" {
			check.Errors = append(check.Errors, issue("MISSING_STATE", result, "Execution state must be RED, AMBER or GREEN before audit."))
			continue
		}
		fields := []string{"REVISION", "EVIDENCE"}
		if result.State == "AMBER" {
			fields = []string{"REASON"}
		} else if strings.HasPrefix(test.ID, "RT") {
			fields = append(fields, "COMMAND")
		}
		for _, field := range fields {
			if strings.TrimSpace(result.Properties[field]) == "" {
				check.Errors = append(check.Errors, issue("MISSING_EVIDENCE", result, "State requires a non-empty "+field+" property."))
			}
		}
		if mode == "delivery-code" && !strings.HasPrefix(test.ID, "UT") && result.State != "GREEN" {
			check.Errors = append(check.Errors, issue("GREEN_REQUIRED", result, "RT and OT results must be GREEN before delivery-code audit."))
		}
	}
	check.Ready = len(check.Errors) == 0
	r.Audit.Checks[mode] = check
	return r, nil
}

func issue(code string, record Record, message string) Issue {
	return Issue{Code: code, ID: record.ID, File: record.File, Line: record.Line, Message: message}
}

var (
	headline   = regexp.MustCompile(`^\*+ (.+)$`)
	tagSuffix  = regexp.MustCompile(`\s+(:[A-Za-z0-9_@#%:~-]+:)\s*$`)
	identifier = regexp.MustCompile(`^((?:AC|RT|UT|OT)-?[0-9]+(?:[.-][0-9]+)*)(?:\s+|$)(.*)$`)
	candidate  = regexp.MustCompile(`(?i)^(?:\S+\s+)?(?:AC|RT|UT|OT)-?[0-9]`)
	property   = regexp.MustCompile(`^:([A-Za-z0-9_]+):\s*(.*)$`)
)

// parse recognizes real Org headings outside blocks, comments and drawers.
// This is deliberately a documented subset, not a general Org exporter.
func parse(path string, validation bool) ([]Record, []Issue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", path, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(io.LimitReader(file, 16*1024*1024+1))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	records, issues := []Record{}, []Issue{}
	seen := map[string]bool{}
	current, line, size, todoCount := -1, 0, 0, 0
	block, drawer := "", ""
	commentLevel := 0
	seenHeading := false
	add := func(code, message string) {
		issues = append(issues, Issue{Code: code, File: path, Line: line, Message: message})
	}
	for scanner.Scan() {
		line++
		raw := scanner.Text()
		size += len(raw) + 1
		text := strings.TrimSpace(raw)
		lower := strings.ToLower(text)
		if block != "" {
			if lower == "#+end_"+block {
				block = ""
			}
			continue
		}
		if strings.HasPrefix(lower, "#+begin_") {
			parts := strings.Fields(strings.TrimPrefix(lower, "#+begin_"))
			if len(parts) == 0 {
				add("INVALID_BLOCK", "Block start requires a type.")
			} else {
				block = parts[0]
			}
			continue
		}
		if strings.HasPrefix(lower, "#+end_") {
			add("INVALID_BLOCK", "Block end without a matching start.")
			continue
		}
		if strings.HasPrefix(raw, "# ") || raw == "#" {
			continue
		}
		h := headline.FindStringSubmatch(raw)
		if len(h) > 0 {
			seenHeading = true
			level := strings.IndexByte(raw, ' ')
			if commentLevel > 0 && level > commentLevel {
				continue
			}
			commentLevel = 0
			if strings.HasPrefix(h[1], "COMMENT ") {
				commentLevel = level
				current = -1
				continue
			}
			if drawer != "" {
				add("INVALID_DRAWER", "Drawer not closed before heading.")
				drawer = ""
			}
			current = -1
			body, tags := h[1], ""
			if match := tagSuffix.FindStringSubmatchIndex(body); match != nil {
				tags = body[match[2]:match[3]]
				body = strings.TrimSpace(body[:match[0]])
			}
			testTag, acTag := strings.Contains(tags, ":testdef:"), strings.Contains(tags, ":ac:")
			if !testTag && !acTag && !candidate.MatchString(body) {
				continue
			}
			state := ""
			id := identifier.FindStringSubmatch(body)
			if len(id) == 0 {
				first, rest, _ := strings.Cut(body, " ")
				state, id = first, identifier.FindStringSubmatch(strings.TrimSpace(rest))
			}
			if len(id) == 0 {
				add("INVALID_ID", "Tagged heading requires an AC/RT/OT/UT numeric ID and title.")
				continue
			}
			record := Record{ID: id[1], Title: strings.TrimSpace(strings.TrimPrefix(id[2], "-")), State: state, File: path, Line: line, Properties: map[string]string{}}
			ac := strings.HasPrefix(record.ID, "AC")
			if seen[record.ID] {
				issues = append(issues, issue("DUPLICATE_ID", record, "ID has more than one heading."))
			}
			seen[record.ID] = true
			if record.Title == "" {
				issues = append(issues, issue("MISSING_TITLE", record, "Heading requires a descriptive title."))
			}
			if ac && (!acTag || testTag || validation) {
				issues = append(issues, issue("INVALID_AC_TAG", record, "AC headings belong in spec.org and require :ac: only."))
			}
			if !ac && (!testTag || acTag) {
				issues = append(issues, issue("MISSING_TEST_TAG", record, "Test heading requires :testdef: and must not carry :ac:."))
			}
			if state != "" && (!validation || (state != "RED" && state != "GREEN" && state != "AMBER")) {
				issues = append(issues, issue("INVALID_STATE", record, "Execution state belongs before the ID in validation.org; use RED, AMBER or GREEN."))
			}
			records = append(records, record)
			current = len(records) - 1
			continue
		}
		if commentLevel > 0 {
			continue
		}
		if text == ":END:" {
			if drawer == "" {
				add("INVALID_DRAWER", "Drawer end without a start.")
			}
			drawer = ""
			continue
		}
		if drawer != "" {
			if drawer == "PROPERTIES" && current >= 0 {
				if p := property.FindStringSubmatch(text); p != nil {
					if _, exists := records[current].Properties[p[1]]; exists {
						add("DUPLICATE_PROPERTY", "Duplicate property "+p[1]+".")
					}
					records[current].Properties[p[1]] = p[2]
				}
			}
			continue
		}
		if p := property.FindStringSubmatch(text); p != nil && p[2] == "" {
			drawer = p[1]
			continue
		}
		if strings.HasPrefix(lower, "#+todo:") || strings.HasPrefix(lower, "#+seq_todo:") || strings.HasPrefix(lower, "#+typ_todo:") {
			if validation {
				todoCount++
				if seenHeading || strings.HasPrefix(lower, "#+typ_todo:") {
					add("INVALID_STATES", "Declare workflow states in frontmatter before headings.")
				}
				_, value, _ := strings.Cut(text, ":")
				if strings.Join(strings.Fields(value), " ") != "AMBER RED | GREEN" {
					add("INVALID_STATES", "Declare #+TODO: AMBER RED | GREEN.")
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if size > 16*1024*1024 {
		return nil, nil, fmt.Errorf("%s exceeds 16 MiB", path)
	}
	if block != "" {
		add("INVALID_BLOCK", "Unclosed Org block.")
	}
	if drawer != "" {
		add("INVALID_DRAWER", "Unclosed Org drawer.")
	}
	if validation && todoCount != 1 {
		add("INVALID_STATES", "Require exactly one #+TODO: AMBER RED | GREEN declaration.")
	}
	return records, issues, nil
}
