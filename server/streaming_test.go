package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/savaki/twin-in-disguise/types"
)

func TestRespondStream(t *testing.T) {
	resp := &types.AnthropicResponse{
		ID:    "msg_123",
		Type:  "message",
		Role:  "assistant",
		Model: "gemini-2.0-flash",
		Usage: types.AnthropicUsage{
			InputTokens:  10,
			OutputTokens: 20,
		},
		Content: []types.AnthropicContentBlock{
			{
				Type: "text",
				Text: "Hello world",
			},
		},
		StopReason: "end_turn",
	}

	w := httptest.NewRecorder()
	respondStream(w, resp)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", contentType)
	}

	body := w.Body.String()
	events := strings.Split(strings.TrimSpace(body), "\n\n")

	expectedEvents := []string{
		"message_start",
		"content_block_start",
		"content_block_delta",
		"content_block_stop",
		"message_delta",
		"message_stop",
	}

	if len(events) != len(expectedEvents) {
		t.Errorf("expected %d events, got %d", len(expectedEvents), len(events))
	}

	for i, event := range events {
		if !strings.HasPrefix(event, "event: "+expectedEvents[i]) {
			t.Errorf("event %d: expected type %s, got %s", i, expectedEvents[i], event)
		}
	}
}

func TestRespondStream_ToolUse(t *testing.T) {
	resp := &types.AnthropicResponse{
		ID:    "msg_456",
		Type:  "message",
		Role:  "assistant",
		Model: "gemini-2.0-flash",
		Content: []types.AnthropicContentBlock{
			{
				Type: "tool_use",
				ID:   "toolu_1",
				Name: "get_weather",
				Input: map[string]interface{}{
					"location": "Paris",
				},
			},
		},
		StopReason: "tool_use",
	}

	w := httptest.NewRecorder()
	respondStream(w, resp)

	body := w.Body.String()
	if !strings.Contains(body, "input_json_delta") {
		t.Error("expected input_json_delta in response")
	}
	if !strings.Contains(body, "partial_json") {
		t.Error("expected partial_json in response")
	}
}
