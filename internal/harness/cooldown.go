// ABOUTME: Extracts a reset deadline from Claude's native session-limit diagnostic.
// ABOUTME: Accepts only the known clock-and-timezone format, never audit prose.
package harness

import (
	"strings"
	"time"
)

func claudeReset(message, provider string, now time.Time) time.Time {
	if provider != "claude" {
		return time.Time{}
	}
	for _, line := range strings.Split(message, "\n") {
		clockAndZone, ok := strings.CutPrefix(strings.TrimSpace(line), "You've hit your session limit · resets ")
		if !ok {
			continue
		}
		clock, zone, ok := strings.Cut(clockAndZone, " (")
		if !ok || !strings.HasSuffix(zone, ")") {
			continue
		}
		location, err := time.LoadLocation(strings.TrimSuffix(zone, ")"))
		if err != nil {
			continue
		}
		local := now.In(location)
		deadline, err := time.ParseInLocation("2006-01-02 3:04pm", local.Format("2006-01-02")+" "+clock, location)
		if err != nil {
			continue
		}
		if !deadline.After(now) {
			deadline = deadline.AddDate(0, 0, 1)
		}
		return deadline.UTC()
	}
	return time.Time{}
}
