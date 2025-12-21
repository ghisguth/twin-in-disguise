package translator

import (
	"encoding/json"
	"testing"
)

func TestCleanSchemaForGemini_RemovesExclusiveFields(t *testing.T) {
	// Input schema with exclusiveMinimum and exclusiveMaximum
	inputJSON := `{
		"type": "object",
		"properties": {
			"age": {
				"type": "integer",
				"minimum": 0,
				"exclusiveMinimum": true
			},
			"score": {
				"type": "number",
				"maximum": 100,
				"exclusiveMaximum": true
			}
		}
	}`

	var inputSchema map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &inputSchema); err != nil {
		t.Fatalf("failed to unmarshal input json: %v", err)
	}

	// Clean the schema
	cleaned := CleanSchemaForGemini(inputSchema)

	// Verify exclusiveMinimum is gone
	props := cleaned["properties"].(map[string]interface{})
	age := props["age"].(map[string]interface{})
	if _, ok := age["exclusiveMinimum"]; ok {
		t.Error("expected exclusiveMinimum to be removed")
	}
	if _, ok := age["minimum"]; !ok {
		t.Error("expected minimum to be preserved")
	}

	// Verify exclusiveMaximum is gone
	score := props["score"].(map[string]interface{})
	if _, ok := score["exclusiveMaximum"]; ok {
		t.Error("expected exclusiveMaximum to be removed")
	}
	if _, ok := score["maximum"]; !ok {
		t.Error("expected maximum to be preserved")
	}
}
