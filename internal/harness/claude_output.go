// ABOUTME: Extracts Claude's final response from single JSON or streamed JSON records.
// ABOUTME: Rejects provider error envelopes rather than treating them as audit results.
package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

func parseClaudeResponse(data []byte) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	response := ""
	seen := false
	for {
		var event struct {
			Type    string `json:"type"`
			Subtype string `json:"subtype"`
			Result  string `json:"result"`
			IsError bool   `json:"is_error"`
		}
		if err := decoder.Decode(&event); err != nil {
			if errors.Is(err, io.EOF) {
				return strings.TrimSpace(response), nil
			}
			return "", fmt.Errorf("parsing Claude result: %w", err)
		}
		if event.IsError || event.Type == "error" || strings.HasPrefix(event.Subtype, "error_") {
			return "", errors.New("Claude returned an error envelope; see provider diagnostics on stderr")
		}
		if event.Type == "result" || (event.Type == "" && event.Result != "") {
			if seen {
				return "", errors.New("Claude returned multiple final results")
			}
			seen = true
			response = event.Result
		}
	}
}
