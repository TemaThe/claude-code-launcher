package resolve

import (
	"fmt"
	"slices"
	"time"

	"claude-code-launcher/internal/config"
)

type Selection struct {
	ProfileName string
	ExtraArgs   []string
	PassThrough []string
}

type SessionSelection = Selection

type ResolvedProfile struct {
	Name        string
	HookProfile string
	MCPProfile  string
	Plugins     map[string]config.TriState
	Flags       config.Flags
	Env         map[string]string
	Settings    map[string]any
}

type ResolvedSettings struct {
	Hooks       map[string]any
	Plugins     map[string]bool
	Passthrough map[string]any
}

type ResolvedMCP struct {
	Name    string
	Servers map[string]map[string]any
}

type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type LaunchSpec struct {
	Profile       ResolvedProfile
	Settings      ResolvedSettings
	MCP           ResolvedMCP
	Warnings      []Warning
	CWD           string
	ClaudeBinary  string
	SettingsPath  string
	MCPPath       string
	Argv          []string
	EnvOverrides  map[string]string
	SettingsJSON  []byte
	MCPJSON       []byte
	ManifestJSON  []byte
	DryRun        bool
	NoExec        bool
	PrintOnly     bool
	JSONOnly      bool
	Timestamp     time.Time
	PID           int
	CCXVersion    string
	SelectedHooks string
	SelectedMCP   string
}

func Build(cfg *config.Config, sel Selection) (LaunchSpec, error) {
	name := sel.ProfileName
	if name == "" {
		name = "min"
	}
	profile, ok := cfg.Profiles[name]
	if !ok {
		return LaunchSpec{}, fmt.Errorf("unknown profile %q", name)
	}
	hookProfile := profile.HookProfile
	if hookProfile == "" {
		hookProfile = "none"
	}
	mcpProfileName := profile.MCPProfile
	if mcpProfileName == "" {
		mcpProfileName = "none"
	}
	if _, ok := cfg.Hooks[hookProfile]; !ok {
		return LaunchSpec{}, fmt.Errorf("missing hook profile %q", hookProfile)
	}
	if _, ok := cfg.MCP[mcpProfileName]; !ok {
		return LaunchSpec{}, fmt.Errorf("missing mcp profile %q", mcpProfileName)
	}
	resolvedProfile := ResolvedProfile{
		Name:        name,
		HookProfile: hookProfile,
		MCPProfile:  mcpProfileName,
		Plugins:     map[string]config.TriState{},
		Flags:       profile.Flags,
		Env:         cloneStringMap(profile.Env),
		Settings:    cloneAnyMap(profile.Settings),
	}
	settings := ResolvedSettings{
		Hooks:       cloneAnyMap(cfg.Hooks[hookProfile].Hooks),
		Plugins:     map[string]bool{},
		Passthrough: cloneAnyMap(profile.Settings),
	}
	for name, plugin := range cfg.Plugins {
		state := plugin.Default
		if override, ok := profile.Plugins[name]; ok && override != nil {
			state = *override
		}
		resolvedProfile.Plugins[name] = state
		if state == config.TriStateTrue {
			settings.Plugins[name] = true
		}
		if state == config.TriStateFalse {
			settings.Plugins[name] = false
		}
	}
	for name, override := range profile.Plugins {
		if _, seen := resolvedProfile.Plugins[name]; seen {
			continue
		}
		if override == nil {
			continue
		}
		resolvedProfile.Plugins[name] = *override
		if *override == config.TriStateTrue {
			settings.Plugins[name] = true
		}
		if *override == config.TriStateFalse {
			settings.Plugins[name] = false
		}
	}
	mcpProfile := cfg.MCP[mcpProfileName]
	passthroughArgs := append([]string(nil), sel.ExtraArgs...)
	passthroughArgs = append(passthroughArgs, sel.PassThrough...)
	return LaunchSpec{
		Profile:       resolvedProfile,
		Settings:      settings,
		MCP:           ResolvedMCP{Name: mcpProfileName, Servers: cloneServerMap(mcpProfile.Servers)},
		ClaudeBinary:  cfg.Defaults.ClaudeBinary,
		EnvOverrides:  cloneStringMap(profile.Env),
		Argv:          passthroughArgs,
		SelectedHooks: hookProfile,
		SelectedMCP:   mcpProfileName,
		CCXVersion:    "dev",
		Timestamp:     time.Now().UTC(),
	}, nil
}

func FinalArgv(spec LaunchSpec, settingsPath, mcpPath string) []string {
	argv := []string{spec.ClaudeBinary, "--settings", settingsPath, "--mcp-config", mcpPath}
	if spec.Profile.Flags.StrictMCPConfig {
		argv = append(argv, "--strict-mcp-config")
	}
	if spec.Profile.Flags.ExcludeDynamicSystemPromptSections {
		argv = append(argv, "--exclude-dynamic-system-prompt-sections")
	}
	argv = append(argv, spec.Argv...)
	return argv
}

func SortedProfiles(cfg *config.Config) []string {
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func cloneAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneServerMap(in map[string]map[string]any) map[string]map[string]any {
	if in == nil {
		return map[string]map[string]any{}
	}
	out := make(map[string]map[string]any, len(in))
	for k, v := range in {
		clone := make(map[string]any, len(v))
		for key, value := range v {
			clone[key] = value
		}
		out[k] = clone
	}
	return out
}
