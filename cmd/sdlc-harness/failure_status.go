// ABOUTME: Reports bounded recovery transitions without relaying provider bodies.
// ABOUTME: Keeps failure diagnosis visible while leaving auditor repair operator-owned.
package main

import (
	"fmt"
	"io"
)

func reportAuditStatus(writer io.Writer, message string) error {
	if writer == nil {
		return nil
	}
	if reporter, ok := writer.(interface{ report(string) error }); ok {
		return reporter.report(message)
	}
	_, err := fmt.Fprintln(writer, message)
	return err
}

func failureSummary(kind string) string {
	switch kind {
	case "request-too-large":
		return "request exceeds the model's context limit"
	case "quota-exhausted":
		return "provider usage quota is exhausted"
	case "authentication-failed":
		return "provider authentication failed"
	case "model-unavailable":
		return "configured model is unavailable or its identifier is unsupported"
	case "rate-limited":
		return "provider rate limit was reached"
	case "provider-unavailable":
		return "provider service is unavailable"
	case "tls-verification-failed":
		return "provider TLS certificate verification failed"
	case "request-rejected":
		return "provider rejected the request format or policy"
	case "provider-permanent-error":
		return "provider reported a non-retryable error"
	case "response-start-timeout":
		return "model did not begin responding before the response-start deadline"
	case "response-idle-timeout":
		return "model stopped responding before completion"
	case "timeout":
		return "provider turn exceeded its deadline"
	case "total-timeout":
		return "overall audit deadline expired"
	case "context-lost":
		return "auditor lost mandatory standards context"
	case "session-unavailable":
		return "retained auditor context is unavailable"
	case "executable-unavailable":
		return "configured harness executable is unavailable"
	case "start-failed", "resume-failed":
		return "harness invocation failed; details remain in the private diagnostic"
	case "unexpected-tool-use":
		return "auditor attempted a prohibited native tool operation"
	case "response-malformed", "response-empty", "response-missing", "response-standards-request":
		return "auditor returned an unusable response"
	default:
		return "provider did not return a usable audit result; details remain in the private diagnostic"
	}
}

const auditStopStatus = "Recovery stopped. No verdict obtained. Caller: report this blocker; do not troubleshoot the auditor or change its configuration. An operator-authorized --reset restores harness retry eligibility."
