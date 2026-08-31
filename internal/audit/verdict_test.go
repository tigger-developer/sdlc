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

func TestParseVerdictAcceptsPassWithAdvisory(t *testing.T) {
	report := "AUDIT: audit-design\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: PASS\n\n1. [ADVISORY] Document the rejected cache alternative.\n"
	verdict, err := ParseVerdict(report)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Result != "PASS" || len(verdict.Findings) != 1 {
		t.Fatalf("verdict = %#v", verdict)
	}
}

func TestParseVerdictAcceptsFailWithBlockingAndAdvisoryFindings(t *testing.T) {
	report := "AUDIT: audit-spec\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: FAIL\n\n1. [BLOCKING] Requirement R1, named export behavior, has no failure case.\n2. [ADVISORY] Add a short rationale for the chosen terminology.\n"
	verdict, err := ParseVerdict(report)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Result != "FAIL" || len(verdict.Findings) != 2 {
		t.Fatalf("verdict = %#v", verdict)
	}
}

func TestParseVerdictAcceptsProvisionalWithConditionAndAdvisory(t *testing.T) {
	report := "AUDIT: audit-design\nAUDITOR_PROVIDER: openai-codex\nAUDITOR_MODEL: gpt-5.6-luna\nVERDICT: PROVISIONAL\n\n1. [CONDITION] Add the selected timeout to the deployment table. | VERIFY: The table value equals config/defaults.yaml.\n2. [ADVISORY] Record the rejected cache alternative.\n"
	verdict, err := ParseVerdict(report)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Result != "PROVISIONAL" || len(verdict.Findings) != 2 {
		t.Fatalf("verdict = %#v", verdict)
	}
}

func TestParseVerdictFailsClosed(t *testing.T) {
	cases := []string{
		"VERDICT: PASS\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: MAYBE\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PASS\n1. [BLOCKING] unsafe behavior\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PASS\n1. [CONDITION] update the reference | VERIFY: reference resolves\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PROVISIONAL\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PROVISIONAL\n1. [BLOCKING] unsafe behavior\n2. [CONDITION] update the reference | VERIFY: reference resolves\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PROVISIONAL\n1. [CONDITION] update the reference\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: FAIL\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: FAIL\n1. [ADVISORY] optional improvement\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: FAIL\n1. [BLOCKING] unsafe behavior\n2. [CONDITION] update the reference | VERIFY: reference resolves\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: FAIL\n1. unclassified finding\n",
		"AUDIT: audit-code\nAUDIT: audit-tests\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PASS\n",
		"AUDIT: audit-code\nAUDITOR_PROVIDER: provider\nAUDITOR_MODEL: model\nVERDICT: PASS\nunstructured explanation\n",
	}
	for _, report := range cases {
		if _, err := ParseVerdict(report); err == nil {
			t.Errorf("ParseVerdict(%q) succeeded", report)
		}
	}
}
