// Copyright 2025 Matt Ho
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/savaki/twin-in-disguise/types"
)

func respondStream(w http.ResponseWriter, resp *types.AnthropicResponse) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 1. message_start
	msgStart := map[string]interface{}{
		"type": "message_start",
		"message": map[string]interface{}{
			"id":            resp.ID,
			"type":          "message",
			"role":          "assistant",
			"model":         resp.Model,
			"content":       []interface{}{},
			"stop_reason":   nil,
			"stop_sequence": nil,
			"usage": map[string]int{
				"input_tokens":  resp.Usage.InputTokens,
				"output_tokens": 0, // Initial output tokens
			},
		},
	}
	sendSSE(w, "message_start", msgStart)

	// 2. content blocks
	for i, block := range resp.Content {
		// content_block_start
		blockStart := map[string]interface{}{
			"type":  "content_block_start",
			"index": i,
			"content_block": map[string]interface{}{
				"type": block.Type,
			},
		}

		if block.Type == types.ContentTypeToolUse {
			blockStart["content_block"].(map[string]interface{})["id"] = block.ID
			blockStart["content_block"].(map[string]interface{})["name"] = block.Name
			blockStart["content_block"].(map[string]interface{})["input"] = map[string]interface{}{} // Empty initial input
		} else if block.Type == types.ContentTypeText {
			blockStart["content_block"].(map[string]interface{})["text"] = "" // Empty initial text
		}

		sendSSE(w, "content_block_start", blockStart)

		// content_block_delta
		delta := map[string]interface{}{
			"type":  "content_block_delta",
			"index": i,
			"delta": map[string]interface{}{},
		}

		if block.Type == types.ContentTypeText {
			delta["delta"].(map[string]interface{})["type"] = "text_delta"
			delta["delta"].(map[string]interface{})["text"] = block.Text
			sendSSE(w, "content_block_delta", delta)
		} else if block.Type == types.ContentTypeToolUse {
			// For tools, we send input_json_delta
			inputJSON, _ := json.Marshal(block.Input)
			delta["delta"].(map[string]interface{})["type"] = "input_json_delta"
			delta["delta"].(map[string]interface{})["partial_json"] = string(inputJSON)
			sendSSE(w, "content_block_delta", delta)
		} else if block.ThoughtSignature != "" {
			// If it's a block with thought signature (and possibly no text/tool?), handle it?
			// Actually thought signature is usually attached to tool use or text.
			// Currently AnthropicContentBlock stores ThoughtSignature separately but it's part of the block.
			// The translator puts it on the block.
			// We don't need to stream it separately unless it's a thinking block (which Anthropic doesn't support officially yet in this format?)
			// Ignore for now.
		}

		// content_block_stop
		stop := map[string]interface{}{
			"type":  "content_block_stop",
			"index": i,
		}
		sendSSE(w, "content_block_stop", stop)
	}

	// 3. message_delta
	msgDelta := map[string]interface{}{
		"type": "message_delta",
		"delta": map[string]interface{}{
			"stop_reason":   resp.StopReason,
			"stop_sequence": nil,
		},
		"usage": map[string]int{
			"output_tokens": resp.Usage.OutputTokens,
		},
	}
	sendSSE(w, "message_delta", msgDelta)

	// 4. message_stop
	msgStop := map[string]interface{}{
		"type": "message_stop",
	}
	sendSSE(w, "message_stop", msgStop)
}

func sendSSE(w http.ResponseWriter, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
