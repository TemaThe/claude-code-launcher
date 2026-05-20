package render

import (
	"encoding/json"

	"claude-code-launcher/internal/resolve"
)

func Settings(resolved resolve.ResolvedSettings) ([]byte, error) {
	payload := map[string]any{
		"hooks":   resolved.Hooks,
		"plugins": resolved.Plugins,
	}
	for k, v := range resolved.Passthrough {
		payload[k] = v
	}
	return json.MarshalIndent(payload, "", "  ")
}
