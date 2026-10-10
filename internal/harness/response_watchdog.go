// ABOUTME: Cancels silent provider calls without confusing liveness with model activity.
// ABOUTME: Uses native response events to switch from start to idle deadlines.
package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type responseWatchdog struct {
	mu         sync.Mutex
	closed     bool
	kind       string
	timer      *time.Timer
	cancel     context.CancelFunc
	idle       time.Duration
	generation uint64
}

func watchResponse(parent context.Context, request Request) (context.Context, *responseWatchdog) {
	ctx, cancel := context.WithCancel(parent)
	w := &responseWatchdog{cancel: cancel, idle: request.ResponseIdleTimeout}
	if request.ResponseStartTimeout > 0 {
		w.timer = time.AfterFunc(request.ResponseStartTimeout, func() { w.expire("response-start-timeout", 0) })
	}
	return ctx, w
}

func (w *responseWatchdog) expire(kind string, generation uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed || w.kind != "" || generation != w.generation {
		return
	}
	w.kind = kind
	w.cancel()
}

func (w *responseWatchdog) activity() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed || w.kind != "" {
		return
	}
	w.generation++
	if w.timer != nil {
		w.timer.Stop()
	}
	if w.idle > 0 {
		generation := w.generation
		w.timer = time.AfterFunc(w.idle, func() { w.expire("response-idle-timeout", generation) })
	}
}

func (w *responseWatchdog) stop() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closed = true
	if w.timer != nil {
		w.timer.Stop()
	}
	w.cancel()
	return w.kind
}

func responseActivity(provider string, line []byte) bool {
	var event struct {
		Type  string `json:"type"`
		Event struct {
			Type  string `json:"type"`
			Delta struct {
				Text     string `json:"text"`
				Thinking string `json:"thinking"`
			} `json:"delta"`
			ContentBlock struct {
				Text     string `json:"text"`
				Thinking string `json:"thinking"`
			} `json:"content_block"`
		} `json:"event"`
		Item struct {
			Type string `json:"type"`
		} `json:"item"`
		Message struct {
			Content []struct {
				Type     string `json:"type"`
				Text     string `json:"text"`
				Thinking string `json:"thinking"`
			} `json:"content"`
		} `json:"message"`
		Content string `json:"content"`
	}
	if json.Unmarshal(line, &event) != nil {
		return false
	}
	switch provider {
	case "claude":
		if event.Type == "stream_event" {
			return event.Event.Delta.Text != "" || event.Event.Delta.Thinking != "" || event.Event.ContentBlock.Text != "" || event.Event.ContentBlock.Thinking != ""
		}
		if event.Type == "assistant" {
			for _, c := range event.Message.Content {
				if c.Text != "" || c.Thinking != "" {
					return true
				}
			}
		}
	case "codex":
		return (event.Type == "item.started" || event.Type == "item.updated" || event.Type == "item.completed") && (event.Item.Type == "reasoning" || event.Item.Type == "agent_message")
	case "copilot":
		return event.Type == "assistant.message_delta" || event.Type == "assistant.reasoning_delta" || ((event.Type == "assistant.message" || event.Type == "assistant") && event.Content != "")
	}
	return false
}

func responseTimeoutIncident(kind string, request Request, identity string) error {
	limit := request.ResponseStartTimeout
	if kind == "response-idle-timeout" {
		limit = request.ResponseIdleTimeout
	}
	return newIncident(kind, request.Harness, identity, fmt.Errorf("no model response activity within %s", limit))
}
