package render

import (
	"encoding/json"

	"claude-code-launcher/internal/resolve"
)

func MCP(resolved resolve.ResolvedMCP) ([]byte, error) {
	return json.MarshalIndent(map[string]any{
		"mcpServers": resolved.Servers,
	}, "", "  ")
}
