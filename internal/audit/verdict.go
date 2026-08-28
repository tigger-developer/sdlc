package audit

import (
	"bufio"
	"fmt"
	"strings"
)

// Verdict is the machine-checkable header emitted by every audit skill.
type Verdict struct {
	Audit    string
	Provider string
	Model    string
	Result   string
	Findings []string
}

// ParseVerdict fails closed on missing headers, unknown results, or a PASS
// accompanied by findings.
func ParseVerdict(report string) (Verdict, error) {
	verdict := Verdict{}
	seen := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(report))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "AUDIT:"):
			if seen["AUDIT"] {
				return Verdict{}, fmt.Errorf("audit verdict repeats AUDIT header")
			}
			seen["AUDIT"] = true
			verdict.Audit = strings.TrimSpace(strings.TrimPrefix(line, "AUDIT:"))
		case strings.HasPrefix(line, "AUDITOR_PROVIDER:"):
			if seen["AUDITOR_PROVIDER"] {
				return Verdict{}, fmt.Errorf("audit verdict repeats AUDITOR_PROVIDER header")
			}
			seen["AUDITOR_PROVIDER"] = true
			verdict.Provider = strings.TrimSpace(strings.TrimPrefix(line, "AUDITOR_PROVIDER:"))
		case strings.HasPrefix(line, "AUDITOR_MODEL:"):
			if seen["AUDITOR_MODEL"] {
				return Verdict{}, fmt.Errorf("audit verdict repeats AUDITOR_MODEL header")
			}
			seen["AUDITOR_MODEL"] = true
			verdict.Model = strings.TrimSpace(strings.TrimPrefix(line, "AUDITOR_MODEL:"))
		case strings.HasPrefix(line, "VERDICT:"):
			if seen["VERDICT"] {
				return Verdict{}, fmt.Errorf("audit verdict repeats VERDICT header")
			}
			seen["VERDICT"] = true
			verdict.Result = strings.TrimSpace(strings.TrimPrefix(line, "VERDICT:"))
		case isNumberedFinding(line):
			verdict.Findings = append(verdict.Findings, line)
		default:
			return Verdict{}, fmt.Errorf("audit verdict contains unstructured content %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return Verdict{}, fmt.Errorf("reading audit verdict: %w", err)
	}
	if verdict.Audit == "" || verdict.Provider == "" || verdict.Model == "" || verdict.Result == "" {
		return Verdict{}, fmt.Errorf("audit verdict is missing a required header")
	}
	if verdict.Result != "PASS" && verdict.Result != "FAIL" {
		return Verdict{}, fmt.Errorf("unsupported verdict %q", verdict.Result)
	}
	if verdict.Result == "PASS" && len(verdict.Findings) != 0 {
		return Verdict{}, fmt.Errorf("PASS verdict contains findings")
	}
	if verdict.Result == "FAIL" && len(verdict.Findings) == 0 {
		return Verdict{}, fmt.Errorf("FAIL verdict contains no numbered findings")
	}
	return verdict, nil
}

func isNumberedFinding(line string) bool {
	dot := strings.IndexByte(line, '.')
	if dot < 1 || dot == len(line)-1 {
		return false
	}
	for _, character := range line[:dot] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return strings.TrimSpace(line[dot+1:]) != ""
}
