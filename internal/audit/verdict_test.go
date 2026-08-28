package audit

import "testing"

func TestParseVerdictAcceptsPass(t *testing.T) {
	report := "AUDIT: audit-design\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: PASS\n"
	verdict, err := ParseVerdict(report)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Audit != "audit-design" || verdict.Result != "PASS" {
		t.Fatalf("verdict = %#v", verdict)
	}
}

func TestParseVerdictAcceptsFailWithFinding(t *testing.T) {
	report := "AUDIT: audit-spec\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: FAIL\n\n1. Requirement R1, named export behavior, has no failure case.\n"
	verdict, err := ParseVerdict(report)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Result != "FAIL" || len(verdict.Findings) != 1 {
		t.Fatalf("verdict = %#v", verdict)
	}
}

func TestParseVerdictFailsClosed(t *testing.T) {
	cases := []string{
		"VERDICT: PASS\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: MAYBE\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PASS\n1. finding\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: FAIL\n",
		"AUDIT: audit-code\nAUDIT: audit-tests\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PASS\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PASS\nunstructured explanation\n",
	}
	for _, report := range cases {
		if _, err := ParseVerdict(report); err == nil {
			t.Errorf("ParseVerdict(%q) succeeded", report)
		}
	}
}
