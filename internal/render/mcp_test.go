package render

import (
	"encoding/json"
	"testing"

	"claude-code-launcher/internal/resolve"
)

func TestRenderMCPPrettyJSON(t *testing.T) {
	data, err := MCP(resolve.ResolvedMCP{
		Servers: map[string]map[string]any{
			"serena": {
				"command": "serena",
				"args":    []string{"serve"},
			},
		},
	})
	if err != nil {
		t.Fatalf("MCP() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json invalid: %v", err)
	}
	if _, ok := decoded["mcpServers"]; !ok {
		t.Fatalf("expected mcpServers field")
	}
}
