package preflight

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"claude-code-launcher/internal/resolve"
)

func Validate(tempRoot, home string, spec resolve.LaunchSpec) ([]resolve.Warning, error) {
	if _, err := exec.LookPath(spec.ClaudeBinary); err != nil {
		return nil, fmt.Errorf("claude binary %q not found: %w", spec.ClaudeBinary, err)
	}
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		return nil, err
	}
	testFile := filepath.Join(tempRoot, ".ccx-write-test")
	if err := os.WriteFile(testFile, []byte("ok"), 0o644); err != nil {
		return nil, err
	}
	_ = os.Remove(testFile)
	for _, data := range [][]byte{spec.SettingsJSON, spec.MCPJSON} {
		var parsed any
		if err := json.Unmarshal(data, &parsed); err != nil {
			return nil, fmt.Errorf("invalid generated json: %w", err)
		}
	}
	for _, server := range spec.MCP.Servers {
		cmd, _ := server["command"].(string)
		if cmd == "" {
			continue
		}
		if _, err := exec.LookPath(cmd); err != nil {
			return nil, fmt.Errorf("mcp command %q not found: %w", cmd, err)
		}
	}
	return readDisabledMCPWarnings(filepath.Join(home, ".claude.json"), spec.MCP), nil
}

func readDisabledMCPWarnings(path string, mcp resolve.ResolvedMCP) []resolve.Warning {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return []resolve.Warning{{Code: "claude-config-read", Message: err.Error()}}
	}
	var payload struct {
		Disabled []string `json:"disabledMcpServers"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return []resolve.Warning{{Code: "claude-config-parse", Message: err.Error()}}
	}
	disabled := map[string]struct{}{}
	for _, name := range payload.Disabled {
		disabled[name] = struct{}{}
	}
	warnings := make([]resolve.Warning, 0)
	for name := range mcp.Servers {
		if _, ok := disabled[name]; ok {
			warnings = append(warnings, resolve.Warning{
				Code:    "disabled-mcp-server",
				Message: fmt.Sprintf("global Claude config disables MCP server %q", name),
			})
		}
	}
	return warnings
}
