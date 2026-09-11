//go:build unix

// ABOUTME: Checks the prototype's local structured-result boundary without model calls.
// ABOUTME: Native identity and malformed responses must not become accepted audit results.
package main

import (
	"strings"
	"testing"
)

func TestParseResult(t *testing.T) {
	valid := `{"type":"result","subtype":"success","is_error":false,"session_id":"native-id","structured_output":{"verdict":"FAIL","findings":["Return value contradicts the requirement."]}}`
	for _, tc := range []struct {
		name, input, session string
		bad                  bool
	}{
		{"start", valid, "", false},
		{"resume", valid, "native-id", false},
		{"wrong session", valid, "different", true},
		{"empty", "", "", true},
		{"provider error", strings.Replace(valid, `"is_error":false`, `"is_error":true`, 1), "", true},
		{"no native identity", strings.Replace(valid, `"native-id"`, `""`, 1), "", true},
		{"wrong verdict", strings.Replace(valid, `"FAIL"`, `"PROVISIONAL"`, 1), "", true},
		{"missing structured result", `{"type":"result","subtype":"success","session_id":"native-id"}`, "", true},
		{"empty finding", strings.Replace(valid, `"Return value contradicts the requirement."`, `""`, 1), "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseResult([]byte(tc.input), tc.session)
			if (err != nil) != tc.bad {
				t.Fatalf("result=%+v error=%v", got, err)
			}
			if !tc.bad && (got.SessionID != "native-id" || got.Response.Verdict != "FAIL") {
				t.Fatalf("unexpected result %+v", got)
			}
		})
	}
}
