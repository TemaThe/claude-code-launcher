package render

import (
	"encoding/json"
	"testing"

	"claude-code-launcher/internal/resolve"
)

func TestRenderSettingsPrettyJSON(t *testing.T) {
	data, err := Settings(resolve.ResolvedSettings{
		Hooks: map[string]any{
			"pre": map[string]any{"command": "echo hi"},
		},
		Plugins: map[string]bool{
			"serena@claude-plugins-official": false,
		},
		Passthrough: map[string]any{
			"model": "opus",
		},
	})
	if err != nil {
		t.Fatalf("Settings() error = %v", err)
	}
	if len(data) == 0 || data[0] != '{' {
		t.Fatalf("expected json object")
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json invalid: %v", err)
	}
	if _, ok := decoded["hooks"]; !ok {
		t.Fatalf("expected hooks field")
	}
}
