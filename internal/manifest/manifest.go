package manifest

import (
	"encoding/json"
	"sort"
	"time"

	"claude-code-launcher/internal/redact"
	"claude-code-launcher/internal/resolve"
)

type Data struct {
	Profile         string            `json:"profile"`
	CWD             string            `json:"cwd"`
	Timestamp       time.Time         `json:"timestamp"`
	PID             int               `json:"pid"`
	ClaudeBinary    string            `json:"claude_binary"`
	SettingsPath    string            `json:"settings_path"`
	MCPPath         string            `json:"mcp_path"`
	Argv            []string          `json:"argv"`
	EnvOverrides    map[string]string `json:"env_overrides"`
	SelectedPlugins []string          `json:"selected_plugins"`
	HookProfile     string            `json:"hook_profile"`
	MCPProfile      string            `json:"mcp_profile"`
	DryRun          bool              `json:"dry_run"`
	CCXVersion      string            `json:"ccx_version"`
}

func Build(spec resolve.LaunchSpec, patterns []string) Data {
	selected := make([]string, 0, len(spec.Profile.Plugins))
	for name, state := range spec.Profile.Plugins {
		if state == "true" {
			selected = append(selected, name)
		}
	}
	sort.Strings(selected)
	return Data{
		Profile:         spec.Profile.Name,
		CWD:             spec.CWD,
		Timestamp:       spec.Timestamp,
		PID:             spec.PID,
		ClaudeBinary:    spec.ClaudeBinary,
		SettingsPath:    spec.SettingsPath,
		MCPPath:         spec.MCPPath,
		Argv:            append([]string(nil), spec.Argv...),
		EnvOverrides:    redact.Env(spec.EnvOverrides, patterns),
		SelectedPlugins: selected,
		HookProfile:     spec.Profile.HookProfile,
		MCPProfile:      spec.Profile.MCPProfile,
		DryRun:          spec.DryRun,
		CCXVersion:      spec.CCXVersion,
	}
}

func JSON(data Data) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}
