package preflight

import (
	"os"
	"path/filepath"
	"testing"

	"claude-code-launcher/internal/resolve"
)

func TestDisabledMCPWarningParser(t *testing.T) {
	home := t.TempDir()
	claudePath := filepath.Join(home, ".claude.json")
	content := `{"disabledMcpServers":["serena","other"]}`
	if err := os.WriteFile(claudePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	warnings := readDisabledMCPWarnings(claudePath, resolve.ResolvedMCP{
		Name: "serena",
		Servers: map[string]map[string]any{
			"serena": {"command": "serena"},
		},
	})
	if len(warnings) != 1 {
		t.Fatalf("warnings len = %d", len(warnings))
	}
}
